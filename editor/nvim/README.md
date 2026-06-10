# GSX Neovim Plugin

Neovim support for `.t2` files: tree-sitter syntax highlighting, Go language injection, and LSP integration.

## Requirements

- Neovim 0.9+ (0.11+ recommended for native LSP config)
- [nvim-treesitter](https://github.com/nvim-treesitter/nvim-treesitter) for syntax highlighting
- The `tui` CLI installed and on your `$PATH` (for the LSP)
- A C compiler for building the tree-sitter parser (`cc` or `gcc`)

## Install

### lazy.nvim

Since the plugin lives in a subdirectory of the repo, use the `init` callback to prepend the correct path:

```lua
{
  "grindlemire/go-tui",
  init = function()
    -- Add the nvim plugin subdirectory to the runtimepath
    vim.opt.rtp:prepend(vim.fn.stdpath("data") .. "/lazy/go-tui/editor/nvim")
  end,
  config = function()
    require("t2").setup()
  end,
  ft = "t2",
}
```

### packer.nvim

```lua
use {
  "grindlemire/go-tui",
  rtp = "editor/nvim",
  config = function()
    require("t2").setup()
  end,
  ft = { "t2" },
}
```

### vim-plug

```vim
Plug 'grindlemire/go-tui', { 'rtp': 'editor/nvim' }
```

Then in your `init.lua`:

```lua
require("t2").setup()
```

### Manual

Clone the repo and add the plugin path to your runtimepath:

```lua
vim.opt.rtp:prepend("/path/to/go-tui/editor/nvim")
require("t2").setup()
```

## Setup

After installing, call `setup()` in your Neovim config:

```lua
require("t2").setup()
```

This does three things:

1. Registers `.t2` as a filetype
2. Registers the tree-sitter parser with nvim-treesitter (so `:TSInstall t2` works)
3. Configures and starts the `tui lsp` language server for `.t2` files

## Options

```lua
require("t2").setup({
  lsp = {
    enabled = true,                       -- set to false to disable the LSP
    cmd = { "tui", "lsp" },              -- command to start the LSP server
    log = "/tmp/t2-lsp.log",            -- optional: enable LSP debug logging
  },
})
```

## Installing the tree-sitter parser

After setup, install the parser:

```vim
:TSInstall t2
```

If that doesn't work (the grammar isn't upstream in nvim-treesitter yet), the plugin's parser registration should still let nvim-treesitter build it from the GitHub repo. You can verify with:

```vim
:TSInstallInfo
```

Look for `t2` in the list.

## LSP features

The `tui lsp` language server provides:

- Real-time diagnostics
- Hover documentation
- Auto-completion (elements, attributes, Tailwind classes, Go expressions via gopls)
- Go-to-definition
- Find references
- Document and workspace symbols
- Semantic token highlighting
- Code formatting

Make sure the `tui` binary is installed and on your `$PATH`:

```bash
go install github.com/senforsce/tndr-tui/cmd/tui@latest
```

## Troubleshooting

**No syntax highlighting**: Run `:TSInstall t2` and restart Neovim. Check `:TSInstallInfo` to confirm the parser is installed.

**LSP not starting**: Check that `tui lsp` runs from your shell. Enable logging with `lsp = { log = "/tmp/t2-lsp.log" }` and inspect the log. Run `:LspInfo` (or `:lua vim.print(vim.lsp.get_clients())` on 0.11+) to see if the client attached.

**Wrong filetype**: Run `:set ft?` in a `.t2` buffer. It should say `filetype=t2`. If not, make sure the plugin loaded (`:scriptnames` should show `t2`).
