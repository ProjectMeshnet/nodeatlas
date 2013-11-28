package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"database/sql"
	"github.com/coocood/jas"
	"github.com/dchest/captcha"
	"html"
	"html/template"
	"math/rand"
	"net"
	"net/http"
	"path"
	"regexp"
	"time"
)

const (
	APIDocs = "https://github.com/ProjectMeshnet/nodeatlas/blob/master/API.md"
)

var (
	PGPRegexp   = regexp.MustCompilePOSIX("^[0-9a-z]{8}{0,2}$")
	EmailRegexp = regexp.MustCompilePOSIX("^[a-z0-9._%+-]+@([a-z0-9-]+\\.)+[a-z]+$")
)

// Api is the JAS-required type which is passed to all API-related
// functions.
type Api struct{}

var (
	ActiveTokens = make(map[uint32]token)
)

type token struct {
	IP     string
	Issued time.Time
}

var (
	ReadOnlyError = jas.NewRequestError("database in readonly mode")
)

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

// Get responds on the root API handler ("/api/") with 303 SeeOther
// and a link to the API documentation on the project homepage.
func (*Api) Get(ctx *jas.Context) {
	ctx.Status = http.StatusSeeOther
	ctx.ResponseHeader.Set("Location", APIDocs)
	ctx.Data = http.StatusText(http.StatusSeeOther) + ": " + APIDocs
}

// GetEcho responds with the remote address of the user
func (*Api) GetEcho(ctx *jas.Context) {
	if Conf.Verify.Netmask != nil {
		i := net.ParseIP(ctx.RemoteAddr)
		if (*net.IPNet)(Conf.Verify.Netmask).Contains(i) {
			ctx.Data = ctx.RemoteAddr
		} else {
			ctx.Error = jas.NewRequestError("remote address not in subnet")
		}
	} else {
		ctx.Error = jas.NewRequestError("netmask not set")
	}
}

// GetStatus responds with a status summary of the map, including the
// map name, total number of nodes, number available (pingable), etc.
// (Not yet implemented.)
func (*Api) GetStatus(ctx *jas.Context) {
	localNodes := Db.LenNodes(false)
	ctx.Data = map[string]interface{}{
		"Name":        Conf.Name,
		"LocalNodes":  localNodes,
		"CachedNodes": Db.LenNodes(true) - localNodes,
		"CachedMaps":  len(Conf.ChildMaps),
	}
}

// GetToken generates a short random token and stores it in an
// in-memory map with its generation time. (See CheckToken.)
func (*Api) GetToken(ctx *jas.Context) {
	tokenid := rand.Uint32()
	ActiveTokens[tokenid] = token{ctx.RemoteAddr, time.Now()}
	ctx.Data = tokenid
}

// GetKey generates a CAPTCHA ID and returns it. This can be combined
// with the solution to the returned CAPTCHA to authenticate certain
// API functions. The CAPTCHAs can be accessed at /captcha/<id>.png or
// /captcha/<id>.wav.
func (*Api) GetKey(ctx *jas.Context) {
	ctx.Data = captcha.New()
}

// GetNode retrieves a single node from the database, removes
// sensitive data (such as an email address) and sets ctx.Data to
// it. If `?geojson` is set, then it returns it in geojson.Feature
// form.
func (*Api) GetNode(ctx *jas.Context) {
	ip := IP(net.ParseIP(ctx.RequireStringLen(0, 40, "address")))
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
	var err error

	// Require a token, because this is mildly sensitive.
	RequireToken(ctx)

	// Initialize the node and retrieve fields.
	node := new(Node)

	ip := IP(net.ParseIP(ctx.RequireStringLen(0, 40, "address")))
	if ip == nil {
		// If the address is invalid, return that error.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}
	node.Addr = ip
	node.Latitude = ctx.RequireFloat("latitude")
	node.Longitude = ctx.RequireFloat("longitude")
	node.OwnerName = html.EscapeString(ctx.RequireString("name"))
	node.OwnerEmail = ctx.RequireStringMatch(EmailRegexp, "email")

	node.Contact, _ = ctx.FindString("contact")
	node.Contact = html.EscapeString(node.Contact)

	node.Details, _ = ctx.FindString("details")
	node.Details = html.EscapeString(node.Details)

	// If Contact, Details, or OwnerName are too long to fit in
	// the database, produce an error.
	if len(node.Contact) > 255 {
		ctx.Error = jas.NewRequestError("contactTooLong")
		return
	}
	if len(node.Details) > 255 {
		ctx.Error = jas.NewRequestError("detailsTooLong")
		return
	}
	if len(node.OwnerName) > 255 {
		ctx.Error = jas.NewRequestError("ownerNameTooLong")
		return
	}

	// Validate the PGP ID, if given. It can be an lowercase hex
	// string of length 0, 8, or 16.
	pgpstr, _ := ctx.FindStringMatch(PGPRegexp, "pgp")
	if node.PGP, err = DecodePGPID([]byte(pgpstr)); err != nil {
		ctx.Error = jas.NewRequestError("pgpInvalid")
		return
	}
	status, _ := ctx.FindPositiveInt("status")
	node.Status = uint32(status)

	// Ensure that the node is correct and usable.
	if err = Db.VerifyRegistrant(node); err != nil {
		ctx.Error = jas.NewRequestError(err.Error())
		return
	}

	// TODO(DuoNoxSol): Authenticate/limit node registration.

	// If SMTP is missing from the config, we cannot continue.
	if Conf.SMTP == nil {
		ctx.Error = jas.NewInternalError(SMTPDisabledError)
		l.Err(SMTPDisabledError)
		return
	}

	// If SMTP verification is not explicitly disabled, and the
	// connecting address is not an admin, send an email.
	if !Conf.SMTP.VerifyDisabled && !IsAdmin(ctx.Request) {
		id := rand.Int63() // Pseudo-random positive int64

		emailsent := true
		if err := SendVerificationEmail(id, node.OwnerEmail); err != nil {
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

		// Add the new node to the RSS feed.
		AddNodeToRSS(node, time.Now())

		ctx.Data = "node registered"
		l.Infof("Node %q registered\n", ip)
	}
}

// PostUpdateNode removes a Node of a given IP from the database and
// re-adds it with the supplied information. It is the equivalent of
// removing a Node from the database, then invoking PostNode() with
// its information, with the exception that it does not send a
// verification email, and requires that the request be sent by the
// Node that is being update.
func (*Api) PostUpdateNode(ctx *jas.Context) {
	if Db.ReadOnly {
		// If the database is readonly, set that as the error and
		// return.
		ctx.Error = ReadOnlyError
		return
	}
	var err error

	// Require a token, because this is a very sensitive endpoint.
	RequireToken(ctx)

	// Retrieve the given IP address, check that it's sane, and check
	// that it exists in the *local* database.
	ip := IP(net.ParseIP(ctx.RequireStringLen(0, 40, "address")))
	if ip == nil {
		// If the address is invalid, return that error.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}

	node, err := Db.GetNode(ip)
	if err != nil {
		ctx.Error = jas.NewInternalError(err.Error())
		return
	}

	if node == nil || len(node.OwnerEmail) == 0 {
		ctx.Error = jas.NewRequestError("no matching local node")
		return
	}

	// Check to make sure that the Node is the one sending the
	// address, or an admin. If not, return an error.
	if !net.IP(ip).Equal(net.ParseIP(ctx.RemoteAddr)) &&
		!IsAdmin(ctx.Request) {
		ctx.Error = jas.NewRequestError(
			RemoteAddressDoesNotMatchError.Error())
		return
	}

	node.Addr = ip
	node.Latitude = ctx.RequireFloat("latitude")
	node.Longitude = ctx.RequireFloat("longitude")
	node.OwnerName = html.EscapeString(ctx.RequireString("name"))
	node.Contact, err = ctx.FindString("contact")
	if err != nil {
		ctx.Error = jas.NewRequestError("someError")
		return
	}
	node.Contact = html.EscapeString(node.Contact)
	node.Details, err = ctx.FindString("details")
	if err != nil {
		ctx.Error = jas.NewRequestError("someError")
		return
	}
	node.Details = html.EscapeString(node.Details)

	// If Contact, Details, or OwnerName are too long to fit in
	// the database, produce an error.
	if len(node.Contact) > 255 {
		ctx.Error = jas.NewRequestError("contactTooLong")
		return
	}
	if len(node.Details) > 255 {
		ctx.Error = jas.NewRequestError("detailsTooLong")
		return
	}
	if len(node.OwnerName) > 255 {
		ctx.Error = jas.NewRequestError("ownerNameTooLong")
		return
	}

	// Validate the PGP ID, if given. It can be an lowercase hex
	// string of length 0, 8, or 16.
	pgpstr, _ := ctx.FindStringMatch(PGPRegexp, "pgp")
	if node.PGP, err = DecodePGPID([]byte(pgpstr)); err != nil {
		ctx.Error = jas.NewRequestError("pgpInvalid")
		return
	}
	status, _ := ctx.FindPositiveInt("status")
	node.Status = uint32(status)

	// Note that we do not perform a verification step here, or send
	// an email. Because the Node was already verified once, we can
	// assume that it remains usable.

	// Update the Node in the database, replacing the one of matching
	// IP.
	err = Db.UpdateNode(node)
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Errf("Error updating %q: %s", node.Addr, err)
		return
	}

	// If we reach this point, all was successful.
	ctx.Data = "successful"
}

// PostDeleteNode removes a node with the given address from the
// database. This must be done from that node's address, or an admin
// address.
func (*Api) PostDeleteNode(ctx *jas.Context) {
	if Db.ReadOnly {
		// If the database is readonly, set that as the error and
		// return.
		ctx.Error = ReadOnlyError
		return
	}
	var err error

	// Require a token, because this is a very sensitive endpoint.
	RequireToken(ctx)

	// Retrieve the given IP address, check that it's sane, and check
	// that it exists in the *local* database.
	ip := IP(net.ParseIP(ctx.RequireStringLen(0, 40, "address")))
	if ip == nil {
		// If the address is invalid, return that error.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}

	// Check to make sure that the Node is the one sending the
	// address, or an admin. If not, return an error.
	if !net.IP(ip).Equal(net.ParseIP(ctx.RemoteAddr)) &&
		!IsAdmin(ctx.Request) {
		ctx.Error = jas.NewRequestError(
			RemoteAddressDoesNotMatchError.Error())
		return
	}

	// If all is well, then delete it.
	err = Db.DeleteNode(ip)
	if err == sql.ErrNoRows {
		// If there are no rows with that IP, explain that in the
		// error.
		//
		// I'm not actually sure this can happen. (DuoNoxSol)
		ctx.Error = jas.NewRequestError("no matching node")
	} else if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Errf("Error deleting node: %s\n")
	} else {
		l.Infof("Node %q deleted\n", ip)
		ctx.Data = "deleted"
	}
}

// GetVerify moves a node from the verification queue to the normal
// database, as identified by its long random ID.
func (*Api) GetVerify(ctx *jas.Context) {
	id := ctx.RequireInt("id")
	ip, verifyerr, err := Db.VerifyQueuedNode(id, ctx.Request)
	if verifyerr != nil {
		// If there was an error inverification, there was no internal
		// error, but the circumstances of the verification were
		// incorrect. It has not been removed from the database.
		ctx.Error = jas.NewRequestError(verifyerr.Error())
		return
	} else if err == sql.ErrNoRows {
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
// ones. If the form value `since` is supplied with a valid RFC3339
// timestamp, only nodes updated or cached more recently than that
// will be dumped. If 'geojson' is present, then the "data" field
// contains the dump in GeoJSON compliant form.
func (*Api) GetAll(ctx *jas.Context) {
	// We must invoke ParseForm() so that we can access ctx.Form.
	ctx.ParseForm()

	// In order to access this at the end, we need to declare nodes
	// here, so the results from the dump don't go out of scope.
	var nodes []*Node
	var err error

	// If the form value "since" was supplied, we will be doing a dump
	// based on update/retrieve time.
	if tstring := ctx.FormValue("since"); len(tstring) > 0 {
		var t time.Time
		t, err = time.Parse(time.RFC3339, tstring)
		if err != nil {
			ctx.Data = err.Error()
			ctx.Error = jas.NewRequestError("invalidTime")
			return
		}

		// Now, perform the time-based dump. Errors will be handled
		// outside the if block.
		nodes, err = Db.DumpChanges(t)
	} else {
		// If there was no "since," provide a simple full-database
		// dump.
		nodes, err = Db.DumpNodes()
	}

	// Handle any database errors here.
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Err(err)
		return
	}

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

// PostMessage emails the given message to the email address owned by
// the node with the given IP. It requires a correct and non-expired
// CAPTCHA pair be given.
func (*Api) PostMessage(ctx *jas.Context) {
	// Because this is a somewhat sensitive endpoint, require a token.
	RequireToken(ctx)

	// Ensure that the given CAPTCHA pair is correct. If it is not,
	// then return the explanation. This is bypassed if the request
	// comes from an admin address.
	if !IsAdmin(ctx.Request) {
		err := VerifyCAPTCHA(ctx.Request)
		if err != nil {
			ctx.Error = jas.NewRequestError(err.Error())
			return
		}
	}

	// Next, retrieve the IP of the node the user is attempting to
	// contact.
	ip := IP(net.ParseIP(ctx.RequireStringLen(0, 40, "address")))
	if ip == nil {
		// If the address is invalid, return that error.
		ctx.Error = jas.NewRequestError("addressInvalid")
		return
	}

	// Find the appropriate variables. If any of these are not
	// found, JAS will return a request error.
	replyto := ctx.RequireStringMatch(EmailRegexp, "from")
	message := ctx.RequireStringLen(0, 1000, "message")

	// Retrieve the appropriate node from the database.
	node, err := Db.GetNode(ip)
	if err != nil {
		// If we encounter an error here, it was a database error.
		ctx.Error = jas.NewInternalError(err)
		l.Err("Error getting node %q: %s", ip, err)
		return
	} else if node == nil {
		// If the IP wasn't found, explain that there was no node with
		// that IP.
		ctx.Error = jas.NewRequestError("address unknown")
		return
	} else if len(node.OwnerEmail) == 0 {
		// If there was no email on the node, that probably means that
		// it was cached.
		ctx.Error = jas.NewRequestError("address belongs to cached node")
		return
	}

	// Create and send an email. Log any errors.
	e := &Email{
		To:      node.OwnerEmail,
		From:    Conf.SMTP.EmailAddress,
		Subject: "Message via " + Conf.Name,
	}
	e.Data = map[string]interface{}{
		"ReplyTo": replyto,
		"Message": template.HTML(message),
		"Name":    Conf.Name,
		"Link": template.HTML(Conf.Web.Hostname + Conf.Web.Prefix +
			"/node/" + ip.String()),
		"AdminContact": Conf.AdminContact,

		// Generate a random number for use as a boundary marker in the
		// multipart/alternative email.
		"Boundary": rand.Int31(),
	}

	err = e.Send("message.txt")
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Errf("Error messaging %q from %q: %s",
			node.OwnerEmail, replyto, err)
		return
	}

	// Even if there is no error, log the to and from info, in case it
	// is abusive or spam.
	l.Noticef("IP %q sent a message to %q from %q",
		ctx.Request.RemoteAddr, node.OwnerEmail, replyto)
}

func (*Api) GetChildMaps(ctx *jas.Context) {
	var err error
	ctx.Data, err = Db.DumpChildMaps()
	if err != nil {
		ctx.Error = jas.NewInternalError(err)
		l.Errf("Error dumping child maps: %s", err)
	}
	return
}

// RequireToken uses the finder to retrieve a value named "token", and
// panics with "tokenInvalid" if there is either no such value, or it
// is invalid or expired.
func RequireToken(ctx *jas.Context) {
	tokeni, err := ctx.FindInt("token")
	if err != nil || !CheckToken(ctx.RemoteAddr, uint32(tokeni)) {
		panic(jas.NewRequestError("tokenInvalid"))
	}
}

// CheckToken ensures that a particular token is valid, meaning that
// it is in the list, and has not expired. If so, it removes the token
// and returns true. If the token is expired, it is removed, and the
// function returns false.
func CheckToken(IP string, token uint32) bool {
	t, ok := ActiveTokens[token]
	if ok {
		delete(ActiveTokens, token)
	}

	if !ok || time.Now().After(t.Issued.Add(time.Minute*5)) ||
		t.IP != IP {
		return false
	}
	return true
}

// IsAdmin is a small wrapper function to check if the given address
// belongs to an admin, as specified in Conf.AdminAddresses.
func IsAdmin(req *http.Request) bool {
	remoteIP := net.ParseIP(req.RemoteAddr)
	for _, adminAddr := range Conf.AdminAddresses {
		if net.IP(adminAddr).Equal(remoteIP) {
			return true
		}
	}
	return false
}
