package json

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

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
			name: "common type",
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
			name: "type reference to entity",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: types.EntityType("User")},
					ast.EntityNode{
						Name: types.EntityType("Doc"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "owner", Type: ast.TypeRef{Name: "User"}},
							},
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
			name: "set of entities",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					ast.EntityNode{Name: types.EntityType("User")},
					ast.EntityNode{
						Name: types.EntityType("Group"),
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "members", Type: ast.SetType{Element: ast.EntityTypeRef{Name: "User"}}},
							},
						},
					},
				},
			},
		},
		{
			name: "extension type",
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Marshal to JSON
			s := (*Schema)(tt.schema)
			jsonData, err := s.MarshalJSON()
			testutil.OK(t, err)

			// Unmarshal back
			var s2 Schema
			err = s2.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
			schema2 := (*ast.Schema)(&s2)

			// Marshal again
			s3 := (*Schema)(schema2)
			jsonData2, err := s3.MarshalJSON()
			testutil.OK(t, err)

			// JSON should be identical
			testutil.Equals(t, string(jsonData), string(jsonData2))
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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

	// Create new namespace
	ns1 := getOrCreateNamespace(namespaces, "App")
	testutil.Equals(t, true, ns1 != nil)
	testutil.Equals(t, true, ns1.EntityTypes != nil)
	testutil.Equals(t, true, ns1.Actions != nil)
	testutil.Equals(t, true, ns1.CommonTypes != nil)

	// Get existing namespace
	ns2 := getOrCreateNamespace(namespaces, "App")
	testutil.Equals(t, ns1, ns2)
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

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	testutil.OK(t, err)

	// Verify structure
	defaultNS, ok := parsed[""].(map[string]interface{})
	testutil.Equals(t, true, ok)
	testutil.Equals(t, true, defaultNS["entityTypes"] != nil)
	testutil.Equals(t, true, defaultNS["actions"] != nil)
}

func TestAttrTypesInRecord(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attrType ast.IsType
	}{
		{"string", ast.StringType{}},
		{"long", ast.LongType{}},
		{"bool", ast.BoolType{}},
		{"extension", ast.ExtensionType{Name: "ipaddr"}},
		{"entity ref", ast.EntityTypeRef{Name: "User"}},
		{"type ref", ast.TypeRef{Name: "MyType"}},
		{"set", ast.SetType{Element: ast.StringType{}}},
		{"record", ast.RecordType{Pairs: []ast.Pair{{Key: "x", Type: ast.StringType{}}}}},
	}

	for _, tt := range tests {
		tt := tt
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
		})
	}
}

func TestUnknownAttrType(t *testing.T) {
	t.Parallel()

	// Test attr with unknown type - should return empty string for type
	json := `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"field": {"type": ""}}}}}, "actions": {}}}`
	var s Schema
	err := s.UnmarshalJSON([]byte(json))
	if err == nil {
		t.Fatal("expected error for unknown attr type")
	}
}

// TestCommonTypePrimitives tests common types with all primitive types
func TestCommonTypePrimitives(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeVal  ast.IsType
		jsonType string
	}{
		{"long", ast.LongType{}, "Long"},
		{"bool", ast.BoolType{}, "Boolean"},
		{"extension", ast.ExtensionType{Name: "ipaddr"}, "Extension"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema := &ast.Schema{
				Nodes: []ast.IsNode{
					ast.CommonTypeNode{
						Name: "MyType",
						Type: tt.typeVal,
					},
				},
			}

			s := (*Schema)(schema)
			jsonData, err := s.MarshalJSON()
			testutil.OK(t, err)

			// Verify JSON contains expected type
			if !contains(string(jsonData), tt.jsonType) {
				t.Errorf("expected JSON to contain %q, got %s", tt.jsonType, jsonData)
			}

			// Round-trip
			var s2 Schema
			err = s2.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
			schema2 := (*ast.Schema)(&s2)

			s3 := (*Schema)(schema2)
			jsonData2, err := s3.MarshalJSON()
			testutil.OK(t, err)

			testutil.Equals(t, string(jsonData), string(jsonData2))
		})
	}
}

// TestTypeRefToNonEntity tests TypeRef that refers to a common type (not an entity)
func TestTypeRefToNonEntity(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.CommonTypeNode{
				Name: "Address",
				Type: ast.StringType{},
			},
			ast.CommonTypeNode{
				Name: "Person",
				Type: ast.RecordType{
					Pairs: []ast.Pair{
						{Key: "addr", Type: ast.TypeRef{Name: "Address"}},
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Verify it uses type reference, not Entity
	if contains(string(jsonData), `"type": "Entity"`) {
		t.Errorf("expected type reference, not Entity type: %s", jsonData)
	}

	// Round-trip
	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)

	s3 := (*Schema)(schema2)
	jsonData2, err := s3.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(jsonData), string(jsonData2))
}

// TestNestedRecordWithOptional tests nested records with optional fields
func TestNestedRecordWithOptional(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name: "Config",
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
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Verify JSON contains required: false
	if !contains(string(jsonData), `"required": false`) {
		t.Errorf("expected JSON to contain required: false, got %s", jsonData)
	}

	// Round-trip
	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)

	s3 := (*Schema)(schema2)
	jsonData2, err := s3.MarshalJSON()
	testutil.OK(t, err)

	testutil.Equals(t, string(jsonData), string(jsonData2))
}

// TestUnmarshalPrimitiveTypes tests unmarshalling all primitive types
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)
			schema := (*ast.Schema)(&s)

			// Re-marshal and verify
			s2 := (*Schema)(schema)
			jsonData, err := s2.MarshalJSON()
			testutil.OK(t, err)

			if !contains(string(jsonData), tt.wantType) {
				t.Errorf("expected JSON to contain %q, got %s", tt.wantType, jsonData)
			}
		})
	}
}

// TestUnmarshalRecordWithOptional tests unmarshalling records with optional attributes
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

	// Verify the schema was parsed correctly
	testutil.Equals(t, 1, len(schema.Nodes))

	entity, ok := schema.Nodes[0].(ast.EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, 2, len(entity.ShapeVal.Pairs))

	// Check that email is optional
	for _, pair := range entity.ShapeVal.Pairs {
		if string(pair.Key) == "email" {
			testutil.Equals(t, true, pair.Optional)
		}
	}
}

// TestUnmarshalNamedNamespace tests unmarshalling a named namespace
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

// TestUnmarshalAttrTypes tests unmarshalling all attribute types in records
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)
			schema := (*ast.Schema)(&s)

			// Re-marshal and verify round-trip
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

// TestEntityWithTags tests entity with tags marshalling
func TestEntityWithTagsRoundTrip(t *testing.T) {
	t.Parallel()

	// Test with Long tags type
	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name:    "Resource",
				TagsVal: ast.LongType{},
			},
		},
	}

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

// TestSetNestedInSet tests Set type nested within another Set
func TestSetNestedInSet(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
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
	}

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

// TestTypeRefToEntityInCommonType tests that TypeRef to an entity name
// is correctly serialized as Entity type in JSON.
func TestTypeRefToEntityInCommonType(t *testing.T) {
	t.Parallel()

	// Create a common type that references an entity using TypeRef
	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name: "User",
			},
			ast.CommonTypeNode{
				Name: "UserAlias",
				Type: ast.TypeRef{Name: "User"},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// The JSON should have User as Entity type
	testutil.Equals(t, strings.Contains(string(jsonData), `"type": "Entity"`), true)
	testutil.Equals(t, strings.Contains(string(jsonData), `"name": "User"`), true)
}

// TestTypeRefToEntityInRecordAttr tests that TypeRef to an entity name
// in a record attribute is correctly serialized as Entity type.
func TestTypeRefToEntityInRecordAttr(t *testing.T) {
	t.Parallel()

	// Create an entity with attribute that has TypeRef to entity
	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.EntityNode{
				Name: "User",
			},
			ast.EntityNode{
				Name: "Doc",
				ShapeVal: &ast.RecordType{
					Pairs: []ast.Pair{
						{
							Key:  "owner",
							Type: ast.TypeRef{Name: "User"}, // TypeRef not EntityTypeRef
						},
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Verify it round-trips
	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)

	s3 := (*Schema)(schema2)
	jsonData2, err := s3.MarshalJSON()
	testutil.OK(t, err)

	// Verify JSON contains Entity type for owner
	testutil.Equals(t, strings.Contains(string(jsonData), `"type": "Entity"`), true)
	testutil.Equals(t, string(jsonData), string(jsonData2))
}

// TestUnmarshalNestedSetError tests error propagation for invalid Set element
func TestUnmarshalNestedSetError(t *testing.T) {
	t.Parallel()

	// JSON with Set containing invalid nested Set (missing element)
	jsonData := `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Set", "element": {"type": "Set"}}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestUnmarshalNestedRecordError tests error propagation for invalid Record attribute
func TestUnmarshalNestedRecordError(t *testing.T) {
	t.Parallel()

	// JSON with Record containing attribute with invalid nested Set
	jsonData := `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Record", "attributes": {"x": {"type": "Set"}}}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestUnmarshalEmptyTypeName tests error for empty TypeName
func TestUnmarshalEmptyTypeName(t *testing.T) {
	t.Parallel()

	// JSON with empty type name (creates error)
	jsonData := `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": ""}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "unknown type"), true)
}

// TestUnmarshalAttrEmptyTypeName tests error for empty TypeName in attribute
func TestUnmarshalAttrEmptyTypeName(t *testing.T) {
	t.Parallel()

	// JSON with attribute having empty type name
	jsonData := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": ""}}}}}, "actions": {}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "unknown type"), true)
}

// TestUnmarshalAttrNestedSetError tests error propagation for Set in attribute
func TestUnmarshalAttrNestedSetError(t *testing.T) {
	t.Parallel()

	// JSON with attribute containing Set with missing element
	jsonData := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set"}}}}}, "actions": {}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestUnmarshalAttrNestedRecordError tests error in nested Record attribute
func TestUnmarshalAttrNestedRecordError(t *testing.T) {
	t.Parallel()

	// JSON with nested Record containing attribute with invalid Set
	jsonData := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Record", "attributes": {"y": {"type": "Set"}}}}}}}, "actions": {}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestUnmarshalNamespaceError tests error propagation from named namespace
func TestUnmarshalNamespaceError(t *testing.T) {
	t.Parallel()

	// JSON with named namespace containing error
	jsonData := `{"MyApp": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Set"}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestUnmarshalAttrSetWithBadElement tests error in Set attribute with bad element
func TestUnmarshalAttrSetWithBadElement(t *testing.T) {
	t.Parallel()

	// JSON with attribute containing Set whose element is another Set without element
	jsonData := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set", "element": {"type": "Set"}}}}}}, "actions": {}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.Error(t, err)
	testutil.Equals(t, strings.Contains(err.Error(), "set type missing element"), true)
}

// TestMarshalNilTypeInCommonType tests marshalling a common type with nil type
func TestMarshalNilTypeInCommonType(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.CommonTypeNode{
				Name: "BadType",
				Type: nil, // nil type - hits default case
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)
	// Should produce JSON with null type
	testutil.Equals(t, strings.Contains(string(jsonData), "BadType"), true)
}

// TestMarshalNilTypeInRecordAttr tests marshalling record with nil attribute type
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
							Type: nil, // nil type - hits default case in attrToJSON
						},
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)
	// Should produce JSON with the entity
	testutil.Equals(t, strings.Contains(string(jsonData), "badAttr"), true)
}

// TestEntityOrCommonFormat tests parsing Rust CLI v4.8+ EntityOrCommon JSON format
func TestEntityOrCommonFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		json string
	}{
		{
			name: "EntityOrCommon with __cedar::String",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "EntityOrCommon", "name": "__cedar::String"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with __cedar::Long",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"age": {"type": "EntityOrCommon", "name": "__cedar::Long"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with __cedar::Bool",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"active": {"type": "EntityOrCommon", "name": "__cedar::Bool"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with __cedar::Boolean",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"active": {"type": "EntityOrCommon", "name": "__cedar::Boolean"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with __cedar::ipaddr extension",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"ip": {"type": "EntityOrCommon", "name": "__cedar::ipaddr"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with __cedar::datetime extension",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"time": {"type": "EntityOrCommon", "name": "__cedar::datetime"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with unprefixed String",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "EntityOrCommon", "name": "String"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with unprefixed Long",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"age": {"type": "EntityOrCommon", "name": "Long"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with unprefixed Bool",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"flag": {"type": "EntityOrCommon", "name": "Bool"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with unprefixed Boolean",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"flag": {"type": "EntityOrCommon", "name": "Boolean"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon with type reference",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"data": {"type": "EntityOrCommon", "name": "MyType"}}}}}, "actions": {}}}`,
		},
		{
			name: "EntityOrCommon in commonTypes",
			json: `{"": {"commonTypes": {"X": {"type": "EntityOrCommon", "name": "__cedar::String"}}, "entityTypes": {}, "actions": {}}}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.OK(t, err)
			schema := (*ast.Schema)(&s)

			// Re-marshal and verify round-trip
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

// TestRecordMarshalJSONWithNilAttributes tests Record type MarshalJSON with nil attributes
func TestRecordMarshalJSONWithNilAttributes(t *testing.T) {
	t.Parallel()

	// Create a Type directly with nil Attributes to test MarshalJSON
	recordType := Type{
		TypeName:   "Record",
		Attributes: nil, // nil, not empty map
	}

	jsonData, err := json.Marshal(recordType)
	testutil.OK(t, err)

	// Should output attributes: {}
	testutil.Equals(t, strings.Contains(string(jsonData), `"attributes":{}`), true)
}

// TestAttrRecordMarshalJSONWithNilAttributes tests Attr Record type MarshalJSON with nil attributes
func TestAttrRecordMarshalJSONWithNilAttributes(t *testing.T) {
	t.Parallel()

	// Create an Attr directly with nil Attributes to test MarshalJSON
	recordAttr := Attr{
		TypeName:   "Record",
		Attributes: nil, // nil, not empty map
	}

	jsonData, err := json.Marshal(recordAttr)
	testutil.OK(t, err)

	// Should output attributes: {}
	testutil.Equals(t, strings.Contains(string(jsonData), `"attributes":{}`), true)
}

// TestActionMemberOfWithEmptyType tests action memberOf with empty type defaults to Action
func TestActionMemberOfWithEmptyType(t *testing.T) {
	t.Parallel()

	// JSON with action memberOf where type is empty (Rust CLI v4.8+ format)
	jsonData := `{"": {"entityTypes": {}, "actions": {"read": {}, "write": {"memberOf": [{"id": "read"}]}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	// Marshal back
	s2 := (*Schema)(schema)
	jsonOut, err := s2.MarshalJSON()
	testutil.OK(t, err)

	// Should have type: Action in output (check both possible formats)
	hasType := strings.Contains(string(jsonOut), `"type":"Action"`) ||
		strings.Contains(string(jsonOut), `"type": "Action"`)
	testutil.Equals(t, hasType, true)
}

// TestActionMemberOfEmptyTypeInSchema tests marshalling action with empty memberOf type
func TestActionMemberOfEmptyTypeInSchema(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Nodes: []ast.IsNode{
			ast.ActionNode{
				Name: "write",
				MemberOfVal: []ast.EntityRef{
					{
						Type: ast.EntityTypeRef{Name: ""}, // empty type
						ID:   "read",
					},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Should have type: Action (default for empty type) - check both formats
	hasType := strings.Contains(string(jsonData), `"type":"Action"`) ||
		strings.Contains(string(jsonData), `"type": "Action"`)
	if !hasType {
		t.Errorf("expected type:Action in JSON, got: %s", string(jsonData))
	}
}

// TestEmptyAppliesToNotEmitted tests that empty appliesTo is not emitted
func TestEmptyAppliesToNotEmitted(t *testing.T) {
	t.Parallel()

	// JSON with action having empty appliesTo (Rust CLI format)
	jsonData := `{"": {"entityTypes": {}, "actions": {"view": {"appliesTo": {"principalTypes": [], "resourceTypes": []}}}}}`

	var s Schema
	err := s.UnmarshalJSON([]byte(jsonData))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	// Marshal back
	s2 := (*Schema)(schema)
	jsonOut, err := s2.MarshalJSON()
	testutil.OK(t, err)

	// Should NOT have appliesTo in output (since it's empty)
	testutil.Equals(t, !strings.Contains(string(jsonOut), "appliesTo"), true)
}

// TestResolveEntityOrCommonDirectly tests resolveEntityOrCommon function directly
func TestResolveEntityOrCommonDirectly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantType string
	}{
		// __cedar:: prefixed primitives
		{"__cedar::String", "__cedar::String", "StringType"},
		{"__cedar::Long", "__cedar::Long", "LongType"},
		{"__cedar::Bool", "__cedar::Bool", "BoolType"},
		{"__cedar::Boolean", "__cedar::Boolean", "BoolType"},
		// __cedar:: prefixed extension types
		{"__cedar::ipaddr", "__cedar::ipaddr", "ExtensionType"},
		{"__cedar::datetime", "__cedar::datetime", "ExtensionType"},
		{"__cedar::decimal", "__cedar::decimal", "ExtensionType"},
		{"__cedar::duration", "__cedar::duration", "ExtensionType"},
		// Unprefixed primitives
		{"String", "String", "StringType"},
		{"Long", "Long", "LongType"},
		{"Bool", "Bool", "BoolType"},
		{"Boolean", "Boolean", "BoolType"},
		// Type references
		{"MyType", "MyType", "TypeRef"},
		{"User", "User", "TypeRef"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := resolveEntityOrCommon(tt.input)
			testutil.OK(t, err)

			typeName := getTypeName(result)
			testutil.Equals(t, tt.wantType, typeName)
		})
	}
}

// getTypeName returns the type name of an ast.IsType for testing
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
