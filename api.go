package main

import (
	"github.com/coocood/jas"
	"net"
	"net/http"
	"path"
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

	l.Debug("API paths:\n", router.HandledPaths(true))

	// Handle "<prefix>/api/". Note that it must begin and end with /.
	http.Handle(path.Join(prefix, "api")+"/", router)
}

// GetStatus responds with a status summary of the map, including the
// map name, total number of nodes, number available (pingable), etc.
// (Not yet implemented.)
func (*Api) GetStatus(ctx *jas.Context) {
	ctx.Data = apiStatus{
		Name:  Conf.Name,
		Nodes: Db.LenNodes(false),
	}
}

// GetNode retrieves a single node from the database, removes
// sensitive data (such as an email address) and sets ctx.Data to it.
func (*Api) GetNode(ctx *jas.Context) {
	ip := net.ParseIP(ctx.RequireString("address"))
	if ip == nil {
		// If this is encountered, the address was incorrectly
		// formatted.
		// TODO: report an actual error
		return
	}
	node, err := Db.GetNode(ip)
	if err != nil {
		// If there has been a database error, we should not return
		// it. Instead, log it.
		// TODO: report an ambiguous "server error"
		l.Err(err)
		return
	}
	if node == nil {
		// If there are simply no matching nodes, set the error and
		// return.
		ctx.Error = jas.NewRequestError("No matching node")
		return
	}

	// Remove any sensitive data.
	node.OwnerEmail = ""

	// Finally, set the data and exit.
	ctx.Data = node
}
