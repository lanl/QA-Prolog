// Preprocess an AST before generating code.

package main

import "sort"

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
