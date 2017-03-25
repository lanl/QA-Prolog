// Preprocess an AST before generating code.

package main

import (
	"fmt"
	"sort"
)

// BitsNeeded reports the number of bits needed to represent a given
// nonnegative integer.
func BitsNeeded(n int) uint {
	b := uint(0)
	for ; n > 0; n >>= 1 {
		b++
	}
	return b
}

// RejectUnimplemented rejects the AST (i.e., aborts the program) if it
// contains elements we do not currently know how to process.
func (a *ASTNode) RejectUnimplemented(p *Parameters) {
	if n := a.FindByType(ListType); len(n) > 0 {
		ParseError(n[0].Pos, "Lists are not currently supported")
	}
	if n := a.FindByType(StructureType); len(n) > 0 {
		ParseError(n[0].Pos, "Structures are not currently supported")
	}
}

// FindByType walks an AST and returns a list of all nodes of a given type.
func (a *ASTNode) FindByType(t ASTNodeType) []*ASTNode {
	nodes := make([]*ASTNode, 0, 8)
	var walker func(n *ASTNode)
	walker = func(n *ASTNode) {
		if n.Type == t {
			nodes = append(nodes, n)
		}
		for _, c := range n.Children {
			walker(c)
		}
	}
	walker(a)
	return nodes
}

// StoreAtomNames stores both a forward and reverse map between all atoms named
// in an AST (except predicate names) and integers.
func (a *ASTNode) StoreAtomNames(p *Parameters) {
	// Construct a map from integers to symbols.
	nmSet := make(map[string]Empty)
	a.uniqueAtomNames(nmSet, false)
	p.IntToSym = make([]string, 0, len(nmSet))
	for nm := range nmSet {
		p.IntToSym = append(p.IntToSym, nm)
	}
	sort.Strings(p.IntToSym)

	// Construct a map from symbols to integers.
	p.SymToInt = make(map[string]int, len(p.IntToSym))
	for i, s := range p.IntToSym {
		p.SymToInt[s] = i
	}
	p.SymBits = BitsNeeded(len(p.IntToSym) - 1)
	if p.SymBits == 0 {
		p.SymBits = 1 // Need at least one bit
	}
}

// uniqueAtomNames constructs a set of all atoms named in an AST except
// predicate names.  It performs most of the work for AtomNames.
func (a *ASTNode) uniqueAtomNames(names map[string]Empty, skip1 bool) {
	// Process the current AST node.
	if a.Type == AtomType {
		nm, ok := a.Value.(string)
		if !ok {
			notify.Fatalf("Internal error parsing %#v", *a)
		}
		names[nm] = Empty{}
	}

	// Recursively process the current node's children.  If the current
	// node is a clause or a query, skip its first child's first child (the
	// name of the clause/query itself).
	kids := a.Children
	if skip1 {
		kids = kids[1:]
	}
	skip1 = (a.Type == ClauseType || a.Type == QueryType)
	for _, aa := range kids {
		aa.uniqueAtomNames(names, skip1)
	}
}

// maxNumeral returns the maximum-valued numeric literal.
func (a *ASTNode) maxNumeral() int {
	// Process the current node.
	max := 0
	if a.Type == NumeralType {
		m := a.Value.(int)
		if m > max {
			max = m
		}
	}

	// Recursively process each of the node's children.
	for _, aa := range a.Children {
		m := aa.maxNumeral()
		if m > max {
			max = m
		}
	}
	return max
}

// AdjustIntBits increments the integer width to accomodate both the
// maximum-valued numeric literal and the number of symbol literals.
// This function assumes that StoreAtomNames has already been called.
func (a *ASTNode) AdjustIntBits(p *Parameters) {
	// Ensure we can store the maximum integer literal.
	b := BitsNeeded(a.maxNumeral())
	if p.IntBits < b {
		p.IntBits = b
	}

	// We can't handle 0-bit integers so round up to 1 if necessary.
	if p.IntBits == 0 {
		p.IntBits = 1
	}
}

// BinClauses groups all of the clauses in the program by name and arity.  The
// function returns a map with keys are of the form "<name>/<arity>" and values
// being the corresponding lists of clauses.
func (a *ASTNode) BinClauses(p *Parameters) {
	bins := make(map[string][]*ASTNode, 8)
	csAndQs := append(a.FindByType(ClauseType), a.FindByType(QueryType)...)
	for _, cl := range csAndQs {
		// Perform a lot of error-checking as we search for the clause
		// name.
		if len(cl.Children) == 0 {
			notify.Fatal("Internal error: Clause with no children")
		}
		pr := cl.Children[0]
		if pr.Type != PredicateType {
			notify.Fatal("Internal error: Clause with no predicate first child")
		}
		if len(pr.Children) == 0 {
			notify.Fatal("Internal error: Predicate with no children")
		}

		// Extract the symbol name (<name>/<arity>).
		nm := pr.Children[0].Value
		ar := len(pr.Children[1:])
		sym := fmt.Sprintf("%s/%d", nm, ar)

		// Associate the current clause with the symbol name.
		bins[sym] = append(bins[sym], cl)
	}
	p.TopLevel = bins
}

// numToVerVar converts a parameter number from 0-701 (e.g., 5) to a
// lettered Verilog variable (e.g., "E").
func numToVerVar(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const nChars = len(chars)
	switch {
	case n < nChars:
		return chars[n : n+1]
	case n < nChars*(nChars+1):
		n0 := n % nChars
		n1 := n / nChars
		return chars[n1-1:n1] + chars[n0:n0+1]
	default:
		notify.Fatal("Too many parameters")
	}
	return "" // Will never get here.
}

// augmentVerilogVars associates Verilog variables with Prolog variables unless
// they already appear in the given map.
func (a *ASTNode) augmentVerilogVars(minN int, p2v map[string]string) map[string]string {
	varNodes := a.FindByType(VariableType)
	new_p2v := make(map[string]string, len(varNodes))
	for _, pv := range varNodes {
		pVar := pv.Text
		if _, seen := p2v[pVar]; seen {
			continue
		}
		if _, seen := new_p2v[pVar]; seen {
			continue
		}
		new_p2v[pVar] = numToVerVar(minN)
		minN++
	}
	return new_p2v
}
