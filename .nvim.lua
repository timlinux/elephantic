-- Elephantic project-specific Neovim configuration

local wk = require("which-key")

wk.register({
  p = {
    name = "Project",
    b = { "<cmd>!go build -o elephantic .<cr>", "Build" },
    r = { "<cmd>!./elephantic<cr>", "Run" },
    t = { "<cmd>!go test -v ./...<cr>", "Test" },
    f = { "<cmd>!go fmt ./...<cr>", "Format" },
    m = { "<cmd>!go mod tidy<cr>", "Mod Tidy" },
    v = { "<cmd>!go vet ./...<cr>", "Vet" },
    i = { "<cmd>!go install<cr>", "Install" },
  },
}, { prefix = "<leader>" })

-- Go-specific settings
vim.opt_local.tabstop = 4
vim.opt_local.shiftwidth = 4
vim.opt_local.expandtab = false

-- Auto format on save for Go files
vim.api.nvim_create_autocmd("BufWritePre", {
  pattern = "*.go",
  callback = function()
    vim.lsp.buf.format({ async = false })
  end,
})
