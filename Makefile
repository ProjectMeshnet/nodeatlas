PROGRAM_NAME := nodeatlas
GOCOMPILER := go build
GOFLAGS	+= -ldflags "-X main.Version $(shell git describe --dirty=+)"


.PHONY: all clean

all: $(PROGRAM_NAME)

$(PROGRAM_NAME):
	$(GOCOMPILER) $(GOFLAGS)

clean:
	@- $(RM) $(PROGRAM_NAME)
