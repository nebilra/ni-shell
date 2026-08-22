package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

func loadPath() []string {
	path := os.Getenv("PATH")
	pathDirs := strings.Split(path, ":")

	return pathDirs
}

func commandInPath(target string, pathDirs []string) (bool, string) {
	for _, dir := range pathDirs {
		dirInfo, statErr := os.Stat(dir)

		if statErr != nil {
			continue
		}

		if !dirInfo.Mode().IsDir() {
			if target == dirInfo.Name() {
				return true, fmt.Sprintf("%s", dir)
			}
			return false, ""
		}

		entry, readErr := os.ReadDir(dir)

		if readErr != nil {
			continue
		}

		for _, cmd := range entry {
			if cmd.IsDir() || cmd.Type().Perm()&0111 != 0 {
				continue
			}

			if target == cmd.Name() {
				return true, fmt.Sprintf("%s/%s", dir, cmd.Name())
			}
		}

	}

	return false, ""

}

func main() {
	var builtin = []string{"exit", "echo", "type"}
	path := loadPath()

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
			if slices.Contains(builtin, cmd[1]) {
				fmt.Printf("%s is a shell builtin\n", cmd[1])
			} else if inPath, dir := commandInPath(cmd[1], path); inPath {
				fmt.Printf("%s is %s\n", cmd[1], dir)
			} else {
				fmt.Printf("%s: not found\n", cmd[1])
			}
		default:
			fmt.Printf("%s: command not found\n", command)
		}
	}

}
