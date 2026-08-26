package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// requireFFmpeg ignore le test si ffmpeg/ffprobe ne sont pas disponibles
// dans le PATH de la machine executant les tests.
func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg introuvable dans le PATH, test d'integration ignore")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe introuvable dans le PATH, test d'integration ignore")
	}
}

// generateTestMP3 genere, via ffmpeg, un fichier MP3 de test (sinusoide)
// de la duree demandee.
func generateTestMP3(t *testing.T, dir, filename string, durationSeconds int) string {
	t.Helper()
	path := filepath.Join(dir, filename)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=440:duration=%d", durationSeconds),
		"-c:a", "libmp3lame",
		"-b:a", "64k",
		path,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generation du fichier de test %s impossible : %v\n%s", path, err, stderr.String())
	}
	return path
}

func TestProbeDurationIntegration(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	path := generateTestMP3(t, dir, "test.mp3", 12)

	duration, err := ProbeDuration(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if duration < 11 || duration > 13 {
		t.Errorf("duration = %v, want ~12s", duration)
	}
}

func TestProbeDurationMissingFile(t *testing.T) {
	requireFFmpeg(t)
	_, err := ProbeDuration(filepath.Join(t.TempDir(), "missing.mp3"))
	if err == nil {
		t.Fatal("expected error probing a missing file")
	}
}

func TestSplitSegmentIntegration(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	path := generateTestMP3(t, dir, "album.mp3", 12)

	seg := Segment{Track: Track{Title: "Excerpt"}, Start: 2, End: 8}
	out := filepath.Join(dir, "out.mp3")
	if err := SplitSegment(path, seg, out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(out)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("output file is empty")
	}

	duration, err := ProbeDuration(out)
	if err != nil {
		t.Fatalf("failed to probe output: %v", err)
	}
	if duration < 5 || duration > 7 {
		t.Errorf("segment duration = %v, want ~6s", duration)
	}
}

func TestSplitSegmentInvalidSource(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	bogus := filepath.Join(dir, "not-audio.mp3")
	if err := os.WriteFile(bogus, []byte("this is definitely not mp3 data"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	seg := Segment{Track: Track{Title: "X"}, Start: 0, End: 1}
	out := filepath.Join(dir, "out.mp3")
	if err := SplitSegment(bogus, seg, out); err == nil {
		t.Fatal("expected error splitting an invalid source file")
	}
}

func TestCheckExternalToolsMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := checkExternalTools(); err == nil {
		t.Fatal("expected error when ffmpeg/ffprobe are not in PATH")
	}
}

func TestRunEndToEnd(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	mp3Path := generateTestMP3(t, dir, "MonAlbum.mp3", 12)

	tracksContent := "00:00:00 - Track One\n00:00:05 - Track Two\n00:00:09 - Untitled\n"
	tracksPath := filepath.Join(dir, "tracks.txt")
	if err := os.WriteFile(tracksPath, []byte(tracksContent), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := run([]string{mp3Path, tracksPath}); err != nil {
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
	requireFFmpeg(t)
	dir := t.TempDir()
	mp3Path := generateTestMP3(t, dir, "Album.mp3", 5)

	tracksContent := "00:00:00 - Track One\n00:00:10 - Track Two\n"
	tracksPath := filepath.Join(dir, "tracks.txt")
	if err := os.WriteFile(tracksPath, []byte(tracksContent), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	if err := run([]string{mp3Path, tracksPath}); err == nil {
		t.Fatal("expected error when last timestamp is after the real end of the file")
	}
}
