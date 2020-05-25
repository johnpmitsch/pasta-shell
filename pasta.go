package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

type ShellCommand struct {
	command    string
	args       []string
	background bool
}

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

func handleInputErr(err error) {
	if err != nil {
		// Handle Ctrl+D to stop program.
		if err == io.EOF {
			fmt.Println("\nArrivederci!")
			os.Exit(1)
		}
		fmt.Println("Mama Mia! Something went wrong!")
		fmt.Fprintln(os.Stderr, err)
	}
}

func buildCommand(input string, background bool) ShellCommand {
	command := ShellCommand{}
	input = strings.Trim(input, " \n")
	parsedInput := strings.Split(input, " ")
	command.command = parsedInput[0]
	command.args = parsedInput[1:]
	command.background = background
	return command
}

func buildAllCommands(input string) []ShellCommand {
	allCommands := make([]ShellCommand, 0)
	// matching all " & " but ideally should change since one could pass in a quoted " & " as
	// part of a command that would be ignored by a typical shell
	if strings.HasSuffix(input, "&") || strings.Contains(input, " & ") {
		commands := strings.Split(input, "&")
		lastJobIndex := len(commands) - 1
		lastJob := commands[lastJobIndex]
		backgroundJobs := commands[:lastJobIndex]
		foregroundJobs := strings.Split(lastJob, "&&")
		for _, bg := range backgroundJobs {
			allCommands = append(allCommands, buildCommand(bg, true))
		}
		for _, fg := range foregroundJobs {
			allCommands = append(allCommands, buildCommand(fg, false))
		}
	} else {
		userCommands := strings.Split(input, "&&")
		for _, cmd := range userCommands {
			allCommands = append(allCommands, buildCommand(cmd, false))
		}
	}
	return allCommands
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("🍝 ")
		input, err := reader.ReadString('\n')
		handleInputErr(err)
		for _, command := range buildAllCommands(input) {
			if runBuiltIns(command.command, command.args) {
				continue
			}

			cmd := exec.Command(command.command, command.args...)
			cmd.Env = append(os.Environ())
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Print("👨‍🍳 ")
			startErr := cmd.Start()
			if startErr != nil {
				fmt.Printf("Something went wrong!\n %s\n", startErr)
			}

			ctx := context.Background()

			// trap Ctrl+C and call cancel on the context
			ctx, cancel := context.WithCancel(ctx)
			c := make(chan os.Signal, 1)
			signal.Notify(c, os.Interrupt)
			defer func() {
				signal.Stop(c)
				cancel()
			}()
			go func() {
				select {
				case <-c:
					cancel()
				case <-ctx.Done():
				}
			}()

			waitErr := cmd.Wait()

			if waitErr != nil {
				fmt.Println("")
				break
			}
		}
	}
}
