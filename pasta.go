package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func runBuiltIns(command string, args []string) bool {
	switch command {
	case "cd":
		syscall.Chdir(args[0])
		return true
	case "exit":
		if len(args) > 0 {
			code, err := strconv.Atoi(args[0])
			fmt.Println(code)
			if err != nil {
				fmt.Println("Please pass an integer!")
			} else {
				syscall.Exit(code)
			}
		} else {
			syscall.Exit(0)
		}
		return true
	default:
		return false
	}
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("🍝 ")
		input, err := reader.ReadString('\n')
		input = strings.TrimRight(input, "\n")
		if err != nil {
			// Handle Ctrl+D to stop program.
			if err == io.EOF {
				fmt.Println("\nArrivederci!")
				os.Exit(1)
			}
			fmt.Println("Mama Mia! Something went wrong!")
			fmt.Fprintln(os.Stderr, err)
		}

		parsedInput := strings.Split(input, " ")
		userCmd := parsedInput[0]
		args := parsedInput[1:]

		if runBuiltIns(userCmd, args) {
			continue
		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			os.Exit(1)
		}()
		cmd := exec.Command(userCmd, args...)
		cmd.Env = append(os.Environ())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Print("👨‍🍳 ")
		if err := cmd.Run(); err != nil {
			fmt.Printf("Something went wrong!\n %s\n", err)
		}
	}
}
