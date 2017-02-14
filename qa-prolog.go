// This program implements a compiler for Quantum-Annealing Prolog.  It accepts
// a small subset of Prolog and generates weights for a Hamiltonian expression,
// which can be fed to a quantum annealer such as the D-Wave supercomputer.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

var notify *log.Logger // Help notify the user of warnings and errors.

// BaseName returns a file path with the directory and extension removed.
func BaseName(filename string) string {
	return path.Base(strings.TrimSuffix(filename, path.Ext(filename)))
}

// Parameters encapsulates all program parameters.
type Parameters struct {
	ProgName   string // Name of this program
	InFileName string // Name of the input file
	IntBits    uint   // Number of bits to use for each program integer
}

func main() {
	// Parse the command line.
	p := Parameters{}
	p.ProgName = BaseName(os.Args[0])
	notify = log.New(os.Stderr, p.ProgName+": ", 0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [<options>] [<infile.pl>]\n\n", p.ProgName)
		flag.PrintDefaults()
	}
	flag.UintVar(&p.IntBits, "int-bits", 0, "Minimum integer width in bits")
	flag.Parse()
	if flag.NArg() == 0 {
		p.InFileName = "<stdin>"
	} else {
		p.InFileName = flag.Arg(0)
	}

	// Parse the input file into an AST.
	var r io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(p.InFileName)
		if err != nil {
			notify.Fatal(err)
		}
		defer f.Close()
		r = f
	}
	ast, err := ParseReader(p.InFileName, r)
	if err != nil {
		notify.Fatal(err)
	}

	// Temporary
	a := ast.(*ASTNode)
	fmt.Println(a)
	fmt.Printf("ATOMS: %v\n", a.AtomNames())
	fmt.Printf("MAX NUM: %d\n", a.MaxNumeral())
}
