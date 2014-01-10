package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"github.com/SashaCrofter/staticdir"
	"io/ioutil"
)

// CompileStatic uses staticdir to translate a resource directory into
// static content. It returns the directory into which it copies.
func CompileStatic(dir string, conf *Config) (tmpdir string, err error) {
	// Create a directory in the OS's tempdir, (`/tmp`), with the
	// prefix being the name of the map.
	tmpdir, err = ioutil.TempDir("", conf.Name)
	if err != nil {
		return
	}

	// Create a translator via package staticdir.
	t := staticdir.New(dir, tmpdir)
	t.CopyFunc = staticdir.TemplateCopy
	t.CopyData = struct {
		*Config
		Version string
	}{conf, Version}

	return tmpdir, t.Translate()
}
