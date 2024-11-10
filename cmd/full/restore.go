package full

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

func Restore(ctx context.Context, dir string, backupDir string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		utils.PrintError("creating target directory", err)
		return err
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) > 0 {
		if !utils.AskForConfirmation(fmt.Sprintf("Directory %s is not empty, if you proceed, files inside will be deleted. Proceed?", dir)) {
			utils.PrintError("", utils.ErrAborted)
			return err
		}
		fmt.Println("Cleaning up target directory...")
		err = file.ClearDir(ctx, dir)
		if err != nil {
			utils.PrintError("cleaning up old files", err)
			return err
		}
	}
	fmt.Println("Copying data...")
	err = file.CopyFolder(ctx, backupDir, dir)
	if err != nil {
		utils.PrintError("copying files", err)
		return err
	}
	fmt.Println("Deleting backup metadata...")
	err = os.Remove(filepath.Join(dir, utils.Metadata))
	if err != nil {
		utils.PrintError("Deleting backup metadata", err)
		fmt.Printf("Restore successful, but file %s failed to be deleted, consider deleting it manually\n", utils.Metadata)
		return err
	}
	fmt.Println("Restore successful")
	return nil
}
