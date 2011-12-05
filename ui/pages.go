package ui

import (
	"bytes"
	"fmt"
	"html"
	"http"
	"io/ioutil"
	"strings"
	"json"
	"os"

	"appengine"
	"fictex"
)

// Set up the handlers

func init() {
	http.Handle("/", Wrapper(Edit))
	http.Handle("/edit/", Wrapper(Edit))
	http.Handle("/read/", Wrapper(Read))
	http.Handle("/ajax", Wrapper(Ajax))
	http.Handle("/save", Wrapper(Save))
}

// Set up the pages

func Edit(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	w.Header().Set("Content-Type", "application/xhtml+xml; charset=UTF-8")

	var id string
	if strings.HasPrefix(r.URL.Path, "/edit/") {
		id = r.URL.Path[len("/edit/"):]
	} else if r.URL.Path == "/" {
		id = "autosave"
	} else {
		return NotFound(r.URL.Path)
	}

	type metadata struct{
		Id string
		Label string
		Value string
	}
	type maindata struct{
		// Story
		Id    string
		Title string
		Meta  []metadata

		// Story list
		Stories string

		// Preview
		Source      string
		PreviewHTML   string
		PreviewSource string
	}

	var data maindata

	_, k := UserKey(c)
	s := NewStory(c, id, k)

	if err := s.Get(c); err != nil {
		if id != "autosave" {
			return NotFound(r.URL.Path)
		}
	} else {
		data.Title = html.EscapeString(s.Title)
		data.Id = html.EscapeString(id)
	}

	for name, prop := range s.Meta {
		if len(name) == 0 || len(prop.Name) == 0 {
			c.Warningf("Zero-length property name?")
			continue
		}
		data.Meta = append(data.Meta, metadata{
			Id: html.EscapeString(name),
			Label: html.EscapeString(strings.ToUpper(prop.Name[:1])+prop.Name[1:]),
			Value: html.EscapeString(prop.Value),
		})
	}

	data.Source = html.EscapeString(string(s.Source))
	if node, err := fictex.ParseBytes(s.Source); err == nil {
		b := new(bytes.Buffer)
		if err := fictex.HTMLRenderer.Render(b, node); err == nil {
			data.PreviewHTML = b.String()
		}
	}
	data.PreviewSource = html.EscapeString(data.PreviewHTML)

	if js, err := JSONStoryList(c, k); err != nil {
		c.Warningf("Failed to load story list: %s", err)
	} else {
		data.Stories = string(js)
	}

	return templates.Execute(w, "edit.html", data)
}

func Read(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
	w.Header().Set("Content-Type", "application/xhtml+xml; charset=UTF-8")

	var id string
	if strings.HasPrefix(r.URL.Path, "/read/") {
		id = r.URL.Path[len("/read/"):]
	} else {
		return NotFound(r.URL.Path)
	}

	type metadata struct{
		Label string
		Value string
	}
	type renderdata struct{
		Title string
		Meta  []metadata
		HTML  string
	}

	var data renderdata

	s, err := GetStory(c, id)
	if err != nil {
		return err
	}

	data.Title = html.EscapeString(s.Title)

	for name, prop := range s.Meta {
		if len(name) == 0 || len(prop.Name) == 0 {
			c.Warningf("Zero-length property name?")
			continue
		}
		data.Meta = append(data.Meta, metadata{
			Label: html.EscapeString(strings.ToUpper(prop.Name[:1])+prop.Name[1:]),
			Value: html.EscapeString(prop.Value),
		})
	}

	if node, err := fictex.ParseBytes(s.Source); err == nil {
		b := new(bytes.Buffer)
		if err := fictex.HTMLRenderer.Render(b, node); err == nil {
			data.HTML = b.String()
		}
	} else {
		data.HTML = html.EscapeString(fmt.Sprintf("Error: %s", err))
	}

	return templates.Execute(w, "render.html", data)
}

func Ajax(c appengine.Context, w http.ResponseWriter, r *http.Request) os.Error {
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

func Save(c appengine.Context, w http.ResponseWriter, r *http.Request) (e os.Error) {
	out := map[string]string{}
	in := map[string]interface{}{}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &in); err != nil {
		return err
	}

	c.Infof("meta: %#v", in)

	id, _ := in["id"].(string)
	source, _ := in["source"].(string)

	meta, _ := in["meta"].(map[string]interface{})
	refreshStories := false
	switch id {
	case "", "autosave":
		id = "autosave"
		if title, _ := meta["title"].(string); title != "" {
			id = GenID(title)
			out["id"] = id
			refreshStories = true
		}
	default:
		out["id"] = id
	}

	c.Debugf("Meta: %#v", in["meta"])

	_, k := UserKey(c)
	s := NewStory(c, id, k)
	s.Source = []byte(source)

	if id != "autosave" {
		for prop, raw := range meta {
			switch prop {
			case "title":
				title, _ := raw.(string)
				if title == "" {
					break
				}
				// TODO(kevlar): Use memcache to figure out if a story's name changes
				/*
				if s.Title != "" && s.Title != title {
					refreshStories = true
				}
				*/
				s.Title = title
			default:
				val, _ := raw.(string)
				if val == "" {
					break
				}
				s.NewProperty(c, prop, val)
			}
		}
	}

	if err := s.Put(c); err != nil {
		return err
	}

	// Send a new list of stories
	if refreshStories {
		js, err := JSONStoryList(c, k)
		if err == nil {
			out["stories"] = string(js)
		}
	}

	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if _, err := w.Write(encoded); err != nil {
		return err
	}
	return nil
}
