package json

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// testRoundTrip is a helper that marshals a schema to JSON, unmarshals it back,
// marshals again, and verifies both JSON outputs are identical.
func testRoundTrip(t *testing.T, schema *ast.Schema) {
	t.Helper()
	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)

	s3 := (*Schema)(schema2)
	jsonData2, err := s3.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(jsonData), string(jsonData2))
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema *ast.Schema
	}{
		{
			name: "empty schema",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{},
			},
		},
		{
			name: "simple entity",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: types.EntityType("User")},
				},
			},
		},
		{
			name: "entity with memberOf",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: types.EntityType("User"),
						MemberOfVal: []ast.EntityTypeRef{
							{Name: types.EntityType("Group")},
						},
					},
					ast.EntityNode{Name: types.EntityType("Group")},
				},
			},
		},
		{
			name: "entity with shape",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: types.EntityType("User"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "name", Type: ast.StringType{}},
								{Key: "age", Type: ast.LongType{}},
								{Key: "active", Type: ast.BoolType{}},
								{Key: "email", Type: ast.StringType{}, Optional: true},
							},
						},
					},
				},
			},
		},
		{
			name: "entity with tags",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name:    types.EntityType("Resource"),
						TagsVal: ast.StringType{},
					},
				},
			},
		},
		{
			name: "enum entity",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EnumNode{
						Name:   types.EntityType("Status"),
						Values: []types.String{"active", "inactive", "pending"},
					},
				},
			},
		},
		{
			name: "simple action",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: types.EntityType("User")},
					ast.EntityNode{Name: types.EntityType("Doc")},
					ast.ActionNode{
						Name: "view",
						AppliesToVal: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
							ResourceTypes:  []ast.EntityTypeRef{{Name: "Doc"}},
						},
					},
				},
			},
		},
		{
			name: "action with memberOf",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.ActionNode{
						Name: "view",
						MemberOfVal: []ast.EntityRef{
							{Type: ast.EntityTypeRef{Name: "Action"}, ID: "readOnly"},
						},
					},
				},
			},
		},
		{
			name: "action with context",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: types.EntityType("User")},
					ast.EntityNode{Name: types.EntityType("Doc")},
					ast.ActionNode{
						Name: "view",
						AppliesToVal: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
							ResourceTypes:  []ast.EntityTypeRef{{Name: "Doc"}},
							Context: ast.RecordType{
								Pairs: []ast.Pair{
									{Key: "ip", Type: ast.ExtensionType{Name: "ipaddr"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "common type string",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{
						Name: "Name",
						Type: ast.StringType{},
					},
				},
			},
		},
		{
			name: "common type set",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{
						Name: "Names",
						Type: ast.SetType{Element: ast.StringType{}},
					},
				},
			},
		},
		{
			name: "common type record",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{
						Name: "Address",
						Type: ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "street", Type: ast.StringType{}},
								{Key: "zip", Type: ast.LongType{}, Optional: true},
							},
						},
					},
				},
			},
		},
		{
			name: "namespace",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.NamespaceNode{
						Name: "App",
						Declarations: []ast.IsDeclaration{
							ast.EntityNode{Name: types.EntityType("User")},
							ast.ActionNode{Name: "view"},
						},
					},
				},
			},
		},
		{
			name: "namespace with all declaration types",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.NamespaceNode{
						Name: "App",
						Declarations: []ast.IsDeclaration{
							ast.CommonTypeNode{Name: "Name", Type: ast.StringType{}},
							ast.EntityNode{Name: types.EntityType("User")},
							ast.EnumNode{Name: "Status", Values: []types.String{"active"}},
							ast.ActionNode{Name: "view"},
						},
					},
				},
			},
		},
		{
			name: "type reference to common type",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{Name: "Name", Type: ast.StringType{}},
					ast.EntityNode{
						Name: types.EntityType("User"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "name", Type: ast.TypeRef{Name: "Name"}},
							},
						},
					},
				},
			},
		},
		{
			name: "nested records",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: types.EntityType("Config"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{
									Key: "nested",
									Type: ast.RecordType{
										Pairs: []ast.Pair{
											{Key: "value", Type: ast.StringType{}},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "extension types",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: types.EntityType("Request"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "ip", Type: ast.ExtensionType{Name: "ipaddr"}},
								{Key: "time", Type: ast.ExtensionType{Name: "datetime"}},
							},
						},
					},
				},
			},
		},
		{
			name: "nested record with optional",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: types.EntityType("Config"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{
									Key: "nested",
									Type: ast.RecordType{
										Pairs: []ast.Pair{
											{Key: "required_field", Type: ast.StringType{}},
											{Key: "optional_field", Type: ast.LongType{}, Optional: true},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "set nested in set",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{
						Name: "NestedSets",
						Type: ast.SetType{
							Element: ast.SetType{
								Element: ast.StringType{},
							},
						},
					},
				},
			},
		},
		{
			name: "entity with long tags",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name:    "Resource",
						TagsVal: ast.LongType{},
					},
				},
			},
		},
		{
			name: "common type with primitives",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{Name: "MyLong", Type: ast.LongType{}},
					ast.CommonTypeNode{Name: "MyBool", Type: ast.BoolType{}},
					ast.CommonTypeNode{Name: "MyExt", Type: ast.ExtensionType{Name: "ipaddr"}},
				},
			},
		},
		{
			name: "type ref to non-entity",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{Name: "Address", Type: ast.StringType{}},
					ast.CommonTypeNode{
						Name: "Person",
						Type: ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "addr", Type: ast.TypeRef{Name: "Address"}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testRoundTrip(t, tt.schema)
		})
	}
}

func TestUnmarshalErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: "invalid character",
		},
		{
			name:    "set without element",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "unknown type",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": ""}}}, "actions": {}}}`,
			wantErr: "unknown type",
		},
		{
			name:    "common type error",
			json:    `{"": {"commonTypes": {"Bad": {"type": "Set"}}, "entityTypes": {}, "actions": {}}}`,
			wantErr: "parsing common type",
		},
		{
			name:    "entity shape error",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "parsing entity",
		},
		{
			name:    "entity tags error",
			json:    `{"": {"entityTypes": {"User": {"tags": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "parsing entity",
		},
		{
			name:    "action context error",
			json:    `{"": {"entityTypes": {}, "actions": {"view": {"appliesTo": {"context": {"type": "Set"}}}}}}`,
			wantErr: "parsing action",
		},
		{
			name:    "attr set without element",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"data": {"type": "Set"}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "nested attr error",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"nested": {"type": "Record", "attributes": {"bad": {"type": "Set"}}}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "empty type name in commonTypes",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": ""}}}}`,
			wantErr: "unknown type",
		},
		{
			name:    "empty type name in attribute",
			json:    `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": ""}}}}}, "actions": {}}}`,
			wantErr: "unknown type",
		},
		{
			name:    "nested set error",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Set", "element": {"type": "Set"}}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "nested record error",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Record", "attributes": {"x": {"type": "Set"}}}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "attr nested set error",
			json:    `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set", "element": {"type": "Set"}}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "namespace error",
			json:    `{"MyApp": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Set"}}}}`,
			wantErr: "set type missing element",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.Error(t, err)
			testutil.Equals(t, strings.Contains(err.Error(), tt.wantErr), true)
		})
	}
}

func TestRecordAttributeTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attrType ast.IsType
	}{
		{"string", ast.StringType{}},
		{"long", ast.LongType{}},
		{"bool", ast.BoolType{}},
		{"extension", ast.ExtensionType{Name: "ipaddr"}},
		{"type ref", ast.TypeRef{Name: "MyType"}},
		{"set", ast.SetType{Element: ast.StringType{}}},
		{"record", ast.RecordType{Pairs: []ast.Pair{{Key: "x", Type: ast.StringType{}}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema := &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{
						Name: "User",
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: tt.attrType},
							},
						},
					},
				},
			}

			testRoundTrip(t, schema)
		})
	}
}

func TestUnmarshalPrimitiveTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		json     string
		wantType string
	}{
		{
			name:     "Long type",
			json:     `{"": {"commonTypes": {"X": {"type": "Long"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "Long",
		},
		{
			name:     "Bool type",
			json:     `{"": {"commonTypes": {"X": {"type": "Bool"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "Bool",
		},
		{
			name:     "Boolean type",
			json:     `{"": {"commonTypes": {"X": {"type": "Boolean"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "Boolean",
		},
		{
			name:     "Extension type",
			json:     `{"": {"commonTypes": {"X": {"type": "Extension", "name": "ipaddr"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "Extension",
		},
		{
			name:     "Entity type",
			json:     `{"": {"commonTypes": {"X": {"type": "Entity", "name": "User"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "Entity",
		},
		{
			name:     "Type reference",
			json:     `{"": {"commonTypes": {"X": {"type": "MyType"}}, "entityTypes": {}, "actions": {}}}`,
			wantType: "MyType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)
			schema := (*ast.Schema)(&s)

			s2 := (*Schema)(schema)
			jsonData, err := s2.MarshalJSON()
			testutil.OK(t, err)

			testutil.Equals(t, strings.Contains(string(jsonData), tt.wantType), true)
		})
	}
}

func TestUnmarshalAttrLegacyBoolType(t *testing.T) {
	t.Parallel()

	// Test legacy "Bool" type format in record attributes
	json := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"active": {"type": "Bool"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(json))
	testutil.OK(t, err)

	// Verify the Bool type was parsed correctly
	schema := (*ast.Schema)(&s)
	testutil.FatalIf(t, len(schema.Nodes) == 0, "expected at least one node")

	entity, ok := schema.Nodes[0].(ast.EntityNode)
	testutil.FatalIf(t, !ok, "expected EntityNode")
	testutil.FatalIf(t, entity.ShapeVal == nil, "expected shape")
	testutil.FatalIf(t, len(entity.ShapeVal.Pairs) == 0, "expected at least one pair")

	// Verify it's a BoolType
	_, ok = entity.ShapeVal.Pairs[0].Type.(ast.BoolType)
	testutil.FatalIf(t, !ok, "expected BoolType")
}

func TestUnmarshalRecordWithOptional(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String"},
							"email": {"type": "String", "required": false}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	testutil.Equals(t, 1, len(schema.Nodes))

	entity, ok := schema.Nodes[0].(ast.EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, 2, len(entity.ShapeVal.Pairs))

	for _, pair := range entity.ShapeVal.Pairs {
		if string(pair.Key) == "email" {
			testutil.Equals(t, true, pair.Optional)
		}
	}
}

func TestUnmarshalNamedNamespace(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"MyApp": {
			"entityTypes": {
				"User": {}
			},
			"actions": {
				"view": {}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	testutil.Equals(t, 1, len(schema.Nodes))

	ns, ok := schema.Nodes[0].(ast.NamespaceNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, "MyApp", string(ns.Name))
	testutil.Equals(t, 2, len(ns.Declarations))
}

func TestUnmarshalAttrTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
	}{
		{
			name: "Long attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Long"}}}}}, "actions": {}}}`,
		},
		{
			name: "Boolean attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Boolean"}}}}}, "actions": {}}}`,
		},
		{
			name: "Extension attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Extension", "name": "ipaddr"}}}}}, "actions": {}}}`,
		},
		{
			name: "Entity attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Entity", "name": "User"}}}}}, "actions": {}}}`,
		},
		{
			name: "Type ref attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "MyType"}}}}}, "actions": {}}}`,
		},
		{
			name: "Set attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set", "element": {"type": "String"}}}}}}, "actions": {}}}`,
		},
		{
			name: "Record attr",
			json: `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Record", "attributes": {"y": {"type": "String"}}}}}}}, "actions": {}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)

			schema := (*ast.Schema)(&s)
			s2 := (*Schema)(schema)
			jsonData, err := s2.MarshalJSON()
			testutil.OK(t, err)

			var s3 Schema
			err = s3.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
			schema2 := (*ast.Schema)(&s3)

			s4 := (*Schema)(schema2)
			jsonData2, err := s4.MarshalJSON()
			testutil.OK(t, err)

			testutil.Equals(t, string(jsonData), string(jsonData2))
		})
	}
}

func TestEntityOrCommonFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
	}{
		{
			name: "__cedar::String",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "EntityOrCommon", "name": "__cedar::String"}}}}}, "actions": {}}}`,
		},
		{
			name: "__cedar::Long",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"age": {"type": "EntityOrCommon", "name": "__cedar::Long"}}}}}, "actions": {}}}`,
		},
		{
			name: "__cedar::Bool",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"active": {"type": "EntityOrCommon", "name": "__cedar::Bool"}}}}}, "actions": {}}}`,
		},
		{
			name: "__cedar::Boolean",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"active": {"type": "EntityOrCommon", "name": "__cedar::Boolean"}}}}}, "actions": {}}}`,
		},
		{
			name: "__cedar::ipaddr extension",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"ip": {"type": "EntityOrCommon", "name": "__cedar::ipaddr"}}}}}, "actions": {}}}`,
		},
		{
			name: "__cedar::datetime extension",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"time": {"type": "EntityOrCommon", "name": "__cedar::datetime"}}}}}, "actions": {}}}`,
		},
		{
			name: "unprefixed String",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "EntityOrCommon", "name": "String"}}}}}, "actions": {}}}`,
		},
		{
			name: "unprefixed Long",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"age": {"type": "EntityOrCommon", "name": "Long"}}}}}, "actions": {}}}`,
		},
		{
			name: "unprefixed Bool",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"flag": {"type": "EntityOrCommon", "name": "Bool"}}}}}, "actions": {}}}`,
		},
		{
			name: "unprefixed Boolean",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"flag": {"type": "EntityOrCommon", "name": "Boolean"}}}}}, "actions": {}}}`,
		},
		{
			name: "type reference",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"data": {"type": "EntityOrCommon", "name": "MyType"}}}}}, "actions": {}}}`,
		},
		{
			name: "in commonTypes",
			json: `{"": {"commonTypes": {"X": {"type": "EntityOrCommon", "name": "__cedar::String"}}, "entityTypes": {}, "actions": {}}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)

			schema := (*ast.Schema)(&s)
			s2 := (*Schema)(schema)
			jsonData, err := s2.MarshalJSON()
			testutil.OK(t, err)

			var s3 Schema
			err = s3.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
			schema2 := (*ast.Schema)(&s3)

			s4 := (*Schema)(schema2)
			jsonData2, err := s4.MarshalJSON()
			testutil.OK(t, err)

			testutil.Equals(t, string(jsonData), string(jsonData2))
		})
	}
}

func TestResolveEntityOrCommon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		{"__cedar::String", "__cedar::String", "StringType"},
		{"__cedar::Long", "__cedar::Long", "LongType"},
		{"__cedar::Bool", "__cedar::Bool", "BoolType"},
		{"__cedar::ipaddr", "__cedar::ipaddr", "ExtensionType"},
		{"__cedar::datetime", "__cedar::datetime", "ExtensionType"},
		{"__cedar::decimal", "__cedar::decimal", "ExtensionType"},
		{"__cedar::duration", "__cedar::duration", "ExtensionType"},
		{"String", "String", "StringType"},
		{"Long", "Long", "LongType"},
		{"Bool", "Bool", "BoolType"},
		{"Boolean", "Boolean", "TypeRef"},
		{"MyType", "MyType", "TypeRef"},
		{"User", "User", "TypeRef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := resolveEntityOrCommon(tt.input)
			testutil.OK(t, err)

			typeName := getTypeName(result)
			testutil.Equals(t, tt.wantType, typeName)
		})
	}
}

func getTypeName(t ast.IsType) string {
	switch t.(type) {
	case ast.StringType:
		return "StringType"
	case ast.LongType:
		return "LongType"
	case ast.BoolType:
		return "BoolType"
	case ast.ExtensionType:
		return "ExtensionType"
	case ast.TypeRef:
		return "TypeRef"
	default:
		return "Unknown"
	}
}

func TestActionMemberOfWithEmptyType(t *testing.T) {
	t.Parallel()

	jsonData := `{"": {"entityTypes": {}, "actions": {"read": {}, "write": {"memberOf": [{"id": "read"}]}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	s2 := (*Schema)(schema)
	jsonOut, err := s2.MarshalJSON()
	testutil.OK(t, err)

	hasType := strings.Contains(string(jsonOut), `"type":"Action"`) ||
		strings.Contains(string(jsonOut), `"type": "Action"`)
	testutil.Equals(t, hasType, true)
}

func TestEmptyAppliesToNotEmitted(t *testing.T) {
	t.Parallel()

	jsonData := `{"": {"entityTypes": {}, "actions": {"view": {"appliesTo": {"principalTypes": [], "resourceTypes": []}}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	s2 := (*Schema)(schema)
	jsonOut, err := s2.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, !strings.Contains(string(jsonOut), "appliesTo"), true)
}

func TestRecordMarshalJSONWithNilAttributes(t *testing.T) {
	t.Parallel()

	recordType := Type{
		TypeName:   "Record",
		Attributes: nil,
	}

	jsonData, err := json.Marshal(recordType)
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"attributes":{}`), true)
}

func TestAttrRecordMarshalJSONWithNilAttributes(t *testing.T) {
	t.Parallel()

	recordAttr := Attr{
		TypeName:   "Record",
		Attributes: nil,
	}

	jsonData, err := json.Marshal(recordAttr)
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"attributes":{}`), true)
}

func TestMarshalNilTypeInCommonType(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.CommonTypeNode{
				Name: "BadType",
				Type: nil,
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, strings.Contains(string(jsonData), "BadType"), true)
}

func TestMarshalNilTypeInRecordAttr(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name: "E",
				ShapeVal: &ast.RecordType{
					Pairs: []ast.Pair{
						{
							Key:  "badAttr",
							Type: nil,
						},
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, strings.Contains(string(jsonData), "badAttr"), true)
}

func TestMarshalJSONOutput(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{Name: "User"},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	testutil.OK(t, err)

	defaultNS, ok := parsed[""].(map[string]interface{})
	testutil.Equals(t, true, ok)
	testutil.Equals(t, true, defaultNS["entityTypes"] != nil)
	testutil.Equals(t, true, defaultNS["actions"] != nil)
}

func TestCollectEntityNames(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{Name: "User"},
			ast.EnumNode{Name: "Status"},
			ast.NamespaceNode{
				Name: "App",
				Declarations: []ast.IsDeclaration{
					ast.EntityNode{Name: "Admin"},
					ast.EnumNode{Name: "Role"},
				},
			},
		},
	}

	names := collectEntityNames(schema)
	testutil.Equals(t, true, names["User"])
	testutil.Equals(t, true, names["Status"])
	testutil.Equals(t, true, names["Admin"])
	testutil.Equals(t, true, names["Role"])
	testutil.Equals(t, false, names["Unknown"])
}

func TestGetOrCreateNamespace(t *testing.T) {
	t.Parallel()

	namespaces := make(map[string]*Namespace)

	ns1 := getOrCreateNamespace(namespaces, "App", nil)
	testutil.Equals(t, true, ns1 != nil)
	testutil.Equals(t, true, ns1.EntityTypes != nil)
	testutil.Equals(t, true, ns1.Actions != nil)
	testutil.Equals(t, true, ns1.CommonTypes != nil)

	ns2 := getOrCreateNamespace(namespaces, "App", nil)
	testutil.Equals(t, ns1, ns2)
}

func TestTypeRefToEntity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema *ast.Schema
	}{
		{
			name: "in common type",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: "User"},
					ast.CommonTypeNode{
						Name: "UserAlias",
						Type: ast.TypeRef{Name: "User"},
					},
				},
			},
		},
		{
			name: "in record attribute",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: "User"},
					ast.EntityNode{
						Name: "Doc",
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "owner", Type: ast.TypeRef{Name: "User"}},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := (*Schema)(tt.schema)
			jsonData, err := s.MarshalJSON()
			testutil.OK(t, err)

			// TypeRef to an entity should marshal as "Entity"
			testutil.Equals(t, strings.Contains(string(jsonData), `"type": "Entity"`), true)
			testutil.Equals(t, strings.Contains(string(jsonData), `"name": "User"`), true)

			// When unmarshaled, "Entity" becomes EntityTypeRef which marshals to "EntityOrCommon"
			// So we verify the unmarshal works but don't expect round-trip equality
			var s2 Schema
			err = s2.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
		})
	}
}

func TestActionMemberOfEmptyTypeInSchema(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.ActionNode{
				Name: "write",
				MemberOfVal: []ast.EntityRef{
					{
						Type: ast.EntityTypeRef{Name: ""},
						ID:   "read",
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	hasType := strings.Contains(string(jsonData), `"type":"Action"`) ||
		strings.Contains(string(jsonData), `"type": "Action"`)
	testutil.Equals(t, hasType, true)
}

func TestEntityAnnotations(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name: "User",
				Annotations: []ast.Annotation{
					{Key: "doc", Value: "A user entity"},
					{Key: "author", Value: "Alice"},
				},
			},
		},
	}

	testRoundTrip(t, schema)

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"doc"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"A user entity"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"author"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"Alice"`), true)
}

func TestEnumAnnotations(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EnumNode{
				Name:   "Status",
				Values: []types.String{"active", "inactive"},
				Annotations: []ast.Annotation{
					{Key: "doc", Value: "Status enum"},
				},
			},
		},
	}

	testRoundTrip(t, schema)

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"doc"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"Status enum"`), true)
}

func TestActionAnnotations(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.ActionNode{
				Name: "view",
				Annotations: []ast.Annotation{
					{Key: "doc", Value: "View action"},
					{Key: "version", Value: "1.0"},
				},
			},
		},
	}

	testRoundTrip(t, schema)

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"doc"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"View action"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"version"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"1.0"`), true)
}

func TestCommonTypeAnnotations(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.CommonTypeNode{
				Name: "Address",
				Type: ast.RecordType{
					Pairs: []ast.Pair{
						{Key: "street", Type: ast.StringType{}},
					},
				},
				Annotations: []ast.Annotation{
					{Key: "doc", Value: "Address type"},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Debug: print the JSON
	t.Logf("JSON: %s", string(jsonData))

	testutil.Equals(t, strings.Contains(string(jsonData), `"doc"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"Address type"`), true)

	// Test unmarshaling preserves annotations
	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)

	ct := schema2.Nodes[0].(ast.CommonTypeNode)
	testutil.Equals(t, len(ct.Annotations), 1)
	testutil.Equals(t, string(ct.Annotations[0].Key), "doc")
	testutil.Equals(t, string(ct.Annotations[0].Value), "Address type")
}

func TestAnnotationsUnmarshal(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"": {
			"entityTypes": {
				"User": {
					"annotations": {
						"doc": "User entity",
						"author": "Bob"
					}
				}
			},
			"actions": {
				"view": {
					"annotations": {
						"doc": "View action"
					}
				}
			},
			"commonTypes": {
				"Address": {
					"type": "Record",
					"attributes": {
						"street": {"type": "String"}
					},
					"annotations": {
						"doc": "Address type"
					}
				}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	// Find nodes by type since order may vary
	var entity ast.EntityNode
	var action ast.ActionNode
	var commonType ast.CommonTypeNode
	var foundEntity, foundAction, foundCommonType bool

	for _, node := range schema.Nodes {
		switch n := node.(type) {
		case ast.EntityNode:
			if string(n.Name) == "User" {
				entity = n
				foundEntity = true
			}
		case ast.ActionNode:
			if string(n.Name) == "view" {
				action = n
				foundAction = true
			}
		case ast.CommonTypeNode:
			if string(n.Name) == "Address" {
				commonType = n
				foundCommonType = true
			}
		}
	}

	// Check entity annotations
	testutil.FatalIf(t, !foundEntity, "User entity not found")
	testutil.Equals(t, len(entity.Annotations), 2)

	// Check action annotations
	testutil.FatalIf(t, !foundAction, "view action not found")
	testutil.Equals(t, len(action.Annotations), 1)
	testutil.Equals(t, string(action.Annotations[0].Value), "View action")

	// Check common type annotations
	testutil.FatalIf(t, !foundCommonType, "Address common type not found")
	testutil.Equals(t, len(commonType.Annotations), 1)
	testutil.Equals(t, string(commonType.Annotations[0].Value), "Address type")
}

func TestNamespaceAnnotations(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.NamespaceNode{
				Name: "MyApp",
				Declarations: []ast.IsDeclaration{
					ast.EntityNode{
						Name: "User",
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "User in MyApp"},
						},
					},
					ast.ActionNode{
						Name: "view",
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "View in MyApp"},
						},
					},
					ast.CommonTypeNode{
						Name: "Address",
						Type: ast.StringType{},
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "Address in MyApp"},
						},
					},
					ast.EnumNode{
						Name:   "Status",
						Values: []types.String{"active"},
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "Status in MyApp"},
						},
					},
				},
			},
		},
	}

	testRoundTrip(t, schema)

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, strings.Contains(string(jsonData), `"User in MyApp"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"View in MyApp"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"Address in MyApp"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"Status in MyApp"`), true)
}

func TestEntityTypeRefInCommonType(t *testing.T) {
	t.Parallel()

	// Test that ast.EntityTypeRef marshals to EntityOrCommon in a common type
	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{Name: "User"},
			ast.CommonTypeNode{
				Name: "UserRef",
				Type: ast.EntityTypeRef{Name: "User"},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Should marshal to EntityOrCommon
	testutil.Equals(t, strings.Contains(string(jsonData), `"EntityOrCommon"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"name": "User"`), true)
}

func TestEntityTypeRefInRecordAttribute(t *testing.T) {
	t.Parallel()

	// Test that ast.EntityTypeRef marshals to EntityOrCommon in record attributes
	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{Name: "User"},
			ast.EntityNode{
				Name: "Doc",
				ShapeVal: &ast.RecordType{
					Pairs: []ast.Pair{
						{Key: "owner", Type: ast.EntityTypeRef{Name: "User"}},
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Should marshal to EntityOrCommon
	testutil.Equals(t, strings.Contains(string(jsonData), `"EntityOrCommon"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"name": "User"`), true)
}
