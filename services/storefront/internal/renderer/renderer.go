package renderer

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

//go:embed templates/*
var templatesFS embed.FS

type TemplateCache map[string]*template.Template

type TemplateRenderer struct {
	templates TemplateCache
}

func NewTemplateCache() (TemplateCache, error) {
	cache := TemplateCache{}

	layouts, err := fs.Glob(templatesFS, "templates/layouts/*.html")
	if err != nil {
		return nil, err
	}

	pages, err := fs.Glob(templatesFS, "templates/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		files := append(layouts, page)
		name := filepath.Base(page)
		tmpl, err := template.ParseFS(templatesFS, files...)
		if err != nil {
			return nil, err
		}
		cache[name] = tmpl
	}

	return cache, nil
}

func NewTemplateRenderer() *TemplateRenderer {
	templates, err := NewTemplateCache()
	if err != nil {
		log.Fatal("failed to create TemplateRenderer: ", err)
	}
	return &TemplateRenderer{templates: templates}
}

func (r *TemplateRenderer) Render(w http.ResponseWriter, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return errors.New("template not found: " + name)
	}
	return tmpl.ExecuteTemplate(w, "base", data)
}
