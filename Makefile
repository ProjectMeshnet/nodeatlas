PROGRAM_NAME := $(shell basename $(shell pwd))
VERSION := $(shell git describe --dirty=+)

ifndef GOCOMPILER
GOCOMPILER = go build $(GOFLAGS)
endif

# If the root and prefix are not yet defined, define them here.
ifndef DESTDIR
DESTDIR = /
endif

ifndef prefix
prefix = usr/local
endif

GOFLAGS	+= -ldflags "-X main.Version $(VERSION) \
	-X main.defaultResLocation $(DESTDIR)/$(prefix)/share/nodeatlas/ \
	-X main.defaultConfLocation $(DESTDIR)etc/nodeatlas.conf"

.PHONY: all install clean deps


# DEPS are non-hidden files found in the assets directory. Because we
# are using the glob rather than `wildcard`, this rule is *not*
# skipped when there are no visible files in the deps directory.
DEPSFILE := depslist
DEPS := res/web/assets/*

all: $(DEPS) $(PROGRAM_NAME)

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER)

# Download dependencies if the dependency list has changed more
# recently than the directory (or the directory is empty).
$(DEPS): $(DEPSFILE)
	@- $(RM) $(wildcard $(DEPS))
	- ./getdeps.sh

install: all
	test -d $(DESTDIR)/$(prefix)/bin || mkdir -p $(DESTDIR)/$(prefix)/bin
	test -d $(DESTDIR)/$(prefix)/share/nodeatlas || \
		mkdir -p $(DESTDIR)/$(prefix)/share/nodeatlas

	install -m 0755 $(PROGRAM_NAME) $(DESTDIR)/$(prefix)/bin/nodeatlas
	rm -rf $(DESTDIR)/$(prefix)/share/nodeatlas
	cp --no-preserve=all -r res $(DESTDIR)/$(prefix)/share/nodeatlas

clean:
	@- $(RM) $(PROGRAM_NAME) $(DEPS)

pkg_arch:
	mkdir -p build
	cp Makefile build/Makefile
	sed "s/pkgver=.*/pkgver=$(shell git describe | sed \
's/-/_/g')/" < packaging/PKGBUILD | sed "s/_gitver=.*/_gitver=\
$(shell git rev-parse HEAD)/" > build/PKGBUILD
	updpkgsums build/PKGBUILD

# vim: set noexpandtab:
