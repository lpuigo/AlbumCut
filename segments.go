package main

import "fmt"

// Segment decrit la portion du fichier MP3 source correspondant a une piste.
type Segment struct {
	Track Track
	Start float64
	End   float64
}

// BuildSegments calcule, pour chaque piste, sa borne de depart et de fin au
// sein du fichier MP3 source. La derniere piste va jusqu'a totalDuration.
// Une erreur est retournee si le timestamp de la derniere piste est
// posterieur ou egal a la duree reelle du fichier (ce qui produirait un
// fichier de sortie vide ou de duree nulle).
func BuildSegments(tracks []Track, totalDuration float64) ([]Segment, error) {
	segments := make([]Segment, 0, len(tracks))

	for i, t := range tracks {
		start := float64(t.Start)

		var end float64
		if i+1 < len(tracks) {
			end = float64(tracks[i+1].Start)
		} else {
			end = totalDuration
			if end <= start {
				return nil, fmt.Errorf(
					"timestamp de la derniere piste %q (ligne %d, %s) invalide : "+
						"posterieur ou egal a la duree reelle du fichier MP3 (%s), "+
						"aucun fichier de duree nulle ne sera cree",
					t.Title, t.LineNumber, formatSeconds(t.Start), formatSeconds(int(totalDuration)),
				)
			}
		}

		segments = append(segments, Segment{Track: t, Start: start, End: end})
	}

	return segments, nil
}
