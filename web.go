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
	data := make(map[string]string, 1)
	data["Name"] = Conf.Name
	data["Version"] = Version
	LogError(t.ExecuteTemplate(w, "index.html", data), req)
}

// HandleRes serves static files from the resources directory (*fRes).
func HandleRes(w http.ResponseWriter, req *http.Request) {
	// Serve the file within the resources directory, but slice off
	// len("/res") from the req.URL.Path first.
	http.ServeFile(w, req, path.Join(*fRes, req.URL.Path[4:]))
}

// HandleIcon responds to requests for favicon.ico by serving icon.png
// from the resources directory.
func HandleIcon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join(*fRes, "icon.png"))
}
