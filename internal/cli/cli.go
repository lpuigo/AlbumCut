// Package cli est la couche de commande : parsing des arguments, validation
// des entrees, et presentation (progression, rapport final). Elle orchestre
// la couche pipeline mais n'implemente aucune logique metier elle-meme.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"AlbumCut/internal/duration"
	"AlbumCut/internal/pipeline"
)

const usage = "usage: albumcut [-end-padding=N] <album.mp3> <tracks.txt>"

const helpText = usage + `

Options:
  -end-padding=N   secondes ajoutees a la fin de chaque piste (sauf la
                   derniere) pour eviter une coupe trop courte (defaut 0)
  -h, --help       affiche cette aide et quitte`

type config struct {
	mp3Path        string
	timestampsPath string
	endPadding     float64
}

// Run execute le pipeline complet en ligne de commande : validation des
// arguments et des prerequis, calcul du plan de decoupage, decoupage piste
// par piste avec compte-rendu de progression, et rapport final. Toute la
// sortie utilisateur est ecrite sur out.
func Run(args []string, out io.Writer) error {
	cfg, err := parseArgs(args, out)
	if err != nil {
		return err
	}
	if err := checkFileExists(cfg.mp3Path); err != nil {
		return err
	}
	if err := checkFileExists(cfg.timestampsPath); err != nil {
		return err
	}
	if err := pipeline.CheckPrerequisites(); err != nil {
		return err
	}

	plan, err := pipeline.BuildPlan(cfg.mp3Path, cfg.timestampsPath, cfg.endPadding)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "Album MP3     : %s (duree %s)\n", cfg.mp3Path, duration.Format(int(plan.TotalDuration)))
	fmt.Fprintf(out, "Timestamps    : %s\n", cfg.timestampsPath)
	fmt.Fprintf(out, "Pistes a creer: %d\n\n", len(plan.Segments))

	var results []pipeline.TrackResult
	for i, s := range plan.Segments {
		fmt.Fprintf(out, "[%02d/%02d] %s (%s -> %s) ... ", i+1, len(plan.Segments), s.Track.Title, duration.Format(int(s.Start)), duration.Format(int(s.End)))
		result, err := pipeline.SplitTrack(cfg.mp3Path, i, s)
		if err != nil {
			fmt.Fprintln(out, "ECHEC")
			return err
		}
		fmt.Fprintln(out, "OK")
		results = append(results, result)
	}

	printReport(out, results)
	return nil
}

func printReport(out io.Writer, results []pipeline.TrackResult) {
	fmt.Fprintf(out, "\n%d piste(s) creee(s) avec succes :\n", len(results))
	for _, r := range results {
		line := "  - " + r.OutputPath
		if r.Warning != "" {
			line += " [attention: " + r.Warning + "]"
		}
		fmt.Fprintln(out, line)
	}
}

func parseArgs(args []string, out io.Writer) (config, error) {
	fs := flag.NewFlagSet("albumcut", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	endPadding := fs.Float64("end-padding", 0,
		"secondes ajoutees a la fin de chaque piste (sauf la derniere) pour eviter une coupe trop courte")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprintln(out, helpText)
			return config{}, flag.ErrHelp
		}
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
