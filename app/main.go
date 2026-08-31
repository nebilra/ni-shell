package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
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

func parseCommand(command string) (string, []string) {
	command = strings.TrimSpace(command)

	var parts []string
	var cur string
	var quote rune
	var escape bool

	for _, arg := range command {
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
		return "", []string{}
	}

	if len(parts) == 1 {
		return parts[0], []string{}
	}

	return parts[0], parts[1:]
}

func handleExecutable(cmd string, args []string, idx int) {
	isExec, _ := isExecutable(cmd)
	if !isExec {
		output(fmt.Sprintf("%s: command not found\n", cmd), idx, args)
		return
	}

	command := exec.Command(cmd, args...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if idx != -1 {
		redirectOutput(func() {
			err := command.Run()
			if err != nil {
				output(fmt.Sprintf("Error running command:", err.Error()), idx, args)
				return
			}
		}, args[idx], args[idx+1])
	} else {
		err := command.Run()
		if err != nil {
			fmt.Println("Error running command:", err.Error())
			return
		}
	}

}

func output(value string, idx int, args []string) {
	if idx != -1 {
		redirectOutput(func() { fmt.Println(value) }, args[idx], args[idx+1])
		return
	}
	fmt.Println(value)
}

func redirectOutput(callback func(), sign string, path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		fmt.Println("Error writing to file:", err)
	}

	defer file.Close()
	stdout := os.Stdout
	stderr := os.Stderr

	switch sign {
	case "1>":
		fallthrough
	case ">":
		os.Stdout = file
	case "2>1":
		os.Stderr = os.Stdout
	case "2>":
		os.Stderr = file
	}

	callback()

	os.Stdout = stdout
	os.Stderr = stderr
}

func isRedirected(args []string) int {
	return slices.IndexFunc(args, func(val string) bool {
		return strings.Contains(
			val,
			">",
		)
	})
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
		idx := isRedirected(args)
		redirected := idx != -1

		switch cmd {
		case "exit":
			os.Exit(0)
		case "echo":
			if redirected {
				redirectOutput(func() {
					fmt.Println(strings.Join(args[:idx], " "))
				}, args[idx], args[idx+1])
			} else {
				fmt.Println(strings.Join(args, " "))
			}
		case "cd":
			dir := args[0]
			if args[0] == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					output(fmt.Sprintf("cd: %s: No such file or directory", args[0]), idx, args)
					break
				}
				dir = home
			}
			err := os.Chdir(dir)

			if err != nil {
				output(fmt.Sprintf("cd: %s: No such file or directory", args[0]), idx, args)
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				output(err.Error(), idx, args)
				break
			}
			output(dir, idx, args)
		case "type":
			if slices.Contains(builtin, args[0]) {
				output(fmt.Sprintf("%s is a shell builtin", args[0]), idx, args)
			} else if dir, err := exec.LookPath(args[0]); err == nil {
				output(fmt.Sprintf("%s is %s", args[0], dir), idx, args)
			} else {
				output(fmt.Sprintf("%s: not found", args[0]), idx, args)
			}
		default:
			handleExecutable(cmd, args, idx)
		}
	}

}
