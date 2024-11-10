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
	err = os.Remove(filepath.Join(dir, utils.Metadata))
	if err != nil {
		utils.PrintError("Deleting backup metadata", err)
		fmt.Printf("Restore successful, but file %s failed to be deleted, consider deleting it manually\n", utils.Metadata)
		return
	}
	fmt.Println("Restore successful")
}

// applyChanged applies incremental backup to full
func applyChanged(ctx context.Context, src string, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == utils.DeletedExt { // file deleted
			err = os.RemoveAll(filepath.Join(dest, entry.Name()[:len(entry.Name())-len(utils.DeletedExt)]))
			if err != nil {
				return err
			}
			continue
		}
		if !entry.IsDir() { //file changed
			err = file.CopyFile(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
			if err != nil {
				return err
			}
			continue
		}
		err = os.MkdirAll(filepath.Join(dest, entry.Name()), os.ModePerm) //directory changed
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
