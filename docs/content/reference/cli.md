# CLI Reference

## Installation

```bash
go install github.com/senforsce/tndr-tui/cmd/tui@latest
```

This installs the `tui` binary, which compiles `.t2` files to Go, formats them, validates syntax, and runs the language server for editor integration.

## Commands

### tui generate

```bash
tui generate [options] [path...]
```

Compiles `.t2` files into Go source files. Each `input.t2` produces a corresponding `input_t2.go` in the same directory. Hyphens in filenames are converted to underscores (e.g., `my-app.t2` becomes `my_app_t2.go`).

Never hand-edit the generated `_t2.go` files. They get overwritten on the next run.

**Options:**

| Flag | Description |
|------|-------------|
| `-v` | Verbose output (lists files found and processed) |

**Path formats:**

| Pattern | Behavior |
|---------|----------|
| `./...` | Recursively find all `.t2` files |
| `./components` | Process `.t2` files in that directory (non-recursive) |
| `header.t2` | Process a single file |
| *(none)* | Defaults to current directory (`.`) |

```bash
tui generate ./...              # all .t2 files recursively
tui generate ./components       # one directory
tui generate header.t2         # one file
tui generate -v ./...           # verbose
```

The command exits with code 1 if any file has errors. Error messages include the filename, line, and column.

### tui check

```bash
tui check [options] [path...]
```

Parses and analyzes `.t2` files without generating any output. Validates syntax, element names, attribute types, and imports. Same path formats as `generate`.

**Options:**

| Flag | Description |
|------|-------------|
| `-v` | Verbose output |

```bash
tui check ./...                 # check all files
tui check header.t2            # check one file
```

Exits with code 0 if all files pass. Exits with code 1 and prints errors to stderr if any file has problems.

### tui fmt

```bash
tui fmt [options] [path...]
```

Formats `.t2` files. By default, modifies files in place. Runs files in parallel for speed.

**Options:**

| Flag | Description |
|------|-------------|
| `--check` | Check if files are formatted without modifying them. Exits with code 1 if any file needs formatting. |
| `--stdout` | Print formatted output to stdout instead of writing back to disk. When processing multiple files, each is prefixed with `// filename`. |

```bash
tui fmt ./...                   # format all files in place
tui fmt --check ./...           # CI check: fail if any file isn't formatted
tui fmt --stdout file.t2       # preview formatted output
```

### tui lsp

```bash
tui lsp [options]
```

Starts the go-tui language server, communicating over stdin/stdout using the Language Server Protocol (JSON-RPC). Editors connect to this process for features like:

- Syntax error diagnostics
- Autocompletion for elements, attributes, and Tailwind classes
- Hover documentation
- Go-to-definition
- Find references
- Semantic token highlighting
- Document formatting

**Options:**

| Flag | Description |
|------|-------------|
| `--log FILE` | Write debug logs to the given file. Useful for troubleshooting LSP issues. |

```bash
tui lsp                         # start on stdio
tui lsp --log /tmp/tui-lsp.log  # start with debug logging
```

### tui version

```bash
tui version
```

Prints the version string, e.g., `tui version 0.1.0`.

### tui help

```bash
tui help
```

Prints the full usage message with all commands and examples. Also triggered by `-h` or `--help`.

## Editor Integration

The `tui lsp` command provides a Language Server Protocol server. Here's how to set it up in common editors.

### VS Code

Install the official go-tui extension, which bundles the LSP client, syntax highlighting, and file associations:

- [VS Code Marketplace](https://marketplace.visualstudio.com/items?itemName=grindlemire.go-tui)
- [Open VSX](https://open-vsx.org/extension/grindlemire/go-tui) (for VS Code forks like Cursor)

The extension automatically runs `tui lsp` for `.t2` files. No manual configuration needed.

### Neovim

With [nvim-lspconfig](https://github.com/neovim/nvim-lspconfig), add a custom server configuration:

```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.tui then
    configs.tui = {
        default_config = {
            cmd = { 'tui', 'lsp' },
            filetypes = { 't2' },
            root_dir = lspconfig.util.root_pattern('go.mod'),
        },
    }
end

lspconfig.tui.setup({})
```

You'll also want to associate `.t2` files with a filetype:

```lua
vim.filetype.add({
    extension = {
        t2 = 't2',
    },
})
```

### Debugging the LSP

If the language server isn't working as expected, start it with logging enabled:

```bash
tui lsp --log /tmp/tui-lsp.log
```

Then tail the log file while editing to see requests, responses, and errors:

```bash
tail -f /tmp/tui-lsp.log
```

## Cross-References

- [GSX Syntax Reference](t2-syntax.md) — the file format that `tui generate` compiles
- [Getting Started Guide](../guides/01-getting-started.md) — project setup walkthrough using the CLI
