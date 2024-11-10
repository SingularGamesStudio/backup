package incremental

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/SingularGamesStudio/backup/cmd/backup"
	"github.com/SingularGamesStudio/backup/cmd/utils"
	"github.com/SingularGamesStudio/backup/cmd/utils/file"
)

// latestFull gets last full backup in dir
func latestFull(ctx context.Context, dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	latest := time.Time{}
	res := ""
	for _, entry := range entries {
		if entry.IsDir() {
			when, err := time.Parse("2006-01-02_15-04-05", entry.Name())
			if err != nil {
				continue
			}
			if !latest.IsZero() && latest.After(when) {
				continue
			}
			exists, err := backup.CheckJson(filepath.Join(dir, entry.Name()))
			if err != nil || !exists {
				continue
			}
			info, err := backup.GetJson(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			if info.Type == "full" {
				latest = when
				res = filepath.Join(dir, entry.Name())
			}
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
	}
	if res == "" {
		return "", errors.New("no valid full backup found")
	}
	return res, nil
}

// saveChanged saves files that changed in new, compared with old (except deleted ones)
func saveChanged(ctx context.Context, old string, new string, dest string) error {
	entries, err := os.ReadDir(new)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		change, err := changed(entry, old, new)
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			if change != "false" { // file changed
				err = os.MkdirAll(dest, os.ModePerm)
				if err != nil {
					return err
				}
				err = file.CopyFile(filepath.Join(new, entry.Name()), filepath.Join(dest, entry.Name()))
				if err != nil {
					return err
				}
			}
			continue
		}
		if change != "false" { // directory changed
			err = os.MkdirAll(filepath.Join(dest, entry.Name()), os.ModePerm)
			if err != nil {
				return err
			}
		}
		if change == "new" { // directory created
			err = file.CopyFolder(ctx, filepath.Join(new, entry.Name()), filepath.Join(dest, entry.Name()))
			if err != nil {
				return err
			}
			continue
		}
		// directory contents might be changed
		err = saveChanged(ctx, filepath.Join(old, entry.Name()), filepath.Join(new, entry.Name()), filepath.Join(dest, entry.Name()))
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

// saveDeleted saves files that were deleted in new, compared with old
func saveDeleted(ctx context.Context, old string, new string, dest string, root bool) error {
	entries, err := os.ReadDir(old)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if root && entry.Name() == utils.Metadata { //.backup.json is considered deleted, but it is not
			continue
		}
		change, err := changed(entry, new, old) //reversed argument order, so "new" would mean that file was deleted
		if err != nil {
			return err
		}
		if change == "new" {
			err = os.MkdirAll(dest, os.ModePerm)
			if err != nil {
				return err
			}
			file, err := os.Create(filepath.Join(dest, entry.Name()+utils.DeletedExt))
			if err != nil {
				return err
			}
			file.Close()
			continue
		}
		if entry.IsDir() {
			err = saveDeleted(ctx, filepath.Join(old, entry.Name()), filepath.Join(new, entry.Name()), filepath.Join(dest, entry.Name()), false)
			if err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	return nil
}

// changed returns how entry was changed in new, compared to old ("new" - created, "true" - modified, "false" - no change)
func changed(entry fs.DirEntry, old string, new string) (string, error) {
	oldStat, err := os.Lstat(filepath.Join(old, entry.Name()))
	if errors.Is(err, os.ErrNotExist) {
		return "new", nil
	}
	if err != nil {
		return "", err
	}
	newStat, err := os.Lstat(filepath.Join(new, entry.Name()))
	if err != nil {
		return "", err
	}
	if !newStat.ModTime().After(oldStat.ModTime()) {
		return "false", nil
	}
	if newStat.Size() != oldStat.Size() || entry.IsDir() {
		return "true", nil
	}
	return "false", nil
}
