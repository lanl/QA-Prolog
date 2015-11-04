# Build the Quantum Computing Prolog compiler
# By Scott Pakin <pakin@lanl.gov>

GO = go
PIGEON = pigeon

GO_SOURCES = qc-prolog.go astnodetype_string.go

all: qc-prolog

qc-prolog: $(GO_SOURCES)
	$(GO) build $(GO_SOURCES)

qc-prolog.go: qc-prolog.peg
	$(PIGEON) qc-prolog.peg > qc-prolog.tmp
	goimports qc-prolog.tmp | gofmt > qc-prolog.go
	$(RM) qc-prolog.tmp

astnodetype_string.go: qc-prolog.go
	$(GO) generate

clean:
	$(RM) qc-prolog qc-prolog.go qc-prolog.tmp
	$(RM) astnodetype_string.go
