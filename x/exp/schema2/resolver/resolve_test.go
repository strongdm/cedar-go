package resolver_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/resolver"
)

// TestEnumNodeEntityUIDsEarlyTermination tests early termination of EntityUIDs iterator
func TestEnumNodeEntityUIDsEarlyTermination(t *testing.T) {
	t.Parallel()

	t.Run("EnumNode.EntityUIDs early break", func(t *testing.T) {
		enum := ast.Enum("Status", "active", "inactive", "pending")

		count := 0
		for range enum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}

		if count != 1 {
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})

	t.Run("ResolvedEnum.EntityUIDs early break", func(t *testing.T) {
		schema := ast.NewSchema(ast.Enum("Status", "active", "inactive", "pending"))
		resolved, err := resolver.Resolve(schema)
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
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})
}

func TestEntityTypeResolution(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupSchema   func() *ast.Schema
		wantErr       bool
		checkResolved func(*testing.T, *resolver.ResolvedSchema)
	}{
		{
			"EntityType without namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("User"))
			},
			false,
			func(t *testing.T, rs *resolver.ResolvedSchema) {
				if _, found := rs.Entities["User"]; !found {
					t.Error("expected 'User' entity in resolved schema")
				}
			},
		},
		{
			"EntityType with namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("MyApp", ast.Entity("User")))
			},
			false,
			func(t *testing.T, rs *resolver.ResolvedSchema) {
				if _, found := rs.Entities["MyApp::User"]; !found {
					t.Error("expected 'MyApp::User' entity in resolved schema")
				}
			},
		},
		{
			"EntityType with nested namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("MyApp::Models", ast.Entity("User")))
			},
			false,
			func(t *testing.T, rs *resolver.ResolvedSchema) {
				if _, found := rs.Entities["MyApp::Models::User"]; !found {
					t.Error("expected 'MyApp::Models::User' entity in resolved schema")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		enumKey     types.EntityType
		wantCount   int
		wantType    types.EntityType
		wantIDs     []types.String
	}{
		{
			"EntityUIDs without namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Enum("Status", "active", "inactive", "pending"))
			},
			types.EntityType("Status"),
			3,
			"Status",
			[]types.String{"active", "inactive", "pending"},
		},
		{
			"EntityUIDs with namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("MyApp", ast.Enum("Status", "active", "inactive", "pending")))
			},
			types.EntityType("MyApp::Status"),
			3,
			"MyApp::Status",
			[]types.String{"active", "inactive", "pending"},
		},
		{
			"EntityUIDs with nested namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("MyApp::Models", ast.Enum("Status", "active", "inactive", "pending")))
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
			resolved, err := resolver.Resolve(schema)
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
		schema := ast.NewSchema(ast.Enum("Status", "active", "inactive", "pending"))
		resolved, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		actionUID   types.EntityUID
	}{
		{
			"EntityUID without namespace",
			func() *ast.Schema { return ast.NewSchema(ast.Action("view")) },
			types.NewEntityUID("Action", "view"),
		},
		{
			"EntityUID with namespace",
			func() *ast.Schema { return ast.NewSchema(ast.Namespace("Bananas", ast.Action("view"))) },
			types.NewEntityUID("Bananas::Action", "view"),
		},
		{
			"EntityUID with nested namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("MyApp::Resources", ast.Action("view")))
			},
			types.NewEntityUID("MyApp::Resources::Action", "view"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{"resolve Long type", func() *ast.Schema {
			return ast.NewSchema(ast.Entity("User").Shape(ast.Attribute("age", ast.Long())))
		}, false},
		{"resolve Bool type", func() *ast.Schema {
			return ast.NewSchema(ast.Entity("User").Shape(ast.Attribute("active", ast.Bool())))
		}, false},
		{
			"resolve ExtensionType",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("User").Shape(
					ast.Attribute("ip", ast.IPAddr()),
					ast.Attribute("amount", ast.Decimal()),
					ast.Attribute("created", ast.Datetime()),
					ast.Attribute("timeout", ast.Duration()),
				))
			},
			false,
		},
		{"resolve SetType", func() *ast.Schema {
			return ast.NewSchema(ast.Entity("User").Shape(ast.Attribute("tags", ast.Set(ast.String()))))
		}, false},
		{
			"resolve nested RecordType",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("User").Shape(
					ast.Attribute("address", ast.Record(
						ast.Attribute("street", ast.String()),
						ast.Attribute("city", ast.String()),
					)),
				))
			},
			false,
		},
		{"resolve entity with tags", func() *ast.Schema {
			return ast.NewSchema(ast.Entity("Document").Tags(ast.Record(ast.Attribute("classification", ast.String()))))
		}, false},
		{
			"resolve entity with MemberOf",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("Group"), ast.Entity("User").MemberOf(ast.Ref("Group")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
		checkCount  func(*testing.T, *resolver.ResolvedSchema)
	}{
		{
			"top-level common type",
			func() *ast.Schema {
				ct := ast.CommonType("Address", ast.Record(
					ast.Attribute("street", ast.String()),
					ast.Attribute("city", ast.String()),
				))
				entity := ast.Entity("User").Shape(ast.Attribute("address", ast.Type("Address")))
				return ast.NewSchema(ct, entity)
			},
			false,
			func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 1 {
					t.Errorf("expected 1 entity, got %d", len(rs.Entities))
				}
			},
		},
		{
			"mixed top-level and namespaced common types",
			func() *ast.Schema {
				topLevelCT := ast.CommonType("TopLevelType", ast.String())
				namespaceCT := ast.CommonType("NamespaceType", ast.Long())
				topLevelEntity := ast.Entity("TopEntity").Shape(ast.Attribute("field", ast.Type("TopLevelType")))
				namespaceEntity := ast.Entity("NsEntity").Shape(ast.Attribute("field", ast.Type("NamespaceType")))
				ns := ast.Namespace("App", namespaceCT, namespaceEntity)
				return ast.NewSchema(topLevelCT, topLevelEntity, ns)
			},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"action with principal",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("User"), ast.Action("view").Principal(ast.Ref("User")))
			},
			false,
		},
		{
			"action with resource",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("Document"), ast.Action("view").Resource(ast.Ref("Document")))
			},
			false,
		},
		{
			"action with context",
			func() *ast.Schema {
				return ast.NewSchema(ast.Action("view").Context(ast.Record(ast.Attribute("ip", ast.String()))))
			},
			false,
		},
		{
			"action with MemberOf",
			func() *ast.Schema {
				return ast.NewSchema(ast.Action("parent"), ast.Action("child").MemberOf(ast.UID("parent")))
			},
			false,
		},
		{
			"action in namespace with empty name",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("", ast.Action("view")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"empty namespace resolves without errors",
			func() *ast.Schema { return ast.NewSchema(ast.Namespace("Empty")) },
			false,
		},
		{
			"entity with invalid type reference in shape",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("User").Shape(ast.Attribute("field", ast.Type("NonExistent"))))
			},
			true,
		},
		{
			"entity with invalid type reference in tags",
			func() *ast.Schema { return ast.NewSchema(ast.Entity("User").Tags(ast.Type("NonExistent"))) },
			true,
		},
		{
			"action with invalid context type",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("App", ast.Action("view").Context(ast.Type("NonExistent"))))
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"lazy resolution in namespace",
			func() *ast.Schema {
				ct := ast.CommonType("MyType", ast.String())
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("MyType")))
				return ast.NewSchema(ast.Namespace("App", ct, entity))
			},
			false,
		},
		{
			"lazy resolution cached on second use",
			func() *ast.Schema {
				ct := ast.CommonType("MyType", ast.String())
				entity1 := ast.Entity("User").Shape(ast.Attribute("field1", ast.Type("MyType")))
				entity2 := ast.Entity("Admin").Shape(ast.Attribute("field2", ast.Type("MyType")))
				return ast.NewSchema(ast.Namespace("App", ct, entity1, entity2))
			},
			false,
		},
		{
			"qualified type reference across namespaces",
			func() *ast.Schema {
				ct := ast.CommonType("SharedType", ast.String())
				ns1 := ast.Namespace("App1", ct)
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("App1::SharedType")))
				ns2 := ast.Namespace("App2", entity)
				return ast.NewSchema(ns1, ns2)
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"qualified entity type ref",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("App", ast.Entity("Group"), ast.Entity("User").MemberOf(ast.Ref("App::Group"))))
			},
			false,
		},
		{
			"entity type ref without namespace",
			func() *ast.Schema {
				return ast.NewSchema(ast.Entity("Group"), ast.Entity("User").MemberOf(ast.Ref("Group")))
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// func TestWithNamespace(t *testing.T) {
// 	t.Parallel()

// 	t.Run("withNamespace returns same rd when namespace matches", func(t *testing.T) {
// 		ct := ast.CommonType("Type1", ast.String())
// 		ns := ast.Namespace("App", ct)
// 		schema := ast.NewSchema(ns)
// 		rd := newResolveData(schema, &ns)
// 		rd2 := rd.withNamespace(&ns)
// 		if rd != rd2 {
// 			t.Error("expected withNamespace to return same resolveData when namespace matches")
// 		}
// 	})
// }

func TestAnnotations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *ast.Schema
		wantErr     bool
		checkFunc   func(*testing.T, *resolver.ResolvedSchema)
	}{
		{
			"namespace with annotations",
			func() *ast.Schema {
				return ast.NewSchema(ast.Namespace("App").Annotate("key1", "value1").Annotate("key2", "value2"))
			},
			false,
			func(t *testing.T, rs *resolver.ResolvedSchema) {
				ns := rs.Namespaces["App"]
				if len(ns.Annotations) != 2 {
					t.Errorf("expected 2 annotations, got %d", len(ns.Annotations))
				}
			},
		},
		{"common type with annotations", func() *ast.Schema {
			return ast.NewSchema(ast.CommonType("MyType", ast.String()).Annotate("doc", "documentation"))
		}, false, nil},
		{"entity with annotations", func() *ast.Schema {
			return ast.NewSchema(ast.Entity("User").Annotate("doc", "User entity"))
		}, false, nil},
		{"enum with annotations", func() *ast.Schema {
			return ast.NewSchema(ast.Enum("Status", "active").Annotate("doc", "Status enum"))
		}, false, nil},
		{"action with annotations", func() *ast.Schema { return ast.NewSchema(ast.Action("view").Annotate("doc", "View action")) }, false, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			resolved, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{"top-level action with invalid type reference", func() *ast.Schema { return ast.NewSchema(ast.Action("view").Context(ast.Type("NonExistent"))) }, true},
		{"top-level common type with invalid nested type", func() *ast.Schema {
			return ast.NewSchema(ast.CommonType("MyType", ast.Type("NonExistent")))
		}, true},
		{"common type with error in nested record", func() *ast.Schema {
			return ast.NewSchema(ast.CommonType("MyType", ast.Record(ast.Attribute("field", ast.Type("NonExistent")))))
		}, true},
		{"common type with error in set type", func() *ast.Schema {
			return ast.NewSchema(ast.CommonType("MyType", ast.Set(ast.Type("NonExistent"))))
		}, true},
		{"namespace common type with error", func() *ast.Schema {
			return ast.NewSchema(ast.Namespace("App", ast.CommonType("MyType", ast.Type("NonExistent"))))
		}, true},
		{"namespace entity with error", func() *ast.Schema {
			return ast.NewSchema(ast.Namespace("App", ast.Entity("User").Shape(ast.Attribute("field", ast.Type("NonExistent")))))
		}, true},
		{"namespace action with error", func() *ast.Schema {
			return ast.NewSchema(ast.Namespace("App", ast.Action("view").Context(ast.Type("NonExistent"))))
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"type ref that resolves to cached already-resolved type",
			func() *ast.Schema {
				ct := ast.CommonType("CachedType", ast.String())
				entity1 := ast.Entity("User").Shape(ast.Attribute("field1", ast.Type("CachedType")))
				entity2 := ast.Entity("Admin").Shape(ast.Attribute("field2", ast.Type("CachedType")))
				return ast.NewSchema(ct, entity1, entity2)
			},
			false,
		},
		{
			"type ref in namespace that resolves to namespace type cached",
			func() *ast.Schema {
				ct := ast.CommonType("NsType", ast.String())
				entity1 := ast.Entity("User").Shape(ast.Attribute("field1", ast.Type("NsType")))
				entity2 := ast.Entity("Admin").Shape(ast.Attribute("field2", ast.Type("NsType")))
				return ast.NewSchema(ast.Namespace("App", ct, entity1, entity2))
			},
			false,
		},
		{
			"type ref with error during lazy resolution in namespace",
			func() *ast.Schema {
				ct := ast.CommonType("MyType", ast.Type("NonExistent"))
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("MyType")))
				return ast.NewSchema(ast.Namespace("App", ct, entity))
			},
			true,
		},
		{
			"type ref with error during lazy resolution at schema level",
			func() *ast.Schema {
				ct := ast.CommonType("MyType", ast.Type("NonExistent"))
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("App::MyType")))
				ns1 := ast.Namespace("App", ct)
				ns2 := ast.Namespace("Other", entity)
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
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
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"action with all features in namespace",
			func() *ast.Schema {
				user := ast.Entity("User")
				doc := ast.Entity("Document")
				parent := ast.Action("parent")
				child := ast.Action("child").
					MemberOf(ast.UID("parent")).
					Principal(ast.Ref("User")).
					Resource(ast.Ref("Document")).
					Context(ast.Record(ast.Attribute("requestTime", ast.String())))
				return ast.NewSchema(ast.Namespace("App", user, doc, parent, child))
			},
			false,
		},
		{
			"entity with all features in namespace",
			func() *ast.Schema {
				group := ast.Entity("Group")
				user := ast.Entity("User").
					MemberOf(ast.Ref("Group")).
					Shape(ast.Attribute("name", ast.String())).
					Tags(ast.Record(ast.Attribute("department", ast.String())))
				return ast.NewSchema(ast.Namespace("App", group, user))
			},
			false,
		},
		{
			"cross-namespace type reference",
			func() *ast.Schema {
				ct := ast.CommonType("SharedType", ast.String())
				ns1 := ast.Namespace("Shared", ct)
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("Shared::SharedType")))
				ns2 := ast.Namespace("App", entity)
				return ast.NewSchema(ns1, ns2)
			},
			false,
		},
		{
			"deeply nested type references",
			func() *ast.Schema {
				ct1 := ast.CommonType("Type1", ast.String())
				ct2 := ast.CommonType("Type2", ast.Type("Type1"))
				ct3 := ast.CommonType("Type3", ast.Type("Type2"))
				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("Type3")))
				return ast.NewSchema(ct1, ct2, ct3, entity)
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// // errorType is a test-only type that always returns an error during resolution
// type errorType struct{}

// func (errorType) isType() {}

// func (errorType) resolve(rd *resolveData) (IsType, error) {
// 	return nil, fmt.Errorf("intentional test error")
// }

// func TestUnreachableErrorPaths(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name        string
// 		setupSchema func() *ast.Schema
// 		wantErr     bool
// 	}{
// 		{
// 			"common type with error in nested type",
// 			func() *ast.Schema {
// 				ct := ast.CommonTypeNode{Name: "ErrorType", Type: errorType{}}
// 				return ast.NewSchema(ast.Namespace("App", ct))
// 			},
// 			true,
// 		},
// 		{
// 			"entity shape with error in nested type",
// 			func() *ast.Schema {
// 				entity := ast.EntityNode{
// 					Name:     "User",
// 					ShapeVal: &ast.RecordType{Pairs: []ast.Pair{{Key: "field", Type: errorType{}}}},
// 				}
// 				return ast.NewSchema(entity)
// 			},
// 			true,
// 		},
// 		{
// 			"entity tags with error type",
// 			func() *ast.Schema {
// 				entity := ast.EntityNode{Name: "User", TagsVal: errorType{}}
// 				return ast.NewSchema(entity)
// 			},
// 			true,
// 		},
// 		{
// 			"action context with error type",
// 			func() *ast.Schema {
// 				action := ast.ActionNode{Name: "view", AppliesToVal: &ast.AppliesTo{Context: errorType{}}}
// 				return ast.NewSchema(action)
// 			},
// 			true,
// 		},
// 		{
// 			"namespace type reference with error during lazy resolution",
// 			func() *ast.Schema {
// 				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("BadType")))
// 				ct := ast.CommonTypeNode{Name: "BadType", Type: errorType{}}
// 				return ast.NewSchema(ast.Namespace("App", entity, ct))
// 			},
// 			true,
// 		},
// 		{
// 			"schema-wide type reference with error during lazy resolution",
// 			func() *ast.Schema {
// 				entity := ast.Entity("User").Shape(ast.Attribute("field", ast.Type("Zoo::BadType")))
// 				ct := ast.CommonTypeNode{Name: "BadType", Type: errorType{}}
// 				ns1 := ast.Namespace("App", entity)
// 				ns2 := ast.Namespace("Zoo", ct)
// 				return ast.NewSchema(ns1, ns2)
// 			},
// 			true,
// 		},
// 		{
// 			"EntityTypeRef.resolve() called through IsType interface",
// 			func() *ast.Schema {
// 				return ast.NewSchema(ast.Namespace("App", ast.Entity("User").Shape(ast.Attribute("groups", ast.Set(ast.Ref("Group")))), ast.Entity("Group")))
// 			},
// 			false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			schema := tt.setupSchema()
// 			_, err := resolver.Resolve(schema)
// 			if (err != nil) != tt.wantErr {
// 				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func TestGlobalEntityResolution(t *testing.T) {
// 	t.Parallel()

// 	tests := []struct {
// 		name        string
// 		setupSchema func() *ast.Schema
// 		wantErr     bool
// 		checkFunc   func(*testing.T, *resolver.ResolvedSchema)
// 	}{
// 		{
// 			"action referencing global enum",
// 			func() *ast.Schema {
// 				globalEnum := ast.Enum("Status", "active", "inactive")
// 				localEntity := ast.Entity("User")
// 				action := ast.Action("view").Principal(ast.Ref("User")).Resource(ast.Ref("Status"))
// 				ns := ast.Namespace("App", localEntity, action)
// 				return ast.NewSchema(globalEnum, ns)
// 			},
// 			false,
// 			func(t *testing.T, rs *resolver.ResolvedSchema) {
// 				if _, found := rs.Enums["Status"]; !found {
// 					t.Error("expected global enum 'Status' to be present")
// 				}
// 			},
// 		},
// 		{
// 			"entity referencing already qualified name",
// 			func() *ast.Schema {
// 				baseEntity := ast.Entity("Base")
// 				ns1 := ast.Namespace("Core", baseEntity)
// 				derivedEntity := ast.Entity("Derived").MemberOf(ast.Ref("Core::Base"))
// 				ns2 := ast.Namespace("App", derivedEntity)
// 				return ast.NewSchema(ns1, ns2)
// 			},
// 			false,
// 			func(t *testing.T, rs *resolver.ResolvedSchema) {
// 				if _, found := rs.Entities["Core::Base"]; !found {
// 					t.Error("expected 'Core::Base' entity")
// 				}
// 				if _, found := rs.Entities["App::Derived"]; !found {
// 					t.Error("expected 'App::Derived' entity")
// 				}
// 			},
// 		},
// 		{
// 			"entity exists check with non-entity nodes",
// 			func() *ast.Schema {
// 				commonType := ast.CommonType("MyType", ast.String())
// 				globalEntity := ast.Entity("GlobalEntity")
// 				localAction := ast.Action("view").Resource(ast.Ref("GlobalEntity"))
// 				ns := ast.Namespace("App", localAction)
// 				return ast.NewSchema(commonType, globalEntity, ns)
// 			},
// 			false,
// 			func(t *testing.T, rs *resolver.ResolvedSchema) {
// 				if _, found := rs.Entities["GlobalEntity"]; !found {
// 					t.Error("expected global entity 'GlobalEntity' to be present")
// 				}
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			schema := tt.setupSchema()
// 			resolved, err := resolver.Resolve(schema)
// 			if (err != nil) != tt.wantErr {
// 				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if !tt.wantErr && tt.checkFunc != nil {
// 				tt.checkFunc(t, resolved)
// 			}
// 		})
// 	}

// 	t.Run("entity exists check with nil schema", func(t *testing.T) {
// 		rd := &resolveData{
// 			schema:               nil,
// 			namespace:            &ast.NamespaceNode{Name: "Test"},
// 			schemaCommonTypes:    make(map[string]*commonTypeEntry),
// 			namespaceCommonTypes: make(map[string]*commonTypeEntry),
// 		}
// 		exists := rd.entityExistsInEmptyast.Namespace("SomeEntity")
// 		if exists {
// 			t.Error("expected false for nil schema")
// 		}
// 	})
// }

func TestNamingConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupSchema func() *ast.Schema
		wantErr     bool
	}{
		{
			"entity conflict - nested namespace vs qualified name",
			func() *ast.Schema {
				ns1 := ast.Namespace("Goat::Gorilla", ast.Entity("Cows"))
				ns2 := ast.Namespace("Goat", ast.Entity("Gorilla::Cows"))
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"enum conflict - nested namespace vs qualified name",
			func() *ast.Schema {
				ns1 := ast.Namespace("Goat::Gorilla", ast.Enum("Status", "active"))
				ns2 := ast.Namespace("Goat", ast.Enum("Gorilla::Status", "active"))
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"entity vs enum conflict - entity first in namespace",
			func() *ast.Schema {
				ns1 := ast.Namespace("App", ast.Entity("Thing"))
				ns2 := ast.Namespace("App", ast.Enum("Thing", "value"))
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"entity vs enum conflict - enum first in namespace",
			func() *ast.Schema {
				ns1 := ast.Namespace("App", ast.Enum("Thing", "value"))
				ns2 := ast.Namespace("App", ast.Entity("Thing"))
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
		{"top-level entity vs enum conflict - enum first", func() *ast.Schema { return ast.NewSchema(ast.Enum("Status", "active"), ast.Entity("Status")) }, true},
		{"top-level enum vs entity conflict - entity first", func() *ast.Schema { return ast.NewSchema(ast.Entity("Status"), ast.Enum("Status", "active")) }, true},
		{
			"top-level vs namespaced entity conflict",
			func() *ast.Schema {
				topLevel := ast.Entity("A::B")
				ns := ast.Namespace("A", ast.Entity("B"))
				return ast.NewSchema(topLevel, ns)
			},
			true,
		},
		{
			"action conflict - nested namespace vs qualified name",
			func() *ast.Schema {
				ns1 := ast.Namespace("Goat::Gorilla", ast.Action("view"))
				ns2 := ast.Namespace("Goat::Gorilla", ast.Action("view"))
				return ast.NewSchema(ns1, ns2)
			},
			true,
		},
		{
			"no conflict - different namespaces",
			func() *ast.Schema {
				ns1 := ast.Namespace("Goat", ast.Entity("Cows"))
				ns2 := ast.Namespace("Sheep", ast.Entity("Cows"))
				return ast.NewSchema(ns1, ns2)
			},
			false,
		},
		{"top-level duplicate entity", func() *ast.Schema { return ast.NewSchema(ast.Entity("User"), ast.Entity("User")) }, true},
		{"top-level duplicate enum", func() *ast.Schema {
			return ast.NewSchema(ast.Enum("Status", "active"), ast.Enum("Status", "inactive"))
		}, true},
		{"top-level duplicate action", func() *ast.Schema { return ast.NewSchema(ast.Action("view"), ast.Action("view")) }, true},
		{"namespace duplicate entity", func() *ast.Schema {
			return ast.NewSchema(ast.Namespace("App", ast.Entity("User"), ast.Entity("User")))
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := tt.setupSchema()
			_, err := resolver.Resolve(schema)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
