package main

// Copyright (C) 2013 Alexander Bauer, Luke Evers, Daniel Supernault,
// Dylan Whichard, and contributors; (GPLv3) see LICENSE or doc.go

import (
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// metaData is a convenience type which wraps a *Config, but adds
// information such as Version.
type metaData struct {
	*Config

	Version string
}

// CompileStatic traverses a directory, gathering a list of all
// files. It then performs actions based on their filenames, and
// copies them to a temporary directory. It returns the name of the
// temporary directory.
//
// Files with the ".tmpl" extention are run through 'html/template'
// engine with the data argument being the supplied *Config.
func CompileStatic(dir string, conf *Config) (tmpdir string, err error) {
	// Create a directory in the OS's tempdir, (`/tmp`), with the
	// prefix being the name of the map.
	tmpdir, err = ioutil.TempDir("", conf.Name)
	if err != nil {
		return
	}

	// Next, get a slice of all files in the directory, and any in its
	// subdirectories (recursively).
	files := make([]string, 0)
	err = crawlDirectories(&files, tmpdir, dir)
	if err != nil {
		return
	}

	// Now, process the files one at a time, writing the result to
	// `<tmpdir>/path/to/file`. The path is the path to the file
	// within the given dir.
	err = processFiles(files, metaData{conf, Version}, tmpdir, dir)
	if err != nil {
		return
	}

	return
}

// processFiles uses the file extension(s) to transform each file,
// placing the result in the outdir with the same path, with the
// prefix sliced off from the filename. Templates are executed with
// conf as data.It will panic if any of the files are shorter than
// len(prefix).
func processFiles(files []string, data metaData,
	outdir, prefix string) error {
	for _, filename := range files {
		// Now, process based on file extension.
		switch path.Ext(filename) {
		case ".tmpl":
			// First, open the outfile. html/template handles the
			// infile. Note that it strips out the ".tmpl" extension.
			out, err := os.Create(path.Join(outdir,
				filename[len(prefix):len(filename)-5]))
			if err != nil {
				return err
			}
			defer out.Close()

			// Next, parse the template from the file.
			t, err := template.ParseFiles(filename)
			if err != nil {
				return err
			}

			// Finally, write it to the file using conf as data.
			err = t.Execute(out, data)
			if err != nil {
				return err
			}
		default:
			// Begin by opening the in file and creating the out file.
			in, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer in.Close()
			out, err := os.Create(path.Join(outdir, filename[len(prefix):]))
			if err != nil {
				return err
			}
			defer out.Close()

			// If the file extension was not recognized, just copy it
			// directly.
			_, err = io.Copy(out, in)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// crawlDirectories uses os.Readdir() on the given dir to retrieve a
// list of os.FileInfo types. It goes through that list, sorting them
// by file extension and inserting them in the given list of files. If
// it encounters a directory, it creates that directory under the
// tmpdir, and recursively calls itself until it has a listed every
// file.
func crawlDirectories(files *[]string, tmpdir, dir string) (err error) {
	// First, open the file so that we can use Readdir().
	f, err := os.Open(dir)
	if err != nil {
		return
	}

	// Next, get a list of os.FileInfo types, and close the file as
	// soon as possible.
	fis, err := f.Readdir(0)
	f.Close()
	if err != nil {
		return
	}

	// Initialize a slice for new filenames with length zero, and
	// capacity len(fi). That way, actual files can be appended easily
	// without leaving blanks.
	newFiles := make([]string, 0, len(fis))

	// Loop through the slice of os.FileInfo types and check if each
	// is a directory. If it isn't, append it to newFiles. If it is,
	// call crawlDirectories() on it again.
	for _, fi := range fis {
		if len(fi.Name()) > 0 &&
			(strings.HasPrefix(fi.Name(), ".") ||
				strings.HasPrefix(fi.Name(), "#") ||
				strings.HasSuffix(fi.Name(), "~")) {
			// Make sure not to include files that start with . or #
			// Also make sure not to include files that end in ~
		} else if !fi.IsDir() {
			// If a file...
			newFiles = append(newFiles, path.Join(dir, fi.Name()))
		} else {
			// If a directory...
			newdir := path.Join(tmpdir, fi.Name())
			err = os.Mkdir(newdir, 0777)
			if err != nil {
				return
			}
			// Here, we tell the next call that the tmpdir is the one
			// we just created in our tmpdir, and the directory to
			// crawl is the one we just discovered.
			err = crawlDirectories(files,
				newdir, path.Join(dir, fi.Name()))
			if err != nil {
				return
			}
		}
	}

	// Finally, append newFiles to the target slice.
	*files = append(*files, newFiles...)
	return
}
