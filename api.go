package main

import (
	"database/sql"
	"github.com/coocood/jas"
	"math/rand"
	"net"
	"net/http"
	"path"
	"time"
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
	router.BasePath = path.Join("/", prefix)
	// Disable automatic internal error logging.
	router.InternalErrorLogger = nil

	l.Debug("API paths:\n", router.HandledPaths(true))

	// Seed the random number generator with the current Unix
	// time. This is not random, but it should be Good Enough.
	rand.Seed(time.Now().Unix())

	// Handle "<prefix>/api/". Note that it must begin and end with /.
	http.Handle(path.Join("/", prefix, "api")+"/", router)
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
// sensitive data (such as an email address) and sets ctx.Data to
// it. If `?geojson` is set, then it returns it in geojson.Feature
// form.
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

	// We must invoke ParseForm() so that we can access ctx.Form.
	ctx.ParseForm()

	// If the form value 'geojson' is included, dump in GeoJSON
	// form. Otherwise, just dump with normal marhshalling.
	if _, ok := ctx.Form["geojson"]; ok {
		ctx.Data = node.Feature()
		return
	} else {
		// Only after removing any sensitive data, though.
		node.OwnerEmail = ""

		// Finally, set the data and exit.
		ctx.Data = node
	}
}

// PostNode creates a *Node from the submitted form and queues it for
// addition with a positive 64 bit integer as an ID.
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
	status, _ := ctx.FindPositiveInt("status")
	node.Status = uint32(status)

	// TODO(DuoNoxSol): Authenticate/limit node registration.

	// If SMTP is missing from the config, we cannot continue.
	if Conf.SMTP == nil {
		ctx.Error = jas.NewInternalError(SMTPDisabledError)
		l.Err(SMTPDisabledError)
		return
	}

	// If SMTP verification is not explicitly disabled, send an email.
	if !Conf.SMTP.VerifyDisabled {
		id := rand.Int63() // Pseudo-random positive int64

		emailsent := true
		if err := SendVerificationEmail(id, node); err != nil {
			// If the sending of the email fails, set the internal
			// error and log it, then set a bool so that email can be
			// resent. If email continues failing to send, it will
			// eventually expire and be removed from the database.
			ctx.Error = jas.NewInternalError(err)
			l.Err(err)
			emailsent = false
			// Note that we do *not* return here.
		}

		// Once we have attempted to send the email, queue the node
		// for verification. If the email has not been sent, it will
		// be recorded in the database.
		if err := Db.QueueNode(id, emailsent,
			Conf.VerificationExpiration, node); err != nil {
			// If there is a database failure, report it as an
			// internal error.
			ctx.Error = jas.NewInternalError(err)
			l.Err(err)
			return
		}

		// If the email could be sent successfully, report
		// it. Otherwise, report that it is in the queue, and the
		// email will be resent.
		if emailsent {
			ctx.Data = "verification email sent"
			l.Infof("Node %q entered, waiting for verification", ip)
		} else {
			ctx.Data = "verification email will be resent"
			l.Infof("Node %q entered, verification email will be resent",
				ip)
		}
	} else {
		err := Db.AddNode(node)
		if err != nil {
			// If there was an error, log it and report the failure.
			ctx.Error = jas.NewInternalError(err)
			l.Err(err)
			return
		}
		ctx.Data = "node registered"
		l.Infof("Node %q registered\n", ip)
	}
}

// GetVerify moves a node from the verification queue to the normal
// database, as identified by its long random ID.
func (*Api) GetVerify(ctx *jas.Context) {
	id := ctx.RequireInt("id")
	ip, err := Db.VerifyQueuedNode(id)
	if err == sql.ErrNoRows {
		// If we encounter a ErrNoRows, then there was no node with
		// that ID. Report it.
		ctx.Error = jas.NewRequestError("invalid id")
		l.Noticef("%q attempted to verify invalid ID\n", ctx.RemoteAddr)
		return
	} else if err != nil {
		// If we encounter any other database error, it is an internal
		// error and needs to be logged.
		ctx.Error = jas.NewInternalError(err)
		l.Err(err)
		return
	}
	// If there was no error, inform the user that it was successful,
	// and log it.
	ctx.Data = "successful"
	l.Infof("Node %q verified", ip)
}

// GetAll dumps the entire database of nodes, including cached
// ones. If the form value 'geojson' is present, then the "data" field
// contains the dump in GeoJSON compliant form.
func (*Api) GetAll(ctx *jas.Context) {
	nodes, err := Db.DumpNodes()
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Err(err)
		return
	}

	// We must invoke ParseForm() so that we can access ctx.Form.
	ctx.ParseForm()

	// If the form value 'geojson' is included, dump in GeoJSON
	// form. Otherwise, just dump with normal marhshalling.
	if _, ok := ctx.Form["geojson"]; ok {
		ctx.Data = FeatureCollectionNodes(nodes)
	} else {
		mappedNodes, err := Db.CacheFormatNodes(nodes)
		if err != nil {
			ctx.Error = jas.NewInternalError(err)
			l.Err(err)
			return
		}
		ctx.Data = mappedNodes
	}
}
