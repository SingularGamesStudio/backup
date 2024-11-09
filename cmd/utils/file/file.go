package file

import (
	"context"
	"io"
	"os"
	"path/filepath"
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
		err = os.MkdirAll(filepath.Join(dest, entry.Name()), os.ModePerm)
		if err != nil {
			return err
		}
		err = CopyFolder(ctx, filepath.Join(src, entry.Name()), filepath.Join(dest, entry.Name()))
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

// CopyFolder copies a file or symlink
func CopyFile(src string, dest string) error {
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
