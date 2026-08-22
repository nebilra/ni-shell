package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func main() {
	path := []string{"exit", "echo", "type"}
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
		case "type":
			if slices.Contains(path, cmd[1]) {
				fmt.Printf("%s is a shell builtin\n", cmd[1])
			} else {
				fmt.Printf("%s: not found\n", cmd[1])
			}
		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}

}
