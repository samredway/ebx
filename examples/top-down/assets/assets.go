package assets

import "embed"

//go:embed *.tmx *.tsx *.png
var GameFS embed.FS
