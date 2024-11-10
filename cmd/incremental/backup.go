package incremental

import (
	"context"
	"fmt"
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
	found, err := backup.CheckJson(dir)
	if err == nil && found {
		if !utils.AskForConfirmation(".backup.json found in source directory, it will be deleted in backup. Proceed?") {
			err = utils.ErrAborted
		}
	}
	if err != nil {
		utils.PrintError("checking for .backup.json in source", err)
		return
	}
	fmt.Println("Looking for latest full backup...")
	base, err := latestFull(ctx, targetDir)
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
	err = saveDeleted(ctx, base, dir, backupDir)
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
