package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

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

type stack struct{ s []rune }

func (s *stack) Push(value rune) {
	s.s = append(s.s, value)
}

func (s *stack) Pop() (rune, error) {
	n := len(s.s)
	if n == 0 {
		return 0, errors.New("Empty stack")

	}
	res := s.s[n-1]
	s.s = s.s[:n-1]
	return res, nil
}

func parseCommand(command string) (string, []string) {
	command = strings.TrimSpace(command)
	parts := strings.SplitN(command, " ", 2)

	if len(parts) == 0 {
		return "", []string{}
	}

	if len(parts) == 1 {
		return parts[0], []string{}
	}

	cmd := parts[0]
	rest := parts[1]

	var args []string
	var cur string
	stack := stack{}
	var escape bool

	for _, arg := range rest {
		if arg == '\\' {
			escape = true
			continue
		}

		if escape {
			cur += string(arg)
			escape = false
			continue
		}

		if len(stack.s) > 0 {
			if arg == stack.s[len(stack.s)-1] {
				stack.Pop()
				continue
			}
			cur += string(arg)
			continue
		}

		if arg == '\'' || arg == '"' {
			stack.Push(arg)
			continue
		}

		if arg == ' ' {
			if len(cur) > 0 {
				args = append(args, cur)
				cur = ""
			}
			continue
		}
		cur += string(arg)
	}
	if len(cur) > 0 {
		args = append(args, cur)
	}

	return cmd, args
}

func handleExecutable(cmd string, args []string) {
	isExec, _ := isExecutable(cmd)
	if !isExec {
		fmt.Printf("%s: command not found\n", cmd)
		return
	}

	command := exec.Command(cmd, args...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()

	if err != nil {
		fmt.Println("Error running command:", err.Error())
		return
	}
}

func main() {
	var builtin = []string{"exit", "echo", "type", "pwd", "cd"}

	for {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		cmd, args := parseCommand(command)

		if err != nil {
			panic(err)
		}
		switch cmd {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "cd":
			dir := args[0]
			if args[0] == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Printf("cd: %s: No such file or directory\n", args[0])
					break
				}
				dir = home
			}
			err := os.Chdir(dir)

			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", args[0])
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			fmt.Println(dir)
		case "type":
			if slices.Contains(builtin, args[0]) {
				fmt.Printf("%s is a shell builtin\n", args[0])
			} else if dir, err := exec.LookPath(args[0]); err == nil {
				fmt.Printf("%s is %s\n", args[0], dir)
			} else {
				fmt.Printf("%s: not found\n", args[0])
			}
		default:
			handleExecutable(cmd, args)
		}
	}

}
