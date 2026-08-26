package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

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
