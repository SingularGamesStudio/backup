package backup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

type Info struct {
	Type string `json:"Type"`
	Base string `json:"Base"`
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

func GetJson(path string) (Info, error) {
	file, err := os.Open(filepath.Join(path, ".backup.json"))
	if err != nil {
		return Info{}, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return Info{}, err
	}
	res := Info{}
	err = json.Unmarshal(data, &res)
	return res, err
}

// TryAbort tries to delete everything in dir
func TryAbort(dir string) {
	fmt.Println("Attempting to clean up failed backup file copies...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := file.ClearDir(ctx, dir)
	if err != nil {
		fmt.Printf("Failed to abort backup (%s), files in %s must be deleted manually\n", err.Error(), dir)
	} else {
		fmt.Println("Backup aborted")
	}
}

// SaveInfo saves backup metadata in .backup.json
func SaveInfo(dir string, info Info) error {
	file, err := os.Create(filepath.Join(dir, ".backup.json"))
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	return err
}
