package photo_org

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// Options controls organizer behavior.
type Options struct {
	// DryRun previews operations without creating directories or moving files.
	DryRun bool
}

// Action represents a single file organization attempt.
type Action struct {
	Source      string
	Destination string
	Dated       bool
	Moved       bool
}

// Result summarizes a run.
type Result struct {
	Actions       []Action
	MovedCount    int
	UndatedCount  int
	ProcessedFile int
}

// Organize scans the top-level directory for files and moves each file into
// YYYY-MM (from metadata date-created) or undated when no date can be found.
//
// Only Linux and Windows are supported.
//
// Default behavior is dry-run when opts is nil.
func Organize(topDir string, opts *Options) (Result, error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		return Result{}, fmt.Errorf("unsupported OS %q: only linux and windows are supported", runtime.GOOS)
	}

	entries, err := os.ReadDir(topDir)
	if err != nil {
		return Result{}, err
	}

	dryRun := true
	if opts != nil {
		dryRun = opts.DryRun
	}

	result := Result{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		src := filepath.Join(topDir, entry.Name())
		createdAt, ok := dateExtractor(src)
		folder := "undated"
		if ok {
			folder = createdAt.Format("2006-01")
		} else {
			result.UndatedCount++
		}

		destDir := filepath.Join(topDir, folder)
		destPath, err := nextAvailablePath(destDir, entry.Name())
		if err != nil {
			return result, err
		}

		action := Action{
			Source:      src,
			Destination: destPath,
			Dated:       ok,
			Moved:       false,
		}

		if !dryRun {
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return result, err
			}

			if err := moveFile(src, destPath); err != nil {
				return result, err
			}
			action.Moved = true
			result.MovedCount++
		}

		result.Actions = append(result.Actions, action)
		result.ProcessedFile++
	}

	return result, nil
}

var dateExtractor = ExtractDateCreated

// ExtractDateCreated attempts to read date-created from metadata.
func ExtractDateCreated(path string) (time.Time, bool) {
	f, err := os.Open(path)
	if err != nil {
		return time.Time{}, false
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return time.Time{}, false
	}

	for _, tagName := range []exif.FieldName{
		exif.FieldName("DateTimeOriginal"),
		exif.FieldName("CreateDate"),
		exif.FieldName("DateTimeDigitized"),
		exif.DateTime,
	} {
		t, ok := parseExifTagTime(x, tagName)
		if ok {
			return t, true
		}
	}

	return time.Time{}, false
}

func parseExifTagTime(x *exif.Exif, tagName exif.FieldName) (time.Time, bool) {
	tag, err := x.Get(tagName)
	if err != nil {
		return time.Time{}, false
	}

	v, err := tag.StringVal()
	if err != nil {
		return time.Time{}, false
	}

	v = strings.TrimSpace(strings.Trim(v, "\""))
	if v == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{
		"2006:01:02 15:04:05",
		"2006:01:02 15:04:05-07:00",
		"2006:01:02 15:04:05-0700",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z07:00",
	} {
		if t, err := time.Parse(layout, v); err == nil {
			return t, true
		}
	}

	return time.Time{}, false
}

func nextAvailablePath(dir string, base string) (string, error) {
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	if name == "" {
		name = base
	}

	candidate := filepath.Join(dir, base)
	if _, err := os.Stat(candidate); os.IsNotExist(err) {
		return candidate, nil
	} else if err != nil {
		if os.IsNotExist(err) {
			return candidate, nil
		}
		if os.IsPermission(err) {
			return "", err
		}
	}

	for i := 1; ; i++ {
		candidate = filepath.Join(dir, name+"_"+strconv.Itoa(i)+ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil && !os.IsNotExist(err) {
			return "", err
		}
	}
}

func moveFile(src string, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(destFile, srcFile)
	closeErr := destFile.Close()
	if copyErr != nil {
		_ = os.Remove(dest)
		return copyErr
	}
	if closeErr != nil {
		_ = os.Remove(dest)
		return closeErr
	}

	return os.Remove(src)
}
