package internal

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Ask prompts the user with a question and only accepts 'y' or 'n'.
// It loops until a valid key is pressed and keeps the prompt on a single line.
func Ask(prompt string) (bool, error) {
	// Save current terminal state and put terminal into raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return false, err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState) // restore terminal on exit

	fmt.Printf("%s [y/n]: ", prompt)

	for {
		var b [1]byte
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			return false, err
		}

		switch b[0] {
		case 'y', 'Y':
			fmt.Printf("%s\r\n", string(b[0]))
			return true, nil
		case 'n', 'N':
			fmt.Printf("%s\r\n", string(b[0]))
			return false, nil
		default:
			// Move cursor back and overwrite invalid input
			fmt.Printf("\r%s [y/n]: ", prompt)
		}
	}
}
