package main

import "testing"

func TestBuildSegments(t *testing.T) {
	tracks := []Track{
		{Title: "A", Start: 0, LineNumber: 1},
		{Title: "B", Start: 60, LineNumber: 2},
		{Title: "C", Start: 150, LineNumber: 3},
	}

	segments, err := BuildSegments(tracks, 200)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(segments))
	}

	want := []struct{ start, end float64 }{
		{0, 60},
		{60, 150},
		{150, 200},
	}
	for i, w := range want {
		if segments[i].Start != w.start || segments[i].End != w.end {
			t.Errorf("segment %d: got [%v, %v], want [%v, %v]", i, segments[i].Start, segments[i].End, w.start, w.end)
		}
	}
}

func TestBuildSegmentsLastTrackAfterEnd(t *testing.T) {
	tracks := []Track{
		{Title: "A", Start: 0, LineNumber: 1},
		{Title: "B", Start: 250, LineNumber: 2},
	}

	if _, err := BuildSegments(tracks, 200); err == nil {
		t.Fatal("expected error when last track starts after real file duration")
	}
}

func TestBuildSegmentsLastTrackZeroDuration(t *testing.T) {
	tracks := []Track{
		{Title: "A", Start: 0, LineNumber: 1},
		{Title: "B", Start: 200, LineNumber: 2},
	}

	if _, err := BuildSegments(tracks, 200); err == nil {
		t.Fatal("expected error when last track would have zero duration")
	}
}

func TestBuildSegmentsSingleTrack(t *testing.T) {
	tracks := []Track{{Title: "Only", Start: 0, LineNumber: 1}}

	segments, err := BuildSegments(tracks, 120)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(segments) != 1 || segments[0].End != 120 {
		t.Fatalf("unexpected segments: %+v", segments)
	}
}
