package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	// "errors"
)

var activeCommands atomic.Int32
var jobs map[int]*exec.Cmd
var jobsMu sync.Mutex
var jobsCond *sync.Cond

func main() {
	jobs = make(map[int]*exec.Cmd)
	jobsCond = sync.NewCond(&jobsMu)
	//initialize the scanner to look for input
	scanner := bufio.NewScanner(os.Stdin)
	//Make a buffered channel to catch the SIGINT/ Ctrl+C.
	//Notify catches the signal and puts it in the channel
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go checksignal(sigchan)

	for {
		fmt.Print("tish>")
		//fetch the input
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		status, err := parseline(line)
		if err != nil {
			fmt.Println(err)
		}
		if status == 0 {
			break
		}

	}
}

type command struct {
	cmd        string
	env        []string
	args       []string
	stdinFile  string
	stdoutFile string
	isAppend   bool
}
type pipeline struct {
	cmds         []command
	isBackground bool
}

type tokenType int

const (
	tokenWord tokenType = iota
	tokenPipe
	tokenRedirectIn
	tokenRedirectOut
	tokenRedirectAppend
	tokenBackground
)

type token struct {
	kind  tokenType
	value string
}

func tokenize(line string) ([]token, error) {
	runes := []rune(line)
	var isSingleQuote bool
	var isDoubleQuote bool
	var currentToken strings.Builder
	var tokens []token
	flushWord := func() {
		if currentToken.Len() > 0 {
			tokens = append(tokens, token{kind: tokenWord, value: currentToken.String()})
			currentToken.Reset()
		}
	}
	for i := 0; i < len(runes); i++ {
		if runes[i] == '"' && !isSingleQuote {
			if isDoubleQuote == true {
				isDoubleQuote = false
			} else {
				isDoubleQuote = true
			}
		} else if runes[i] == '\'' && !isDoubleQuote {
			if isSingleQuote == true {
				isSingleQuote = false
			} else {
				isSingleQuote = true
			}
		} else if runes[i] == '|' {
			if !isDoubleQuote && !isSingleQuote {
				flushWord()
				tokens = append(tokens, token{kind: tokenPipe, value: "|"})
			} else {
				currentToken.WriteRune(runes[i])
			}
		} else if runes[i] == '&' {
			if !isDoubleQuote && !isSingleQuote {
				flushWord()
				tokens = append(tokens, token{kind: tokenBackground, value: "&"})
			} else {
				currentToken.WriteRune(runes[i])
			}
		} else if runes[i] == '>' {
			if !isDoubleQuote && !isSingleQuote {
				if i != len(runes)-1 && runes[i+1] == '>' {
					flushWord()
					tokens = append(tokens, token{kind: tokenRedirectAppend, value: ">>"})
					i++
				} else {
					flushWord()
					tokens = append(tokens, token{kind: tokenRedirectOut, value: ">"})
				}

			} else {
				currentToken.WriteRune(runes[i])
			}
		} else if runes[i] == '<' {
			if !isDoubleQuote && !isSingleQuote {
				flushWord()
				tokens = append(tokens, token{kind: tokenRedirectIn, value: "<"})
			} else {
				currentToken.WriteRune(runes[i])
			}
		} else if runes[i] == '$' && !isSingleQuote {
			i++
			isBraced := false
			isClosingBrace := false
			if i < len(runes) && runes[i] == '{' {
				isBraced = true
				i++
			}
			var varName strings.Builder
			for i < len(runes) {
				r := runes[i]
				if isBraced {
					if r == ' ' {
						return nil, fmt.Errorf("Bad substitution")
					} else if r == '}' {
						i++
						isClosingBrace = true
						break
					} else if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
						varName.WriteRune(r)
						i++
					} else {
						return nil, fmt.Errorf("Bad substitution")
					}
				} else {
					if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
						varName.WriteRune(r)
						i++
					} else {
						break
					}
				}
			}
			if !isClosingBrace && isBraced {
				return nil, fmt.Errorf("Missing closing brace")
			}
			env := os.Getenv(varName.String())
			if env != "" {
				currentToken.WriteString(env)
			}
			i--
		} else if (runes[i] == ' ' || runes[i] == '\t') && !isDoubleQuote && !isSingleQuote {
			flushWord()
		} else {
			currentToken.WriteRune(runes[i])
		}
	}
	if isDoubleQuote || isSingleQuote {
		return nil, fmt.Errorf("Missing quotes")
	}
	flushWord()
	return tokens, nil
}

func parse(tokens []token) (pipeline, error) {
	cmdPipeline := pipeline{}
	if len(tokens) == 0 {
		return pipeline{}, nil
	}
	if tokens[0].kind != tokenWord {
		return pipeline{}, fmt.Errorf("Invalid syntax")
	}
	if tokens[len(tokens)-1].kind == tokenBackground {
		cmdPipeline.isBackground = true
		tokens = tokens[:len(tokens)-1]
	}
	j := 0
	cmdPipeline.cmds = append(cmdPipeline.cmds, command{})
	for i := 0; i < len(tokens); i++ {
		if tokens[i].kind == tokenWord {
			if strings.Contains(tokens[i].value, "=") && cmdPipeline.cmds[j].cmd == "" {
				cmdPipeline.cmds[j].env = append(cmdPipeline.cmds[j].env, tokens[i].value)
			} else if cmdPipeline.cmds[j].cmd == "" {
				cmdPipeline.cmds[j].cmd = tokens[i].value
			} else {
				cmdPipeline.cmds[j].args = append(cmdPipeline.cmds[j].args, tokens[i].value)
			}
		} else if tokens[i].kind == tokenPipe {
			if i == 0 {
				return pipeline{}, fmt.Errorf("Invalid pipe")
			} else if i == len(tokens)-1 {
				return pipeline{}, fmt.Errorf("Invalid pipe")
			} else if tokens[i+1].kind != tokenPipe {
				cmdPipeline.cmds = append(cmdPipeline.cmds, command{})
				j++
			} else {
				return pipeline{}, fmt.Errorf("Invalid pipe")
			}
		} else if tokens[i].kind == tokenBackground {
			return pipeline{}, fmt.Errorf("Invalid background token")
		} else if tokens[i].kind == tokenRedirectIn {
			if i == len(tokens)-1 || tokens[i+1].kind != tokenWord {
				return pipeline{}, fmt.Errorf("Missing filename")
			} else {
				cmdPipeline.cmds[j].stdinFile = tokens[i+1].value
				i++
			}
		} else if tokens[i].kind == tokenRedirectOut {
			if i == len(tokens)-1 || tokens[i+1].kind != tokenWord {
				return pipeline{}, fmt.Errorf("Missing filename")
			} else {
				cmdPipeline.cmds[j].stdoutFile = tokens[i+1].value
				i++
			}
		} else if tokens[i].kind == tokenRedirectAppend {
			if i == len(tokens)-1 || tokens[i+1].kind != tokenWord {
				return pipeline{}, fmt.Errorf("Missing filename")
			} else {
				cmdPipeline.cmds[j].stdoutFile = tokens[i+1].value
				cmdPipeline.cmds[j].isAppend = true
				i++
			}
		}
	}
	for i := 0; i < len(cmdPipeline.cmds); i++ {
		if cmdPipeline.cmds[i].cmd == "" {
			return pipeline{}, fmt.Errorf("Command not found")
		}
	}
	return cmdPipeline, nil
}

func parseline(line string) (int, error) {
	tokens, err := tokenize(line)
	if err != nil {
		return 1, err
	}
	p, err := parse(tokens)
	if err != nil {
		return 1, err
	}
	status, err := executePipeline(p)
	return status, err
}

func executePipeline(p pipeline) (status int, err error) {
	//Keep track of all spawned commands to wait for them at the end
	if len(p.cmds) == 1 {
		switch p.cmds[0].cmd {
		case "exit":
			return 0, nil
		case "cd":
			return cd(p.cmds[0].args)
		case "jobs":
			return printJobs()
		case "wait":
			return wait()
		case "export":
			return export(p.cmds[0].args)
		}
	}
	var firstPid int
	var prevReader *os.File
	var cms []*exec.Cmd
	for i, c := range p.cmds {

		inFile := c.stdinFile
		outFile := c.stdoutFile
		appendStdout := c.isAppend
		comd := exec.Command(c.cmd, c.args...)
		if len(c.env) > 0 {
			comd.Env = append(os.Environ(), c.env...)
		}

		if inFile != "" {
			file, err := os.Open(inFile)
			if err != nil {
				return 1, err
			}
			comd.Stdin = file
			defer file.Close() //Maybe need to close at the end of the loop
		} else if prevReader != nil {
			comd.Stdin = prevReader
		} else {
			comd.Stdin = os.Stdin
		}
		var nextReader *os.File
		if i < len(p.cmds)-1 {
			r, w, err := os.Pipe()
			if err != nil {
				return 1, err
			}
			nextReader = r
			comd.Stdout = w
		} else if outFile != "" {
			var file *os.File
			var err error
			if appendStdout {
				file, err = os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			} else {
				file, err = os.OpenFile(outFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			}
			if err != nil {
				return 1, err
			}
			comd.Stdout = file
			defer file.Close()
		} else {
			comd.Stdout = os.Stdout
		}
		comd.Stderr = os.Stderr
		if p.isBackground {
			if i == 0 {
				comd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			} else {
				comd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: firstPid}
			}
		}
		err = comd.Start()
		if err != nil {
			return 1, err
		}
		if p.isBackground && i == 0 {
			firstPid = comd.Process.Pid
		}
		if i < len(p.cmds)-1 {
			comd.Stdout.(*os.File).Close()
		}
		if prevReader != nil {
			prevReader.Close()
		}
		prevReader = nextReader
		if p.isBackground {
			pid := comd.Process.Pid
			jobsMu.Lock()
			jobs[pid] = comd
			fmt.Printf("[%d]%d\n", i, pid)
			jobsMu.Unlock()
		} else {
			activeCommands.Add(1)
			defer activeCommands.Add(-1)
		}
		cms = append(cms, comd)
	}
	if !p.isBackground {
		for _, c := range cms {
			c.Wait()
		}
	} else {
		go func() {
			for _, c := range cms {
				c.Wait()
				pid := c.Process.Pid
				jobsMu.Lock()
				delete(jobs, pid)
				jobsCond.Broadcast()
				jobsMu.Unlock()
			}
		}()
	}
	return 1, nil
}

// cd makes the directory change to the desired directory with the builtin os.Chdir()
func cd(args []string) (int, error) {
	if len(args) == 0 {
		return 1, fmt.Errorf("Empty path arguments for cd")
	} else {
		err := os.Chdir(args[0])
		if err != nil {
			return 1, fmt.Errorf("%w", err)
		}
	}
	return 1, nil
}

func printJobs() (int, error) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	for key := range jobs {
		fmt.Println(key)
	}
	return 1, nil
}
func export(arg []string) (int, error) {
	if len(arg) == 0 {
		return 1, fmt.Errorf("Missing env declaration")
	}
	key, value, found := strings.Cut(arg[0], "=")
	if !found {
		return 1, nil
	}
	return 1, os.Setenv(key, value)
}
func wait() (int, error) {
	jobsMu.Lock()
	defer jobsMu.Unlock()
	for len(jobs) > 0 {
		jobsCond.Wait()
	}
	return 1, nil
}

// Checks if any child processes is active with the activeCommands counter
//
//	if not only the parent shell is active
func checksignal(sigchan chan os.Signal) {
	for range sigchan {
		if activeCommands.Load() == 0 {
			fmt.Print("\ntish>")
		}
	}
}
