// Code generated by "stringer -type=ASTNodeType"; DO NOT EDIT.

package main

import "strconv"

const _ASTNodeType_name = "UnknownTypeNumeralTypeAtomTypeVariableTypeTermTypeTermListTypeListTailTypeListTypePrimaryExprTypeUnaryExprTypeUnaryOpTypeMultiplicativeExprTypeMultiplicativeOpTypeAdditiveExprTypeAdditiveOpTypeRelationOpTypeRelationTypePredicateTypeStructureTypePredicateListTypeClauseTypeClauseListTypeQueryTypeProgramType"

var _ASTNodeType_index = [...]uint16{0, 11, 22, 30, 42, 50, 62, 74, 82, 97, 110, 121, 143, 163, 179, 193, 207, 219, 232, 245, 262, 272, 286, 295, 306}

func (i ASTNodeType) String() string {
	if i < 0 || i >= ASTNodeType(len(_ASTNodeType_index)-1) {
		return "ASTNodeType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ASTNodeType_name[_ASTNodeType_index[i]:_ASTNodeType_index[i+1]]
}
