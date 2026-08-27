// Package testutil fournit des fixtures partagees par les tests
// d'integration qui dependent d'un vrai binaire ffmpeg.
package testutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"
)

// RequireFFmpeg ignore le test si ffmpeg/ffprobe ne sont pas disponibles
// dans le PATH de la machine executant les tests.
func RequireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg introuvable dans le PATH, test d'integration ignore")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe introuvable dans le PATH, test d'integration ignore")
	}
}

// GenerateTestMP3 genere, via ffmpeg, un fichier MP3 de test (sinusoide) de
// la duree demandee.
func GenerateTestMP3(t *testing.T, dir, filename string, durationSeconds int) string {
	t.Helper()
	path := filepath.Join(dir, filename)

	cmd := exec.Command("ffmpeg",
		"-y",
		"-f", "lavfi",
		"-i", fmt.Sprintf("sine=frequency=440:duration=%d", durationSeconds),
		"-c:a", "libmp3lame",
		"-b:a", "64k",
		path,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("generation du fichier de test %s impossible : %v\n%s", path, err, stderr.String())
	}
	return path
}
