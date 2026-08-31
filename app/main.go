package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

type Command struct {
	input      string
	cmd        string
	args       []string
	cmdArgs    []string
	outputArgs []string
	outputSign string
	outputFile *os.File
	redirected bool
	appended   bool
}

func isExecutable(path string) (bool, string) {
	if cmd, err := exec.LookPath(path); err == nil {
		return true, cmd
	}
	info, err := os.Stat(path)

	if err != nil {
		return false, ""
	}

	return info.Mode()&0111 != 0, path
}

func (c *Command) parseCommand() {
	c.input = strings.TrimSpace(c.input)

	var parts []string
	var cur string
	var quote rune
	var escape bool

	for idx, arg := range c.input {
		if escape {
			cur += string(arg)
			escape = false
			continue
		}

		if arg == '\\' && quote != '\'' {
			escape = true
			continue
		}

		if quote != 0 {
			if arg == quote {
				quote = 0
				continue
			}
			cur += string(arg)
			continue
		}

		if arg == '\'' || arg == '"' {
			quote = arg
			continue
		}

		if arg == ' ' {
			if len(cur) > 0 {
				parts = append(parts, cur)
				cur = ""
			}
			continue
		}
		if arg == '>' {
			if rune(c.input[idx-1]) == arg {
				c.appended = true
				continue
			}

			if len(cur) > 0 {
				num, err := strconv.Atoi(string(cur[len(cur)-1]))
				if err == nil && num < 3 {
					if len(cur[:len(cur)-1]) > 0 {
						parts = append(parts, cur[:len(cur)-1])
					}
					parts = append(parts, string(cur[len(cur)-1])+string(arg))
					cur = ""
					continue
				}
				parts = append(parts, cur)
				parts = append(parts, string(arg))
				cur = ""
				continue
			}

		}
		cur += string(arg)
	}
	if len(cur) > 0 {
		parts = append(parts, cur)
	}

	if len(parts) == 0 {
		c.cmd = ""
		c.args = []string{}
		return
	}

	if len(parts) == 1 {
		c.cmd = parts[0]
		c.args = []string{}
	}

	c.cmd = parts[0]
	c.args = parts[1:]
	c.isRedirected()
}

func (c *Command) handleExecutable() {
	isExec, _ := isExecutable(c.cmd)
	if !isExec {
		c.output(fmt.Sprintf("%s: command not found", c.cmd))
		return
	}

	command := exec.Command(c.cmd, c.cmdArgs...)
	command.Stdout, command.Stderr = c.outputChannel()

	err := command.Run()

	if err != nil {
		return
	}
}

func (c *Command) output(value string) {
	if c.redirected {
		stdout := os.Stdout
		stderr := os.Stderr

		os.Stdout, os.Stderr = c.outputChannel()
		fmt.Println(value)

		os.Stdout = stdout
		os.Stderr = stderr
		return
	}
	fmt.Println(value)
}

func (c *Command) outputChannel() (*os.File, *os.File) {
	var fileErr error
	if c.redirected {
		if c.appended {
			c.outputFile, fileErr = os.OpenFile(c.outputArgs[0], os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		} else {
			c.outputFile, fileErr = os.OpenFile(c.outputArgs[0], os.O_CREATE|os.O_WRONLY, 0644)
		}

		if fileErr != nil {
			fmt.Println("Error writing to file:", fileErr)
		}
	}

	switch c.outputSign {
	case "1>":
		fallthrough
	case ">":
		return c.outputFile, os.Stderr
	case "2>1":
		return os.Stdout, os.Stdout
	case "2>":
		return os.Stdout, c.outputFile
	}

	return os.Stdout, os.Stderr
}

func (c *Command) isRedirected() {
	idx := slices.IndexFunc(c.args, func(val string) bool {
		return strings.Contains(
			val,
			">",
		)
	})
	c.redirected = idx != -1
	c.cmdArgs = c.args

	if c.redirected {
		c.outputSign = c.args[idx]
		c.cmdArgs = c.args[:idx]
		c.outputArgs = c.args[idx+1:]
	}
}

var builtin = []string{"exit", "echo", "type", "pwd", "cd"}

type AutoCompleter struct {
}

func (ac *AutoCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	var out [][]rune
	var completions []string
	completions = append(completions, builtin...)

	path := filepath.SplitList(os.Getenv("PATH"))

	for _, dir := range path {
		dir = filepath.Clean(dir)
		contents, dirErr := os.ReadDir(dir)
		if dirErr != nil {
			continue
		}
		for _, file := range contents {
			if isExec, _ := isExecutable(file.Name()); isExec {
				completions = append(completions, file.Name())
			}
		}
	}

	for _, cmd := range completions {
		if trimmed, ok := strings.CutPrefix(cmd, string(line)); ok {
			out = append(out, []rune(trimmed+" "))
		}
	}
	if len(out) == 0 {
		fmt.Print("\a")
	}

	return out, pos
}

func main() {
	for {
		rl, rlerr := readline.New("$ ")

		if rlerr != nil {
			panic(rlerr)
		}

		defer rl.Close()

		rl.Config.AutoComplete = &AutoCompleter{}

		line, err := rl.Readline()
		if err != nil {
			break
		}

		c := Command{
			input: line,
		}
		c.parseCommand()
		switch c.cmd {
		case "exit":
			os.Exit(0)
		case "echo":
			c.output(strings.Join(c.cmdArgs, " "))
		case "cd":
			dir := c.cmdArgs[0]
			if c.cmdArgs[0] == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					c.output(fmt.Sprintf("cd: %s: No such file or directory", c.cmdArgs[0]))
					break
				}
				dir = home
			}
			err := os.Chdir(dir)

			if err != nil {
				c.output(fmt.Sprintf("cd: %s: No such file or directory", c.cmdArgs[0]))
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				c.output(err.Error())
				break
			}
			c.output(dir)
		case "type":
			if slices.Contains(builtin, c.cmdArgs[0]) {
				c.output(fmt.Sprintf("%s is a shell builtin", c.cmdArgs[0]))
			} else if dir, err := exec.LookPath(c.cmdArgs[0]); err == nil {
				c.output(fmt.Sprintf("%s is %s", c.cmdArgs[0], dir))
			} else {
				c.output(fmt.Sprintf("%s: not found", c.cmdArgs[0]))
			}
		default:
			c.handleExecutable()
		}
		if c.redirected {
			c.outputFile.Close()
		}
	}
}
