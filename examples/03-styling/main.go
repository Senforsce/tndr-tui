// Package main demonstrates the Styling and Colors guide example.
//
// To build and run:
//
//	go run ../../cmd/tui generate styling.t2
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate styling.t2

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(StatusApp()),
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
