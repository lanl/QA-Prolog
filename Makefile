# Build the Quantum Computing Prolog compiler
# By Scott Pakin <pakin@lanl.gov>

GO = go
PIGEON = pigeon

all: qc-prolog

qc-prolog: qc-prolog.go
	$(GO) build qc-prolog.go

qc-prolog.go: qc-prolog.peg
	$(PIGEON) qc-prolog.peg > qc-prolog.tmp
	goimports qc-prolog.tmp | gofmt > qc-prolog.go
	$(RM) qc-prolog.tmp

clean:
	$(RM) qc-prolog qc-prolog.go qc-prolog.tmp
