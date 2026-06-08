-- Minimal auto-setup: register the t2 filetype on load.
-- Users still need to call require("t2").setup() from their config
-- to enable tree-sitter parser registration and the LSP.
vim.filetype.add({
	extension = {
		t2 = "t2",
	},
})
