# NodeAtlas
## Federated node mapping for mesh networks

[![Build Status](https://travis-ci.org/ProjectMeshnet/nodeatlas.png?branch=master)](https://travis-ci.org/ProjectMeshnet/nodeatlas)

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


## Install

TODO

## Configuration

NodeAtlas needs a configuration file. By default, NodeAtlas looks for
`conf.json` in the current directory. There is a file called
`conf.json.example` in the repository, which is a template for what
the configuration file should look like.

You can tell NodeAtlas to use a configuration file from anywhere else
by using the `--conf` flag. For example:

```
nodeatlas --res res/ --conf /etc/nodeatlas.json
```

For documentation on what exactly every line in your configuration
file does, see [CONFIGURATION][] in the `doc` folder.

  [CONFIGURATION]: ./doc/CONFIGURATION.md

## Contributing

If you see something that needs fixing, or you can think of something
that could make NodeAtlas better, please feel free to open an issue or
submit a pull request. Check the open issues before doing so! If there
is already an issue open for what you want to help with, don't open
another; one will suffice. All issues and pull requests are welcome,
and encouraged.

## Copyright & License

&copy; Alexander Bauer, Luke Evers, Dylan Whichard, and
contributors. NodeAtlas is licensed under GPLv3. See [LICENSE][] for
full detials.

  [LICENSE]: ./LICENSE
