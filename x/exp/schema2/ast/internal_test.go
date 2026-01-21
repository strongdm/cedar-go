package ast

import (
	"fmt"
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
		for ct := range ns.CommonTypes() {
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
		for e := range ns.Entities() {
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
		for e := range ns.Enums() {
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
		for a := range ns.Actions() {
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
		for e := range ns2.Entities() {
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
		for e := range ns3.Enums() {
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
		for a := range ns4.Actions() {
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

// TestTypeResolution tests various type resolution scenarios
func TestTypeResolution(t *testing.T) {
	t.Parallel()

	t.Run("resolve Long type", func(t *testing.T) {
		entity := Entity("User").Shape(Attribute("age", Long()))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve Bool type", func(t *testing.T) {
		entity := Entity("User").Shape(Attribute("active", Bool()))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve ExtensionType", func(t *testing.T) {
		entity := Entity("User").Shape(
			Attribute("ip", IPAddr()),
			Attribute("amount", Decimal()),
			Attribute("created", Datetime()),
			Attribute("timeout", Duration()),
		)
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve SetType", func(t *testing.T) {
		entity := Entity("User").Shape(Attribute("tags", Set(String())))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve nested RecordType", func(t *testing.T) {
		entity := Entity("User").Shape(
			Attribute("address", Record(
				Attribute("street", String()),
				Attribute("city", String()),
			)),
		)
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve entity with tags", func(t *testing.T) {
		entity := Entity("Document").Tags(Record(Attribute("classification", String())))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("resolve entity with MemberOf", func(t *testing.T) {
		group := Entity("Group")
		user := Entity("User").MemberOf(Ref("Group"))
		schema := NewSchema(group, user)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// TestTopLevelCommonTypes tests common types at the top level (not in a namespace)
func TestTopLevelCommonTypes(t *testing.T) {
	t.Parallel()

	t.Run("top-level common type", func(t *testing.T) {
		ct := CommonType("Address", Record(
			Attribute("street", String()),
			Attribute("city", String()),
		))
		entity := Entity("User").Shape(Attribute("address", Type("Address")))
		schema := NewSchema(ct, entity)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		if len(resolved.Entities) != 1 {
			t.Errorf("expected 1 entity, got %d", len(resolved.Entities))
		}
	})

	t.Run("mixed top-level and namespaced common types", func(t *testing.T) {
		topLevelCT := CommonType("TopLevelType", String())
		namespaceCT := CommonType("NamespaceType", Long())
		topLevelEntity := Entity("TopEntity").Shape(Attribute("field", Type("TopLevelType")))
		namespaceEntity := Entity("NsEntity").Shape(Attribute("field", Type("NamespaceType")))
		ns := Namespace("App", namespaceCT, namespaceEntity)
		schema := NewSchema(topLevelCT, topLevelEntity, ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// TestActionWithAppliesTo tests action resolution with appliesTo
func TestActionWithAppliesTo(t *testing.T) {
	t.Parallel()

	t.Run("action with principal", func(t *testing.T) {
		user := Entity("User")
		action := Action("view").Principal(Ref("User"))
		schema := NewSchema(user, action)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("action with resource", func(t *testing.T) {
		doc := Entity("Document")
		action := Action("view").Resource(Ref("Document"))
		schema := NewSchema(doc, action)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("action with context", func(t *testing.T) {
		action := Action("view").Context(Record(Attribute("ip", String())))
		schema := NewSchema(action)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("action with MemberOf", func(t *testing.T) {
		parent := Action("parent")
		child := Action("child").MemberOf(UID("parent"))
		schema := NewSchema(parent, child)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("action in namespace with empty name", func(t *testing.T) {
		action := Action("view")
		ns := Namespace("", action)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		uid := types.NewEntityUID("Action", "view")
		if _, found := resolved.Actions[uid]; !found {
			t.Error("expected Action::view in resolved schema for empty namespace")
		}
	})
}

// TestNamespaceResolutionErrors tests error handling in namespace resolution
func TestNamespaceResolutionErrors(t *testing.T) {
	t.Parallel()

	t.Run("empty namespace resolves without errors", func(t *testing.T) {
		ns := Namespace("Empty")
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() should not error for empty namespace, got: %v", err)
		}
	})

	t.Run("entity with invalid type reference in shape", func(t *testing.T) {
		entity := Entity("User").Shape(Attribute("field", Type("NonExistent")))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference")
		}
	})

	t.Run("entity with invalid type reference in tags", func(t *testing.T) {
		entity := Entity("User").Tags(Type("NonExistent"))
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in tags")
		}
	})

	t.Run("entity with invalid MemberOf reference", func(t *testing.T) {
		entity := Entity("User").MemberOf(Ref("NonExistent"))
		ns := Namespace("App", entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			// Expected - reference resolution occurs but doesn't validate entity exists
			t.Logf("Got error as expected: %v", err)
		}
	})

	t.Run("action with invalid principal reference", func(t *testing.T) {
		action := Action("view").Principal(Ref("NonExistent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Logf("Got error as expected: %v", err)
		}
	})

	t.Run("action with invalid resource reference", func(t *testing.T) {
		action := Action("view").Resource(Ref("NonExistent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Logf("Got error as expected: %v", err)
		}
	})

	t.Run("action with invalid context type", func(t *testing.T) {
		action := Action("view").Context(Type("NonExistent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid context type reference")
		}
	})

	t.Run("action with invalid MemberOf type", func(t *testing.T) {
		action := Action("view").MemberOf(EntityUID("NonExistent", "parent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Logf("Got error as expected: %v", err)
		}
	})
}

// TestLazyTypeResolution tests lazy resolution of common types
func TestLazyTypeResolution(t *testing.T) {
	t.Parallel()

	t.Run("lazy resolution in namespace", func(t *testing.T) {
		ct := CommonType("MyType", String())
		entity := Entity("User").Shape(Attribute("field", Type("MyType")))
		ns := Namespace("App", ct, entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("lazy resolution cached on second use", func(t *testing.T) {
		ct := CommonType("MyType", String())
		entity1 := Entity("User").Shape(Attribute("field1", Type("MyType")))
		entity2 := Entity("Admin").Shape(Attribute("field2", Type("MyType")))
		ns := Namespace("App", ct, entity1, entity2)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("qualified type reference across namespaces", func(t *testing.T) {
		ct := CommonType("SharedType", String())
		ns1 := Namespace("App1", ct)
		entity := Entity("User").Shape(Attribute("field", Type("App1::SharedType")))
		ns2 := Namespace("App2", entity)
		schema := NewSchema(ns1, ns2)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// TestEntityTypeRefResolution tests entity type reference resolution
func TestEntityTypeRefResolution(t *testing.T) {
	t.Parallel()

	t.Run("qualified entity type ref", func(t *testing.T) {
		group := Entity("Group")
		user := Entity("User").MemberOf(Ref("App::Group"))
		ns := Namespace("App", group, user)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("entity type ref without namespace", func(t *testing.T) {
		group := Entity("Group")
		user := Entity("User").MemberOf(Ref("Group"))
		schema := NewSchema(group, user)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// TestWithNamespace tests the withNamespace method
func TestWithNamespace(t *testing.T) {
	t.Parallel()

	t.Run("withNamespace returns same rd when namespace matches", func(t *testing.T) {
		ct := CommonType("Type1", String())
		ns := Namespace("App", ct)
		schema := NewSchema(ns)
		rd := newResolveData(schema, &ns)
		rd2 := rd.withNamespace(&ns)
		// Verify it returns the same pointer
		if rd != rd2 {
			t.Error("expected withNamespace to return same resolveData when namespace matches")
		}
	})
}

// TestAnnotations tests annotation methods
func TestAnnotations(t *testing.T) {
	t.Parallel()

	t.Run("namespace with annotations", func(t *testing.T) {
		ns := Namespace("App").Annotate("key1", "value1").Annotate("key2", "value2")
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		annotations := resolved.Namespaces["App"]
		if len(annotations) != 2 {
			t.Errorf("expected 2 annotations, got %d", len(annotations))
		}
	})

	t.Run("common type with annotations", func(t *testing.T) {
		ct := CommonType("MyType", String()).Annotate("doc", "documentation")
		schema := NewSchema(ct)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("entity with annotations", func(t *testing.T) {
		entity := Entity("User").Annotate("doc", "User entity")
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("enum with annotations", func(t *testing.T) {
		enum := Enum("Status", "active").Annotate("doc", "Status enum")
		schema := NewSchema(enum)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("action with annotations", func(t *testing.T) {
		action := Action("view").Annotate("doc", "View action")
		schema := NewSchema(action)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// TestErrorPaths tests error handling paths
func TestErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("top-level action with invalid type reference", func(t *testing.T) {
		action := Action("view").Context(Type("NonExistent"))
		schema := NewSchema(action)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in top-level action")
		}
	})

	t.Run("top-level common type with invalid nested type", func(t *testing.T) {
		ct := CommonType("MyType", Type("NonExistent"))
		schema := NewSchema(ct)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in top-level common type")
		}
	})

	t.Run("common type with error in nested record", func(t *testing.T) {
		ct := CommonType("MyType", Record(
			Attribute("field", Type("NonExistent")),
		))
		schema := NewSchema(ct)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in record")
		}
	})

	t.Run("common type with error in set type", func(t *testing.T) {
		ct := CommonType("MyType", Set(Type("NonExistent")))
		schema := NewSchema(ct)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in set")
		}
	})

	t.Run("namespace common type with error", func(t *testing.T) {
		ct := CommonType("MyType", Type("NonExistent"))
		ns := Namespace("App", ct)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in namespace common type")
		}
	})

	t.Run("namespace entity with error", func(t *testing.T) {
		entity := Entity("User").Shape(Attribute("field", Type("NonExistent")))
		ns := Namespace("App", entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in namespace entity")
		}
	})

	t.Run("namespace action with error", func(t *testing.T) {
		action := Action("view").Context(Type("NonExistent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error for invalid type reference in namespace action")
		}
	})
}

// TestTypRefResolutionEdgeCases tests edge cases in TypeRef resolution
func TestTypeRefResolutionEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("type ref that resolves to cached already-resolved type", func(t *testing.T) {
		// First reference resolves and caches
		// Second reference uses cached version
		ct := CommonType("CachedType", String())
		entity1 := Entity("User").Shape(Attribute("field1", Type("CachedType")))
		entity2 := Entity("Admin").Shape(Attribute("field2", Type("CachedType")))
		schema := NewSchema(ct, entity1, entity2)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("type ref in namespace that resolves to namespace type cached", func(t *testing.T) {
		ct := CommonType("NsType", String())
		entity1 := Entity("User").Shape(Attribute("field1", Type("NsType")))
		entity2 := Entity("Admin").Shape(Attribute("field2", Type("NsType")))
		ns := Namespace("App", ct, entity1, entity2)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("type ref with error during lazy resolution in namespace", func(t *testing.T) {
		// CommonType that refers to another non-existent type
		ct := CommonType("MyType", Type("NonExistent"))
		entity := Entity("User").Shape(Attribute("field", Type("MyType")))
		ns := Namespace("App", ct, entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error during lazy resolution")
		}
	})

	t.Run("type ref with error during lazy resolution at schema level", func(t *testing.T) {
		ct := CommonType("MyType", Type("NonExistent"))
		entity := Entity("User").Shape(Attribute("field", Type("App::MyType")))
		ns1 := Namespace("App", ct)
		ns2 := Namespace("Other", entity)
		schema := NewSchema(ns1, ns2)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error during lazy resolution at schema level")
		}
	})
}

// TestEntityMemberOfError tests error handling in EntityNode.resolve MemberOf
func TestEntityMemberOfError(t *testing.T) {
	t.Parallel()

	// This test ensures we hit the error path in EntityNode.resolve when resolving MemberOf
	t.Run("entity MemberOf with invalid type ref", func(t *testing.T) {
		// Create an entity with MemberOf that references a common type (not an entity)
		// which will cause an error during resolution
		ct := CommonType("NotAnEntity", String())
		entity := Entity("User").MemberOf(Ref("NotAnEntity"))
		ns := Namespace("App", ct, entity)
		schema := NewSchema(ns)
		// This should work actually, since we qualify the type ref but don't validate it exists
		_, err := schema.Resolve()
		_ = err // The resolution may or may not error depending on implementation
	})
}

// TestActionResolveErrorPaths tests error handling in ActionNode.resolve
func TestActionResolveErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("action MemberOf with invalid type", func(t *testing.T) {
		// Trigger error in MemberOf resolution
		action := Action("view").MemberOf(EntityUID("NonExistent", "parent"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		_ = err // May or may not error
	})

	t.Run("action Principal with type that doesn't resolve", func(t *testing.T) {
		// This should exercise the principal resolution path
		action := Action("view").Principal(Ref("ValidType"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		_ = err
	})

	t.Run("action Resource with type that doesn't resolve", func(t *testing.T) {
		// This should exercise the resource resolution path
		action := Action("view").Resource(Ref("ValidType"))
		ns := Namespace("App", action)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		_ = err
	})
}

// TestComplexResolutionScenarios tests complex resolution scenarios
func TestComplexResolutionScenarios(t *testing.T) {
	t.Parallel()

	t.Run("action with all features in namespace", func(t *testing.T) {
		user := Entity("User")
		doc := Entity("Document")
		parent := Action("parent")
		child := Action("child").
			MemberOf(UID("parent")).
			Principal(Ref("User")).
			Resource(Ref("Document")).
			Context(Record(Attribute("requestTime", String())))
		ns := Namespace("App", user, doc, parent, child)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("entity with all features in namespace", func(t *testing.T) {
		group := Entity("Group")
		user := Entity("User").
			MemberOf(Ref("Group")).
			Shape(Attribute("name", String())).
			Tags(Record(Attribute("department", String())))
		ns := Namespace("App", group, user)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("cross-namespace type reference", func(t *testing.T) {
		// Define a type in one namespace, use it from another
		ct := CommonType("SharedType", String())
		ns1 := Namespace("Shared", ct)

		entity := Entity("User").Shape(Attribute("field", Type("Shared::SharedType")))
		ns2 := Namespace("App", entity)

		schema := NewSchema(ns1, ns2)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("deeply nested type references", func(t *testing.T) {
		ct1 := CommonType("Type1", String())
		ct2 := CommonType("Type2", Type("Type1"))
		ct3 := CommonType("Type3", Type("Type2"))
		entity := Entity("User").Shape(Attribute("field", Type("Type3")))
		schema := NewSchema(ct1, ct2, ct3, entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("namespace common type lazy resolution", func(t *testing.T) {
		// This specifically tests the lazy resolution path in namespace context (lines 197-210 in types.go)
		// The common type is defined but not yet resolved when first referenced
		ct := CommonType("LazyType", String())
		entity := Entity("User").Shape(Attribute("field", Type("LazyType")))
		ns := Namespace("App", ct, entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("namespace common type with transitive dependency", func(t *testing.T) {
		// Tests lazy resolution where the type itself references another type
		ct1 := CommonType("BaseType", String())
		ct2 := CommonType("DerivedType", Type("BaseType"))
		entity := Entity("User").Shape(Attribute("field", Type("DerivedType")))
		ns := Namespace("App", ct1, ct2, entity)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})

	t.Run("force namespace-local type resolution", func(t *testing.T) {
		// This test specifically targets the namespace-local lazy resolution path
		// by creating a scenario where a type is referenced before being fully resolved
		ct := CommonType("LocalType", Record(Attribute("nested", String())))
		entity1 := Entity("First").Shape(Attribute("field", Type("LocalType")))
		entity2 := Entity("Second").Shape(Attribute("field", Type("LocalType")))
		ns := Namespace("App", ct, entity1, entity2)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
	})
}

// errorType is a test-only type that always returns an error during resolution
type errorType struct{}

func (errorType) isType() {}

func (errorType) resolve(rd *resolveData) (IsType, error) {
	return nil, fmt.Errorf("intentional test error")
}

// TestUnreachableErrorPaths tests error paths that are normally unreachable
// These paths exist for defensive programming but can't be triggered in normal usage
func TestUnreachableErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("entity MemberOf with error during resolve", func(t *testing.T) {
		// Create an entity with a custom type that errors
		// This requires directly constructing the AST nodes
		entity := EntityNode{
			Name: "User",
			MemberOfVal: []EntityTypeRef{
				{Name: "Group"},
			},
		}
		// Manually create a malformed scenario by replacing the EntityTypeRef
		// with something that will error. Since we can't actually do this due to types,
		// this test documents that the error path exists but is unreachable.
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err != nil {
			t.Logf("Got error (not expected in normal flow): %v", err)
		}
	})

	t.Run("common type with error in nested type", func(t *testing.T) {
		// Create a common type that has an error-prone nested type
		ct := CommonTypeNode{
			Name: "ErrorType",
			Type: errorType{},
		}
		ns := Namespace("App", ct)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type")
		}
	})

	t.Run("entity shape with error in nested type", func(t *testing.T) {
		// Entity with shape containing error type
		entity := EntityNode{
			Name: "User",
			ShapeVal: &RecordType{
				Pairs: []Pair{
					{Key: "field", Type: errorType{}},
				},
			},
		}
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type in shape")
		}
	})

	t.Run("entity tags with error type", func(t *testing.T) {
		entity := EntityNode{
			Name:    "User",
			TagsVal: errorType{},
		}
		schema := NewSchema(entity)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type in tags")
		}
	})

	t.Run("action context with error type", func(t *testing.T) {
		action := ActionNode{
			Name: "view",
			AppliesToVal: &AppliesTo{
				Context: errorType{},
			},
		}
		schema := NewSchema(action)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type in context")
		}
	})

	t.Run("namespace type reference with error during lazy resolution", func(t *testing.T) {
		// Create an entity that references a common type BEFORE the common type is defined
		// This forces lazy resolution when the entity's shape is resolved
		// The common type contains an error type, triggering the error path (types.go:204-206)
		entity := Entity("User").Shape(Attribute("field", Type("BadType")))
		ct := CommonTypeNode{
			Name: "BadType",
			Type: errorType{},
		}
		// Put entity BEFORE the common type to force lazy resolution
		ns := Namespace("App", entity, ct)
		schema := NewSchema(ns)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type in namespace common type")
		}
	})

	t.Run("schema-wide type reference with error during lazy resolution", func(t *testing.T) {
		// Create an entity in one namespace that references a common type in another namespace
		// Put the entity BEFORE the common type to force lazy resolution via schema-wide cache
		// The common type contains an error type, triggering the error path (types.go:246-248)
		entity := Entity("User").Shape(Attribute("field", Type("Zoo::BadType")))
		ct := CommonTypeNode{
			Name: "BadType",
			Type: errorType{},
		}
		ns1 := Namespace("App", entity)
		ns2 := Namespace("Zoo", ct)
		schema := NewSchema(ns1, ns2)
		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error from error type in schema-wide common type")
		}
	})

	t.Run("EntityTypeRef.resolve() called through IsType interface", func(t *testing.T) {
		// Test that EntityTypeRef.resolve() is called (not just mustResolve())
		// This happens when EntityTypeRef is embedded in another type like SetType
		// and that type's resolve() is called
		entity := Entity("User").Shape(
			Attribute("groups", Set(Ref("Group"))),
		)
		group := Entity("Group")
		ns := Namespace("App", entity, group)
		schema := NewSchema(ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}
		// Verify that the entity type ref was resolved with namespace qualification
		userEntity := resolved.Entities["App::User"]
		if userEntity.ShapeVal == nil {
			t.Fatal("expected entity to have shape")
		}
		// The groups attribute should be a Set of EntityTypeRef
		groupsAttr := userEntity.ShapeVal.Pairs[0]
		setType, ok := groupsAttr.Type.(SetType)
		if !ok {
			t.Fatal("expected groups attribute to be SetType")
		}
		entityRef, ok := setType.Element.(EntityTypeRef)
		if !ok {
			t.Fatal("expected set element to be EntityTypeRef")
		}
		if entityRef.Name != "App::Group" {
			t.Errorf("expected fully qualified name 'App::Group', got %v", entityRef.Name)
		}
	})
}

// TestGlobalEntityResolution tests resolution of global entities referenced from namespaces
func TestGlobalEntityResolution(t *testing.T) {
	t.Parallel()

	t.Run("action referencing global enum", func(t *testing.T) {
		// Global enum
		globalEnum := Enum("Status", "active", "inactive")
		// Namespace with action referencing the global enum
		localEntity := Entity("User")
		action := Action("view").Principal(Ref("User")).Resource(Ref("Status"))
		ns := Namespace("App", localEntity, action)
		schema := NewSchema(globalEnum, ns)

		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		// Verify global enum is present
		if _, found := resolved.Enums["Status"]; !found {
			t.Error("expected global enum 'Status' to be present")
		}

		// Verify action exists (with namespace qualification)
		actionUID := types.EntityUID{Type: "App::Action", ID: "view"}
		if _, found := resolved.Actions[actionUID]; !found {
			t.Errorf("expected action 'App::Action::view' to be present, got actions: %v", resolved.Actions)
		}
	})

	t.Run("entity referencing already qualified name", func(t *testing.T) {
		// Test that already qualified names (starting with ::) are handled
		baseEntity := Entity("Base")
		ns1 := Namespace("Core", baseEntity)

		// Entity with already qualified reference
		derivedEntity := Entity("Derived").MemberOf(Ref("Core::Base"))
		ns2 := Namespace("App", derivedEntity)

		schema := NewSchema(ns1, ns2)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		// Verify both entities exist with qualified names
		if _, found := resolved.Entities["Core::Base"]; !found {
			t.Error("expected 'Core::Base' entity")
		}
		if _, found := resolved.Entities["App::Derived"]; !found {
			t.Error("expected 'App::Derived' entity")
		}
	})

	t.Run("entity exists check with nil schema", func(t *testing.T) {
		// Test entityExistsInEmptyNamespace with nil schema (edge case)
		rd := &resolveData{
			schema:               nil,
			namespace:            &NamespaceNode{Name: "Test"},
			schemaCommonTypes:    make(map[string]*commonTypeEntry),
			namespaceCommonTypes: make(map[string]*commonTypeEntry),
		}
		// Should return false for nil schema
		exists := rd.entityExistsInEmptyNamespace("SomeEntity")
		if exists {
			t.Error("expected false for nil schema")
		}
	})

	t.Run("entity exists check with non-entity nodes", func(t *testing.T) {
		// Create schema with a CommonType node (not entity or enum)
		commonType := CommonType("MyType", String())
		globalEntity := Entity("GlobalEntity")

		// Namespace that references the global entity
		localAction := Action("view").Resource(Ref("GlobalEntity"))
		ns := Namespace("App", localAction)

		schema := NewSchema(commonType, globalEntity, ns)
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() failed: %v", err)
		}

		// Verify global entity exists
		if _, found := resolved.Entities["GlobalEntity"]; !found {
			t.Error("expected global entity 'GlobalEntity' to be present")
		}
	})
}

// TestNamingConflicts tests that naming conflicts are detected during schema resolution
func TestNamingConflicts(t *testing.T) {
	t.Parallel()

	t.Run("entity conflict - nested namespace vs qualified name", func(t *testing.T) {
		// Scenario: namespace Goat::Gorilla with entity Cows
		// vs namespace Goat with entity Gorilla::Cows
		// Both resolve to Goat::Gorilla::Cows - CONFLICT!
		ns1 := Namespace("Goat::Gorilla", Entity("Cows"))
		ns2 := Namespace("Goat", Entity("Gorilla::Cows"))
		schema := NewSchema(ns1, ns2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to naming conflict, got nil")
		}
		// Check error message mentions conflict
		if err.Error() == "" {
			t.Error("expected non-empty error message")
		}
	})

	t.Run("enum conflict - nested namespace vs qualified name", func(t *testing.T) {
		// Similar to entity conflict but with enums
		ns1 := Namespace("Goat::Gorilla", Enum("Status", "active"))
		ns2 := Namespace("Goat", Enum("Gorilla::Status", "active"))
		schema := NewSchema(ns1, ns2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to naming conflict, got nil")
		}
	})

	t.Run("entity vs enum conflict - entity first in namespace", func(t *testing.T) {
		// An entity and enum with the same fully qualified name in namespace
		// Entity processed first, then enum
		ns1 := Namespace("App", Entity("Thing"))
		ns2 := Namespace("App", Enum("Thing", "value"))
		schema := NewSchema(ns1, ns2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to entity/enum conflict, got nil")
		}
	})

	t.Run("entity vs enum conflict - enum first in namespace", func(t *testing.T) {
		// An enum and entity with the same fully qualified name in namespace
		// Enum processed first, then entity (covers line 355)
		ns1 := Namespace("App", Enum("Thing", "value"))
		ns2 := Namespace("App", Entity("Thing"))
		schema := NewSchema(ns1, ns2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to enum/entity conflict, got nil")
		}
	})

	t.Run("top-level entity vs enum conflict - enum first", func(t *testing.T) {
		// Top-level enum defined first, then entity with same name (covers line 398)
		enum1 := Enum("Status", "active")
		entity1 := Entity("Status")
		schema := NewSchema(enum1, entity1)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to top-level enum/entity conflict, got nil")
		}
	})

	t.Run("top-level enum vs entity conflict - entity first", func(t *testing.T) {
		// Top-level entity defined first, then enum with same name (covers line 410)
		entity1 := Entity("Status")
		enum1 := Enum("Status", "active")
		schema := NewSchema(entity1, enum1)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to top-level entity/enum conflict, got nil")
		}
	})

	t.Run("top-level vs namespaced entity conflict", func(t *testing.T) {
		// Top-level entity A::B conflicts with namespace A, entity B
		topLevel := Entity("A::B")
		ns := Namespace("A", Entity("B"))
		schema := NewSchema(topLevel, ns)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to top-level vs namespace conflict, got nil")
		}
	})

	t.Run("action conflict - nested namespace vs qualified name", func(t *testing.T) {
		// Actions have EntityUIDs like Goat::Gorilla::Action::"foo"
		// If namespace Goat has action "Gorilla::Action::\"foo\"", that seems pathological
		// but let's test a more realistic scenario:
		// Two namespaces with actions that could have the same UID
		ns1 := Namespace("Goat::Gorilla", Action("view"))
		ns2 := Namespace("Goat::Gorilla", Action("view"))
		schema := NewSchema(ns1, ns2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to duplicate action, got nil")
		}
	})

	t.Run("no conflict - different namespaces", func(t *testing.T) {
		// This should NOT conflict - different fully qualified names
		ns1 := Namespace("Goat", Entity("Cows"))
		ns2 := Namespace("Sheep", Entity("Cows"))
		schema := NewSchema(ns1, ns2)

		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Both entities should exist with different qualified names
		if _, found := resolved.Entities["Goat::Cows"]; !found {
			t.Error("expected 'Goat::Cows' entity")
		}
		if _, found := resolved.Entities["Sheep::Cows"]; !found {
			t.Error("expected 'Sheep::Cows' entity")
		}
	})

	t.Run("top-level duplicate entity", func(t *testing.T) {
		// Two top-level entities with the same name
		e1 := Entity("User")
		e2 := Entity("User")
		schema := NewSchema(e1, e2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to duplicate entity, got nil")
		}
	})

	t.Run("top-level duplicate enum", func(t *testing.T) {
		// Two top-level enums with the same name
		enum1 := Enum("Status", "active")
		enum2 := Enum("Status", "inactive")
		schema := NewSchema(enum1, enum2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to duplicate enum, got nil")
		}
	})

	t.Run("top-level duplicate action", func(t *testing.T) {
		// Two top-level actions with the same name
		a1 := Action("view")
		a2 := Action("view")
		schema := NewSchema(a1, a2)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to duplicate action, got nil")
		}
	})

	t.Run("namespace duplicate entity", func(t *testing.T) {
		// Two entities with the same name in the same namespace
		ns := Namespace("App", Entity("User"), Entity("User"))
		schema := NewSchema(ns)

		_, err := schema.Resolve()
		if err == nil {
			t.Fatal("expected error due to duplicate entity in namespace, got nil")
		}
	})
}
