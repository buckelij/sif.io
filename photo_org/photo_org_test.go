package photo_org

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOrganizeDefaultsToDryRun(t *testing.T) {
	root := t.TempDir()
	photo := filepath.Join(root, "photo.jpg")
	video := filepath.Join(root, "video.mp4")
	if err := os.WriteFile(photo, []byte("photo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(video, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}

	prevExtractor := dateExtractor
	dateExtractor = func(path string) (time.Time, bool) {
		if strings.HasSuffix(path, "photo.jpg") {
			return time.Date(2024, 7, 11, 12, 0, 0, 0, time.UTC), true
		}
		return time.Time{}, false
	}
	defer func() { dateExtractor = prevExtractor }()

	result, err := Organize(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	if result.ProcessedFile != 2 {
		t.Fatalf("expected 2 processed files, got %d", result.ProcessedFile)
	}
	if result.MovedCount != 0 {
		t.Fatalf("expected no moved files in dry-run, got %d", result.MovedCount)
	}
	if result.UndatedCount != 1 {
		t.Fatalf("expected 1 undated file, got %d", result.UndatedCount)
	}

	if _, err := os.Stat(photo); err != nil {
		t.Fatalf("photo should remain in place during dry-run: %v", err)
	}
	if _, err := os.Stat(video); err != nil {
		t.Fatalf("video should remain in place during dry-run: %v", err)
	}
}

func TestOrganizeMovesFiles(t *testing.T) {
	root := t.TempDir()
	photo := filepath.Join(root, "photo.jpg")
	video := filepath.Join(root, "video.mp4")
	if err := os.WriteFile(photo, []byte("photo"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(video, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}

	prevExtractor := dateExtractor
	dateExtractor = func(path string) (time.Time, bool) {
		if strings.HasSuffix(path, "photo.jpg") {
			return time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC), true
		}
		return time.Time{}, false
	}
	defer func() { dateExtractor = prevExtractor }()

	result, err := Organize(root, &Options{DryRun: false})
	if err != nil {
		t.Fatal(err)
	}

	if result.MovedCount != 2 {
		t.Fatalf("expected 2 moved files, got %d", result.MovedCount)
	}

	if _, err := os.Stat(filepath.Join(root, "2023-12", "photo.jpg")); err != nil {
		t.Fatalf("expected dated destination file: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "undated", "video.mp4")); err != nil {
		t.Fatalf("expected undated destination file: %v", err)
	}
	if _, err := os.Stat(photo); !os.IsNotExist(err) {
		t.Fatalf("expected source photo to be moved, err=%v", err)
	}
	if _, err := os.Stat(video); !os.IsNotExist(err) {
		t.Fatalf("expected source video to be moved, err=%v", err)
	}
}

func TestOrganizeResolvesNameCollisions(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	datedDir := filepath.Join(root, "2024-01")
	if err := os.MkdirAll(datedDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(datedDir, "photo.jpg"), []byte("existing"), 0o644); err != nil {
		t.Fatal(err)
	}

	prevExtractor := dateExtractor
	dateExtractor = func(path string) (time.Time, bool) {
		return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), true
	}
	defer func() { dateExtractor = prevExtractor }()

	_, err := Organize(root, &Options{DryRun: false})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(datedDir, "photo_1.jpg")); err != nil {
		t.Fatalf("expected collision-safe filename, err=%v", err)
	}
}
