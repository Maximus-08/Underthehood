package main
import(
	"testing"
	"reflect"
) 
func TestTokenizer(t *testing.T){
	var tests = []struct {
        name string
            input string
            want  []token
        }{
            // the table itself
            {"Test-1 echo", "echo hello world", []token{{kind:tokenWord, value:"echo"}, {kind:tokenWord, value:"hello"}, {kind:tokenWord, value:"world"}}},
            {"Test-2 Two command pipeline", "ls -l | grep main", []token{{kind:tokenWord, value:"ls"},{kind:tokenWord, value:"-l"}, {kind:tokenPipe, value:"|"},{kind:tokenWord, value:"grep"}, {kind:tokenWord, value:"main"}}},
            {"Test-3", "cat < input.txt > output.txt", []token{{kind:tokenWord, value:"cat"}, {kind:tokenRedirectIn, value:"<"},{kind:tokenWord, value:"input.txt"},{kind:tokenRedirectOut, value:">"},{kind:tokenWord, value:"output.txt"} }},
            {"Test-4 Env", "VAR=val env", []token{{kind:tokenWord, value:"VAR=val"},{kind:tokenWord, value:"env"}}},
        }
      // The execution loop
        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                tokens,_ := tokenize(tt.input)
                if !reflect.DeepEqual(tokens,tt.want) {
                    t.Errorf("Error")
                }
            })
        }
}
func TestParse(t *testing.T){
	var tests = []struct {
        name string
            input string
            want  pipeline
        }{
            // the table itself
            {"Test-1 echo", "echo hello world", pipeline{cmds:[]command{{cmd:"echo", args:[]string{"hello", "world"}}}}},
            {"Test-2 Two command pipeline", "ls -l | grep main", pipeline{cmds:[]command{{cmd:"ls", args:[]string{"-l"}},{cmd:"grep",args:[]string{"main"}}}}},
            {"Test-3 Redirections", "cat < input.txt > output.txt", pipeline{cmds:[]command{{cmd:"cat", stdinFile:"input.txt", stdoutFile:"output.txt", isAppend:false}}}},
            {"Test-4 Background", "sleep 10 &", pipeline{cmds:[]command{{cmd:"sleep", args:[]string{"10"}, }},isBackground:true}},
            {"Test-5 Env", "VAR=val env", pipeline{cmds:[]command{{cmd:"env", env:[]string{"VAR=val"}, }}}},
        }
      // The execution loop
        for _, tt := range tests {
            t.Run(tt.name, func(t *testing.T) {
                tokens,_ := tokenize(tt.input)
				ans,_ := parse(tokens)
                if !reflect.DeepEqual(ans,tt.want) {
                    t.Errorf("Error")
                }
            })
        }
}