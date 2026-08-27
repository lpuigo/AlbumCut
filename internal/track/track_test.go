package track

import (
	"strings"
	"testing"
)

func linesOf(content string) []string {
	if content == "" {
		return nil
	}
	return strings.Split(strings.TrimRight(content, "\n"), "\n")
}

func TestParseTracksValidMixedFormats(t *testing.T) {
	content := "00:00:00 - My Pulse Got Lost in the Air Vents\n" +
		"00:05:07 - Thunder in the Marrow\n" +
		"9:57 - Exit Wound Waltz\n" +
		"00:14:26 -\n"

	tracks, err := ParseTracks(linesOf(content))
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

func TestParseTracksPreservesInternalSpaces(t *testing.T) {
	tracks, err := ParseTracks(linesOf("00:00:00 - A   Title  With Spaces\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := tracks[0].Title, "A   Title  With Spaces"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestParseTracksMalformedLine(t *testing.T) {
	_, err := ParseTracks(linesOf("00:00:00 - Track One\nnot a timestamp\n"))
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

func TestParseTracksOutOfOrder(t *testing.T) {
	_, err := ParseTracks(linesOf("00:01:00 - Track One\n00:00:30 - Track Two\n"))
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

func TestParseTracksDuplicate(t *testing.T) {
	_, err := ParseTracks(linesOf("00:01:00 - Track One\n00:01:00 - Track Two\n"))
	if err == nil {
		t.Fatal("expected error for duplicate timestamps")
	}
}

func TestParseTracksEmptyFile(t *testing.T) {
	_, err := ParseTracks(linesOf(""))
	if err == nil {
		t.Fatal("expected error for empty file")
	}
}

func TestParseTracksOnlyBlankLines(t *testing.T) {
	_, err := ParseTracks(linesOf("\n   \n\t\n"))
	if err == nil {
		t.Fatal("expected error for file without valid lines")
	}
}

func TestParseTracksInvalidMinutesSeconds(t *testing.T) {
	_, err := ParseTracks(linesOf("00:99:00 - Track One\n"))
	if err == nil {
		t.Fatal("expected error for out-of-range minutes")
	}
}
