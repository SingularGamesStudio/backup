package incremental

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SingularGamesStudio/backup/cmd/full"
	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

func Restore(ctx context.Context, dir string, backupDir string, fullDir string) {
	fmt.Println("Restoring latest full backup...")
	err := full.Restore(ctx, dir, fullDir)
	if err != nil {
		return
	}
	fmt.Println("Applying incremental backup...")
	err = applyChanged(ctx, backupDir, dir)
	if err != nil {
		utils.PrintError("Applying incremental backup", err)
		return
	}
	fmt.Println("Deleting backup metadata...")
	err = os.Remove(filepath.Join(dir, ".backup.json"))
	if err != nil {
		utils.PrintError("Deleting backup metadata", err)
		fmt.Println("Restore successful, but file .backup.json failed to be deleted, consider deleting it manually")
		return
	}
	fmt.Println("Restore successful")
}

func applyChanged(ctx context.Context, src string, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".deleted" {
			err = os.RemoveAll(filepath.Join(dest, entry.Name()[:len(entry.Name())-8]))
			if err != nil {
				return err
			}
			continue
		}
		if !entry.IsDir() {
			err = file.CopyFile(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
			if err != nil {
				return err
			}
			continue
		}
		err = os.MkdirAll(filepath.Join(dest, entry.Name()), os.ModePerm)
		if err != nil {
			return err
		}
		err = applyChanged(ctx, filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
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
