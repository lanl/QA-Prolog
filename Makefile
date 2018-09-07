###############################################
# Build the Quantum Annealing Prolog compiler #
# By Scott Pakin <pakin@lanl.gov>             #
###############################################

prefix = /usr/local
bindir = $(prefix)/bin
INSTALL = install
GO = go
GO_SOURCES = \
	qa-prolog.go \
	parser.go \
	preproc.go \
	run.go \
	verilog.go \
	type-inf.go \
	astnodetype_string.go

all: qa-prolog

qa-prolog: $(GO_SOURCES)
	$(GO) build $(GO_SOURCES)

parser.go astnodetype_string.go: parser.peg
	$(GO) generate

clean:
	$(RM) qa-prolog

maintainer-clean:
	$(RM) parser.go astnodetype_string.go

install: qa-prolog
	$(INSTALL) -m 0755 -D qa-prolog $(DESTDIR)$(bindir)

.PHONY: all clean maintainer-clean install
