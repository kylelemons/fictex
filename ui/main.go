package ui

import (
	"http"
	"os"
	"template"

	"appengine"
)

func Root(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	set, err := template.ParseTemplateGlob("templates/*.html")
	if err != nil {
		return err
	}

	return set.Execute(w, "main.html", nil)
}
