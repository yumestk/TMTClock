// Package frontend embeds the built Vite output (frontend/dist) so the
// whole app ships as one Go binary.
package frontend

import "embed"

//go:embed all:dist
var Dist embed.FS
