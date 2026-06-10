// Package main demonstrates the Timers, Watchers, and Channels guide example.
//
// To build and run:
//
//	go run ../../cmd/tui generate watchers.t2
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/senforsce/tndr-tui"
)

//go:generate go run ../../cmd/tui generate watchers.t2

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(WatcherApp()),
		tui.WithMouse(),
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
