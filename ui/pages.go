package ui

import (
	"bytes"
	"fmt"
	"html"
	"http"
	"os"

	"appengine"
	"fictex"
)

// Set up the handlers

func init() {
	http.Handle("/", Wrapper(Root))
	http.Handle("/read", Wrapper(Read))
	http.Handle("/save", Wrapper(Save))
}

func Root(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	w.Header().Set("Content-Type", "application/xhtml+xml; charset=UTF-8")

	_, k := UserKey(c)
	s := NewStory(c, "autosave", k)
	s.Get()

	data := struct{
		Saved       string
		SavedHTML   string
		SavedSource string
	}{}

	if len(s.Source) > 0 {
		data.Saved = html.EscapeString(string(s.Source))

		if node, err := fictex.ParseBytes(s.Source); err == nil {
			b := new(bytes.Buffer)
			if err := fictex.HTMLRenderer.Render(b, node); err == nil {
				data.SavedHTML = b.String()
			}
		}
		data.SavedSource = html.EscapeString(data.SavedHTML)
	}

	return templates.Execute(w, "main.html", data)
}

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

func Save(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	if err := r.ParseForm(); err != nil {
		return err
	}

	id := r.Form.Get("id")
	if id == "" {
		id = "autosave"
	}

	_, k := UserKey(c)
	s := NewStory(c, id, k)
	s.Title = "Title"
	s.Source = []byte(r.Form.Get("source"))

	return s.Put()
}
