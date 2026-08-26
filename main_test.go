package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cfg, err := parseArgs([]string{"album.mp3", "tracks.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.mp3Path != "album.mp3" || cfg.timestampsPath != "tracks.txt" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParseArgsWrongCount(t *testing.T) {
	for _, args := range [][]string{{}, {"only-one"}, {"a", "b", "c"}} {
		if _, err := parseArgs(args); err == nil {
			t.Fatalf("expected error for args %v", args)
		}
	}
}

func TestCheckFileExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "present.txt")
	if err := os.WriteFile(file, []byte("content"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := checkFileExists(file); err != nil {
		t.Fatalf("unexpected error for existing file: %v", err)
	}

	if err := checkFileExists(filepath.Join(dir, "missing.txt")); err == nil {
		t.Fatalf("expected error for missing file")
	} else if !strings.Contains(err.Error(), "introuvable") {
		t.Fatalf("unexpected error message: %v", err)
	}

	if err := checkFileExists(dir); err == nil {
		t.Fatalf("expected error when path is a directory")
	}
}
