// Package pipeline orchestre le decoupage d'un album : lecture et parsing
// du fichier de timestamps, sondage de la duree reelle du fichier source,
// calcul des segments, et decoupage piste par piste. Il ne fait aucune
// presentation (pas d'impression) : c'est le role de la couche CLI.
package pipeline

import (
	"fmt"

	"AlbumCut/internal/infra/ffmpeg"
	"AlbumCut/internal/infra/textfile"
	"AlbumCut/internal/naming"
	"AlbumCut/internal/segment"
	"AlbumCut/internal/track"
)

// ShortTrackWarningThreshold est la duree en dessous de laquelle une piste
// declenche un avertissement (mais reste creee normalement).
const ShortTrackWarningThreshold = 2.0 // secondes

// Plan decrit le decoupage calcule pour un album : la duree totale reelle du
// fichier source et les segments a produire, un par piste.
type Plan struct {
	TotalDuration float64
	Segments      []segment.Segment
}

// BuildPlan lit le fichier de timestamps, sonde la duree reelle du fichier
// MP3 source, et calcule les segments a produire.
func BuildPlan(mp3Path, timestampsPath string, endPadding float64) (Plan, error) {
	lines, err := textfile.ReadLines(timestampsPath)
	if err != nil {
		return Plan{}, err
	}

	tracks, err := track.ParseTracks(lines)
	if err != nil {
		return Plan{}, err
	}

	totalDuration, err := ffmpeg.ProbeDuration(mp3Path)
	if err != nil {
		return Plan{}, err
	}

	segments, err := segment.BuildSegments(tracks, totalDuration, endPadding)
	if err != nil {
		return Plan{}, err
	}

	return Plan{TotalDuration: totalDuration, Segments: segments}, nil
}

// TrackResult decrit le resultat de la creation du fichier d'une piste.
type TrackResult struct {
	OutputPath string
	Warning    string
}

// SplitTrack decoupe et ecrit le fichier de sortie pour le segment
// d'indice index (base 0) du plan, et signale un avertissement si la piste
// produite est tres courte.
func SplitTrack(mp3Path string, index int, seg segment.Segment) (TrackResult, error) {
	out := naming.OutputPath(mp3Path, index+1, seg.Track.Title)
	if err := ffmpeg.SplitSegment(mp3Path, seg.Start, seg.End, out); err != nil {
		return TrackResult{}, err
	}

	result := TrackResult{OutputPath: out}
	if d := seg.End - seg.Start; d < ShortTrackWarningThreshold {
		result.Warning = fmt.Sprintf("piste tres courte (%.1fs)", d)
	}
	return result, nil
}

// CheckPrerequisites verifie que les outils externes necessaires (ffmpeg,
// ffprobe) sont disponibles.
func CheckPrerequisites() error {
	return ffmpeg.CheckAvailable()
}
