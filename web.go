package main

import (
	"html/template"
	"net/http"
	"path"
)

var (
	t *template.Template
)

// LogError uses the global variable l to log errors. If err is nil,
// it logs nothing.
func LogError(err error, req *http.Request) {
	if err != nil {
		l.Errf("View of %q by %q caused: %s", req.URL.Path, req.RemoteAddr, err)
	}
}

// RegisterTemplates loads templates from <*fRes>/templates/*.html into
// the global variable t.
func RegisterTemplates() (err error) {
	t, err = template.ParseGlob(path.Join(*fRes, "templates/*html"))
	return
}

// HandleRoot serves the "index.html" template.
func HandleRoot(w http.ResponseWriter, req *http.Request) {
	LogError(t.Execute(w, "index.html"), req)
}
