package incremental

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/utils"
)

func Backup(ctx context.Context, dir string, targetDir string) {
	backupDir, err := backup.Setup(ctx, targetDir)
	if err != nil {
		utils.PrintError("setting up backup folder", err)
		return
	}
	foundWrongExt, err := checkExts(ctx, dir)
	if err != nil {
		utils.PrintError("traversing target folder", err)
		return
	}
	if foundWrongExt {
		fmt.Printf("Files with extension %s found in %s, they are not supported for incremental backup. Aborting.\n", utils.DeletedExt, dir)
		return
	}
	found, err := backup.CheckJson(dir)
	if err == nil && found {
		if !utils.AskForConfirmation(fmt.Sprintf("%s found in source directory, it will be deleted in backup. Proceed?", utils.Metadata)) {
			err = utils.ErrAborted
		}
	}
	if err != nil {
		utils.PrintError(fmt.Sprintf("checking for %s in source", utils.Metadata), err)
		return
	}
	fmt.Println("Looking for latest full backup...")
	base, err := Latest(ctx, targetDir, true)
	if err != nil {
		utils.PrintError("looking for latest full backup", err)
		return
	}
	fmt.Println("Saving diff between full backup and current state...")
	err = saveChanged(ctx, base, dir, backupDir)
	if err != nil {
		utils.PrintError("calculating and saving diff", err)
		backup.TryAbort(backupDir)
		return
	}
	fmt.Println("Saving info about deleted files...")
	err = saveDeleted(ctx, base, dir, backupDir, true)
	if err != nil {
		utils.PrintError("calculating and saving diff (deleted files)", err)
		backup.TryAbort(backupDir)
		return
	}
	fmt.Println("Saving backup metadata...")
	err = backup.SaveInfo(backupDir, backup.Info{Type: "incremental", Base: filepath.Base(base)})
	if err != nil {
		utils.PrintError("saving backup metadata", err)
		backup.TryAbort(backupDir)
		return
	}
	fmt.Println("Backup successful")
}

// checkExts checks if there are files with extension .deleted
func checkExts(ctx context.Context, path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == utils.DeletedExt {
			return true, nil
		}
		if entry.IsDir() {
			found, err := checkExts(ctx, filepath.Join(path, entry.Name()))
			if err != nil {
				return false, err
			}
			if found {
				return true, nil
			}
		}
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
	}
	return false, nil
}
