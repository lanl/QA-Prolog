# Build the Quantum Computing Prolog compiler
# By Scott Pakin <pakin@lanl.gov>

GO = go
PIGEON = pigeon

GO_SOURCES = qc-prolog.go parser.go astnodetype_string.go

all: qc-prolog

qc-prolog: $(GO_SOURCES)
	$(GO) build $(GO_SOURCES)

parser.go: parser.peg
	$(PIGEON) parser.peg > parser.tmp
	goimports parser.tmp | gofmt > parser.go
	$(RM) parser.tmp

astnodetype_string.go: parser.go
	$(GO) generate

clean:
	$(RM) qc-prolog qc-prolog.tmp
	$(RM) parser.go astnodetype_string.go
