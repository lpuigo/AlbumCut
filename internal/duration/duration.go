// Package duration formate des durees en secondes pour l'affichage.
package duration

import "fmt"

// Format formate un nombre de secondes en HH:MM:SS.
func Format(totalSeconds int) string {
	h := totalSeconds / 3600
	m := (totalSeconds % 3600) / 60
	s := totalSeconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}
