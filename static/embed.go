// Package static exposes the frontend build output embedded in the executable.
package static

import (
	"embed"
	"io/fs"
)

// content contains the complete frontend dist directory at compile time.
//
//go:embed all:dist
var content embed.FS

// Files returns the dist directory as the root of a read-only filesystem.
func Files() fs.FS {
	files, err := fs.Sub(content, "dist")
	if err != nil {
		panic("open embedded frontend files: " + err.Error())
	}
	return files
}
