package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/SingularGamesStudio/backup/cmd/full"
	"github.com/SingularGamesStudio/backup/cmd/incremental"
)

func main() {
	backupType := os.Args[1]
	dir := os.Args[2]
	backupDir := os.Args[3]
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool, 1)
	go func() {
		switch backupType {
		case "full":
			full.Backup(ctx, dir, backupDir)
		case "incremental":
			incremental.Backup(ctx, dir, backupDir)
		default:
			fmt.Printf("Error: unknown backup type: %s, supported types are incremental and full", backupType)
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
