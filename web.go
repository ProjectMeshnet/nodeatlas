package main
// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"encoding/xml"
	"errors"
	"github.com/baliw/moverss"
	"github.com/dchest/captcha"
	"github.com/russross/blackfriday"
	"html/template"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

var (
	NodeRSS     *moverss.Channel
	NodeXMLName xml.Name
)

var (
	listener net.Listener

	captchaServer = captcha.Server(captcha.StdWidth, captcha.StdHeight)
)

var (
	InvalidBindAddress   = errors.New("invalid address to bind to")
	InvalidCAPTCHAFormat = errors.New("CAPTCHA format invalid")
	IncorrectCAPTCHA     = errors.New("CAPTCHA ID or solution is incorrect")
)

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
	// need to wrap the default Handler with a Deproxier. Otherwise,
	// we just use our Handler.
	if len(Conf.Web.Prefix) > 0 || len(Conf.Web.DeproxyHeaderFields) > 0 {
		s.Handler = &Deproxier{http.DefaultServeMux}
	} else {
		s.Handler = &Handler{http.DefaultServeMux}
	}

	// We need to set the database tile store.
	captcha.SetCustomStore(CAPTCHAStore{})

	http.HandleFunc("/", HandleStatic)
	http.HandleFunc("/node/", HandleMap)
	http.HandleFunc("/verify/", HandleMap)
	http.Handle("/captcha/", captchaServer)

	// Start the HTTP server and return any errors if it crashes.
	l.Infof("Starting HTTP server on %q\n", Conf.Web.Addr)
	return s.Serve(listener)
}

// HandleStatic serves files directly from <StaticDir>/web using
// http.ServeFile().
func HandleStatic(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join(StaticDir, "web", req.URL.Path))
}

// HandleMap always serves <StaticDir>/web/index.html.
func HandleMap(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, path.Join(StaticDir, "web", "index.html"))
}

// Handler acts is a simple http.Handler which performs some cleanup
// on the Request before passing it on to its underlying
// http.ServeMux.
type Handler struct {
	Mux *http.ServeMux
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.RemoteAddr, _, _ = net.SplitHostPort(r.RemoteAddr)
	h.Mux.ServeHTTP(w, r)
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
	// one is present. If so, set the r.RemoteAddr to its first value
	// and break out of the loop.
	for _, fieldname := range Conf.Web.DeproxyHeaderFields {
		if realip, ok := r.Header[fieldname]; ok {
			r.RemoteAddr = realip[0]
			break
		}
	}

	// Finally, pass the request on to the underlying http.ServeMux.
	d.Mux.ServeHTTP(w, r)
}

// RegisterTemplates loads templates from <StaticDir>/email/*.txt into
// the global variable t.
func RegisterTemplates() (err error) {
	t = template.New("")

	t.Funcs(template.FuncMap{
		"markdownify": func(s string) template.HTML {
			return template.HTML(
				string(blackfriday.MarkdownBasic([]byte(s))))
		},
	})

	t, err = t.ParseGlob(path.Join(StaticDir, "email/*.txt"))
	return
}

// VerifyCAPTCHA accepts a *http.Request and verifies that the given
// 'captcha' form is valid. This is a string of the form
// "id:solution". It will return IncorrectCAPTCHAError if the solution
// or ID is invalid.
func VerifyCAPTCHA(req *http.Request) error {
	// Get the "captcha" form value.
	solution := req.FormValue("captcha")

	// Find the point to split the form value at. If it's not found in
	// the string, return the InvalidCAPTCHAFormat error.
	index := strings.Index(solution, ":")
	if index < 0 {
		return InvalidCAPTCHAFormat
	}

	// If that was successful, try to verify it. If it returns false,
	// the ID or solution was invalid.
	if !captcha.VerifyString(solution[:index], solution[index+1:]) {
		return IncorrectCAPTCHA
	}

	// If we get to this point, then it was successfully validated and
	// we can return nil.
	return nil
}

// GenerateNodeRSS creates an RSS feed of nodes from the database and
// returns any errors.
func GenerateNodeRSS() (err error) {
	NodeRSS = moverss.ChannelFactory(
		Conf.Name+" NodeAtlas",
		Conf.Web.Hostname,
		"New local node feed")

	NodeXMLName = xml.Name{
		Space: Conf.Web.Hostname,
		Local: "nodes",
	}

	// We need to use some custom SQL here in order to retrieve the
	// timestamps.
	rows, err := Db.Query(`
SELECT updated,address,owner
FROM nodes;`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize the variables.
		var updated int64
		node := new(Node)

		// Scan all of the values into them.
		err = rows.Scan(&updated, &node.Addr, &node.OwnerName)
		if err != nil {
			return
		}

		// Add the Node to the RSS feed.
		in := node.Item()
		in.SetPubDate(time.Unix(updated, 0))
		NodeRSS.AddItem(in)
	}

	// Write the feed to the file, and log any errors.
	WriteNodeRSS()

	return
}

// AddNodeToRSS adds a Node to the existing NodeRSS channel with the
// given time and invokes WriteNodeRSS to write it to a file and log
// any errors.
func AddNodeToRSS(n *Node, t time.Time) {
	in := n.Item()
	in.SetPubDate(t)
	NodeRSS.AddItem(in)
	WriteNodeRSS()
}

func WriteNodeRSS() {
	f, err := os.Create(StaticDir + "/web/index.rss")
	if err != nil {
		l.Errf("Error writing NodeRSS feed: %s", err)
	}

	f.Write(NodeRSS.Publish())
	f.Close()
}
