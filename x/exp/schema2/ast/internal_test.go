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
	(NamespaceNode{}).isNode()
	(CommonTypeNode{}).isNode()
	(EntityNode{}).isNode()
	(EnumNode{}).isNode()
	(ActionNode{}).isNode()
}

// TestIsDeclarationMarkerMethods tests that all IsDeclaration marker methods are callable for coverage
func TestIsDeclarationMarkerMethods(t *testing.T) {
	t.Parallel()

	// Call all isDeclaration() marker methods for coverage
	(CommonTypeNode{}).isDeclaration()
	(EntityNode{}).isDeclaration()
	(EnumNode{}).isDeclaration()
	(ActionNode{}).isDeclaration()
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
		var commonTypes []CommonTypeNode
		for ct := range ns.commonTypes() {
			commonTypes = append(commonTypes, ct)
		}
		if len(commonTypes) != 2 {
			t.Errorf("expected 2 common types, got %d", len(commonTypes))
		}
		if commonTypes[0].Name != ct1.Name {
			t.Errorf("expected first common type to be ct1")
		}
		if commonTypes[1].Name != ct2.Name {
			t.Errorf("expected second common type to be ct2")
		}
	})

	// Test Entities iterator
	t.Run("Entities", func(t *testing.T) {
		var entities []EntityNode
		for e := range ns.entities() {
			entities = append(entities, e)
		}
		if len(entities) != 2 {
			t.Errorf("expected 2 entities, got %d", len(entities))
		}
		if entities[0].Name != e1.Name {
			t.Errorf("expected first entity to be e1")
		}
		if entities[1].Name != e2.Name {
			t.Errorf("expected second entity to be e2")
		}
	})

	// Test Enums iterator
	t.Run("Enums", func(t *testing.T) {
		var enums []EnumNode
		for e := range ns.enums() {
			enums = append(enums, e)
		}
		if len(enums) != 2 {
			t.Errorf("expected 2 enums, got %d", len(enums))
		}
		if enums[0].Name != enum1.Name {
			t.Errorf("expected first enum to be enum1")
		}
		if enums[1].Name != enum2.Name {
			t.Errorf("expected second enum to be enum2")
		}
	})

	// Test Actions iterator
	t.Run("Actions", func(t *testing.T) {
		var actions []ActionNode
		for a := range ns.actions() {
			actions = append(actions, a)
		}
		if len(actions) != 2 {
			t.Errorf("expected 2 actions, got %d", len(actions))
		}
		if actions[0].Name != a1.Name {
			t.Errorf("expected first action to be a1")
		}
		if actions[1].Name != a2.Name {
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
		for range ns.commonTypes() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 common types, got %d", count)
		}
	})

	t.Run("Entities", func(t *testing.T) {
		count := 0
		for range ns.entities() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 entities, got %d", count)
		}
	})

	t.Run("Enums", func(t *testing.T) {
		count := 0
		for range ns.enums() {
			count++
		}
		if count != 0 {
			t.Errorf("expected 0 enums, got %d", count)
		}
	})

	t.Run("Actions", func(t *testing.T) {
		count := 0
		for range ns.actions() {
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
		for ct := range ns.commonTypes() {
			count++
			if ct.Name == ct1.Name {
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
		for e := range ns2.entities() {
			count++
			if e.Name == e1.Name {
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
		for e := range ns3.enums() {
			count++
			if e.Name == enum1.Name {
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
		for a := range ns4.actions() {
			count++
			if a.Name == a1.Name {
				break
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})
}

// TestCommonTypeFullName tests the FullName method on CommonTypeNode
func TestCommonTypeFullName(t *testing.T) {
	t.Parallel()

	ct := CommonType("MyType", StringType{})

	t.Run("FullName with nil namespace", func(t *testing.T) {
		fullName := ct.FullName(nil)
		if fullName != "MyType" {
			t.Errorf("expected 'MyType', got '%s'", fullName)
		}
	})

	t.Run("FullName with namespace", func(t *testing.T) {
		ns := Namespace("MyApp")
		fullName := ct.FullName(&ns)
		if fullName != "MyApp::MyType" {
			t.Errorf("expected 'MyApp::MyType', got '%s'", fullName)
		}
	})

	t.Run("FullName with nested namespace", func(t *testing.T) {
		ns := Namespace("MyApp::Types")
		fullName := ct.FullName(&ns)
		if fullName != "MyApp::Types::MyType" {
			t.Errorf("expected 'MyApp::Types::MyType', got '%s'", fullName)
		}
	})
}

// TestEntityEntityType tests the EntityType method on EntityNode
func TestEntityEntityType(t *testing.T) {
	t.Parallel()

	t.Run("EntityType without namespace", func(t *testing.T) {
		entity := Entity("User")
		schema := NewSchema(entity)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if _, found := resolved.Entities["User"]; !found {
			t.Error("expected 'User' entity in resolved schema")
		}
	})

	t.Run("EntityType with namespace", func(t *testing.T) {
		entity := Entity("User")
		ns := Namespace("MyApp", entity)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if _, found := resolved.Entities["MyApp::User"]; !found {
			t.Error("expected 'MyApp::User' entity in resolved schema")
		}
	})

	t.Run("EntityType with nested namespace", func(t *testing.T) {
		entity := Entity("User")
		ns := Namespace("MyApp::Models", entity)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if _, found := resolved.Entities["MyApp::Models::User"]; !found {
			t.Error("expected 'MyApp::Models::User' entity in resolved schema")
		}
	})
}

// TestEnumEntityUIDs tests the EntityUIDs iterator on EnumNode
func TestEnumEntityUIDs(t *testing.T) {
	t.Parallel()

	t.Run("EntityUIDs without namespace", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")
		schema := NewSchema(enum)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		resolvedEnum := resolved.Enums["Status"]

		var uids []types.EntityUID
		for uid := range resolvedEnum.EntityUIDs() {
			uids = append(uids, uid)
		}
		if len(uids) != 3 {
			t.Errorf("expected 3 UIDs, got %d", len(uids))
		}
		if uids[0].Type != "Status" {
			t.Errorf("expected type 'Status', got '%s'", uids[0].Type)
		}
		if uids[0].ID != "active" {
			t.Errorf("expected id 'active', got '%s'", uids[0].ID)
		}
		if uids[1].ID != "inactive" {
			t.Errorf("expected id 'inactive', got '%s'", uids[1].ID)
		}
		if uids[2].ID != "pending" {
			t.Errorf("expected id 'pending', got '%s'", uids[2].ID)
		}
	})

	t.Run("EntityUIDs with namespace", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")
		ns := Namespace("MyApp", enum)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		resolvedEnum := resolved.Enums["MyApp::Status"]

		var uids []types.EntityUID
		for uid := range resolvedEnum.EntityUIDs() {
			uids = append(uids, uid)
		}
		if len(uids) != 3 {
			t.Errorf("expected 3 UIDs, got %d", len(uids))
		}
		if uids[0].Type != "MyApp::Status" {
			t.Errorf("expected type 'MyApp::Status', got '%s'", uids[0].Type)
		}
		if uids[0].ID != "active" {
			t.Errorf("expected id 'active', got '%s'", uids[0].ID)
		}
	})

	t.Run("EntityUIDs with nested namespace", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")
		ns := Namespace("MyApp::Models", enum)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		resolvedEnum := resolved.Enums["MyApp::Models::Status"]

		var uids []types.EntityUID
		for uid := range resolvedEnum.EntityUIDs() {
			uids = append(uids, uid)
		}
		if len(uids) != 3 {
			t.Errorf("expected 3 UIDs, got %d", len(uids))
		}
		if uids[0].Type != "MyApp::Models::Status" {
			t.Errorf("expected type 'MyApp::Models::Status', got '%s'", uids[0].Type)
		}
	})

	t.Run("EntityUIDs early break", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")
		schema := NewSchema(enum)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		resolvedEnum := resolved.Enums["Status"]

		count := 0
		for range resolvedEnum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}
		if count != 1 {
			t.Errorf("expected to iterate 1 time before break, got %d", count)
		}
	})
}

// TestActionEntityUID tests the EntityUID method on ActionNode
func TestActionEntityUID(t *testing.T) {
	t.Parallel()

	t.Run("EntityUID without namespace", func(t *testing.T) {
		a := Action("view")
		schema := NewSchema(a)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		uid := types.NewEntityUID("Action", "view")
		if _, found := resolved.Actions[uid]; !found {
			t.Error("expected Action::view in resolved schema")
		}
	})

	t.Run("EntityUID with namespace", func(t *testing.T) {
		a := Action("view")
		ns := Namespace("Bananas", a)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		uid := types.NewEntityUID("Bananas::Action", "view")
		if _, found := resolved.Actions[uid]; !found {
			t.Error("expected Bananas::Action::view in resolved schema")
		}
	})

	t.Run("EntityUID with nested namespace", func(t *testing.T) {
		a := Action("view")
		ns := Namespace("MyApp::Resources", a)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		uid := types.NewEntityUID("MyApp::Resources::Action", "view")
		if _, found := resolved.Actions[uid]; !found {
			t.Error("expected MyApp::Resources::Action::view in resolved schema")
		}
	})
}
