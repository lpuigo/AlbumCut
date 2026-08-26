package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type config struct {
	mp3Path        string
	timestampsPath string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "erreur:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}
	if err := checkFileExists(cfg.mp3Path); err != nil {
		return err
	}
	if err := checkFileExists(cfg.timestampsPath); err != nil {
		return err
	}
	if err := checkExternalTools(); err != nil {
		return err
	}

	tracks, err := ParseTimestamps(cfg.timestampsPath)
	if err != nil {
		return err
	}

	totalDuration, err := ProbeDuration(cfg.mp3Path)
	if err != nil {
		return err
	}

	segments, err := BuildSegments(tracks, totalDuration)
	if err != nil {
		return err
	}

	fmt.Printf("Album MP3     : %s (duree %s)\n", cfg.mp3Path, formatSeconds(int(totalDuration)))
	fmt.Printf("Timestamps    : %s\n", cfg.timestampsPath)
	fmt.Printf("Pistes a creer: %d\n\n", len(segments))

	var results []trackResult
	for i, s := range segments {
		out := OutputPath(cfg.mp3Path, i+1, s.Track.Title)
		fmt.Printf("[%02d/%02d] %s (%s -> %s) ... ", i+1, len(segments), s.Track.Title, formatSeconds(int(s.Start)), formatSeconds(int(s.End)))
		if err := SplitSegment(cfg.mp3Path, s, out); err != nil {
			fmt.Println("ECHEC")
			return err
		}
		fmt.Println("OK")

		result := trackResult{OutputPath: out}
		if duration := s.End - s.Start; duration < shortTrackWarningThreshold {
			result.Warning = fmt.Sprintf("piste tres courte (%.1fs)", duration)
		}
		results = append(results, result)
	}

	printReport(results)
	return nil
}

// shortTrackWarningThreshold est la duree en dessous de laquelle une piste
// declenche un avertissement (mais reste creee normalement).
const shortTrackWarningThreshold = 2.0 // secondes

type trackResult struct {
	OutputPath string
	Warning    string
}

func printReport(results []trackResult) {
	fmt.Printf("\n%d piste(s) creee(s) avec succes :\n", len(results))
	for _, r := range results {
		line := "  - " + r.OutputPath
		if r.Warning != "" {
			line += " [attention: " + r.Warning + "]"
		}
		fmt.Println(line)
	}
}

func parseArgs(args []string) (config, error) {
	if len(args) != 2 {
		return config{}, fmt.Errorf("usage: albumcut <album.mp3> <tracks.txt>")
	}
	return config{mp3Path: args[0], timestampsPath: args[1]}, nil
}

func checkFileExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("fichier introuvable : %s", path)
		}
		return fmt.Errorf("impossible d'acceder au fichier %s : %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s est un repertoire, un fichier est attendu", path)
	}
	return nil
}

func checkExternalTools() error {
	for _, tool := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("%s introuvable dans le PATH : installez ffmpeg (https://ffmpeg.org/download.html) et verifiez qu'il est accessible depuis la ligne de commande", tool)
		}
	}
	return nil
}
