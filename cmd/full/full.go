package full

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/utils"
)

func Backup(ctx context.Context, dir string, target_dir string) {
	backup_dir, err := utils.SetupBackup(ctx, target_dir)
	if err != nil {
		utils.PrintError("setting up backup folder", err)
		return
	}
	found, err := utils.CheckBackupJson(dir)
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
	err = copyFolder(ctx, dir, backup_dir)
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

func copyFolder(ctx context.Context, src string, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			err = copyFile(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
			if err != nil {
				return err
			}
			continue
		}
		err = os.MkdirAll(filepath.Join(dest, entry.Name()), os.ModePerm)
		if err != nil {
			return err
		}
		err = copyFolder(ctx, filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
		if err != nil {
			return err
		} //TODO:copy ownership
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

func copyFile(src string, dest string) error {
	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if stat.Mode()&os.ModeSymlink != 0 {
		file, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(file, dest)
	}

	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()
	to, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

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

func abortBackup(backup_dir string) {
	fmt.Println("Attempting to clean up failed backup file copies...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := utils.ClearDir(ctx, backup_dir)
	if err != nil {
		fmt.Printf("Failed to abort backup (%s), files in %s must be deleted manually\n", err.Error(), backup_dir)
	} else {
		fmt.Println("Backup aborted")
	}
}
