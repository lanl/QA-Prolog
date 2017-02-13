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
	// node is a predicate, skip its first child (the name of the predicate
	// itself).
	kids := a.Children
	if a.Type == PredicateType {
		kids = kids[1:]
	}
	for _, aa := range kids {
		aa.uniqueAtomNames(names)
	}
}
