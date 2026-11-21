//go:build !embedded

package webui

import "io/fs"

func embeddedFiles() (fs.FS, error) { return nil, nil }
