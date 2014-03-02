API
===

NodeAtlas was designed to be entirely usable via just the API, which
is powered by the Go package [JAS][]. The web interface is just a
prettier front end to the API - requests are simply made through
[JQuery][]. This means that it's entirely possible to build
alternative clients for NodeAtlas or access it via the command line.

  [JAS]: http://godoc.org/github.com/coocood/jas
  [JQuery]: http://jquery.com/

For example, when `Verify.FromNode` is `true` in the configuration
file, verification emails are sent with the instruction to, if the
user is trying to add a remote node, execute the command `curl
http://nodeatlas.example.com/api/verify?id=0123456789` from the node
they're adding. This will allow them to verify their remote node from
its own address, which is required on certain instances.

This document describes API behavior as of version
`v0.5.9-17-g0ee47aa`, and possibly later. Major changes will likely
not be left undocumented, but there may be minor discrepancies.


## Accessing the API ##

The API for every NodeAtlas instance (even severely outdated versions)
is accessible at `/api/`. It returns a [JSON][]-encoded response of
the form `{ "data": {}, "error": null }`. 

  [JSON]: http://json.org

It can be accessed by any HTTP client, though a JSON-decoder is needed
to make programmatic use of the response. This is one advantage of
using JQuery to access it. JSON is human-readable, though, so it is
easy enough to use [`curl`][cURL] or [`wget`][wget] to access it by
hand. The API expects only HTTP GET and POST requests.

  [cURL]: http://curl.haxx.se/
  [wget]: https://www.gnu.org/software/wget/

## Endpoints ##

API endpoints are paths such as `/api/status` which return data of the
aforementioned form. All API outputs given below are piped through
`python -mjson.tool` for readability.

### / ###

`GET /api/` redirects (`303 See Other`) to this document as hosted on
the GitHub home page. It attaches the HTTP header `Location` in order
to do so, but also gives the URL in the `data` field.

```json
// curl -s "http://localhost:8077/api/"
{
    "data": "See Other: https://github.com/ProjectMeshnet/nodeatlas/blob/master/doc/API.md", 
    "error": null
}
```

### all ###

`GET /api/all` returns a complete list of nodes, both local and
cached, in native form. The data is given as a map of arrays, with the
key being the link to the parent node, or "local." Private email
addresses are never included.

The only error it will return is `InternalError`, which is usually
related to a database problem.

```json
// curl -s "http://localhost:8077/api/all"
{
    "data": {
        "http://map.maryland.projectmeshnet.org": [
            {
                "Addr": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149c", 
                "Latitude": 39.522979, 
                "Longitude": -76.993403, 
                "OwnerName": "Alexander Bauer", 
                "Status": 385
            }
        ], 
        "local": [
            {
                "Addr": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149b", 
                "Contact": "XMPP: duonoxsol@rows.io", 
                "Details": "Bay node", 
                "Latitude": 39.134321, 
                "Longitude": -76.360474, 
                "OwnerName": "Alexander Bauer", 
                "PGP": "76aad89b", 
                "Status": 257
            }
        ]
    }, 
    "error": null
}
```

Additionally, if the `?geojson` argument is supplied, data will be
dumped in [GeoJSON][] format. This is extremely useful for displaying
nodes.

  [GeoJSON]: http://geojson.org/

```json
// curl -s "http://localhost:8077/api/all?geojson"
{
    "data": {
        "features": [
            {
                "geometry": {
                    "coordinates": [
                        -76.360474, 
                        39.134321
                    ], 
                    "type": "Point"
                }, 
                "id": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149b", 
                "properties": {
                    "Contact": "XMPP: duonoxsol@rows.io", 
                    "Details": "Bay node", 
                    "OwnerName": "Alexander Bauer", 
                    "PGP": "76aad89b", 
                    "Status": 257
                }, 
                "type": "Feature"
            }, 
            {
                "geometry": {
                    "coordinates": [
                        -76.993403, 
                        39.522979
                    ], 
                    "type": "Point"
                }, 
                "id": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149c", 
                "properties": {
                    "OwnerName": "Alexander Bauer", 
                    "SourceID": 1, 
                    "Status": 385
                }, 
                "type": "Feature"
            }
        ], 
        "type": "FeatureCollection"
    }, 
    "error": null
}
```

### child_maps ###

`GET /api/child_maps` returns an array of objects containing the
hostname/link of the child map, its ID local to this instance, and the
name reported by querying `<hostname>/api/status`.

The only error it will return is `InternalError`, which is usually
related to a database problem.

```json
// curl -s "http://localhost:8077/api/child_maps"
{
    "data": [
        {
            "Hostname": "http://map.maryland.projectmeshnet.org", 
            "ID": 1, 
            "Name": "Maryland Mesh"
        }
    ], 
    "error": null
}
```

### key ###

`GET /api/key` generates a new CAPTCHA ID and solution pair in the
database, which will be stored for ten minutes, and returns the ID.

It will never return an error.

```json
// curl -s "http://localhost:8077/api/key"
{
    "data": "XpqCrgvtbyAJnKfYeaXN", 
    "error": null
}
```

### node ###

#### GET ####

`GET /api/node` retrieves data for precisely one node as addressed by
its IP, which can be either local or cached.

If the IP is misformatted or not present, it will return
`addressInvalid` or `No matching node` in the error field,
respectively.

```json
// curl -s "http://localhost:8077/api/node?address=fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149b"
{
    "data": {
        "Addr": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149b",
        "Contact": "XMPP: duonoxsol@rows.io",
        "Details": "Bay node",
        "Latitude": 39.134321,
        "Longitude": -76.360474,
        "OwnerName": "Alexander Bauer",
        "PGP": "76aad89b",
        "Status": 257
    }, 
    "error": null
}
```

It can also be formatted with `?geojson`, but that is currently
outdated and discouraged.

#### POST ####

`POST /api/node` is the means by which nodes are added to the map. If
`SMTP.VerifyDisabled` is `false` in the configuration file, this will
attempt to send a verification email on completion. It requires the
following fields.

```json
{
    "address": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149d",
	"latitude": 40.12345,
	"longitude": -80.54321,
	"name": "Alexander Bauer",
	"email": "duonoxsol@example.com",
}
```

The following fields can also be given, but are not required. The
`contact` and `details` fields must be shorter than 256 characters,
but are otherwise arbitrary plaintext. `pgp` can be 16, 8, or 0 hex
digits, and must be all lowercase, and `status` is a decimal `int32`
composed of single-bit flags, as specified [here][status].

  [status]: https://github.com/ProjectMeshnet/nodeatlas/issues/111

```json
{
    "contact": "XMPP: duonoxsol@rows.io",
	"details": "arbitrary data",
	"pgp": "76AAD89B",
	"status": 385
}
```

In addition, it requires a token.

If there is an error, it will will either be of the form
`<formkey>Invalid`, such as `addressInvalid` or `emailInvalid`. If
there is a database error, then it will return an `InternalError`.

```json
// curl -s -d "address=fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149d" -d "latitude=40.12345" -d "longitude=-80.54321" -d "name=Alexander Bauer" -d "email=duonoxsol@example.com" -d "contact=XMPP: duonoxsol@rows.io" -d "pgp=76AAD89B" -d "status=385" "http://localhost:8077/api/node"
{
    "data": "verification email sent", 
    "error": null
}
```

Or, if `SMTP.VerifyDisabled` is `true` in the configuration file, no
email will be sent, and the response will be:

```json
{
    "data": "successful", 
    "error": null
}
```

### status ###

`GET /api/status` returns simple parameters about the instance.

It will never return an error.

```json
// curl -s "http://localhost:8077/api/status"
{
    "data": {
        "CachedMaps": 1, 
        "CachedNodes": 7, 
        "LocalNodes": 49, 
        "Name": "Project Meshnet"
    }, 
    "error": null
}
```

### verify ###

`GET /api/verify` is used to verify a particular node ID via email. If
`Verify.FromNode` in the configuration is `true`, then it requires
that the request come from the address which is being verified.

If it returns an error, it will be either `verify: remote address does not match Node address` or a database-related `InternalError`.

```json
// curl -s "http://localhost:8077/api/verify?id=5085217136501410721"
{
    "data": "successful", 
    "error": null
}
```

### delete_node ###

`POST /api/delete_node` removes a local node from the database. It
requires that the connecting address match the address to be deleted,
or to be registered as an admin.

In addition, it requires a token.

If it returns an error, it will either be verify: `remote address does
not match Node address` or a database-related InternalError.

```json
// curl -s -d "address=fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149d" http://localhost:8077/api/delete_node
{
    "data": "deleted", 
    "error": null
}
```

### message ###

`POST /api/message` creates and sends an email to the address of the
owner of the given node. The address to which it is sent remains
private, and the IP of the sender is logged. It requires a non-expired
CAPTCHA id and solution pair to be provided, and the message must be
1000 characters or under.

Required fields are as follows.

```json
{
    "captcha": "id:solution",
	"address": "fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149d",
	"from": "me@example.com",
	"message": "Hello, I'd like to peer with your node on the Project Meshnet\n
NodeAtlas instance. Would you please provide peering details?"
}
```

In addition, it requires a token.

If there is an error, it will be a `CAPTCHA ID or solution is
incorrect`, `CAPTCHA format invalid`, `<formkey>Invalid` error, or a
database-related `InternalError`.

```json
// curl -s -d "captcha=n2teCkgMKdceXkEs5HiC:595801" -d "address=fcdf:db8b:fbf5:d3d7:64a:5aa3:f326:149d" -d "from=me@example.com" -d "message=Hello, I'd like to peer with your node on the Project Meshnet\nNodeAtlas instance. Would you please provide peering details?" "http://localhost:8077/api/message"
{
    "data":null,
	"error":null
}
```

### update_node ###

`POST /api/update_node` is very similar to [`POST /api/node`](#post),
except that it does not take the `email` form, and it can only be used
to update existing nodes. It requires that the request be sent from
the address which is being updated, or from an admin address.

In addition, it requires a token.

If there is an error, it will be of the form `<formkey>Invalid` or
`InternalError`.
