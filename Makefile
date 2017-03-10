# Build the Quantum Annealing Prolog compiler
# By Scott Pakin <pakin@lanl.gov>

GO = go
PIGEON = pigeon

GO_SOURCES = \
	qa-prolog.go \
	parser.go \
	preproc.go \
	verilog.go \
	type-inf.go \
	astnodetype_string.go

all: qa-prolog

qa-prolog: $(GO_SOURCES)
	$(GO) build $(GO_SOURCES)

parser.go: parser.peg
	$(PIGEON) parser.peg > parser.tmp
	goimports parser.tmp | gofmt > parser.go
	$(RM) parser.tmp

astnodetype_string.go: parser.go
	$(GO) generate

clean:
	$(RM) qa-prolog qa-prolog.tmp
	$(RM) parser.go astnodetype_string.go
