QA Prolog installation
======================

Dependencies
------------

A number of dependencies must be satisfied before QA Prolog can be built and run.  Specifically, you will need

* a [Go compiler](https://golang.org/),
* the [Yosys Open SYnthesis Suite](http://www.clifford.at/yosys/),
* [edif2qmasm](https://github.com/lanl/edif2qmasm),
* [QMASM](https://github.com/lanl/qmasm), and
* either [SAPI](https://www.dwavesys.com/software) (proprietary, for running on D‑Wave hardware) or [qbsolv](https://github.com/dwavesystems/qbsolv) (for classical solution).

Building QA Prolog
------------------

QA Prolog itself is written in the [Go programming language](https://en.wikipedia.org/wiki/Go_(programming_language)) so you'll need a [Go compiler](https://golang.org/) to build it.  One you have that, a simple
```bash
go get github.com/lanl/QA-Prolog
```
should suffice to download and build the code.  Alternatively, you can clone the GitHub repository and run either `go build` or `make`.

The Makefile additionally supports `install`, `clean`, and `maintainer-clean` targets.  The `install` target honors `DESTDIR`, `prefix`, and `bindir`.  After cleaning with `make maintainer-clean`, you will need to run `go generate` to regenerate a few `.go` files.  Regeneration relies on a couple of additional Go tools:

| Tool                                                          | Installation command                     |
| ------------------------------------------------------------- | ---------------------------------------- |
| [stringer](https://godoc.org/golang.org/x/tools/cmd/stringer) | `go get golang.org/x/tools/cmd/stringer` |
| [pigeon](https://godoc.org/github.com/mna/pigeon)             | `go get github.com/mna/pigeon`           |

