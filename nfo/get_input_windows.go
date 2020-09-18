package nfo

import (
	"bufio"
	"fmt"
	"os"
)

// Gets user input, used during setup and configuration.
func GetInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf(prompt)
	response, _ := reader.ReadString('\n')

	return cleanInput(response)
}
