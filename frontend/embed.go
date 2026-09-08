// Package frontend embeds the built Wails GUI assets so cmd/league-rpc-gui can
package frontend

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Assets returns the built frontend rooted at index.html.
func Assets() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic("frontend: dist not embedded; run the frontend build: " + err.Error())
	}
	return sub
}
