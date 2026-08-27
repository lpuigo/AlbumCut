// Package naming construit le chemin des fichiers de sortie a partir du
// fichier MP3 source et du titre de piste.
package naming

import (
	"fmt"
	"path/filepath"
	"strings"
)

// forbiddenFilenameChars sont les caracteres interdits dans un nom de
// fichier sous Windows (le systeme cible). Les espaces ne sont pas
// touches : ils doivent etre conserves tels quels dans le titre de piste.
var forbiddenFilenameChars = []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}

// AlbumTitle retourne le nom du fichier MP3 source sans son extension.
func AlbumTitle(mp3Path string) string {
	base := filepath.Base(mp3Path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// SanitizeFilenamePart remplace les caracteres interdits dans un nom de
// fichier par "_", sans toucher aux espaces.
func SanitizeFilenamePart(s string) string {
	for _, c := range forbiddenFilenameChars {
		s = strings.ReplaceAll(s, c, "_")
	}
	return s
}

// OutputPath construit le chemin complet du fichier de sortie pour une
// piste donnee, dans le meme repertoire que le fichier MP3 source :
// NN. AlbumTitle_TrackTitle.mp3
func OutputPath(mp3Path string, trackNumber int, trackTitle string) string {
	filename := fmt.Sprintf("%02d. %s_%s.mp3",
		trackNumber,
		SanitizeFilenamePart(AlbumTitle(mp3Path)),
		SanitizeFilenamePart(trackTitle),
	)
	return filepath.Join(filepath.Dir(mp3Path), filename)
}
