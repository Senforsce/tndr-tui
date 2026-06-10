// Package main demonstrates all built-in elements: input, button, ul/li, table, progress, p, hr.
//
// To build and run:
//
//	go run ../../cmd/tui generate elements.t2
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/senforsce/tndr-tui"
)

//go:generate go run ../../cmd/tui generate elements.t2

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Elements()),
		tui.WithMouse(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create app: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
		os.Exit(1)
	}
}
