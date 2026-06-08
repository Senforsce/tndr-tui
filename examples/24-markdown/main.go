// Package main demonstrates the <markdown> element and the Markdown component.
//
// To build and run:
//
//	go run ../../cmd/tui generate markdown.t2
//	go run .
package main

import (
	"fmt"
	"os"

	tui "github.com/grindlemire/go-tui"
)

//go:generate go run ../../cmd/tui generate markdown.t2

// sampleDoc exercises every markdown construct the renderer supports, plus a few
// edge cases. It lives here as plain Go (a double-quoted string can hold the
// backticks that code fences and inline code need) and is referenced by the
// generated component constructor.
const sampleDoc = "# go-tui Markdown (ATX h1)\n" +
	"\n" +
	"## Inline styles (ATX h2)\n" +
	"\n" +
	"This paragraph mixes **bold**, *italic*, ***bold italic***, `inline code`, " +
	"and a [link to go.dev](https://go.dev) on terminals that support OSC 8. " +
	"Underscores work too: __bold__ and _italic_.\n" +
	"\n" +
	"### Edge cases (ATX h3)\n" +
	"\n" +
	"Delimiters without a closer stay literal: \"see **docs\" does not turn bold, " +
	"and math like 3 * 4 keeps its asterisk. This long sentence also shows that " +
	"ordinary paragraph text wraps to the configured width without manual breaks.\n" +
	"\n" +
	"Setext Level 1\n" +
	"==============\n" +
	"\n" +
	"Setext Level 2\n" +
	"--------------\n" +
	"\n" +
	"## Code blocks\n" +
	"\n" +
	"Fenced blocks are syntax-highlighted for Go, JSON, Bash, and JS/TS. Inline " +
	"`fmt.Println` stays styled too, and blank lines inside a fence are kept:\n" +
	"\n" +
	"```go\n" +
	"// Go: keywords, strings, types, and comments\n" +
	"func main() {\n" +
	"    fmt.Println(\"hello\")\n" +
	"\n" +
	"    fmt.Println(\"the blank line above is kept\")\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"```json\n" +
	"{\n" +
	"  \"name\": \"go-tui\",\n" +
	"  \"version\": 2,\n" +
	"  \"stable\": true,\n" +
	"  \"tags\": [\"tui\", \"markdown\"]\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"```bash\n" +
	"# Bash: comments, strings, $VARS, and keywords\n" +
	"for f in *.t2; do\n" +
	"  echo \"generating $f\"\n" +
	"done\n" +
	"```\n" +
	"\n" +
	"```ts\n" +
	"// TypeScript: keywords, template strings, types\n" +
	"async function load(url: string): Promise<Data> {\n" +
	"  const res = await fetch(`${url}/data.json`)\n" +
	"  return res.json()\n" +
	"}\n" +
	"```\n" +
	"\n" +
	"## Table\n" +
	"\n" +
	"| Language | Typed | Notes         |\n" +
	"| -------- | ----- | ------------- |\n" +
	"| Go       | yes   | `gofmt` clean |\n" +
	"| Python   | no    | *dynamic*     |\n" +
	"\n" +
	"## Lists\n" +
	"\n" +
	"- first item\n" +
	"- second item with a deliberately long line so it wraps onto more than one row at the fixed render width\n" +
	"  - nested item one\n" +
	"  - nested item two\n" +
	"- third item\n" +
	"\n" +
	"Ordered:\n" +
	"\n" +
	"1. one\n" +
	"2. two\n" +
	"3. three\n" +
	"\n" +
	"The `*` and `+` markers also start unordered lists:\n" +
	"\n" +
	"* star item\n" +
	"+ plus item\n" +
	"\n" +
	"## Blockquotes\n" +
	"\n" +
	"> A simple quote, long enough to show that quoted text wraps inside the bar column at a fixed width.\n" +
	">\n" +
	"> > A nested quote inside a quote.\n" +
	"\n" +
	"> A quote that contains a list:\n" +
	">\n" +
	"> - quoted item one\n" +
	"> - quoted item two\n"

func main() {
	app, err := tui.NewApp(
		tui.WithRootComponent(Viewer()),
		// Mouse reporting is left off so the terminal keeps native text
		// selection and OSC 8 link clicking. Full-screen mode then enables
		// alternate-scroll, so the mouse wheel still scrolls (the terminal
		// sends arrow keys, which the keymap handles).
		tui.WithoutMouse(),
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
