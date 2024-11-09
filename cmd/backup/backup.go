package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

type Info struct {
	Type string `json:"Type"`
}

// Setup creates path, and asks user to delete everything inside
func Setup(ctx context.Context, path string) (string, error) {
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
		if !utils.AskForConfirmation(fmt.Sprintf("Directory %s is not empty, if you proceed, files inside will be deleted. Proceed?", path)) {
			return "", utils.ErrAborted
		}
		fmt.Println("Cleaning up backup directory...")
		err = file.ClearDir(ctx, path)
		if err != nil {
			return "", err
		}
	}
	return path, nil
}

// CheckJson checks whether there is .backup.json file in path
func CheckJson(path string) (bool, error) {
	if _, err := os.Stat(filepath.Join(path, ".backup.json")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err == nil {
		return true, nil
	} else {
		return false, err
	}
}
