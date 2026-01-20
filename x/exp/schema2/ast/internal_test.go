package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/types"
)

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

// TestNamespaceIterators tests the iterator methods on NamespaceNode
func TestNamespaceIterators(t *testing.T) {
	t.Parallel()

	// Create test declarations
	ct1 := CommonType("MyType1", StringType{})
	ct2 := CommonType("MyType2", LongType{})
	e1 := Entity("User")
	e2 := Entity("Group")
	enum1 := Enum("Status", "active", "inactive")
	enum2 := Enum("Role", "admin", "user")
	a1 := Action("read")
	a2 := Action("write")

	// Create namespace with mixed declarations
	ns := Namespace(
		types.Path("MyApp"),
		ct1, e1, enum1, a1, ct2, e2, enum2, a2,
	)

	// Test CommonTypes iterator
	t.Run("CommonTypes", func(t *testing.T) {
		var commonTypes []*CommonTypeNode
		for ct := range ns.CommonTypes() {
			commonTypes = append(commonTypes, ct)
		}
		if len(commonTypes) != 2 {
			t.Errorf("expected 2 common types, got %d", len(commonTypes))
		}
		if commonTypes[0] != ct1 {
			t.Errorf("expected first common type to be ct1")
		}
		if commonTypes[1] != ct2 {
			t.Errorf("expected second common type to be ct2")
		}
	})

	// Test Entities iterator
	t.Run("Entities", func(t *testing.T) {
		var entities []*EntityNode
		for e := range ns.Entities() {
			entities = append(entities, e)
		}
		if len(entities) != 2 {
			t.Errorf("expected 2 entities, got %d", len(entities))
		}
		if entities[0] != e1 {
			t.Errorf("expected first entity to be e1")
		}
		if entities[1] != e2 {
			t.Errorf("expected second entity to be e2")
		}
	})

	// Test Enums iterator
	t.Run("Enums", func(t *testing.T) {
		var enums []*EnumNode
		for e := range ns.Enums() {
			enums = append(enums, e)
		}
		if len(enums) != 2 {
			t.Errorf("expected 2 enums, got %d", len(enums))
		}
		if enums[0] != enum1 {
			t.Errorf("expected first enum to be enum1")
		}
		if enums[1] != enum2 {
			t.Errorf("expected second enum to be enum2")
		}
	})

	// Test Actions iterator
	t.Run("Actions", func(t *testing.T) {
		var actions []*ActionNode
		for a := range ns.Actions() {
			actions = append(actions, a)
		}
		if len(actions) != 2 {
			t.Errorf("expected 2 actions, got %d", len(actions))
		}
		if actions[0] != a1 {
			t.Errorf("expected first action to be a1")
		}
		if actions[1] != a2 {
			t.Errorf("expected second action to be a2")
		}
	})
}

// TestNamespaceIteratorsEmpty tests iterator methods with empty declarations
func TestNamespaceIteratorsEmpty(t *testing.T) {
	t.Parallel()

	ns := Namespace(types.Path("Empty"))

	// Test that all iterators work with no declarations
	t.Run("CommonTypes", func(t *testing.T) {
		count := 0
		for range ns.CommonTypes() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 common types, got %d", count)
		}
	})

	t.Run("Entities", func(t *testing.T) {
		count := 0
		for range ns.Entities() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 entities, got %d", count)
		}
	})

	t.Run("Enums", func(t *testing.T) {
		count := 0
		for range ns.Enums() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 enums, got %d", count)
		}
	})

	t.Run("Actions", func(t *testing.T) {
		count := 0
		for range ns.Actions() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 actions, got %d", count)
		}
	})
}

// TestNamespaceIteratorsEarlyBreak tests that iterators support early termination
func TestNamespaceIteratorsEarlyBreak(t *testing.T) {
	t.Parallel()

	ct1 := CommonType("Type1", StringType{})
	ct2 := CommonType("Type2", LongType{})
	ct3 := CommonType("Type3", BoolType{})

	ns := Namespace(types.Path("Test"), ct1, ct2, ct3)

	t.Run("CommonTypes early break", func(t *testing.T) {
		count := 0
		for ct := range ns.CommonTypes() {
			count++
			if ct == ct1 {
				break // Early termination after first item
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})

	e1 := Entity("Entity1")
	e2 := Entity("Entity2")
	e3 := Entity("Entity3")
	ns2 := Namespace(types.Path("Test"), e1, e2, e3)

	t.Run("Entities early break", func(t *testing.T) {
		count := 0
		for e := range ns2.Entities() {
			count++
			if e == e1 {
				break
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})

	enum1 := Enum("Enum1", "a")
	enum2 := Enum("Enum2", "b")
	enum3 := Enum("Enum3", "c")
	ns3 := Namespace(types.Path("Test"), enum1, enum2, enum3)

	t.Run("Enums early break", func(t *testing.T) {
		count := 0
		for e := range ns3.Enums() {
			count++
			if e == enum1 {
				break
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})

	a1 := Action("Action1")
	a2 := Action("Action2")
	a3 := Action("Action3")
	ns4 := Namespace(types.Path("Test"), a1, a2, a3)

	t.Run("Actions early break", func(t *testing.T) {
		count := 0
		for a := range ns4.Actions() {
			count++
			if a == a1 {
				break
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})
}
