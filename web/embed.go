package web

import "embed"

// StaticFS embeds the web dashboard files.
//
//go:embed index.html
var StaticFS embed.FS
