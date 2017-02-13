// This program implements a compiler for Quantum-Annealing Prolog.  It accepts
// a small subset of Prolog and generates weights for a Hamiltonian expression,
// which can be fed to a quantum annealer such as the D-Wave supercomputer.
package main

import (
	"fmt"
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

func main() {
	// Parse the input file into an AST.
	progName := BaseName(os.Args[0])
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.pl>\n", progName)
		os.Exit(1)
	}
	notify = log.New(os.Stderr, progName+": ", 0)
	ast, err := ParseFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// For now, we simply output the AST.
	fmt.Println(ast)
}
