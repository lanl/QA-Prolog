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

// CheckError aborts with an error message if an error value is non-nil.
func CheckError(err error) {
	if err != nil {
		notify.Fatal(err)
	}
}

// BaseName returns a file path with the directory and extension removed.
func BaseName(filename string) string {
	return path.Base(strings.TrimSuffix(filename, path.Ext(filename)))
}

// Parameters encapsulates all command-line parameters as well as various
// global values computed from the AST.
type Parameters struct {
	// Command-line parameters
	ProgName   string // Name of this program
	InFileName string // Name of the input file
	WorkDir    string // Directory for holding intermediate files
	IntBits    uint   // Number of bits to use for each program integer
	Verbose    bool   // Whether to output verbose execution information

	// Computed values
	SymToInt      map[string]int        // Map from a symbol to an integer
	IntToSym      []string              // Map from an integer to a symbol
	TopLevel      map[string][]*ASTNode // Top-level clauses, grouped by name and arity
	SymBits       uint                  // Number of bits to use for each symbol
	OutFileBase   string                // Base name (no path or extension) for output files
	DeleteWorkDir bool                  // Whether to delete WorkDir at the end of the program
}

// ParseError reports a parse error at a given position.
var ParseError func(pos position, format string, args ...interface{})

// VerbosePrintf outputs a message only if verbose output is enabled.
func VerbosePrintf(p *Parameters, fmt string, args ...interface{}) {
	if !p.Verbose {
		return
	}
	notify.Printf("INFO: "+fmt, args...)
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
	flag.UintVar(&p.IntBits, "int-bits", 0, "minimum integer width in bits")
	flag.StringVar(&p.WorkDir, "work-dir", "", "directory for storing intermediate files (default: "+path.Join(os.TempDir(), "qap-*")+")")
	flag.BoolVar(&p.Verbose, "verbose", false, "output informational messages during execution")
	flag.BoolVar(&p.Verbose, "v", false, "same as -verbose")
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
		CheckError(err)
		defer f.Close()
		r = f
	}

	// Parse the input file into an AST.
	VerbosePrintf(&p, "Parsing %s as Prolog code", p.InFileName)
	a, err := ParseReader(p.InFileName, r)
	CheckError(err)
	ast := a.(*ASTNode)
	ast.RejectUnimplemented(&p)

	// Preprocess the AST.
	ast.StoreAtomNames(&p)
	ast.AdjustIntBits(&p)
	ast.BinClauses(&p)

	// Create a working directory and switch to it.
	CreateWorkDir(&p)
	err = os.Chdir(p.WorkDir)
	CheckError(err)

	// Output Verilog code.
	p.OutFileBase = BaseName(p.InFileName)
	vName := p.OutFileBase + ".v"
	vf, err := os.Create(vName)
	CheckError(err)
	VerbosePrintf(&p, "Writing Verilog code to %s", vName)
	ast.WriteVerilog(vf, &p)
	vf.Close()

	// Compile the Verilog code to an EDIF netlist.
	CreateYosysScript(&p)
	VerbosePrintf(&p, "Converting Verilog code to an EDIF netlist")
	RunCommand(&p, "yosys", "-q", p.OutFileBase+".v", p.OutFileBase+".ys",
		"-b", "edif", "-o", p.OutFileBase+".edif")

	// Compile the EDIF netlist to QMASM code.
	VerbosePrintf(&p, "Converting the EDIF netlist to QMASM code")
	RunCommand(&p, "edif2qmasm", "-o", p.OutFileBase+".qmasm", p.OutFileBase+".edif")

	// Optionally remove the working directory.
	if p.DeleteWorkDir {
		err = os.RemoveAll(p.WorkDir)
		CheckError(err)
	}
}
