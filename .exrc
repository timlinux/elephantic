" Elephantic project shortcuts
" All shortcuts under <leader>p

" Build
nnoremap <leader>pb :!go build -o elephantic .<CR>

" Run (test mode - just prints version)
nnoremap <leader>pr :!./elephantic<CR>

" Test
nnoremap <leader>pt :!go test -v ./...<CR>

" Format
nnoremap <leader>pf :!go fmt ./...<CR>

" Tidy modules
nnoremap <leader>pm :!go mod tidy<CR>

" Vet
nnoremap <leader>pv :!go vet ./...<CR>

" Install
nnoremap <leader>pi :!go install<CR>
