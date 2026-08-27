package cli

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	cfg, err := parseArgs([]string{"album.mp3", "tracks.txt"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.mp3Path != "album.mp3" || cfg.timestampsPath != "tracks.txt" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	if cfg.endPadding != 0 {
		t.Fatalf("expected default endPadding 0, got %v", cfg.endPadding)
	}
}

func TestParseArgsWrongCount(t *testing.T) {
	for _, args := range [][]string{{}, {"only-one"}, {"a", "b", "c"}} {
		if _, err := parseArgs(args, io.Discard); err == nil {
			t.Fatalf("expected error for args %v", args)
		}
	}
}

func TestParseArgsEndPadding(t *testing.T) {
	cfg, err := parseArgs([]string{"-end-padding=3.5", "album.mp3", "tracks.txt"}, io.Discard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.endPadding != 3.5 {
		t.Fatalf("expected endPadding 3.5, got %v", cfg.endPadding)
	}
	if cfg.mp3Path != "album.mp3" || cfg.timestampsPath != "tracks.txt" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestParseArgsEndPaddingAfterPositional(t *testing.T) {
	// flag.FlagSet only recognizes flags before positional args; this
	// documents that -end-padding must come first.
	if _, err := parseArgs([]string{"album.mp3", "-end-padding=3.5", "tracks.txt"}, io.Discard); err == nil {
		t.Fatal("expected error when flag follows positional arguments")
	}
}

func TestParseArgsEndPaddingNegative(t *testing.T) {
	if _, err := parseArgs([]string{"-end-padding=-1", "album.mp3", "tracks.txt"}, io.Discard); err == nil {
		t.Fatal("expected error for negative -end-padding")
	}
}

func TestParseArgsHelp(t *testing.T) {
	for _, args := range [][]string{{"-h"}, {"--help"}, {"-help"}} {
		var buf bytes.Buffer
		if _, err := parseArgs(args, &buf); !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp for args %v, got %v", args, err)
		}
		if !strings.Contains(buf.String(), usage) {
			t.Fatalf("expected help output to contain usage, got: %q", buf.String())
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
