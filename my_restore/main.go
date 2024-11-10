package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/full"
	"github.com/SingularGamesStudio/backup/cmd/incremental"
	"github.com/SingularGamesStudio/backup/cmd/utils"
)

func main() {
	backupDir := os.Args[1]
	dir := os.Args[2]
	info, err := backup.GetJson(backupDir)
	if err != nil {
		print("Failed to read %s: %s", filepath.Join(backupDir, utils.Metadata), err.Error())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool, 1)
	go func() {
		switch info.Type {
		case "full":
			_ = full.Restore(ctx, dir, backupDir)
		case "incremental":
			incremental.Restore(ctx, dir, backupDir, filepath.Join(filepath.Dir(backupDir), info.Base))
		default:
			fmt.Printf("Error: unknown backup type in %s: %s, supported types are incremental and full", filepath.Join(backupDir, utils.Metadata), info.Type)
		}
		done <- true
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case <-quit:
	case <-done:
		return
	}
	fmt.Println("Shutting down gracefully...")
	cancel()
	<-done
}
