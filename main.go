package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"AlbumCut/internal/cli"
)

func main() {
	err := cli.Run(os.Args[1:], os.Stdout)
	if errors.Is(err, flag.ErrHelp) {
		return
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "erreur:", err)
		os.Exit(1)
	}
}
