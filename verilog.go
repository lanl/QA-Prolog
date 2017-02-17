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

// args retrieves a clause's arguments in both Prolog and Verilog format.
func (c *ASTNode) args() (pArgs, vArgs []string) {
	pred := c.Children[0]
	terms := pred.Children[1:]
	pArgs = make([]string, len(terms)) // Prolog arguments (terms)
	vArgs = make([]string, len(terms)) // Verilog arguments (variables)
	for i, a := range terms {
		pArgs[i] = a.Value.(string)
		vArgs[i] = numToVerVar(i)
	}
	return
}

// process converts each predicate in a clause to an assignment to a valid bit.
func (c *ASTNode) process() []string {
	// Assign validity based on matches on any specified input symbols.
	valid := make([]string, 0, len(c.Children))
	pArgs, vArgs := c.args()
	for i, a := range pArgs {
		r0 := rune(a[0])
		switch {
		case unicode.IsLower(r0):
			// Symbol
			valid = append(valid, fmt.Sprintf("%s == `%s", vArgs[i], a))
		case unicode.IsDigit(r0):
			// Numeral
			valid = append(valid, fmt.Sprintf("%s == %s", vArgs[i], a))
		case unicode.IsUpper(r0):
			// Variable

		default:
			notify.Fatalf("Internal error processing %q", a)
		}
	}
	return valid
}

// writeClauseGroup writes a Verilog module corresponding to a group of clauses
// that have the same name and arity.
func (a *ASTNode) writeClauseGroup(w io.Writer, p *Parameters, nm string, cs []*ASTNode) {
	// Write a module header.
	_, vArgs := cs[0].args()
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

	// Assign validity conditions based on each clause in the clause group.
	for i, c := range cs {
		valid := c.process()
		fmt.Fprintf(w, "  wire [%d:0] $v%d;\n", len(valid)-1, i+1)
		for j, v := range valid {
			fmt.Fprintf(w, "  assign $v%d[%d] = %s;\n", i+1, j, v)
		}
	}

	// Set the final validity bit to the intersection of all predicate
	// validity bits.
	fmt.Fprint(w, "  assign $valid = ")
	for i := range cs {
		if i > 0 {
			fmt.Fprint(w, " | ")
		}
		fmt.Fprintf(w, "&$v%d", i+1)
	}
	fmt.Fprintln(w, ";")
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
