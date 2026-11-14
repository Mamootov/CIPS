package web

import "embed"

// Assets embeds all static frontend files (HTML, CSS, JS).
//
//go:embed *
var Assets embed.FS
