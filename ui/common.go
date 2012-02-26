package ui

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"text/template"

	"appengine"
)

// Templates

var reloadTemplates = !appengine.IsDevAppServer()
var templates = template.New("fictex")

func load(w http.ResponseWriter, ctx appengine.Context) {
	fmt.Printf("Loading...")
	t, err := template.New("fictex").ParseGlob("templates/*.html")
	if err != nil {
		if ctx == nil {
			panic(err)
		}
		ctx.Infof("error parsing templates: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("loaded %#v\n", templates)
	templates = t
}

func init() {
	load(nil, nil)
}

// Infrastructure for the handlers

type Wrapper func(appengine.Context, http.ResponseWriter, *http.Request) error

func (f Wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if reloadTemplates {
		load(w, ctx)
	}
	fmt.Printf("templates %#v\n", templates)

	defer func() {
		if r := recover(); r != nil {
			ctx.Criticalf("panic: (%T) %v\n%s", r, r, debug.Stack())
			http.Error(w, fmt.Sprint(r), http.StatusInternalServerError)
		}
	}()

	if err := f(ctx, w, r); err != nil {
		ctx.Infof("error: %s", err)
		code := http.StatusInternalServerError
		if coder, ok := err.(ErrorCoder); ok {
			code = coder.ErrorCode()
		}
		http.Error(w, err.Error(), code)
	}
}

// Error codes

type ErrorCoder interface {
	ErrorCode() int
}

type Unauthorized string

func (e Unauthorized) Error() string { return string(e) + ": unauthorized" }
func (e Unauthorized) ErrorCode() int { return http.StatusUnauthorized }

type NotFound string

func (e NotFound) Error() string { return string(e) + ": not found" }
func (e NotFound) ErrorCode() int { return http.StatusNotFound }
