package ffmpeg

import (
	"os"
	"path/filepath"
	"testing"

	"AlbumCut/internal/testutil"
)

func TestCheckAvailableMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if err := CheckAvailable(); err == nil {
		t.Fatal("expected error when ffmpeg/ffprobe are not in PATH")
	}
}

func TestProbeDurationIntegration(t *testing.T) {
	testutil.RequireFFmpeg(t)
	dir := t.TempDir()
	path := testutil.GenerateTestMP3(t, dir, "test.mp3", 12)

	duration, err := ProbeDuration(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if duration < 11 || duration > 13 {
		t.Errorf("duration = %v, want ~12s", duration)
	}
}

func TestProbeDurationMissingFile(t *testing.T) {
	testutil.RequireFFmpeg(t)
	_, err := ProbeDuration(filepath.Join(t.TempDir(), "missing.mp3"))
	if err == nil {
		t.Fatal("expected error probing a missing file")
	}
}

func TestSplitSegmentIntegration(t *testing.T) {
	testutil.RequireFFmpeg(t)
	dir := t.TempDir()
	path := testutil.GenerateTestMP3(t, dir, "album.mp3", 12)

	out := filepath.Join(dir, "out.mp3")
	if err := SplitSegment(path, 2, 8, out); err != nil {
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
	testutil.RequireFFmpeg(t)
	dir := t.TempDir()
	bogus := filepath.Join(dir, "not-audio.mp3")
	if err := os.WriteFile(bogus, []byte("this is definitely not mp3 data"), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	out := filepath.Join(dir, "out.mp3")
	if err := SplitSegment(bogus, 0, 1, out); err == nil {
		t.Fatal("expected error splitting an invalid source file")
	}
}
