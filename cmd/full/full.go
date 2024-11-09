package full

import (
	"context"
	"fmt"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
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
	fmt.Println("Copying data...")
	err = file.CopyFolder(ctx, dir, backupDir)
	if err != nil {
		utils.PrintError("copying files", err)
		backup.TryAbort(backupDir)
		return
	}
	fmt.Println("Saving backup metadata...")
	err = backup.SaveInfo(backupDir, backup.Info{Type: "full"})
	if err != nil {
		utils.PrintError("saving backup metadata", err)
		backup.TryAbort(backupDir)
		return
	}
	fmt.Println("Backup successful")
}
