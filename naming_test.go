package main

import (
	"path/filepath"
	"testing"
)

func TestAlbumTitle(t *testing.T) {
	if got, want := AlbumTitle(filepath.Join("C:", "music", "MonAlbum.mp3")), "MonAlbum"; got != want {
		t.Errorf("AlbumTitle() = %q, want %q", got, want)
	}
}

func TestOutputPathBasic(t *testing.T) {
	mp3Path := filepath.Join("C:", "music", "MonAlbum.mp3")
	got := OutputPath(mp3Path, 1, "My Pulse Got Lost in the Air Vents")
	want := filepath.Join("C:", "music", "01. MonAlbum_My Pulse Got Lost in the Air Vents.mp3")
	if got != want {
		t.Errorf("OutputPath() = %q, want %q", got, want)
	}
}

func TestOutputPathTrackNumberPadding(t *testing.T) {
	mp3Path := filepath.Join("music", "Album.mp3")

	if got, want := OutputPath(mp3Path, 1, "T"), filepath.Join("music", "01. Album_T.mp3"); got != want {
		t.Errorf("track 1: got %q, want %q", got, want)
	}
	if got, want := OutputPath(mp3Path, 11, "T"), filepath.Join("music", "11. Album_T.mp3"); got != want {
		t.Errorf("track 11: got %q, want %q", got, want)
	}
	if got, want := OutputPath(mp3Path, 100, "T"), filepath.Join("music", "100. Album_T.mp3"); got != want {
		t.Errorf("track 100: got %q, want %q", got, want)
	}
}

func TestOutputPathSanitizesForbiddenChars(t *testing.T) {
	mp3Path := filepath.Join("music", "Album.mp3")
	got := OutputPath(mp3Path, 2, `Weird: Title / With \ Bad * Chars ? " < > |`)
	want := filepath.Join("music", "02. Album_Weird_ Title _ With _ Bad _ Chars _ _ _ _ _.mp3")
	if got != want {
		t.Errorf("OutputPath() = %q, want %q", got, want)
	}
}

func TestOutputPathKeepsSpaces(t *testing.T) {
	mp3Path := filepath.Join("music", "Album.mp3")
	got := OutputPath(mp3Path, 4, "A   Title  With Spaces")
	want := filepath.Join("music", "04. Album_A   Title  With Spaces.mp3")
	if got != want {
		t.Errorf("OutputPath() = %q, want %q", got, want)
	}
}

func TestOutputPathSameDirAsSource(t *testing.T) {
	mp3Path := filepath.Join("some", "dir", "Album.mp3")
	got := OutputPath(mp3Path, 1, "T")
	if filepath.Dir(got) != filepath.Join("some", "dir") {
		t.Errorf("output dir = %q, want %q", filepath.Dir(got), filepath.Join("some", "dir"))
	}
}
