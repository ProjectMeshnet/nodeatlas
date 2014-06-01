PROGRAM_NAME := nodeatlas
VERSION := $(shell git describe --dirty=+)

ifndef GOCOMPILER
GOCOMPILER = go build $(GOFLAGS)
endif

# If the prefix is not yet defined, define it here.
ifndef prefix
prefix = /usr/local
endif

GOFLAGS	+= -ldflags "-X main.Version $(VERSION) \
	-X main.defaultResLocation $(prefix)/share/$(PROGRAM_NAME)/ \
	-X main.defaultConfLocation /etc/$(PROGRAM_NAME).conf"

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
	test -d $(prefix)/bin || mkdir -p $(prefix)/bin
	test -d $(prefix)/share/$(PROGRAM_NAME) || \
		mkdir -p $(prefix)/share/$(PROGRAM_NAME)

	install -m 0755 $(PROGRAM_NAME) $(prefix)/bin
	rm -rf $(prefix)/share/$(PROGRAM_NAME)
	cp --no-preserve=all -r res $(prefix)/share/$(PROGRAM_NAME)

clean:
	@- $(RM) $(PROGRAM_NAME) $(DEPS)

# vim: set noexpandtab:
