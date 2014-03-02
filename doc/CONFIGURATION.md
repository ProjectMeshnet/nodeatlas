# Configuration

NodeAtlas needs a configuration file. By default, NodeAtlas looks for
`conf.json` in the current directory. There is a file called
`conf.json.example` in the repository, which is a template for what
the configuration file should look like.

You can tell NodeAtlas to use a configuration file from anywhere else
by using the `--conf` flag. For example:

```
nodeatlas --res res/ --conf /etc/nodeatlas.json
```

Below is a list of every config variable and what it is for.

### Name

Name is the string by which this instance of NodeAtlas will be
referred to. It usually describes the entire project name or the
region about which it focuses.

### AdminContact

AdminContact is the structure which contains information relating to
where you can contact the administrator.

#### Name

Name of the administrator.

#### Email

Email of the administrator.

#### PGP

PGP key of the administrator.

### AdminAddresses

AdminAddresses is a slice of addresses which are considered fully
authenticated. Connections originating from those addresses will not
be required to verify or perform any sort of authentication, meaning
that they can edit or register any node. If it is not specified, no
addresses are granted this ability.

### Web

Web is the structure which contains information relating to the
backend of the HTTP webserver.

#### Hostname

Hostname is the address which NodeAtlas should identify itself as. For
example, in a verification email, NodeAtlas would give the
verification link as http://<hostname><prefix>/verify/<long-random-id>

#### Prefix

Prefix is the URL prefix which is required to access the front
end. For example, with a prefix of "/nodeatlas", NodeAtlas would be
able to respond to http://example.com/nodeatlas.

#### Addr

Addr is the network protocol, interface, and port to which NodeAtlas
should bind. For example, "tcp://0.0.0.0:8077" will bind globally to
the 8077 TCP port, and "unix://nodeatlas.sock" will create a UNIX
socket at nodeatlas.sock.

#### DeproxyHeaderFields

DeproxyHeaderFields is a list of HTTP header fields that should be
used instead of the connecting IP when verifying nodes and logging
major errors. They must be in canonicalized form, such as
"X-Forwarded-For" or "X-Real-IP".

#### HeaderSnippet

HeaderSnippet is a snippet of code which is inserted into the <head>
of each page. For example, one could include a script tieing into
Pikwik.

#### AboutSnippet

AboutSnippet is an excerpt that will get put into the /about page for
all to read upon going to the /about page.

#### RSS

RSS is the structure which contains settings for the built-in RSS feed
generator.

##### MaxAge

MaxAge is the duration after which new nodes are considered old, and
should no longer populate the feed.

### ChildMaps

ChildMaps is a list of addresses from which to pull lists of nodes
every heartbeat. Please note that these maps are trusted fully, and
they could easily introduce false nodes to the database temporarily
(until cleared by the CacheExpiration.

### Database

Database is the structure

#### DriverName

DriverName contains the database driver name, such as "sqlite3" or
"mysql."

#### Resource

Resource is the database resource, such as a path to .db file, or
username, password, and name.

#### ReadOnly

ReadOnly is a true/false variable deciding if we can write to the
database or not.

### HeartbeatRate

HeartbeatRate is the amount of time to wait between performing regular
tasks, such as clearing expired nodes from the queue and cache.

### CacheExpiration

CacheExpiration is the amount of time for which to store cached nodes
before considering them outdated, and removing them.

### VerificationExpiration

VerificationExpiration is the amount of time to allow users to verify
nodes by email after initially placing them. See the documentation for
time.ParseDuration for format information.

### ExtraVerificationFlags

ExtraVerificationFlags can be specified to add additional flags (such
as "-6") to the curl and wget instructions in the verification email.

### SMTP

SMTP contains the information necessary to connect to a mail relay, so
as to send verification email to registered nodes.

#### VerifyDisabled

VerifyDisabled controls whether email verification is used for newly
registered nodes. If it is false or omitted, an email will be sent
using the SMTP settings defined in this struct.

#### EmailAddress

EmailAddress will be given as the "From" address when sending email.

#### Username

Username is the username required by the server to login.

#### Password

Password is the password required by the server to login.

#### NoAuthenticate

NoAuthenticate determines whether NodeAtlas should attempt to
authenticate with the SMTP relay or not. Unless the relay is local,
leave this false.

#### ServerAddress

ServerAddress is the address of the SMTP relay, including the port.

### Map

Map contains the information used by NodeAtlas to power the Leaflet.js
map.

#### Favicon

Favicon is the icon to be displayed in the browser when viewing the
map. It is a filename to be loaded from `<*fRes>/icon/`.

#### Tileserver

Tileserver is the URL used for loading tiles. It is of the form
"http://{s}.tile.osm.org/{z}/{x}/{y}.png", so that Leaflet.js can use
it.

#### Center

Center contains the coordinates on which to center the map.

##### Latitude

The latitude of the coordinate on which to center the map.

##### Longitude

The longitude of the coordinates on which to center the map

#### Zoom

Zoom is the Leaflet.js zoom level to start the map at.

#### ClusterRadius

ClusterRadius is the range (in pixels) at which markers on the map
will cluster together.

#### Attibution

Attribution is the "map data" copyright notice placed at the bottom
right of the map, meant to credit the maintainers of the tileserver.

#### AddressType

AddressType is the text that is displayed on the map next to "Address"
when adding a new node or editing a previous node. The default is
"Network-specific IP", but due to how general that is, it should be
changed to whatever is the mosthelpful for people to understand.

### Verify

Verify contains the list of steps used to ensure that new nodes are
valid when registered. They can be enabled or disabled according to
one's needs.

#### Netmask

Netmask, if not nil, is a CIDR-form network mask which requires that
nodes registered have an Addr which matches it. For example,
"fc00::/8" would only allow IPv6 addresses in which the first two
digits are "fc", and "192.168.0.0/16" would only allow IPv4 addresses
in which the first two bytes are "192.168".

#### FromNode

FromNode requires the verification request (GET
/api/verify?id=<long_random_id>) to originate from the address of the
node that is being verified.
