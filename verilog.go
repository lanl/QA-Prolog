// Output an AST as Verilog code.

package main

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"unicode"
)

// Return a random string to use for an instance name.
func generateSuffix() string {
	const nChars = 5 // Number of characters to generate
	const nmChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	suffix := make([]byte, nChars)
	for i := range suffix {
		suffix[i] = nmChars[rand.Intn(len(nmChars))]
	}
	return string(suffix)
}

// writeSymbols defines all of an AST's symbols as Verilog constants.
func (a *ASTNode) writeSymbols(w io.Writer, p *Parameters) {
	// Determine the minimum number of characters needed to represent all
	// symbol names.
	nSymChars := 1
	for _, s := range p.IntToSym {
		if len(s) > nSymChars {
			nSymChars = len(s)
		}
	}

	// Output nicely formatted symbol definitions.
	fmt.Fprintln(w, "// Define all of the symbols used in this program.")
	for i, s := range p.IntToSym {
		fmt.Fprintf(w, "`define %-*s %d'd%d\n", nSymChars, s, p.SymBits, i)
	}
}

// args retrieves a clause's or a query's arguments in both Prolog and Verilog
// format.  In the former case, arguments are renamed to A, B, C, etc.  This is
// needed to handle both non-variable arguments (i.e., numerals or atoms) and
// repeated variable arguments (as in "same(X, X)").  In the latter case, the
// original names are preserved.  This is needed to report results in terms of
// the program-specified variables names (and it is known a priori that clauses
// contain only unique variable names in their argument lists).
func (a *ASTNode) args() (pArgs, vArgs []string) {
	pred := a.Children[0]
	terms := pred.Children[1:]
	pArgs = make([]string, len(terms)) // Prolog arguments (terms)
	vArgs = make([]string, len(terms)) // Verilog arguments (variables)
	for i, t := range terms {
		pArgs[i] = t.Text
		if a.Type == ClauseType {
			vArgs[i] = numToVerVar(i)
		} else {
			vArgs[i] = pArgs[i]
		}
	}
	return
}

// prologToVerilogUnary maps a Prolog unary operator to a Verilog unary
// operator.
var prologToVerilogUnary = map[string]string{"-": "-"}

// prologToVerilogAdd maps a Prolog additive operator to a Verilog additive
// operator.
var prologToVerilogAdd = map[string]string{
	"+": "+",
	"-": "-",
}

// prologToVerilogMult maps a Prolog multiplicative operator to a Verilog
// multiplicative operator.
var prologToVerilogMult = map[string]string{"*": "*"}

// prologToVerilogRel maps a Prolog relational operator to a Verilog relational
// operator.
var prologToVerilogRel = map[string]string{
	"=<":  "<=",
	">=":  ">=",
	"<":   "<",
	">":   ">",
	"=":   "==",
	"\\=": "!=",
	"is":  "==",
}

// toVerilogExpr recursively converts an AST, starting from a clause's body
// predicate, to an expression.
func (a *ASTNode) toVerilogExpr(p *Parameters, p2v map[string]string) string {
	switch a.Type {
	case NumeralType:
		return fmt.Sprintf("%d'd%s", p.IntBits, a.Text)

	case AtomType:
		return "`" + a.Value.(string)

	case VariableType:
		v, ok := p2v[a.Value.(string)]
		if !ok {
			notify.Fatalf("Internal error: Failed to convert variable %s from Prolog to Verilog", a.Value.(string))
		}
		return v

	case UnaryOpType:
		v, ok := prologToVerilogUnary[a.Value.(string)]
		if !ok {
			notify.Fatalf("Internal error: Failed to convert %s %q from Prolog to Verilog", a.Type, a.Value.(string))
		}
		return v

	case AdditiveOpType:
		v, ok := prologToVerilogAdd[a.Value.(string)]
		if !ok {
			notify.Fatalf("Internal error: Failed to convert %s %q from Prolog to Verilog", a.Type, a.Value.(string))
		}
		return v

	case MultiplicativeOpType:
		v, ok := prologToVerilogMult[a.Value.(string)]
		if !ok {
			notify.Fatalf("Internal error: Failed to convert %s %q from Prolog to Verilog", a.Type, a.Value.(string))
		}
		return v

	case RelationOpType:
		v, ok := prologToVerilogRel[a.Value.(string)]
		if !ok {
			notify.Fatalf("Internal error: Failed to convert %s %q from Prolog to Verilog", a.Type, a.Value.(string))
		}
		return v

	case PrimaryExprType:
		c := a.Children[0].toVerilogExpr(p, p2v)
		if a.Value.(string) == "()" {
			return "(" + c + ")"
		}
		return c

	case UnaryExprType:
		if len(a.Children) == 1 {
			return a.Children[0].toVerilogExpr(p, p2v)
		}
		return a.Children[0].toVerilogExpr(p, p2v) + a.Children[1].toVerilogExpr(p, p2v)

	case MultiplicativeExprType:
		if len(a.Children) == 1 {
			return a.Children[0].toVerilogExpr(p, p2v)
		}
		c1 := a.Children[0].toVerilogExpr(p, p2v)
		v := a.Children[1].toVerilogExpr(p, p2v)
		c2 := a.Children[2].toVerilogExpr(p, p2v)
		return c1 + v + c2

	case AdditiveExprType:
		if len(a.Children) == 1 {
			return a.Children[0].toVerilogExpr(p, p2v)
		}
		c1 := a.Children[0].toVerilogExpr(p, p2v)
		v := a.Children[1].toVerilogExpr(p, p2v)
		c2 := a.Children[2].toVerilogExpr(p, p2v)
		return c1 + " " + v + " " + c2

	case RelationType:
		c1 := a.Children[0].toVerilogExpr(p, p2v)
		v := a.Children[1].toVerilogExpr(p, p2v)
		c2 := a.Children[2].toVerilogExpr(p, p2v)
		return c1 + " " + v + " " + c2

	case TermType:
		return a.Children[0].toVerilogExpr(p, p2v)

	case PredicateType:
		// Handle predicate AST nodes that are really just wrappers for
		// expressions.
		if len(a.Children) == 1 {
			return a.Children[0].toVerilogExpr(p, p2v)
		}

		// Ignore atom/1 and integer/1, which exist solely for the type
		// system.
		if len(a.Children) == 2 {
			pName := a.Children[0].Value.(string)
			if pName == "atom" || pName == "integer" {
				return "1'b1"
			}
		}

		cs := make([]string, 0, len(a.Children)*2)
		for i, c := range a.Children {
			switch i {
			case 0:
				vStr := c.toVerilogExpr(p, p2v)
				sfx := generateSuffix()
				arity := len(a.Children) - 1
				cs = append(cs, fmt.Sprintf("\\%s/%d \\%s_%s/%d",
					vStr[1:], arity, vStr[1:], sfx, arity)) // Strip leading "`" from vStr.
			case 1:
				cs = append(cs, " (")
				cs = append(cs, c.toVerilogExpr(p, p2v))
			default:
				cs = append(cs, ", ")
				cs = append(cs, c.toVerilogExpr(p, p2v))
			}
		}
		cs = append(cs, ", %s)")
		return strings.Join(cs, "")

	default:
		notify.Fatalf("Internal error: Unexpected AST node type %s", a.Type)
	}
	return "" // We should never get here.
}

// process converts each predicate in a clause to an assignment to a valid bit.
func (a *ASTNode) process(p *Parameters, p2v map[string]string) []string {
	// Assign validity based on matches on any specified input symbols or
	// numbers.
	valid := make([]string, 0, len(a.Children))
	pArgs, vArgs := a.args()
	for i, pa := range pArgs {
		r0 := rune(pa[0])
		switch {
		case unicode.IsLower(r0):
			// Symbol
			valid = append(valid, fmt.Sprintf("%s == `%s", vArgs[i], pa))
		case unicode.IsDigit(r0):
			// Numeral
			valid = append(valid, fmt.Sprintf("%s == %d'd%s", vArgs[i], p.IntBits, pa))
		case unicode.IsUpper(r0), r0 == '_':
			// Variable

		default:
			notify.Fatalf("Internal error processing %q", pa)
		}
	}

	// Assign validity based on each predicate in the clause's body.
	for _, pred := range a.Children[1:] {
		v := pred.toVerilogExpr(p, p2v)
		if v != "1'b1" {
			valid = append(valid, v)
		}
	}
	return valid
}

// writeClauseGroupHeader is used by writeClauseGroup to write a Verilog module
// header.
func (a *ASTNode) writeClauseGroupHeader(w io.Writer, p *Parameters, nm string, cs []*ASTNode, tys ArgTypes) {
	// Write the module prototype.
	_, vArgs := cs[0].args()
	rawName := strings.Split(nm, "/")[0]
	if len(tys) == 0 {
		// No arguments (rare)
		fmt.Fprintf(w, "// Define %s.\n", rawName)
	} else {
		// At least one argument (common)
		for i, ty := range tys {
			if i == 0 {
				fmt.Fprintf(w, "// Define %s(%v", rawName, ty)
			} else {
				fmt.Fprintf(w, ", %v", ty)
			}
		}
		fmt.Fprintln(w, ").")
	}
	if rawName == "Query" {
		fmt.Fprint(w, "module Query (") // Exclude the arity from the top-level query.
	} else {
		fmt.Fprintf(w, "module \\%s (", nm)
	}
	for i, a := range vArgs {
		if i > 0 {
			fmt.Fprint(w, ", ")
		}
		fmt.Fprint(w, a)
	}
	if len(vArgs) > 0 {
		fmt.Fprintln(w, ", Valid);")
	} else {
		fmt.Fprintln(w, "Valid);")
	}

	// Write the module inputs.
	for i, a := range vArgs {
		bits := p.IntBits
		if tys[i] == InfAtom {
			bits = p.SymBits
		}
		if bits == 1 {
			fmt.Fprintf(w, "  input %s;\n", a)
		} else {
			fmt.Fprintf(w, "  input [%d:0] %s;\n", bits-1, a)
		}
	}

	// Write the module output.
	fmt.Fprintln(w, "  output Valid;")
}

// writeClauseBody is used by writeClauseGroup to assign a Verilog bit for each
// Prolog predicate in a clause's body.  It returns the number of new variables
// introduced.
func (a *ASTNode) writeClauseBody(w io.Writer, p *Parameters, nm string,
	cNum int, nVars int, vTy TypeInfo) int {
	// Construct a map from Prolog variables to Verilog variables.  As we
	// go along, constrain all variables with the same Prolog name to have
	// the same value.
	valid := make([]string, 0)
	pArgs, vArgs := a.args()
	p2v := make(map[string]string, len(pArgs))
	for i, p := range pArgs {
		v, seen := p2v[p]
		if seen {
			valid = append(valid, vArgs[i]+" == "+v)
		} else {
			p2v[p] = vArgs[i]
		}
	}

	// Introduce more Verilog variables for local Prolog variables.
	for pName, vName := range a.augmentVerilogVars(nVars, p2v) {
		bits := p.IntBits
		if vTy[pName] == InfAtom {
			bits = p.SymBits
		}
		if bits == 1 {
			fmt.Fprintf(w, "  (* keep *) wire %s;\n", vName)
		} else {
			fmt.Fprintf(w, "  (* keep *) wire [%d:0] %s;\n", bits-1, vName)
		}
		p2v[pName] = vName
		nVars++
	}

	// Convert the clause body to a list of Boolean Verilog
	// expressions.
	valid = append(valid, a.process(p, p2v)...)
	if len(valid) == 0 {
		// Although not normally used in practice, handle
		// useless clauses that accept all inputs (e.g.,
		// "stupid(A, B, C).").
		valid = append(valid, "1'b1")
	}
	if len(valid) == 1 {
		// Single bit
		fmt.Fprintf(w, "  wire $v%d;\n", cNum+1)
		vBit := fmt.Sprintf("$v%d", cNum+1)
		v := valid[0]
		if strings.Contains(v, "%s") {
			fmt.Fprintf(w, "  "+v+";\n", vBit)
		} else {
			fmt.Fprintf(w, "  assign %s = %s;\n", vBit, v)
		}
	} else {
		// Multiple bits
		fmt.Fprintf(w, "  wire [%d:0] $v%d;\n", len(valid)-1, cNum+1)
		for i, v := range valid {
			vBit := fmt.Sprintf("$v%d[%d]", cNum+1, i)
			if strings.Contains(v, "%s") {
				fmt.Fprintf(w, "  "+v+";\n", vBit)
			} else {
				fmt.Fprintf(w, "  assign %s = %s;\n", vBit, v)
			}
		}
	}
	return nVars
}

// writeClauseGroup writes a Verilog module corresponding to a group of clauses
// that have the same name and arity.
func (a *ASTNode) writeClauseGroup(w io.Writer, p *Parameters, nm string,
	cs []*ASTNode, tys ArgTypes, clVarTys map[*ASTNode]TypeInfo) {
	// Write a module header.
	a.writeClauseGroupHeader(w, p, nm, cs, tys)

	// Assign validity conditions based on each clause in the clause group.
	_, vArgs := cs[0].args()
	nVars := len(vArgs)
	for i, c := range cs {
		nVars += c.writeClauseBody(w, p, nm, i, nVars, clVarTys[c])
	}

	// Set the final validity bit to the intersection of all predicate
	// validity bits.
	fmt.Fprint(w, "  assign Valid = ")
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
func (a *ASTNode) WriteVerilog(w io.Writer, p *Parameters,
	nm2tys map[string]ArgTypes, clVarTys map[*ASTNode]TypeInfo) {
	// Output some header comments.
	fmt.Fprintf(w, "// Verilog version of Prolog program %s\n", p.InFileName)
	fmt.Fprintf(w, "// Conversion by %s, written by Scott Pakin <pakin@lanl.gov>\n", p.ProgName)
	fmt.Fprintln(w, `//
// This program is intended to be passed to edif2qmasm, then to qmasm, and
// finally run on a quantum annealer.
//`)
	fmt.Fprintf(w, "// Note: This program uses %d bit(s) for atoms and %d bit(s) for (unsigned)\n", p.SymBits, p.IntBits)
	fmt.Fprintln(w, "// integers.")
	fmt.Fprintln(w, "")

	// Define constants for all of our symbols.
	a.writeSymbols(w, p)

	// Write each clause in turn.
	for nm, cs := range p.TopLevel {
		fmt.Fprintln(w, "")
		a.writeClauseGroup(w, p, nm, cs, nm2tys[nm], clVarTys)
	}
}
