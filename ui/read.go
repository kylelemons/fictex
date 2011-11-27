package ui

import (
	"fmt"
	"http"
	"os"

	"appengine"
	"fictex"
)

func Read(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	switch action := r.Form.Get("action"); action {
	case "render":
		node, err := fictex.ParseString(r.Form.Get("source"))
		if err != nil {
			return err
		}
		if err := fictex.HTMLRenderer.Render(w, node); err != nil {
			return err
		}
	default:
		fmt.Fprintln(w, "Unknown action", action)
	}

	return nil
}
