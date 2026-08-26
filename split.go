package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// SplitSegment decoupe un segment du fichier MP3 source vers outputPath,
// par copie de flux (sans reencodage), via ffmpeg.
func SplitSegment(mp3Path string, seg Segment, outputPath string) error {
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", mp3Path,
		"-ss", formatFFmpegTime(seg.Start),
		"-to", formatFFmpegTime(seg.End),
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
