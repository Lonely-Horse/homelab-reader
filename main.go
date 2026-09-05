package main

import (
	"embed"
	"homelab-reader/bootstrap"
	"homelab-reader/pkg/handlers"
	"html/template"
	"io/fs"
)

//go:embed templates
var tmplFS embed.FS

func main() {
	sub, _ := fs.Sub(tmplFS, "templates")
	tmpl, _ := template.ParseFS(tmplFS, "templates/*.html")
	s := &handlers.AppServer{Tmpl: tmpl, ViewFS: sub}

	bootstrap.Run(s, tmplFS)
}
