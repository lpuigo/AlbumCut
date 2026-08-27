// Package ffmpeg encapsule tous les appels aux binaires externes ffmpeg et
// ffprobe (sondage de duree, decoupage par copie de flux).
package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CheckAvailable verifie que ffmpeg et ffprobe sont accessibles depuis le
// PATH de la machine.
func CheckAvailable() error {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s introuvable dans le PATH : installez ffmpeg (https://ffmpeg.org/download.html) et verifiez qu'il est accessible depuis la ligne de commande", tool)
		}
	}
	return nil
}

// ProbeDuration retourne la duree totale (en secondes) du fichier audio
// donne, via ffprobe.
func ProbeDuration(mp3Path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		mp3Path,
	)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return 0, fmt.Errorf("ffprobe a echoue sur %s : %s", mp3Path, msg)
	}

	durationStr := strings.TrimSpace(stdout.String())
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("duree renvoyee par ffprobe illisible (%q) pour %s", durationStr, mp3Path)
	}

	return duration, nil
}

// SplitSegment decoupe la portion [start, end] (en secondes) du fichier MP3
// source vers outputPath, par copie de flux (sans reencodage), via ffmpeg.
func SplitSegment(mp3Path string, start, end float64, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", mp3Path,
		"-ss", formatFFmpegTime(start),
		"-to", formatFFmpegTime(end),
		"-c", "copy",
		outputPath,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("echec du decoupage vers %s : %s", outputPath, lastFFmpegErrorLines(stderr.String()))
	}
	return nil
}

func formatFFmpegTime(seconds float64) string {
	return strconv.FormatFloat(seconds, 'f', 3, 64)
}

// lastFFmpegErrorLines extrait les dernieres lignes de la sortie d'erreur
// de ffmpeg (souvent verbeuse) pour ne garder que le message utile.
func lastFFmpegErrorLines(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return "aucun detail fourni par ffmpeg"
	}
	lines := strings.Split(stderr, "\n")
	const maxLines = 5
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, " | ")
}
