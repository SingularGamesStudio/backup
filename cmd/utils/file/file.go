package file

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
)

// ClearDir deletes directory contents
func ClearDir(ctx context.Context, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		err = os.RemoveAll(filepath.Join(path, entry.Name()))
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

// CopyFolder copies directory recursively
func CopyFolder(ctx context.Context, src string, dest string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			err = CopyFile(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
			if err != nil {
				return err
			}
			continue
		}
		err = MkdirAll(filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
		if err != nil {
			return err
		}
		err = CopyFolder(ctx, filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
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

// CopyFile copies a file or symlink
func CopyFile(src string, dest string) error {
	stat, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if stat.Mode()&os.ModeSymlink != 0 { // copy symbolic link
		file, err := os.Readlink(src)
		if err != nil {
			return err
		}
		err = os.Symlink(file, dest)
		if err != nil {
			return err
		}
		return CopyRights(src, dest)
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
	if err != nil {
		return err
	}
	return CopyRights(src, dest)
}

// MkdirAll calls os.MkdirAll(dest) with mode from src
func MkdirAll(src string, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	err = os.MkdirAll(dest, info.Mode())
	if err != nil {
		return err
	}
	return CopyRights(src, dest)
}

// CopyRights copies file uid, gid, and mode
func CopyRights(src string, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if runtime.GOOS != "windows" { //save uid/gid
		gid := reflect.ValueOf(info.Sys()).FieldByName("Gid").Uint()
		uid := reflect.ValueOf(info.Sys()).FieldByName("Uid").Uint()
		err = os.Lchown(dest, int(uid), int(gid))
		if err != nil {
			return err
		}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		err := os.Chmod(dest, info.Mode())
		if err != nil {
			return err
		}
	}
	return nil
}
