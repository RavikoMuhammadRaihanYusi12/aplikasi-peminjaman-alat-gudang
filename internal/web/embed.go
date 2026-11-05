package web

import "embed"

//go:embed templates/layout.html templates/pages/*.html templates/partials/*.html static/app.css static/htmx.min.js
var EmbedFS embed.FS
