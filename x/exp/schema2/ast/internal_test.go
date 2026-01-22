package ast

import (
	"testing"
)

func TestIsTypeMarkerMethods(t *testing.T) {
	t.Parallel()

	typeMarkers := []IsType{
		StringType{},
		LongType{},
		BoolType{},
		ExtensionType{},
		SetType{},
		RecordType{},
		EntityTypeRef{},
		TypeRef{},
	}

	for _, tm := range typeMarkers {
		tm.isType()
	}
}

func TestIsNodeMarkerMethods(t *testing.T) {
	t.Parallel()

	nodeMarkers := []IsNode{
		NamespaceNode{},
		CommonTypeNode{},
		EntityNode{},
		EnumNode{},
		ActionNode{},
	}

	for _, nm := range nodeMarkers {
		nm.isNode()
	}
}

func TestIsDeclarationMarkerMethods(t *testing.T) {
	t.Parallel()

	declarationMarkers := []IsDeclaration{
		CommonTypeNode{},
		EntityNode{},
		EnumNode{},
		ActionNode{},
	}

	for _, dm := range declarationMarkers {
		dm.isDeclaration()
	}
}
