package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrAborted = errors.New("Backup aborted by user")

func AskForConfirmation(s string) bool {
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

func SetupBackup(ctx context.Context, path string) (string, error) {
	path = filepath.Join(path, time.Now().Format("2006-01-02_15-04-05"))
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return "", err
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	if len(entries) > 0 {
		if !AskForConfirmation(fmt.Sprintf("Directory %s is not empty, if you proceed, files inside will be deleted. Proceed?", path)) {
			return "", ErrAborted
		}
		fmt.Println("Cleaning up backup directory...")
		err = ClearDir(ctx, path)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

func CheckBackupJson(path string) (bool, error) {
	if _, err := os.Stat(filepath.Join(path, ".backup.json")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}

func ClearDir(ctx context.Context, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		err = os.RemoveAll(filepath.Join(path, entry.Name()))
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func PrintError(context string, err error) {
	if errors.Is(err, os.ErrPermission) {
		fmt.Println(fmt.Errorf("Permission denied in {%s}: %w", context, err))
	} else if errors.Is(err, ErrAborted) {
		fmt.Println(fmt.Errorf("Aborted by user in {%s}", context))
	} else {
		fmt.Println(fmt.Errorf("Error in {%s}: %w", context, err))
	}
}
