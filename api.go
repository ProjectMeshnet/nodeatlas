package main

import (
	"github.com/coocood/jas"
	"net/http"
)

// Api is the JAS-required type which is passed to all API-related
// functions.
type Api struct{}

// apiStatus is a simple summary of the current NodeAtlas instance
// status.
type apiStatus struct {
	Name      string `json:"name"`
	Nodes     int    `json:"nodes"`
	Available int    `json:"available"`
}

// RegisterAPI invokes http.Handle() with a JAS router using the
// default net/http server. It will respond to any URL "<prefix>/api".
func RegisterAPI(prefix string) {
	// Initialize a JAS router with appropriate attributes.
	router := jas.NewRouter(new(Api))
	router.BasePath = prefix

	// Handle "<prefix>/api".
	http.Handle(prefix, router)
}

// GetStatus responds with a status summary of the map, including the
// map name, total number of nodes, number available (pingable), etc.
// (Not yet implemented.)
func (*Api) GetStatus(ctx *jas.Context) {
	ctx.Data = apiStatus{
		Name: Conf.Name,
	}
}
