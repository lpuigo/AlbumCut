// Package track contient les regles metier de parsing et de validation du
// fichier de timestamps : une seule piste par ligne, timestamps strictement
// croissants.
package track

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"AlbumCut/internal/duration"
)

type Track struct {
	Title      string
	Start      int // secondes depuis le debut de l'album
	LineNumber int
}

type ParseError struct {
	Line    int
	Content string
	Reason  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("fichier de timestamps, ligne %d (%q) : %s", e.Line, e.Content, e.Reason)
}

var timestampLineRe = regexp.MustCompile(`^\s*(\d{1,2}(?::\d{2}){1,2})\s*-\s*(.*)$`)

// ParseTracks parse les lignes d'un fichier de timestamps (dans l'ordre du
// fichier, y compris les lignes vides, pour que les numeros de ligne restent
// corrects) et retourne la liste des pistes qu'elles decrivent. Il doit y
// avoir au moins une ligne valide, et les timestamps doivent etre
// strictement croissants.
func ParseTracks(lines []string) ([]Track, error) {
	var tracks []Track

	for i, rawLine := range lines {
		lineNumber := i + 1
		rawLine = strings.TrimRight(rawLine, "\r")
		if strings.TrimSpace(rawLine) == "" {
			continue
		}

		track, err := parseTimestampLine(rawLine, lineNumber)
		if err != nil {
			return nil, err
		}

		if len(tracks) > 0 {
			prev := tracks[len(tracks)-1]
			if track.Start <= prev.Start {
				reason := fmt.Sprintf(
					"timestamp non strictement croissant (%s), doit etre superieur au precedent (%s, ligne %d)",
					duration.Format(track.Start), duration.Format(prev.Start), prev.LineNumber,
				)
				return nil, &ParseError{Line: lineNumber, Content: rawLine, Reason: reason}
			}
		}

		tracks = append(tracks, track)
	}

	if len(tracks) == 0 {
		return nil, fmt.Errorf("fichier de timestamps vide ou sans ligne valide")
	}

	return tracks, nil
}

func parseTimestampLine(line string, lineNumber int) (Track, error) {
	m := timestampLineRe.FindStringSubmatch(line)
	if m == nil {
		return Track{}, &ParseError{
			Line:    lineNumber,
			Content: line,
			Reason:  "format invalide, attendu 'HH:MM:SS - Titre' ou 'MM:SS - Titre' (titre optionnel)",
		}
	}

	seconds, err := parseTimecode(m[1])
	if err != nil {
		return Track{}, &ParseError{Line: lineNumber, Content: line, Reason: err.Error()}
	}

	title := strings.TrimSpace(m[2])
	if title == "" {
		title = "Untitled"
	} else {
		title = m[2]
	}

	return Track{Title: title, Start: seconds, LineNumber: lineNumber}, nil
}

func parseTimecode(timecode string) (int, error) {
	parts := strings.Split(timecode, ":")

	var h, min, sec int
	var err error
	switch len(parts) {
	case 3:
		if h, err = strconv.Atoi(parts[0]); err != nil {
			return 0, fmt.Errorf("heures invalides dans %q", timecode)
		}
		if min, err = strconv.Atoi(parts[1]); err != nil {
			return 0, fmt.Errorf("minutes invalides dans %q", timecode)
		}
		if sec, err = strconv.Atoi(parts[2]); err != nil {
			return 0, fmt.Errorf("secondes invalides dans %q", timecode)
		}
	case 2:
		if min, err = strconv.Atoi(parts[0]); err != nil {
			return 0, fmt.Errorf("minutes invalides dans %q", timecode)
		}
		if sec, err = strconv.Atoi(parts[1]); err != nil {
			return 0, fmt.Errorf("secondes invalides dans %q", timecode)
		}
	default:
		return 0, fmt.Errorf("format de timestamp invalide %q", timecode)
	}

	if min > 59 || sec > 59 {
		return 0, fmt.Errorf("minutes/secondes hors limites (0-59) dans %q", timecode)
	}

	return h*3600 + min*60 + sec, nil
}
