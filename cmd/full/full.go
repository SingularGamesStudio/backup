package full

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

func Backup(ctx context.Context, dir string, target_dir string) {
	backup_dir, err := backup.Setup(ctx, target_dir)
	if err != nil {
		utils.PrintError("setting up backup folder", err)
		return
	}
	found, err := backup.CheckJson(dir)
	if err != nil {
		utils.PrintError("checking for .backup.json in source", err)
		return
	}
	if found {
		if !utils.AskForConfirmation(".backup.json found in source directory, it will be deleted in backup. Proceed?") {
			utils.PrintError("checking for .backup.json in source", utils.ErrAborted)
			return
		}
	}
	fmt.Println("Copying data...")
	err = file.CopyFolder(ctx, dir, backup_dir)
	if err != nil {
		utils.PrintError("copying files", err)
		abortBackup(backup_dir)
		return
	}
	fmt.Println("Saving backup metadata...")
	err = saveInfo(backup_dir)
	if err != nil {
		utils.PrintError("saving backup metadata", err)
		abortBackup(backup_dir)
		return
	}
	fmt.Println("Backup successful")
}

// saveInfo saves backup metadata in .backup.json
func saveInfo(backup_dir string) error {
	info, err := os.Create(filepath.Join(backup_dir, ".backup.json"))
	if err != nil {
		return err
	}
	defer info.Close()
	data, err := json.Marshal(backup.Info{Type: "full"})
	if err != nil {
		return err
	}
	_, err = info.Write(data)
	return err
}

// abortBackup tries to delete everything in backup_dir
func abortBackup(backup_dir string) {
	fmt.Println("Attempting to clean up failed backup file copies...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := file.ClearDir(ctx, backup_dir)
	if err != nil {
		fmt.Printf("Failed to abort backup (%s), files in %s must be deleted manually\n", err.Error(), backup_dir)
	} else {
		fmt.Println("Backup aborted")
	}
}
