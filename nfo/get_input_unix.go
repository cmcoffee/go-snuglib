//go:build !windows
// +build !windows

package nfo

import (
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"os"
	"syscall"
)

// Gets user input, used during setup and configuration.
func GetInput(prompt string) string {
	unesc := Defer(getEscape())
	defer unesc()

	fmt.Printf(prompt)

	terminal.MakeRaw(int(syscall.Stdin))

	var (
		str string
		err error
	)

	for {
		t := terminal.NewTerminal(os.Stdin, "")
		str, err = t.ReadLine()
		if err == io.EOF {
			signalChan <- syscall.SIGINT
			continue
		}
		break
	}
	return cleanInput(str)
}
