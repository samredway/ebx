package assets

import "embed"

//go:embed *.tmx *.png
var GameFS embed.FS
