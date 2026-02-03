package resolver

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
		EntityType("test"),
	}

	for _, tm := range typeMarkers {
		tm.isType()
	}
}
