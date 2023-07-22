package main

import (
	"html/template"
	"path/filepath"
	"share-my-file/pkg/forms"
	"share-my-file/pkg/models"
)

type templateData struct {
	File        *models.File
	Form        *forms.Form
	CurrentYear int
}

func (app *application) newTemplateCache(dir string) (map[string]*template.Template, error) {
	app.logger.infoLog.Printf("newTemplateCache")
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*page.tmpl.html"))

	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		app.logger.infoLog.Println("name:", name)

		ts, err := template.New(name).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl.html"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl.html"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	app.logger.infoLog.Println("cache:", cache)
	app.logger.infoLog.Println("cache:")
	return cache, nil
}
