// Simple package to get user input from terminal.
package nfo

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"os/signal"
	"strings"
)

var cancel = make(chan struct{})

// Function to restore terminal on event we get an interuption.
func getEscape() {
	s, _ := terminal.GetState(0)
	var signal_chan = make(chan os.Signal)
	signal.Notify(signal_chan)
	go func() {
		BlockShutdown()
		defer UnblockShutdown()
		defer signal.Stop(signal_chan)
		select {
		case <-signal_chan:
			terminal.Restore(0, s)
		case <-cancel:
			return
		}
	}()
}

// Loop until a non-blank answer is given
func NeedAnswer(prompt string, request func(prompt string) string) (output string) {
	for output = request(prompt); output == ""; output = request(prompt) {
	}
	return output
}

// Prompt to press enter.
func PressEnter(prompt string) {
	getEscape()
	defer func() { cancel <- struct{}{} }()

	fmt.Printf("\r%s", prompt)

	var blank_line []rune
	for _ = range prompt {
		blank_line = append(blank_line, ' ')
	}
	terminal.ReadPassword(1)
	fmt.Printf("%s\r", string(blank_line))
}

// Gets user input, used during setup and configuration.
func Input(prompt string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(prompt)
	response, _ := reader.ReadString('\n')
	response = cleanInput(response)

	return response
}

// Get Hidden/Password input, without returning information to the screen.
func Secret(prompt string) string {
	getEscape()
	defer func() { cancel <- struct{}{} }()

	fmt.Printf(prompt)
	resp, _ := terminal.ReadPassword(0)
	output := cleanInput(string(resp))
	if len(output) > 0 {
		return output
	}
	fmt.Printf("\n")

	return output
}

// Get confirmation
func Confirm(prompt string) bool {
	for {
		resp := Input(fmt.Sprintf("%s (y/n): ", prompt))
		resp = strings.ToLower(resp)
		fmt.Println("")
		if resp == "y" || resp == "yes" {
			return true
		} else if resp == "n" || resp == "no" {
			return false
		}
		fmt.Printf("Err: Unrecognized response: %s\n", resp)
		continue
	}
}

// Removes newline characters
func cleanInput(input string) (output string) {
	var output_bytes []rune
	for _, v := range input {
		if v == '\n' || v == '\r' {
			continue
		}
		output_bytes = append(output_bytes, v)
	}
	return strings.TrimSpace(string(output_bytes))
}
