// Perform type inference on an AST.

package main

import (
	"fmt"
	"unicode"
)

// A VarType is the inferred type of a variable.
type VarType int

// We define three different variable types.
const (
	InfUnknown VarType = iota // Unknown type
	InfNumeral                // Inferred numeral
	InfAtom                   // Inferred atom
)

// Convert a VarType to a string.
func (v VarType) String() string {
	switch v {
	case InfUnknown:
		return "*"
	case InfNumeral:
		return "num"
	case InfAtom:
		return "atom"
	default:
		notify.Fatalf("Internal error converting variable type %d to a string", v)
	}
	return "" // Will never get here
}

// TypeInfo represents a mapping from a variable to its type.
type TypeInfo map[string]VarType

// Merge merges two type mappings.
func MergeTypes(t1, t2 TypeInfo) (TypeInfo, error) {
	tm := make(TypeInfo, len(t1)+len(t2))

	// Populate tm with the union of all variables in t1 and t2.
	for k, _ := range t1 {
		tm[k] = InfUnknown
	}
	for k, _ := range t2 {
		tm[k] = InfUnknown
	}

	// Look for type conflicts for each variable in turn.
	for k := range tm {
		// To reduce the number of cases to check, assign each variable
		// a default value of InfUnknown.
		v1, in1 := t1[k]
		if !in1 {
			v1 = InfUnknown
		}
		v2, in2 := t2[k]
		if !in2 {
			v2 = InfUnknown
		}

		// Any type overrides InfUnknown.  Otherwise, types must match.
		switch {
		case v1 == v2:
			// Same type in both t1 and t2: retain that type.
			tm[k] = v1

		case v1 == InfUnknown:
			// v1 has an unknown type: use v2's type.
			tm[k] = v2

		case v2 == InfUnknown:
			// v2 has an unknown type: use v1's type.
			tm[k] = v1

		default:
			// v1 and v2 have incompatible types: complain.
			return nil, fmt.Errorf("Type conflict for variable %s", k)
		}
	}
	return tm, nil
}

// Return a map from clause name (e.g., "my_clause/3") to a list of AST nodes.
func (a *ASTNode) clauseNames() map[string][]*ASTNode {
	nm2node := make(map[string][]*ASTNode)
	var walk func(c *ASTNode)
	walk = func(c *ASTNode) {
		if c.Type == ClauseType || c.Type == QueryType {
			name := c.Value.(string)
			old, seen := nm2node[name]
			if seen {
				nm2node[name] = append(old, c)
			} else {
				nm2node[name] = []*ASTNode{c}
			}
			return
		}
		for _, cc := range c.Children {
			walk(cc)
		}
	}
	walk(a)
	return nm2node
}

// ClauseDependencies maps a clause name to the names of all clauses that it
// depends on.
type ClauseDependencies map[string]map[string]Empty

// findClauseDependencies walks an AST starting from a clause and reports its
// immediate dependencies.
func (a *ASTNode) findClauseDependencies() ClauseDependencies {
	// Find the name and arity of the current dependency (e.g.,
	// "my_clause/3").
	clName := a.Value.(string)
	deps := make(ClauseDependencies)

	// Define a function to recursively search an AST for dependencies.
	var findDeps func(c *ASTNode)
	findDeps = func(c *ASTNode) {
		if c.Type == PredicateType {
			// Ensure we're not an ordinary expression.
			if len(c.Children) <= 1 {
				// Not a reference to another clause
				return
			}

			// We have a dependency.  Store it.
			chName := fmt.Sprintf("%s/%d", c.Children[0].Value.(string), len(c.Children)-1)
			if _, ok := deps[clName]; !ok {
				deps[clName] = make(map[string]Empty)
			}
			deps[clName][chName] = Empty{}
			return
		}
		for _, ch := range c.Children {
			findDeps(ch)
		}
	}

	// Search our children for dependencies.
	for _, ch := range a.Children[1:] {
		findDeps(ch)
	}
	return deps
}

// findRoots returns the roots of a dependency graph.
func (d ClauseDependencies) findRoots() []string {
	// We don't currently support recursive functions.  Search for and
	// reject recursion.
	seen := make(map[string]Empty, len(d))
	var rejRec func(s string)
	rejRec = func(s string) {
		if _, found := seen[s]; found {
			notify.Fatalf("Recursion is not currently supported (%s)", s)
		}
		seen[s] = Empty{}
		for c := range d[s] {
			rejRec(c)
		}
	}

	// Start with every node that depends on another node as a potential
	// root.
	roots := make(map[string]Empty, len(d)*2)
	for r := range d {
		roots[r] = Empty{}
	}

	// Delete every node that another node depends upon.
	for _, cs := range d {
		for c := range cs {
			delete(roots, c)
		}
	}

	// Return the roots as a list.
	rList := make([]string, 0, len(roots))
	for r := range roots {
		rList = append(rList, r)
	}
	return rList
}

// orderedClauses returns a list of all of an AST's clauses sorted in
// dependency order.
func (a *ASTNode) orderedClauses(nm2cls map[string][]*ASTNode) []*ASTNode {
	// Build a complete set of dependencies for all clauses.
	var deps ClauseDependencies = make(ClauseDependencies)
	for _, cls := range nm2cls {
		for _, cl := range cls {
			deps[cl.Value.(string)] = make(map[string]Empty, 0)
		}
	}
	for _, cls := range nm2cls {
		for _, cl := range cls {
			for from, to2 := range cl.findClauseDependencies() {
				if _, ok := deps[from]; !ok {
					// First time we see from
					deps[from] = to2
					continue
				}
				for nm := range to2 {
					// Subsequent times we see from
					deps[from][nm] = Empty{}
				}
			}
		}
	}

	// Find a partial ordering of the dependency graph.
	roots := deps.findRoots()
	nodesSeen := make(map[string]Empty, len(roots)*2)
	ordNames := make([]string, 0, len(roots)*2)
	var makeOrder func(from string)
	makeOrder = func(nm string) {
		if _, seen := nodesSeen[nm]; seen {
			return
		}
		nodesSeen[nm] = Empty{}
		for c := range deps[nm] {
			makeOrder(c)
		}
		ordNames = append(ordNames, nm)
	}
	for _, r := range roots {
		makeOrder(r)
	}

	// Convert from strings to nodes.
	nDeps := len(ordNames)
	order := make([]*ASTNode, 0, nDeps)
	for _, nm := range ordNames {
		order = append(order, nm2cls[nm]...)
	}
	return order
}

// ArgTypes is a list of argument types for a clause.
type ArgTypes []VarType

// MergeArgTypes merges two lists of argument types.
func MergeArgTypes(a1, a2 ArgTypes) (ArgTypes, error) {
	if len(a1) != len(a2) {
		notify.Fatalf("Internal error: Length mismatch between %v and %v", a1, a2)
	}
	aTypes := make(ArgTypes, len(a1))
	for i, t1 := range a1 {
		t2 := a2[i]
		switch {
		case t1 == t2:
			aTypes[i] = t1
		case t1 == InfUnknown:
			aTypes[i] = t2
		case t2 == InfUnknown:
			aTypes[i] = t1
		default:
			return nil, fmt.Errorf("Polymorphic type signatures are not currently supported")
		}
	}
	return aTypes, nil
}

// When applied to a clause node, findClauseTypes augments a mapping from
// clause name to argument types and returns the type of each variable used in
// the clause.
func (a *ASTNode) findClauseTypes(nm2tys map[string]ArgTypes) TypeInfo {
	// Determine the name of each clause argument.
	argNames := make([]string, len(a.Children[0].Children[1:]))
	for i, c := range a.Children[0].Children[1:] {
		argNames[i] = c.Value.(string)
	}

	// Initialize the list of argument types.
	argTypes := make(ArgTypes, len(argNames))
	for i, nm := range argNames {
		r := rune(nm[0])
		switch {
		case unicode.IsLower(r):
			argTypes[i] = InfAtom
		case unicode.IsDigit(r):
			argTypes[i] = InfNumeral
		default:
			argTypes[i] = InfUnknown
		}
	}

	// Update the list of argument types based on what we can infer about
	// all variables that appear in the clause.
	vTypes := a.findVariableTypes(nm2tys)
	for i, ty := range argTypes {
		if ty == InfUnknown {
			if new_ty, ok := vTypes[argNames[i]]; ok {
				argTypes[i] = new_ty
			}
		}
	}

	// Merge the new argument list with the existing list, if any.
	cl := a.Value.(string)
	if oldTys, ok := nm2tys[cl]; ok {
		var err error
		argTypes, err = MergeArgTypes(oldTys, argTypes)
		CheckError(err)
	}

	// Assign the same type to every instance of a variable name.
	var2ty := make(map[string]VarType, len(argTypes))
	for i, v := range argNames {
		ty1 := argTypes[i]
		ty2, seen := var2ty[v]
		if seen {
			switch {
			case ty1 == ty2:
			case ty1 == InfUnknown:
				var2ty[v] = ty2
			case ty2 == InfUnknown:
				var2ty[v] = ty1
			default:
				notify.Fatalf("Type mismatch on variable %s in %s: %v vs. %v", v, cl, ty1, ty2)
			}
		} else {
			var2ty[v] = ty1
		}
	}
	for i, v := range argNames {
		argTypes[i] = var2ty[v]
	}

	// Update the map.
	nm2tys[cl] = argTypes
	return vTypes
}

// When applied to an expression node (specifically, RelationType or below),
// findExprType returns the node's type.
func (a *ASTNode) findExprType() VarType {
	switch a.Type {
	case NumeralType:
		return InfNumeral

	case AtomType:
		return InfAtom

	case VariableType:
		return InfUnknown

	case PrimaryExprType, UnaryExprType, MultiplicativeExprType, AdditiveExprType:
		// Arithmetic applies only to numerals.
		if len(a.Children) == 1 {
			// Trivial wrapper for an underlying expression: Ask
			// they underlying expression for its type.
			return a.Children[0].findExprType()
		}
		return InfNumeral

	case TermType:
		return a.Children[0].findExprType()

	case RelationType:
		// Relations are either numeric or unknown, depending on the
		// specific relation.
		op := a.Children[1].Value.(string)
		if op == "=" || op == "\\=" {
			// Equality and inequality are polymorphic.  See if we
			// can determine the type from our arguments.
			t1 := a.Children[0].findExprType()
			t2 := a.Children[2].findExprType()
			switch {
			case t1 == t2:
				return t1
			case t1 == InfUnknown:
				return t2
			case t2 == InfUnknown:
				return t1
			default:
				notify.Fatalf("Can't apply %q to mixed types (%v and %v)", op, t1, t2)
			}
		} else {
			// All other relations apply only to numerals.
			return InfNumeral
		}

	default:
		notify.Fatalf("Internal error: findExprType doesn't recognize %v", a.Type)
	}
	return InfUnknown // Will never get here.
}

// When applied to any AST node, allVariables returns a set of all variables
// named in that node.
func (a *ASTNode) allVariables() map[string]Empty {
	var m map[string]Empty
	if a.Type == VariableType {
		m = make(map[string]Empty)
		m[a.Value.(string)] = Empty{}
		return m
	}
	for _, c := range a.Children {
		mm := c.allVariables()
		if m == nil {
			// Skip the copy for the first child.
			m = mm
		} else {
			// Merge the remaining children's variable lists.
			for v := range mm {
				m[v] = Empty{}
			}
		}
	}
	return m
}

// When applied to a clause node, findVariableTypes returns a mapping from
// variable name to type.
func (a *ASTNode) findVariableTypes(nm2tys map[string]ArgTypes) TypeInfo {
	// Define a function that assigns the same type to all variables in all
	// of our child nodes.
	var err error
	tm := make(TypeInfo, 1)
	setAllChildren := func(c *ASTNode, ty VarType) {
		// Assign type ty to all child variables.
		vSet := c.allVariables()
		new_tm := make(TypeInfo, len(vSet))
		for k := range vSet {
			new_tm[k] = ty
		}

		// Merge the types with what we already have, then check that
		// the child variables do indeed all have the same type.  One
		// case in which they might not is if ty is InfUnknown but the
		// typemap tm already contains inconsistent types for two of
		// the variables.
		tm, err = MergeTypes(tm, new_tm)
		CheckError(err)
		kid1 := "???"
		kidTy := InfUnknown
		for k := range vSet {
			switch {
			case kidTy == InfUnknown:
				// Initialize kid1 and kidTy.
				kid1 = k
				kidTy = tm[k]
			case tm[k] == InfUnknown:
				// InfUnknown matches everything.
			case kidTy != tm[k]:
				ParseError(c.Pos, "Variables %s and %s have incompatible types", kid1, k)
			}
		}
	}

	// Figure out what to do based on the types of the clause's children.
	for _, p := range a.Children[1:] {
		c := p.Children[0]
		switch c.Type {
		case RelationType, TermType:
			// All variables in a relation or term must have the
			// same type.
			setAllChildren(c, c.findExprType())

		case AtomType:
			// Line up the predicate's arguments with the
			// corresponding clause's argument types.
			name := fmt.Sprintf("%s/%d", c.Value, len(p.Children)-1)
			tys, ok := nm2tys[name]
			if !ok {
				notify.Fatalf("Internal error: Failed to find clause %s", name)
			}
			new_tm := make(TypeInfo, len(tys))
			for i, ty := range tys {
				new_tm[p.Children[i+1].Value.(string)] = ty
			}
			tm, err = MergeTypes(tm, new_tm)
			CheckError(err)

		default:
			notify.Fatalf("Internal error: findVariableTypes doesn't recognize %v", c.Type)
		}
	}
	return tm
}

// PerformTypeInference returns a mapping from clause name to argument types
// for all clauses in the target AST.
func (a *ASTNode) PerformTypeInference() (map[string]ArgTypes, map[*ASTNode]TypeInfo) {
	// Compute a clause order in which to apply type inference.
	nm2cls := a.clauseNames()
	clauses := a.orderedClauses(nm2cls)

	// Populate our mapping from clause name to argument types with a few
	// built-in names.
	nm2tys := make(map[string]ArgTypes, len(clauses)+2)
	nm2tys["integer/1"] = ArgTypes{InfNumeral}
	nm2tys["atom/1"] = ArgTypes{InfAtom}

	// Perform type inference on each clause in turn.
	clVarTys := make(map[*ASTNode]TypeInfo, len(clauses))
	for _, cl := range clauses {
		clVarTys[cl] = cl.findClauseTypes(nm2tys)
	}

	// Ensure that we didn't wind up with any polymorphic clauses.
	for nm, tys := range nm2tys {
		for i, t := range tys {
			if t == InfUnknown {
				notify.Fatalf("%s is polymorphic (in argument %d), which is not currently supported", nm, i+1)
			}
		}
	}
	return nm2tys, clVarTys
}
