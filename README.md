# NodeAtlas
## Federated node mapping for mesh networks

*GPL 3+ Licensed, see LICENSE*  
*© Alexander Bauer, Daniel Supernault, Dylan Whichard, Luke Evers,
and contributors*

NodeAtlas is a high-performance and very portable tool for
geographically mapping mesh networks. It is used and designed by
[Project Meshnet][Atlas].

  [Atlas]: http://atlas.projectmeshnet.org
  [ProjectMeshnet]: https://projectmeshnet.org

It runs as a server which provides a web interface with two parts: a
map, and an API. The mapping portion provides a comfortable and
functional user interface using [Bootstrap][]. The map itself is
provided by [Leafletjs][], which loads tiles from [OpenStreetMap][]
(by default). Nodes are loaded by [JQuery][] from the API.

  [Bootstrap]: http://twitter.github.io/bootstrap/
  [Leafletjs]: http://leafletjs.com
  [JQuery]: http://jquery.com
  [OpenStreetMap]: http://www.openstreetmap.org

The NodeAtlas itself is written in [Go][], and its API is powered by
[JAS][], a RESTful JSON API framework.

  [Go]: http://golang.org
  [JAS]: https://github.com/coocood/jas#jas

In addition to the API, the Go backend provides a simple and powerful
means of federation. Child maps are specified in the configuration,
and NodeAtlas regularly queries their APIs, and pulls a list of node
information, including nodes from sub-children, when are then
displayed on the parent instance. This way, NodeAtlas is capable of
acting as a regional map, incorporating nodes from multiple more
localized instances. (More documentation on this behavior will be
added in the future.)
