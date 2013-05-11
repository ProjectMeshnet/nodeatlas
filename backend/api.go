package main

import (
	"github.com/coocood/jas"
	"net/http"
)

// Atlas is the JAS-required type which is passed to all API-related
// functions.
type Atlas struct{}

// apiStatus is a simple summary of the current NodeAtlas instance
// status.
type apiStatus struct {
	Name      string `json:"name"`
	Nodes     int    `json:"nodes"`
	Available int    `json:"available"`
}

// StartAPI uses net/http to start an HTTP server listening on the
// given address (such as "0.0.0.0:8080" or just ":8080") and only
// responding on the given prefix path. It blocks until the server
// crashes, and returns the error if that occurs.
func StartAPI(addr string, prefix string) (err error) {
	// Initialize a JAS router with appropriate attributes.
	router := jas.NewRouter(new(Atlas))
	router.BasePath = prefix

	l.Info("Initialized API\n")

	// Begin the HTTP server listening appropriately and sending
	// requests containing the prefix path to the router.
	http.Handle(prefix, router)
	return http.ListenAndServe(addr, nil)
}

// GetStatus responds with a status summary of the map, including the
// map name, total number of nodes, number available (pingable), etc.
// (Not yet implemented.)
func (*Atlas) GetStatus(ctx *jas.Context) {
	ctx.Data = apiStatus{}
}
