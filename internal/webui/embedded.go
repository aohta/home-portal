//go:build embedded

package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var bundle embed.FS

func embeddedFiles() (fs.FS, error) { return fs.Sub(bundle, "dist") }
