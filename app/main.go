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
		command = strings.TrimSuffix(command, "\n")

		if err != nil {
			panic(err)
		}
		switch command {
		case "exit":
			os.Exit(0)
		}

		fmt.Printf("%s: command not found\n", command)
	}

}
