// Command termsnake is a polished terminal Snake game.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"

	"github.com/rjeff-sudo/termsnake-go/internal/app"
	"github.com/rjeff-sudo/termsnake-go/internal/storage"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "termsnake: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	noColor := flag.Bool("no-color", false, "disable colors and rely on glyphs only")
	flag.Parse()

	if os.Getenv("NO_COLOR") != "" {
		*noColor = true
	}

	// A broken or unavailable database must never stop the game.
	var store *storage.Store
	if path, err := storage.DefaultPath(); err != nil {
		fmt.Fprintf(os.Stderr, "termsnake: scores unavailable: %v\n", err)
	} else if st, err := storage.Open(path); err != nil {
		fmt.Fprintf(os.Stderr, "termsnake: scores unavailable: %v\n", err)
	} else {
		store = st
		defer store.Close()
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	return app.Run(screen, store, *noColor)
}
