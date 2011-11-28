package ui

import (
	"http"
	"os"
	"template"

	"appengine"
)

var templates = template.SetMust(template.ParseTemplateGlob("templates/*.html"))

func Root(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	return templates.Execute(w, "main.html", nil)
}
