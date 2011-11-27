package ui

import (
	"fmt"
	"http"
	"os"
	"runtime/debug"

	"appengine"
)

// Set up the handlers

func init() {
	http.Handle("/", Wrapper(Root))
	http.Handle("/read", Wrapper(Read))
}

// Infrastructure for the handlers

type Wrapper func(appengine.Context, http.ResponseWriter, *http.Request) os.Error

func (f Wrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

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
		http.Error(w, err.String(), code)
	}
}

// Error codes

type ErrorCoder interface {
	ErrorCode() int
}

type Unauthorized string
func (e Unauthorized) String() string { return string(e) + ": unauthorized" }
func (e Unauthorized) ErrorCode() int { return http.StatusUnauthorized }
