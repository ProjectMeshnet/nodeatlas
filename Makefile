PROGRAM_NAME := nodeatlas
GOCOMPILER := go build
GOFLAGS	+= -ldflags "-X main.Version $(shell git describe --dirty=+)"

# DEPS are non-hidden files found in the assets directory. It's very
# important that we do *not* use the `wildcard` function. That would
# cause the $(DEPS) rule to be skipped if there are no visible files
# in that directory, such as in the case of a `make clean`. By not
# using wildcard, we depend on the '*' file, which causes the rule to
# be executed.
DEPSFILE := depslist
DEPS := res/web/assets/*

.PHONY: all clean deps

all: $(DEPS) $(PROGRAM_NAME)

$(PROGRAM_NAME): $(wildcard *.go)
	$(GOCOMPILER) $(GOFLAGS)

# Download dependencies if the dependency list has changed more
# recently than the directory (or the directory is empty).
$(DEPS): $(DEPSFILE)
	- ./getdeps.sh

clean:
	@- $(RM) $(PROGRAM_NAME) $(DEPS)
