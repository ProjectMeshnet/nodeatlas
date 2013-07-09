package main

import (
	"errors"
	"html/template"
	"net"
	"net/http"
	"path"
	"strings"
)

var (
	listener net.Listener
)

var InvalidBindAddress = errors.New("invalid address to bind to")

// StartServer is a simple helper function to register any handlers
// (such as the API) and start the HTTP server on the configured
// address (Conf.Web.Addr).
//
// If Conf.Web.Prefix or Conf.Web.DeproxyHeaderFields has a length
// greater than zero, it wraps its http.ServeMux with a Deproxier.
//
// On crash, it returns the error.
func StartServer() (err error) {
	// Register any handlers.
	RegisterAPI(Conf.Web.Prefix)
	l.Debug("Registered API handler\n")

	err = RegisterTemplates()
	if err != nil {
		return
	}

	// Parse the address and create an appropriate net.Listener. The
	// Web.Addr will be of the form "protocol://address:port".
	parts := strings.Split(Conf.Web.Addr, "://")
	if len(parts) != 2 {
		return InvalidBindAddress
	}
	listener, err = net.Listen(parts[0], parts[1])
	if err != nil {
		return
	}

	// Create a custom http.Server, so that we can have better control
	// over certain behaviors.
	s := &http.Server{}

	// If either the Prefix or DeproxyHeaderFields are set, then we
	// need to wrap the default Handler with a Deproxier.
	if len(Conf.Web.Prefix) > 0 || len(Conf.Web.DeproxyHeaderFields) > 0 {
		s.Handler = &Deproxier{http.DefaultServeMux}
	}

	http.HandleFunc("/", HandleRoot)
	http.HandleFunc("/res/", HandleRes)
	http.HandleFunc("/favicon", HandleIcon)

	// Start the HTTP server and return any errors if it crashes.
	l.Infof("Starting HTTP server on %q\n", Conf.Web.Addr)
	return s.Serve(listener)
}

// Deproxier implements the http.Handler interface by setting the
// http.Request.RemoteAddr to the appropriate header field, if
// set, then passing the request on to its underlying http.ServeMux.
//
// It interprets the following header fields as a real remote address,
// in this order.
//     X-Proxied-For
//     X-Real-Ip
type Deproxier struct {
	Mux *http.ServeMux
}

func (d *Deproxier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check the acceptable header fields in order, checking if each
	// one is present. If so, set the r.RemoteAddr to its firste value
	// and break out of the loop.
	for _, fieldname := range Conf.Web.DeproxyHeaderFields {
		if realip, ok := r.Header[fieldname]; ok {
			r.RemoteAddr = net.JoinHostPort(realip[0], "0")
			break
		}
	}

	// Finally, pass the request on to the underlying http.ServeMux.
	d.Mux.ServeHTTP(w, r)
}

// HandleRoot serves the "index.html" template.
func HandleRoot(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join(StaticDir, "webpages/index.html"))
}

// HandleRes serves static files from the resources directory (*fRes).
func HandleRes(w http.ResponseWriter, req *http.Request) {
	// Serve the file within the resources directory, but slice off
	// len("/res") from the req.URL.Path first.
	http.ServeFile(w, req, path.Join(StaticDir, req.URL.Path[4:]))
}

// HandleIcon responds to requests for favicon.ico by serving icon.png
// from the resources directory.
func HandleIcon(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join(StaticDir, "icon",
		Conf.Map.Favicon))
}

// RegisterTemplates loads templates from <StaticDir>/emails/*.txt
// into the global variable t.
func RegisterTemplates() (err error) {
	t, err = template.ParseGlob(path.Join(StaticDir, "emails/*.txt"))
	return
}
