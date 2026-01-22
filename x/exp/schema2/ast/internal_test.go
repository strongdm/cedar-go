package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/types"
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

func TestNamespaceIterators(t *testing.T) {
	t.Parallel()

	ct1 := CommonType("MyType1", StringType{})
	ct2 := CommonType("MyType2", LongType{})
	e1 := Entity("User")
	e2 := Entity("Group")
	enum1 := Enum("Status", "active", "inactive")
	enum2 := Enum("Role", "admin", "user")
	a1 := Action("read")
	a2 := Action("write")

	ns := Namespace(types.Path("MyApp"), ct1, e1, enum1, a1, ct2, e2, enum2, a2)

	tests := []struct {
		name      string
		countFunc func() int
		want      int
		checkName func(int) bool
	}{
		{
			"CommonTypes",
			func() int {
				count := 0
				for ct := range ns.CommonTypes() {
					if count == 0 && ct.Name != ct1.Name {
						t.Errorf("expected first common type to be ct1")
					}
					if count == 1 && ct.Name != ct2.Name {
						t.Errorf("expected second common type to be ct2")
					}
					count++
				}
				return count
			},
			2,
			nil,
		},
		{
			"Entities",
			func() int {
				count := 0
				for e := range ns.Entities() {
					if count == 0 && e.Name != e1.Name {
						t.Errorf("expected first entity to be e1")
					}
					if count == 1 && e.Name != e2.Name {
						t.Errorf("expected second entity to be e2")
					}
					count++
				}
				return count
			},
			2,
			nil,
		},
		{
			"Enums",
			func() int {
				count := 0
				for e := range ns.Enums() {
					if count == 0 && e.Name != enum1.Name {
						t.Errorf("expected first enum to be enum1")
					}
					if count == 1 && e.Name != enum2.Name {
						t.Errorf("expected second enum to be enum2")
					}
					count++
				}
				return count
			},
			2,
			nil,
		},
		{
			"Actions",
			func() int {
				count := 0
				for a := range ns.Actions() {
					if count == 0 && a.Name != a1.Name {
						t.Errorf("expected first action to be a1")
					}
					if count == 1 && a.Name != a2.Name {
						t.Errorf("expected second action to be a2")
					}
					count++
				}
				return count
			},
			2,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.countFunc()
			if got != tt.want {
				t.Errorf("expected %d items, got %d", tt.want, got)
			}
		})
	}
}

func TestNamespaceIteratorsEmpty(t *testing.T) {
	t.Parallel()

	ns := Namespace(types.Path("Empty"))

	iteratorTests := []struct {
		name      string
		countFunc func() int
	}{
		{"CommonTypes", func() int { return countNamespaceIter(ns.CommonTypes()) }},
		{"Entities", func() int { return countNamespaceIter(ns.Entities()) }},
		{"Enums", func() int { return countNamespaceIter(ns.Enums()) }},
		{"Actions", func() int { return countNamespaceIter(ns.Actions()) }},
	}

	for _, tt := range iteratorTests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.countFunc()
			if got != 0 {
				t.Errorf("expected 0 items, got %d", got)
			}
		})
	}
}

func TestNamespaceIteratorsEarlyBreak(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ns        NamespaceNode
		countFunc func(NamespaceNode) int
	}{
		{
			"CommonTypes early break",
			Namespace(types.Path("Test"), CommonType("Type1", StringType{}), CommonType("Type2", LongType{}), CommonType("Type3", BoolType{})),
			func(ns NamespaceNode) int {
				count := 0
				for range ns.CommonTypes() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Entities early break",
			Namespace(types.Path("Test"), Entity("Entity1"), Entity("Entity2"), Entity("Entity3")),
			func(ns NamespaceNode) int {
				count := 0
				for range ns.Entities() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Enums early break",
			Namespace(types.Path("Test"), Enum("Enum1", "a"), Enum("Enum2", "b"), Enum("Enum3", "c")),
			func(ns NamespaceNode) int {
				count := 0
				for range ns.Enums() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Actions early break",
			Namespace(types.Path("Test"), Action("Action1"), Action("Action2"), Action("Action3")),
			func(ns NamespaceNode) int {
				count := 0
				for range ns.Actions() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.countFunc(tt.ns)
			if got != 1 {
				t.Errorf("expected to iterate 1 time before break, got %d", got)
			}
		})
	}
}

// Helper functions
func countNamespaceIter[T any](iter func(func(T) bool)) int {
	count := 0
	iter(func(T) bool {
		count++
		return true
	})
	return count
}
