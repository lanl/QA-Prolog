// Preprocess an AST before generating code.

package main

import (
	"fmt"
	"sort"
)

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

// AtomNames returns a sorted list of all atoms named in an AST except
// predicate names.
func (a *ASTNode) AtomNames() []string {
	nmSet := make(map[string]struct{})
	a.uniqueAtomNames(nmSet)
	nmList := make([]string, 0, len(nmSet))
	for nm := range nmSet {
		nmList = append(nmList, nm)
	}
	sort.Strings(nmList)
	return nmList
}

// uniqueAtomNames constructs a set of all atoms named in an AST except
// predicate names.  It performs most of the work for AtomNames.
func (a *ASTNode) uniqueAtomNames(names map[string]struct{}) {
	// Process the current AST node.
	if a.Type == AtomType {
		nm, ok := a.Value.(string)
		if !ok {
			notify.Fatalf("Internal error parsing %#v", *a)
		}
		names[nm] = struct{}{}
	}

	// Recursively process the current node's children.  If the current
	// node is a clause, skip its first child (the name of the clause
	// itself).
	kids := a.Children
	if a.Type == ClauseType {
		kids = kids[1:]
	}
	for _, aa := range kids {
		aa.uniqueAtomNames(names)
	}
}

// MaxNumeral returns the maximum-valued numeric literal.
func (a *ASTNode) MaxNumeral() int {
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
		m := aa.MaxNumeral()
		if m > max {
			max = m
		}
	}
	return max
}

// BinClauses groups all of the clauses in the program by name and arity.  The
// function returns a map with keys are of the form "<name>/<arity>" and values
// being the corresponding lists of clauses.
func (a *ASTNode) BinClauses() map[string][]*ASTNode {
	bins := make(map[string][]*ASTNode, 8)
	for _, cl := range a.FindByType(ClauseType) {
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
		nm := pr.Children[0].Value.(string)
		ar := len(pr.Children[1:])
		sym := fmt.Sprintf("%s/%d", nm, ar)

		// Associate the current clause with the symbol name.
		bins[sym] = append(bins[sym], cl)
	}
	return bins
}
