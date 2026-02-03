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
		ExtensionType("test"),
		SetType{},
		RecordType{},
		EntityTypeRef("test"),
		TypeRef("test"),
	}

	for _, tm := range typeMarkers {
		tm.isType()
	}
}
