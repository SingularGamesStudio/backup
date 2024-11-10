package utils

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	Metadata   = ".backup.json"
	DeletedExt = ".deleted"
)

var ErrAborted = errors.New("Backup aborted by user")

// Yes returns true for any prompts
var Yes = false

// AskForConfirmation creates a yes/no prompt for the user
func AskForConfirmation(s string) bool {
	if Yes {
		return true
	}
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/n]: ", s)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// PrintError prints user-friendly error message caught in given context
func PrintError(context string, err error) { //TODO:
	if errors.Is(err, os.ErrPermission) {
		fmt.Println(fmt.Errorf("Permission denied in {%s}: %w", context, err))
	} else if errors.Is(err, ErrAborted) {
		fmt.Println(fmt.Errorf("Aborted by user in {%s}", context))
	} else {
		fmt.Println(fmt.Errorf("Error in {%s}: %w", context, err))
	}
}
