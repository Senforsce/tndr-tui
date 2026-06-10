// Package main demonstrates the StreamAbove API for character-by-character
// text streaming above an inline widget.
//
// To build and run:
//
//	go run ../../cmd/tui generate stream.t2
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/senforsce/tndr-tui"
)

//go:generate go run ../../cmd/tui generate stream.t2

func main() {
	app, err := tui.NewApp(
		tui.WithInlineHeight(3),
		tui.WithRootComponent(StreamDemo()),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
