package cmd_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/full"
	"github.com/SingularGamesStudio/backup/cmd/incremental"
	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

func TestFull(t *testing.T) {
	utils.Yes = true
	_ = file.ClearDir(context.Background(), "testdata/backup")
	full.Backup(context.Background(), "testdata/src", "testdata/backup")
	folder, _ := incremental.Latest(context.Background(), "testdata/backup", true)
	_ = full.Restore(context.Background(), "testdata/temp", folder)
	defer func() {
		_ = os.RemoveAll("testdata/temp")
	}()
	if !checkSame("testdata/src", "testdata/temp", t) {
		t.Error("dirs different")
	}
}

func TestIncremental(t *testing.T) {
	time.Sleep(2 * time.Second)
	utils.Yes = true
	incremental.Backup(context.Background(), "testdata/src", "testdata/backup")
	inc, _ := incremental.Latest(context.Background(), "testdata/backup", false)
	info, _ := backup.GetJson(inc)
	fmt.Println(inc, filepath.Join("testdata/backup", info.Base))
	incremental.Restore(context.Background(), "testdata/temp", inc, filepath.Join("testdata/backup", info.Base))
	defer func() {
		_ = os.RemoveAll("testdata/temp")
	}()
	if !checkSame("testdata/src", "testdata/temp", t) {
		t.Error("dirs different")
	}
	_ = file.ClearDir(context.Background(), "testdata/backup")
}

func checkSame(src string, dest string, t *testing.T) bool {
	if filepath.Base(src) == ".backup.json" || filepath.Base(dest) == ".backup.json" {
		return true
	}
	info, err := os.Lstat(src)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		if info.Mode()&os.ModeSymlink != 0 {
			info2, err := os.Lstat(dest)
			if err != nil {
				return false
			}
			return info2.Mode()&os.ModeSymlink != 0
		}
		srcFile, err := os.Open(src)
		if err != nil {
			return false
		}
		defer srcFile.Close()
		destFile, err := os.Open(dest)
		if err != nil {
			return false
		}
		defer destFile.Close()
		srcData, err := io.ReadAll(srcFile)
		if err != nil {
			return false
		}
		destData, err := io.ReadAll(destFile)
		if err != nil {
			return false
		}
		return reflect.DeepEqual(srcData, destData)
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !checkSame(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()), t) {
			return false
		}
	}
	entries, err = os.ReadDir(dest)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if _, err := os.Lstat(filepath.Join(src, entry.Name())); err != nil {
			if entry.Name() != ".backup.json" {
				return false
			}
		}
	}
	return true
}
