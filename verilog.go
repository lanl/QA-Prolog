// Output an AST as Verilog code.

package main

import (
	"fmt"
	"io"
)

// writeSymbols defines all of an AST's symbols as Verilog constants.
func (a *ASTNode) writeSymbols(w io.Writer, p *Parameters) {
	// Determine the minimum number of decimal digits needed to represent
	// all symbol values.
	digs := 0
	for n := len(p.IntToSym) - 1; n > 0; n /= 10 {
		digs++
	}

	// Output nicely formatted symbol definitions.
	// TODO: Correct Verilog syntax once I regain Internet access.
	fmt.Fprintln(w, "// Define all of the symbols used in this program.")
	for i, s := range p.IntToSym {
		fmt.Fprintf(w, "const `%-*s = %*d\n", p.IntBits, s, digs, i)
	}
}

// WriteVerilog writes an entire AST as Verilog code.
func (a *ASTNode) WriteVerilog(w io.Writer, p *Parameters) {
	// Output some header comments.
	fmt.Fprintf(w, "// Verilog version of Prolog program %s\n", p.InFileName)
	fmt.Fprintf(w, "// Conversion by %s, written by Scott Pakin <pakin@lanl.gov>\n", p.ProgName)
	fmt.Fprintln(w, `//
// This program is intended to be passed to edif2qmasm, then to qmasm, and
// finally run on a quantum annealer.
//`)
	fmt.Fprintf(w, "// Note: This program uses exclusively %d-bit unsigned integers.\n\n", p.IntBits)

	// Define constants for all of our symbols.
	a.writeSymbols(w, p)
}
