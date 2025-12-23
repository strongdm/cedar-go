package ast

import "testing"

// TestIsTypeMarkerMethods tests that all IsType marker methods are callable for coverage
func TestIsTypeMarkerMethods(t *testing.T) {
	t.Parallel()

	// Call all isType() marker methods for coverage
	StringType{}.isType()
	LongType{}.isType()
	BoolType{}.isType()
	ExtensionType{}.isType()
	SetType{}.isType()
	RecordType{}.isType()
	EntityTypeRef{}.isType()
	TypeRef{}.isType()
}

// TestIsNodeMarkerMethods tests that all IsNode marker methods are callable for coverage
func TestIsNodeMarkerMethods(t *testing.T) {
	t.Parallel()

	// Call all isNode() marker methods for coverage
	(&NamespaceNode{}).isNode()
	(&CommonTypeNode{}).isNode()
	(&EntityNode{}).isNode()
	(&EnumNode{}).isNode()
	(&ActionNode{}).isNode()
}

// TestIsDeclarationMarkerMethods tests that all IsDeclaration marker methods are callable for coverage
func TestIsDeclarationMarkerMethods(t *testing.T) {
	t.Parallel()

	// Call all isDeclaration() marker methods for coverage
	(&CommonTypeNode{}).isDeclaration()
	(&EntityNode{}).isDeclaration()
	(&EnumNode{}).isDeclaration()
	(&ActionNode{}).isDeclaration()
}
