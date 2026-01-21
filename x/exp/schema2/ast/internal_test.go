package ast

import (
	"fmt"
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

func TestEntityTypeResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupSchema   func() *Schema
		wantErr       bool
		checkResolved func(*testing.T, *ResolvedSchema)
	}{
		{
			"EntityType without namespace",
			func() *Schema {
				return NewSchema(Entity("User"))
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Entities["User"]; !found {
					t.Error("expected 'User' entity in resolved schema")
				}
			},
		},
		{
			"EntityType with namespace",
			func() *Schema {
				return NewSchema(Namespace("MyApp", Entity("User")))
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Entities["MyApp::User"]; !found {
					t.Error("expected 'MyApp::User' entity in resolved schema")
				}
			},
		},
		{
			"EntityType with nested namespace",
			func() *Schema {
				return NewSchema(Namespace("MyApp::Models", Entity("User")))
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Entities["MyApp::Models::User"]; !found {
					t.Error("expected 'MyApp::Models::User' entity in resolved schema")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkResolved != nil {
				tt.checkResolved(t, resolved)
			}
		})
	}
}

func TestEnumEntityUIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		enumKey     types.EntityType
		wantCount   int
		wantType    types.EntityType
		wantIDs     []types.String
	}{
		{
			"EntityUIDs without namespace",
			func() *Schema {
				return NewSchema(Enum("Status", "active", "inactive", "pending"))
			},
			types.EntityType("Status"),
			3,
			"Status",
			[]types.String{"active", "inactive", "pending"},
		},
		{
			"EntityUIDs with namespace",
			func() *Schema {
				return NewSchema(Namespace("MyApp", Enum("Status", "active", "inactive", "pending")))
			},
			types.EntityType("MyApp::Status"),
			3,
			"MyApp::Status",
			[]types.String{"active", "inactive", "pending"},
		},
		{
			"EntityUIDs with nested namespace",
			func() *Schema {
				return NewSchema(Namespace("MyApp::Models", Enum("Status", "active", "inactive", "pending")))
			},
			types.EntityType("MyApp::Models::Status"),
			3,
			"MyApp::Models::Status",
			[]types.String{"active", "inactive", "pending"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			resolvedEnum := resolved.Enums[tt.enumKey]

			var uids []types.EntityUID
			for uid := range resolvedEnum.EntityUIDs() {
				uids = append(uids, uid)
			}

			if len(uids) != tt.wantCount {
				t.Errorf("expected %d UIDs, got %d", tt.wantCount, len(uids))
			}
			if len(uids) > 0 && uids[0].Type != tt.wantType {
				t.Errorf("expected type '%s', got '%s'", tt.wantType, uids[0].Type)
			}
			for i, wantID := range tt.wantIDs {
				if i < len(uids) && uids[i].ID != wantID {
					t.Errorf("expected id '%s' at position %d, got '%s'", wantID, i, uids[i].ID)
				}
			}
		})
	}

	t.Run("EntityUIDs early break", func(t *testing.T) {
		schema := NewSchema(Enum("Status", "active", "inactive", "pending"))
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

func TestActionEntityUID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		actionUID   types.EntityUID
	}{
		{
			"EntityUID without namespace",
			func() *Schema { return NewSchema(Action("view")) },
			types.NewEntityUID("Action", "view"),
		},
		{
			"EntityUID with namespace",
			func() *Schema { return NewSchema(Namespace("Bananas", Action("view"))) },
			types.NewEntityUID("Bananas::Action", "view"),
		},
		{
			"EntityUID with nested namespace",
			func() *Schema { return NewSchema(Namespace("MyApp::Resources", Action("view"))) },
			types.NewEntityUID("MyApp::Resources::Action", "view"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if _, found := resolved.Actions[tt.actionUID]; !found {
				t.Errorf("expected action %v in resolved schema", tt.actionUID)
			}
		})
	}
}

func TestTypeResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{"resolve Long type", func() *Schema { return NewSchema(Entity("User").Shape(Attribute("age", Long()))) }, false},
		{"resolve Bool type", func() *Schema { return NewSchema(Entity("User").Shape(Attribute("active", Bool()))) }, false},
		{
			"resolve ExtensionType",
			func() *Schema {
				return NewSchema(Entity("User").Shape(
					Attribute("ip", IPAddr()),
					Attribute("amount", Decimal()),
					Attribute("created", Datetime()),
					Attribute("timeout", Duration()),
				))
			},
			false,
		},
		{"resolve SetType", func() *Schema { return NewSchema(Entity("User").Shape(Attribute("tags", Set(String())))) }, false},
		{
			"resolve nested RecordType",
			func() *Schema {
				return NewSchema(Entity("User").Shape(
					Attribute("address", Record(
						Attribute("street", String()),
						Attribute("city", String()),
					)),
				))
			},
			false,
		},
		{"resolve entity with tags", func() *Schema { return NewSchema(Entity("Document").Tags(Record(Attribute("classification", String())))) }, false},
		{
			"resolve entity with MemberOf",
			func() *Schema {
				return NewSchema(Entity("Group"), Entity("User").MemberOf(Ref("Group")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTopLevelCommonTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
		checkCount  func(*testing.T, *ResolvedSchema)
	}{
		{
			"top-level common type",
			func() *Schema {
				ct := CommonType("Address", Record(
					Attribute("street", String()),
					Attribute("city", String()),
				))
				entity := Entity("User").Shape(Attribute("address", Type("Address")))
				return NewSchema(ct, entity)
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if len(rs.Entities) != 1 {
					t.Errorf("expected 1 entity, got %d", len(rs.Entities))
				}
			},
		},
		{
			"mixed top-level and namespaced common types",
			func() *Schema {
				topLevelCT := CommonType("TopLevelType", String())
				namespaceCT := CommonType("NamespaceType", Long())
				topLevelEntity := Entity("TopEntity").Shape(Attribute("field", Type("TopLevelType")))
				namespaceEntity := Entity("NsEntity").Shape(Attribute("field", Type("NamespaceType")))
				ns := Namespace("App", namespaceCT, namespaceEntity)
				return NewSchema(topLevelCT, topLevelEntity, ns)
			},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkCount != nil {
				tt.checkCount(t, resolved)
			}
		})
	}
}

func TestActionWithAppliesTo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"action with principal",
			func() *Schema {
				return NewSchema(Entity("User"), Action("view").Principal(Ref("User")))
			},
			false,
		},
		{
			"action with resource",
			func() *Schema {
				return NewSchema(Entity("Document"), Action("view").Resource(Ref("Document")))
			},
			false,
		},
		{
			"action with context",
			func() *Schema {
				return NewSchema(Action("view").Context(Record(Attribute("ip", String()))))
			},
			false,
		},
		{
			"action with MemberOf",
			func() *Schema {
				return NewSchema(Action("parent"), Action("child").MemberOf(UID("parent")))
			},
			false,
		},
		{
			"action in namespace with empty name",
			func() *Schema {
				return NewSchema(Namespace("", Action("view")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNamespaceResolutionErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"empty namespace resolves without errors",
			func() *Schema { return NewSchema(Namespace("Empty")) },
			false,
		},
		{
			"entity with invalid type reference in shape",
			func() *Schema { return NewSchema(Entity("User").Shape(Attribute("field", Type("NonExistent")))) },
			true,
		},
		{
			"entity with invalid type reference in tags",
			func() *Schema { return NewSchema(Entity("User").Tags(Type("NonExistent"))) },
			true,
		},
		{
			"action with invalid context type",
			func() *Schema { return NewSchema(Namespace("App", Action("view").Context(Type("NonExistent")))) },
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLazyTypeResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"lazy resolution in namespace",
			func() *Schema {
				ct := CommonType("MyType", String())
				entity := Entity("User").Shape(Attribute("field", Type("MyType")))
				return NewSchema(Namespace("App", ct, entity))
			},
			false,
		},
		{
			"lazy resolution cached on second use",
			func() *Schema {
				ct := CommonType("MyType", String())
				entity1 := Entity("User").Shape(Attribute("field1", Type("MyType")))
				entity2 := Entity("Admin").Shape(Attribute("field2", Type("MyType")))
				return NewSchema(Namespace("App", ct, entity1, entity2))
			},
			false,
		},
		{
			"qualified type reference across namespaces",
			func() *Schema {
				ct := CommonType("SharedType", String())
				ns1 := Namespace("App1", ct)
				entity := Entity("User").Shape(Attribute("field", Type("App1::SharedType")))
				ns2 := Namespace("App2", entity)
				return NewSchema(ns1, ns2)
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEntityTypeRefResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"qualified entity type ref",
			func() *Schema {
				return NewSchema(Namespace("App", Entity("Group"), Entity("User").MemberOf(Ref("App::Group"))))
			},
			false,
		},
		{
			"entity type ref without namespace",
			func() *Schema {
				return NewSchema(Entity("Group"), Entity("User").MemberOf(Ref("Group")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithNamespace(t *testing.T) {
	t.Parallel()

	t.Run("withNamespace returns same rd when namespace matches", func(t *testing.T) {
		ct := CommonType("Type1", String())
		ns := Namespace("App", ct)
		schema := NewSchema(ns)
		rd := newResolveData(schema, &ns)
		rd2 := rd.withNamespace(&ns)
		if rd != rd2 {
			t.Error("expected withNamespace to return same resolveData when namespace matches")
		}
	})
}

func TestAnnotations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
		checkFunc   func(*testing.T, *ResolvedSchema)
	}{
		{
			"namespace with annotations",
			func() *Schema {
				return NewSchema(Namespace("App").Annotate("key1", "value1").Annotate("key2", "value2"))
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				annotations := rs.Namespaces["App"]
				if len(annotations) != 2 {
					t.Errorf("expected 2 annotations, got %d", len(annotations))
				}
			},
		},
		{"common type with annotations", func() *Schema { return NewSchema(CommonType("MyType", String()).Annotate("doc", "documentation")) }, false, nil},
		{"entity with annotations", func() *Schema { return NewSchema(Entity("User").Annotate("doc", "User entity")) }, false, nil},
		{"enum with annotations", func() *Schema { return NewSchema(Enum("Status", "active").Annotate("doc", "Status enum")) }, false, nil},
		{"action with annotations", func() *Schema { return NewSchema(Action("view").Annotate("doc", "View action")) }, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, resolved)
			}
		})
	}
}

func TestErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{"top-level action with invalid type reference", func() *Schema { return NewSchema(Action("view").Context(Type("NonExistent"))) }, true},
		{"top-level common type with invalid nested type", func() *Schema { return NewSchema(CommonType("MyType", Type("NonExistent"))) }, true},
		{"common type with error in nested record", func() *Schema { return NewSchema(CommonType("MyType", Record(Attribute("field", Type("NonExistent"))))) }, true},
		{"common type with error in set type", func() *Schema { return NewSchema(CommonType("MyType", Set(Type("NonExistent")))) }, true},
		{"namespace common type with error", func() *Schema { return NewSchema(Namespace("App", CommonType("MyType", Type("NonExistent")))) }, true},
		{"namespace entity with error", func() *Schema { return NewSchema(Namespace("App", Entity("User").Shape(Attribute("field", Type("NonExistent"))))) }, true},
		{"namespace action with error", func() *Schema { return NewSchema(Namespace("App", Action("view").Context(Type("NonExistent")))) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTypeRefResolutionEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"type ref that resolves to cached already-resolved type",
			func() *Schema {
				ct := CommonType("CachedType", String())
				entity1 := Entity("User").Shape(Attribute("field1", Type("CachedType")))
				entity2 := Entity("Admin").Shape(Attribute("field2", Type("CachedType")))
				return NewSchema(ct, entity1, entity2)
			},
			false,
		},
		{
			"type ref in namespace that resolves to namespace type cached",
			func() *Schema {
				ct := CommonType("NsType", String())
				entity1 := Entity("User").Shape(Attribute("field1", Type("NsType")))
				entity2 := Entity("Admin").Shape(Attribute("field2", Type("NsType")))
				return NewSchema(Namespace("App", ct, entity1, entity2))
			},
			false,
		},
		{
			"type ref with error during lazy resolution in namespace",
			func() *Schema {
				ct := CommonType("MyType", Type("NonExistent"))
				entity := Entity("User").Shape(Attribute("field", Type("MyType")))
				return NewSchema(Namespace("App", ct, entity))
			},
			true,
		},
		{
			"type ref with error during lazy resolution at schema level",
			func() *Schema {
				ct := CommonType("MyType", Type("NonExistent"))
				entity := Entity("User").Shape(Attribute("field", Type("App::MyType")))
				ns1 := Namespace("App", ct)
				ns2 := Namespace("Other", entity)
				return NewSchema(ns1, ns2)
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestComplexResolutionScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"action with all features in namespace",
			func() *Schema {
				user := Entity("User")
				doc := Entity("Document")
				parent := Action("parent")
				child := Action("child").
					MemberOf(UID("parent")).
					Principal(Ref("User")).
					Resource(Ref("Document")).
					Context(Record(Attribute("requestTime", String())))
				return NewSchema(Namespace("App", user, doc, parent, child))
			},
			false,
		},
		{
			"entity with all features in namespace",
			func() *Schema {
				group := Entity("Group")
				user := Entity("User").
					MemberOf(Ref("Group")).
					Shape(Attribute("name", String())).
					Tags(Record(Attribute("department", String())))
				return NewSchema(Namespace("App", group, user))
			},
			false,
		},
		{
			"cross-namespace type reference",
			func() *Schema {
				ct := CommonType("SharedType", String())
				ns1 := Namespace("Shared", ct)
				entity := Entity("User").Shape(Attribute("field", Type("Shared::SharedType")))
				ns2 := Namespace("App", entity)
				return NewSchema(ns1, ns2)
			},
			false,
		},
		{
			"deeply nested type references",
			func() *Schema {
				ct1 := CommonType("Type1", String())
				ct2 := CommonType("Type2", Type("Type1"))
				ct3 := CommonType("Type3", Type("Type2"))
				entity := Entity("User").Shape(Attribute("field", Type("Type3")))
				return NewSchema(ct1, ct2, ct3, entity)
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// errorType is a test-only type that always returns an error during resolution
type errorType struct{}

func (errorType) isType() {}

func (errorType) resolve(rd *resolveData) (IsType, error) {
	return nil, fmt.Errorf("intentional test error")
}

func TestUnreachableErrorPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"common type with error in nested type",
			func() *Schema {
				ct := CommonTypeNode{Name: "ErrorType", Type: errorType{}}
				return NewSchema(Namespace("App", ct))
			},
			true,
		},
		{
			"entity shape with error in nested type",
			func() *Schema {
				entity := EntityNode{
					Name:     "User",
					ShapeVal: &RecordType{Pairs: []Pair{{Key: "field", Type: errorType{}}}},
				}
				return NewSchema(entity)
			},
			true,
		},
		{
			"entity tags with error type",
			func() *Schema {
				entity := EntityNode{Name: "User", TagsVal: errorType{}}
				return NewSchema(entity)
			},
			true,
		},
		{
			"action context with error type",
			func() *Schema {
				action := ActionNode{Name: "view", AppliesToVal: &AppliesTo{Context: errorType{}}}
				return NewSchema(action)
			},
			true,
		},
		{
			"namespace type reference with error during lazy resolution",
			func() *Schema {
				entity := Entity("User").Shape(Attribute("field", Type("BadType")))
				ct := CommonTypeNode{Name: "BadType", Type: errorType{}}
				return NewSchema(Namespace("App", entity, ct))
			},
			true,
		},
		{
			"schema-wide type reference with error during lazy resolution",
			func() *Schema {
				entity := Entity("User").Shape(Attribute("field", Type("Zoo::BadType")))
				ct := CommonTypeNode{Name: "BadType", Type: errorType{}}
				ns1 := Namespace("App", entity)
				ns2 := Namespace("Zoo", ct)
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"EntityTypeRef.resolve() called through IsType interface",
			func() *Schema {
				return NewSchema(Namespace("App", Entity("User").Shape(Attribute("groups", Set(Ref("Group")))), Entity("Group")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGlobalEntityResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
		checkFunc   func(*testing.T, *ResolvedSchema)
	}{
		{
			"action referencing global enum",
			func() *Schema {
				globalEnum := Enum("Status", "active", "inactive")
				localEntity := Entity("User")
				action := Action("view").Principal(Ref("User")).Resource(Ref("Status"))
				ns := Namespace("App", localEntity, action)
				return NewSchema(globalEnum, ns)
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Enums["Status"]; !found {
					t.Error("expected global enum 'Status' to be present")
				}
			},
		},
		{
			"entity referencing already qualified name",
			func() *Schema {
				baseEntity := Entity("Base")
				ns1 := Namespace("Core", baseEntity)
				derivedEntity := Entity("Derived").MemberOf(Ref("Core::Base"))
				ns2 := Namespace("App", derivedEntity)
				return NewSchema(ns1, ns2)
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Entities["Core::Base"]; !found {
					t.Error("expected 'Core::Base' entity")
				}
				if _, found := rs.Entities["App::Derived"]; !found {
					t.Error("expected 'App::Derived' entity")
				}
			},
		},
		{
			"entity exists check with non-entity nodes",
			func() *Schema {
				commonType := CommonType("MyType", String())
				globalEntity := Entity("GlobalEntity")
				localAction := Action("view").Resource(Ref("GlobalEntity"))
				ns := Namespace("App", localAction)
				return NewSchema(commonType, globalEntity, ns)
			},
			false,
			func(t *testing.T, rs *ResolvedSchema) {
				if _, found := rs.Entities["GlobalEntity"]; !found {
					t.Error("expected global entity 'GlobalEntity' to be present")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, resolved)
			}
		})
	}

	t.Run("entity exists check with nil schema", func(t *testing.T) {
		rd := &resolveData{
			schema:               nil,
			namespace:            &NamespaceNode{Name: "Test"},
			schemaCommonTypes:    make(map[string]*commonTypeEntry),
			namespaceCommonTypes: make(map[string]*commonTypeEntry),
		}
		exists := rd.entityExistsInEmptyNamespace("SomeEntity")
		if exists {
			t.Error("expected false for nil schema")
		}
	})
}

func TestNamingConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *Schema
		wantErr     bool
	}{
		{
			"entity conflict - nested namespace vs qualified name",
			func() *Schema {
				ns1 := Namespace("Goat::Gorilla", Entity("Cows"))
				ns2 := Namespace("Goat", Entity("Gorilla::Cows"))
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"enum conflict - nested namespace vs qualified name",
			func() *Schema {
				ns1 := Namespace("Goat::Gorilla", Enum("Status", "active"))
				ns2 := Namespace("Goat", Enum("Gorilla::Status", "active"))
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"entity vs enum conflict - entity first in namespace",
			func() *Schema {
				ns1 := Namespace("App", Entity("Thing"))
				ns2 := Namespace("App", Enum("Thing", "value"))
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"entity vs enum conflict - enum first in namespace",
			func() *Schema {
				ns1 := Namespace("App", Enum("Thing", "value"))
				ns2 := Namespace("App", Entity("Thing"))
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{"top-level entity vs enum conflict - enum first", func() *Schema { return NewSchema(Enum("Status", "active"), Entity("Status")) }, true},
		{"top-level enum vs entity conflict - entity first", func() *Schema { return NewSchema(Entity("Status"), Enum("Status", "active")) }, true},
		{
			"top-level vs namespaced entity conflict",
			func() *Schema {
				topLevel := Entity("A::B")
				ns := Namespace("A", Entity("B"))
				return NewSchema(topLevel, ns)
			},
			true,
		},
		{
			"action conflict - nested namespace vs qualified name",
			func() *Schema {
				ns1 := Namespace("Goat::Gorilla", Action("view"))
				ns2 := Namespace("Goat::Gorilla", Action("view"))
				return NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"no conflict - different namespaces",
			func() *Schema {
				ns1 := Namespace("Goat", Entity("Cows"))
				ns2 := Namespace("Sheep", Entity("Cows"))
				return NewSchema(ns1, ns2)
			},
			false,
		},
		{"top-level duplicate entity", func() *Schema { return NewSchema(Entity("User"), Entity("User")) }, true},
		{"top-level duplicate enum", func() *Schema { return NewSchema(Enum("Status", "active"), Enum("Status", "inactive")) }, true},
		{"top-level duplicate action", func() *Schema { return NewSchema(Action("view"), Action("view")) }, true},
		{"namespace duplicate entity", func() *Schema { return NewSchema(Namespace("App", Entity("User"), Entity("User"))) }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := schema.Resolve()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
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
