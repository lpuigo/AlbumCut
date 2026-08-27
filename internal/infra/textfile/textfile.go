// Package textfile lit des fichiers texte depuis le disque.
package textfile

import (
	"bufio"
	"fmt"
	"os"
)

// ReadLines lit un fichier texte et retourne son contenu ligne par ligne,
// dans l'ordre, sans aucune transformation.
func ReadLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("impossible d'ouvrir le fichier %s : %w", path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erreur de lecture du fichier %s : %w", path, err)
	}

	return lines, nil
}
