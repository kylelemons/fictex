package ui

import (
	"http"
	"os"

	"appengine"
)

func Root(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	return templates.Execute(w, "main.html", nil)
}
