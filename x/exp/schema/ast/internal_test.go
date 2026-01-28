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
