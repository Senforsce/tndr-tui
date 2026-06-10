// Package main demonstrates the Multi-Component Applications guide example.
//
// To build and run:
//
//	go run ../../cmd/tui generate ./...
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/senforsce/tndr-tui"
)

//go:generate go run ../../cmd/tui generate app.t2 sidebar.t2 search.t2

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(MyApp()),
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
