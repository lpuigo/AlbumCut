package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"AlbumCut/internal/cli"
	"AlbumCut/internal/testutil"
)

func TestRunEndToEnd(t *testing.T) {
	testutil.RequireFFmpeg(t)
	dir := t.TempDir()
	mp3Path := testutil.GenerateTestMP3(t, dir, "MonAlbum.mp3", 12)

	tracksContent := "00:00:00 - Track One\n00:00:05 - Track Two\n00:00:09 - Untitled\n"
	tracksPath := filepath.Join(dir, "tracks.txt")
	if err := os.WriteFile(tracksPath, []byte(tracksContent), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := cli.Run([]string{mp3Path, tracksPath}, io.Discard); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantFiles := []string{
		"01. MonAlbum_Track One.mp3",
		"02. MonAlbum_Track Two.mp3",
		"03. MonAlbum_Untitled.mp3",
	}
	for _, name := range wantFiles {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Errorf("expected output file %s: %v", name, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("output file %s is empty", name)
		}
	}
}

func TestRunLastTrackAfterRealEnd(t *testing.T) {
	testutil.RequireFFmpeg(t)
	dir := t.TempDir()
	mp3Path := testutil.GenerateTestMP3(t, dir, "Album.mp3", 5)

	tracksContent := "00:00:00 - Track One\n00:00:10 - Track Two\n"
	tracksPath := filepath.Join(dir, "tracks.txt")
	if err := os.WriteFile(tracksPath, []byte(tracksContent), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := cli.Run([]string{mp3Path, tracksPath}, io.Discard); err == nil {
		t.Fatal("expected error when last timestamp is after the real end of the file")
	}
}
