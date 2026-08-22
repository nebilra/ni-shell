package main

import (
	"bufio"
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

func handleExecutable(cmd []string) {
	isExec, _ := isExecutable(cmd[0])
	if !isExec {
		fmt.Printf("%s: command not found\n", cmd[0])
		return
	}

	command := exec.Command(cmd[0], cmd[1:]...)

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
		command = strings.TrimSpace(command)
		cmd := strings.Split(command, " ")

		if err != nil {
			panic(err)
		}
		switch cmd[0] {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(cmd[1:], " "))
		case "cd":
			dir := cmd[1]
			if cmd[1] == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Printf("cd: %s: No such file or directory\n", cmd[1])
					break
				}
				dir = home
			}
			err := os.Chdir(dir)

			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", cmd[1])
			}
		case "pwd":
			dir, err := os.Getwd()
			if err != nil {
				fmt.Println(err.Error())
				break
			}
			fmt.Println(dir)
		case "type":
			if slices.Contains(builtin, cmd[1]) {
				fmt.Printf("%s is a shell builtin\n", cmd[1])
			} else if dir, err := exec.LookPath(cmd[1]); err == nil {
				fmt.Printf("%s is %s\n", cmd[1], dir)
			} else {
				fmt.Printf("%s: not found\n", cmd[1])
			}
		default:
			handleExecutable(cmd)
		}
	}

}
