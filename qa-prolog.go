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

// Parameters encapsulates all command-line parameters as well as various
// global values computed from the AST.
type Parameters struct {
	// Command-line parameters
	ProgName    string // Name of this program
	InFileName  string // Name of the input file
	OutFileName string // Name of the output file
	IntBits     uint   // Number of bits to use for each program integer

	// Computed values
	SymToInt map[string]int        // Map from a symbol to an integer
	IntToSym []string              // Map from an integer to a symbol
	TopLevel map[string][]*ASTNode // Top-level clauses, grouped by name and arity
	SymBits  uint                  // Number of bits to use for each symbol
}

// ParseError reports a parse error at a given position.
var ParseError func(pos position, format string, args ...interface{})

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
	flag.StringVar(&p.OutFileName, "output", "-", "name of Verilog output file or \"-\" for stdout")
	flag.StringVar(&p.OutFileName, "o", "-", "same as -output")
	flag.Parse()
	if flag.NArg() == 0 {
		p.InFileName = "<stdin>"
	} else {
		p.InFileName = flag.Arg(0)
	}
	ParseError = func(pos position, format string, args ...interface{}) {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: ", p.InFileName, pos.line, pos.col)
		fmt.Fprintf(os.Stderr, format, args...)
		fmt.Fprintln(os.Stderr, "")
		os.Exit(1)
	}

	// Open the input file.
	var r io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(p.InFileName)
		if err != nil {
			notify.Fatal(err)
		}
		defer f.Close()
		r = f
	}

	// Open the output file.
	var out io.Writer
	if p.OutFileName == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(p.OutFileName)
		if err != nil {
			notify.Fatal(err)
		}
		defer f.Close()
		out = f
	}

	// Parse the input file into an AST.
	a, err := ParseReader(p.InFileName, r)
	if err != nil {
		notify.Fatal(err)
	}
	ast := a.(*ASTNode)
	ast.RejectUnimplemented(&p)

	// Preprocess the AST.
	ast.StoreAtomNames(&p)
	ast.AdjustIntBits(&p)
	ast.BinClauses(&p)

	// Output Verilog code.
	ast.WriteVerilog(out, &p)
}
