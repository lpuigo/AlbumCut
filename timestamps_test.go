package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "tracks.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	return path
}

func TestParseTimestampsValidMixedFormats(t *testing.T) {
	content := "00:00:00 - My Pulse Got Lost in the Air Vents\n" +
		"00:05:07 - Thunder in the Marrow\n" +
		"9:57 - Exit Wound Waltz\n" +
		"00:14:26 -\n"

	tracks, err := ParseTimestamps(writeTempFile(t, content))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tracks) != 4 {
		t.Fatalf("expected 4 tracks, got %d", len(tracks))
	}

	want := []struct {
		title string
		start int
	}{
		{"My Pulse Got Lost in the Air Vents", 0},
		{"Thunder in the Marrow", 5*60 + 7},
		{"Exit Wound Waltz", 9*60 + 57},
		{"Untitled", 14*60 + 26},
	}
	for i, w := range want {
		if tracks[i].Title != w.title {
			t.Errorf("track %d: title = %q, want %q", i, tracks[i].Title, w.title)
		}
		if tracks[i].Start != w.start {
			t.Errorf("track %d: start = %d, want %d", i, tracks[i].Start, w.start)
		}
	}
}

func TestParseTimestampsPreservesInternalSpaces(t *testing.T) {
	tracks, err := ParseTimestamps(writeTempFile(t, "00:00:00 - A   Title  With Spaces\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := tracks[0].Title, "A   Title  With Spaces"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestParseTimestampsMalformedLine(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, "00:00:00 - Track One\nnot a timestamp\n"))
	if err == nil {
		t.Fatal("expected error for malformed line")
	}
	perr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if perr.Line != 2 {
		t.Errorf("expected error on line 2, got line %d", perr.Line)
	}
}

func TestParseTimestampsOutOfOrder(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, "00:01:00 - Track One\n00:00:30 - Track Two\n"))
	if err == nil {
		t.Fatal("expected error for out-of-order timestamps")
	}
	perr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T: %v", err, err)
	}
	if perr.Line != 2 {
		t.Errorf("expected error on line 2, got line %d", perr.Line)
	}
}

func TestParseTimestampsDuplicate(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, "00:01:00 - Track One\n00:01:00 - Track Two\n"))
	if err == nil {
		t.Fatal("expected error for duplicate timestamps")
	}
}

func TestParseTimestampsEmptyFile(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, ""))
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseTimestampsOnlyBlankLines(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, "\n   \n\t\n"))
	if err == nil {
		t.Fatal("expected error for file without valid lines")
	}
}

func TestParseTimestampsInvalidMinutesSeconds(t *testing.T) {
	_, err := ParseTimestamps(writeTempFile(t, "00:99:00 - Track One\n"))
	if err == nil {
		t.Fatal("expected error for out-of-range minutes")
	}
}

func TestParseTimestampsMissingFile(t *testing.T) {
	_, err := ParseTimestamps(filepath.Join(t.TempDir(), "missing.txt"))
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
