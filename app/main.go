package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
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
			continue
		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}

}
