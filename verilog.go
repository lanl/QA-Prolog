// Output an AST as Verilog code.

package main

import (
	"fmt"
	"io"
	"unicode"
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
		fmt.Fprintf(w, "`define %-*s %*d\n", p.IntBits, s, digs, i)
	}
}

// numToVerVar converts a parameter number from 0-701 (e.g., 5) to a Verilog
// variable (e.g., "\$E").
func numToVerVar(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const nChars = len(chars)
	switch {
	case n < nChars:
		return "$" + chars[n:n+1]
	case n < nChars*(nChars+1):
		n0 := n % nChars
		n1 := n / nChars
		return "$" + chars[n1-1:n1] + chars[n0:n0+1]
	default:
		notify.Fatal("Too many parameters")
	}
	return "" // Will never get here.
}

// writeClauseGroup writes a Verilog module corresponding to a group of clauses
// that have the same name and arity.
func (a *ASTNode) writeClauseGroup(w io.Writer, p *Parameters, nm string, cs []*ASTNode) {
	// Acquire the function prototype from the first clause.
	pred := cs[0].Children[0]
	terms := pred.Children[1:]
	pArgs := make([]string, len(terms)) // Prolog arguments
	vArgs := make([]string, len(terms)) // Verilog arguments
	for i, a := range terms {
		pArgs[i] = a.Value.(string)
		if unicode.IsUpper(rune(pArgs[i][0])) {
			vArgs[i] = pArgs[i] // Already a variable
		} else {
			vArgs[i] = numToVerVar(i) // Symbol; create a new variable name
		}
	}

	// Write a module header.
	fmt.Fprintf(w, "// Define %s.\n", nm)
	fmt.Fprintf(w, "module \\%s (", nm)
	for i, a := range vArgs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, a)
	}
	fmt.Fprintln(w, ", $valid);")
	if p.IntBits == 1 {
		for _, a := range vArgs {
			fmt.Fprintf(w, "  input %s;\n", a)
		}
	} else {
		for _, a := range vArgs {
			fmt.Fprintf(w, "  input [%d:0] %s;\n", p.IntBits-1, a)
		}
	}
	fmt.Fprintln(w, "  output $valid;")

	// Assign validity conditions based on the input symbols, if any.
	valid := make([]string, 0)
	for i, a := range pArgs {
		if unicode.IsLower(rune(a[0])) {
			valid = append(valid, fmt.Sprintf("%s == `%s", vArgs[i], a))
		}
	}

	// Set the final validity bit to the intersection of all predicate
	// validity bits.
	fmt.Fprintf(w, "  wire [%d:0] $v;\n\n", len(valid)-1)
	for i, v := range valid {
		fmt.Fprintf(w, "  assign $v[%d] = %s;\n", i, v)
	}
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "  assign $valid = &$v;")
	fmt.Fprintln(w, "endmodule")
}

// WriteVerilog writes an entire (preprocessed) AST as Verilog code.
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

	// Write each clause in turn.
	for nm, cs := range p.TopLevel {
		fmt.Fprintln(w, "")
		a.writeClauseGroup(w, p, nm, cs)
	}
}
