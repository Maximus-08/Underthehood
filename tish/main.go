package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	// "errors"
)

func main() {
	//initialize the scanner to look for input
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("tish>")
		//fetch the input
		scanner.Scan()
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

// parseline parses the input into command and args and
// call the commands if built-in or call the execute helper
func parseline(line string) (int, error) {
	cmdStrings := strings.Split(line, "|")
	if len(cmdStrings) == 1 {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			return 1, nil
		}
		args := fields[1:]
		function := fields[0]
		if function == "exit" {
			return 0, nil
		} else if function == "cd" {
			err := cd(args)
			if err != nil {
				return 1, err
			}
			return 1, nil

		} else {
			// Call execute to run external commands
			err := execute(function, args)
			if err != nil {
				return 1, err
			}
		}
		return 1, nil
	} else {
		err := executePipeline(cmdStrings)
		if err != nil {
			return 1, err
		}
		return 1, nil

	}

}
func parseredirection(args []string) (cleaned []string, stdinFile string, stdoutFile string, appendStdout bool, err error) {
	cleaned = []string{}
	stdinFile = ""
	stdoutFile = ""
	appendStdout = false
	err = nil
	for i := 0; i < len(args); i++ {
		if args[i] == "<" {
			if len(args[i+1:]) == 0 {
				err = fmt.Errorf("Missing file name")
			} else {
				stdinFile = args[i+1]
				i++
			}
		} else if args[i] == ">" {
			if len(args[i+1:]) == 0 {
				err = fmt.Errorf("Missing file name")
			} else {
				stdoutFile = args[i+1]
				appendStdout = false
				i++
			}
		} else if args[i] == ">>" {
			if len(args[i+1:]) == 0 {
				err = fmt.Errorf("Missing file name")
			} else {
				stdoutFile = args[i+1]
				appendStdout = true
				i++
			}
		} else {
			cleaned = append(cleaned, args[i])
		}

	}
	return cleaned, stdinFile, stdoutFile, appendStdout, err
}
func executePipeline(cmdStrings []string) (err error) {
	//Keep track of all spawned commands to wait for them at the end
	cmds := []*exec.Cmd{}
	var prevReader *os.File
	for i := 0; i < len(cmdStrings); i++ {
		fields := strings.Fields(cmdStrings[i])
		if len(fields) == 0 {
			return fmt.Errorf("Empty command in pipeline")
		}
		args := fields[1:]
		function := fields[0]
		cleaned, stdinFile, stdoutFile, appendStdout, err := parseredirection(args)
		if err != nil {
			return err
		}
		cmd := exec.Command(function, cleaned...)
		if stdinFile != "" {
			file, err := os.Open(stdinFile)
			if err != nil {
				return err
			}
			cmd.Stdin = file
			defer file.Close()
		} else if prevReader != nil {
			cmd.Stdin = prevReader
		} else {
			cmd.Stdin = os.Stdin
		}
		var nextReader *os.File
		if i < len(cmdStrings)-1 {
			r, w, err := os.Pipe()
			if err != nil {
				return err
			}
			cmd.Stdout = w
			nextReader = r
		} else if stdoutFile != "" {
			var file *os.File
			var err error
			if appendStdout {
				file, err = os.OpenFile(stdoutFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
			} else {
				file, err = os.OpenFile(stdoutFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			}
			if err != nil {
				return err
			}
			cmd.Stdout = file
			defer file.Close()
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			return err
		}
		if i < len(cmdStrings)-1 {
			cmd.Stdout.(*os.File).Close()
		}
		if prevReader != nil {
			prevReader.Close()
		}
		prevReader = nextReader
		cmds = append(cmds, cmd)
	}
	for i := 0; i < len(cmds); i++ {
		err := cmds[i].Wait()
		if err != nil {
			fmt.Println("Command failed:", err)
			return err
		}
	}

	return nil
}

func execute(function string, args []string) error {

	cleaned, stdinFile, stdoutFile, appendStdout, err := parseredirection(args)
	if err != nil {
		return err
	}
	//tie input,output and error streams of current parent process and child
	//which is forked when we use cmd.Run() or cmd.Start() under the hood
	cmd := exec.Command(function, cleaned...)
	if stdinFile != "" {
		file, err := os.Open(stdinFile)
		if err != nil {
			return err
		}
		cmd.Stdin = file
		defer file.Close()
	} else {
		cmd.Stdin = os.Stdin
	}
	if stdoutFile != "" {
		var file *os.File
		var err error
		if appendStdout {
			file, err = os.OpenFile(stdoutFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		} else {
			file, err = os.OpenFile(stdoutFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		}
		if err != nil {
			return err
		}
		cmd.Stdout = file
		defer file.Close()
	} else {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = os.Stderr
	return cmd.Run()

}

// cd makes the directory change to the desired directory with the builtin os.Chdir()
func cd(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("Empty path arguments for cd")
	} else {
		err := os.Chdir(args[0])
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}
	return nil
}
