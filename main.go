package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
)

const usage = "usage: albumcut [-end-padding=N] <album.mp3> <tracks.txt>"

type config struct {
	mp3Path        string
	timestampsPath string
	endPadding     float64
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

	segments, err := BuildSegments(tracks, totalDuration, cfg.endPadding)
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
	fs := flag.NewFlagSet("albumcut", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	endPadding := fs.Float64("end-padding", 0,
		"secondes ajoutees a la fin de chaque piste (sauf la derniere) pour eviter une coupe trop courte")

	if err := fs.Parse(args); err != nil {
		return config{}, fmt.Errorf("%s (%w)", usage, err)
	}

	rest := fs.Args()
	if len(rest) != 2 {
		return config{}, fmt.Errorf(usage)
	}
	if *endPadding < 0 {
		return config{}, fmt.Errorf("-end-padding doit etre positif ou nul (recu %v)", *endPadding)
	}

	return config{mp3Path: rest[0], timestampsPath: rest[1], endPadding: *endPadding}, nil
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
