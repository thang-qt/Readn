package assets

import "embed"

//go:embed *.html
//go:embed service-worker.js
//go:embed graphicarts
//go:embed javascripts
//go:embed stylesheets
var embedded embed.FS

func init() {
	FS.embedded = &embedded
}
