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

var (
	ReadOnlyError = jas.NewRequestError("database in readonly mode")
)

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
	// Disable automatic internal error logging.
	router.InternalErrorLogger = nil

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
	ip := IP(net.ParseIP(ctx.RequireString("address")))
	if ip == nil {
		// If this is encountered, the address was incorrectly
		// formatted.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}
	node, err := Db.GetNode(ip)
	if err != nil {
		// If there has been a database error, log it and report the
		// failure.
		ctx.Error = jas.NewInternalError(err)
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

// PostNode creates a *Node from the submitted form and queues it for
// addition.
func (*Api) PostNode(ctx *jas.Context) {
	if Db.ReadOnly {
		// If the database is readonly, set that as the error and
		// return.
		ctx.Error = ReadOnlyError
		return
	}

	// Initialize the node and retrieve fields.
	node := new(Node)

	ip := IP(net.ParseIP(ctx.RequireString("address")))
	if ip == nil {
		// If the address is invalid, return that error.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}
	node.Addr = ip
	node.Latitude = ctx.RequireFloat("latitude")
	node.Longitude = ctx.RequireFloat("longitude")
	node.OwnerName = ctx.RequireString("name")
	node.OwnerEmail = ctx.RequireString("email")
	status, _ := ctx.FindInt("status")
	node.Status = int(status)

	// TODO(DuoNoxSol): Authenticate/limit node registration.

	err := Db.AddNode(node)
	if err != nil {
		// If there was an error, log it and report the failure.
		ctx.Error = jas.NewInternalError(err)
		l.Err(err)
		return
	}
	ctx.Data = "successful"
	l.Infof("Node %q registered\n", ip)
}

// GetAll dumps the entire database of nodes, including cached ones.
func (*Api) GetAll(ctx *jas.Context) {
	nodes, err := Db.DumpNodes()
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Err(err)
		return
	}
	ctx.Data = nodes
}
