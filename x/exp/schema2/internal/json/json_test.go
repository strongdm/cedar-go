package json

import (
	"encoding/json"
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
					&ast.EntityNode{Name: types.EntityType("User")},
				},
			},
		},
		{
			name: "entity with memberOf",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					&ast.EntityNode{
						Name: types.EntityType("User"),
						MemberOfVal: []ast.EntityTypeRef{
							{Name: types.EntityType("Group")},
						},
					},
					&ast.EntityNode{Name: types.EntityType("Group")},
				},
			},
		},
		{
			name: "entity with shape",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					&ast.EntityNode{
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
					&ast.EntityNode{
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
					&ast.EnumNode{
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
					&ast.EntityNode{Name: types.EntityType("User")},
					&ast.EntityNode{Name: types.EntityType("Doc")},
					&ast.ActionNode{
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
					&ast.ActionNode{
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
					&ast.EntityNode{Name: types.EntityType("User")},
					&ast.EntityNode{Name: types.EntityType("Doc")},
					&ast.ActionNode{
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
					&ast.CommonTypeNode{
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
					&ast.CommonTypeNode{
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
					&ast.CommonTypeNode{
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
					&ast.NamespaceNode{
						Name: "App",
						Declarations: []ast.IsDeclaration{
							&ast.EntityNode{Name: types.EntityType("User")},
							&ast.ActionNode{Name: "view"},
						},
					},
				},
			},
		},
		{
			name: "namespace with all declaration types",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					&ast.NamespaceNode{
						Name: "App",
						Declarations: []ast.IsDeclaration{
							&ast.CommonTypeNode{Name: "Name", Type: ast.StringType{}},
							&ast.EntityNode{Name: types.EntityType("User")},
							&ast.EnumNode{Name: "Status", Values: []types.String{"active"}},
							&ast.ActionNode{Name: "view"},
						},
					},
				},
			},
		},
		{
			name: "type reference to entity",
			schema: &ast.Schema{
				Nodes: []ast.IsNode{
					&ast.EntityNode{Name: types.EntityType("User")},
					&ast.EntityNode{
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
					&ast.CommonTypeNode{Name: "Name", Type: ast.StringType{}},
					&ast.EntityNode{
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
					&ast.EntityNode{
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
					&ast.EntityNode{Name: types.EntityType("User")},
					&ast.EntityNode{
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
					&ast.EntityNode{
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
			jsonData, err := Marshal(tt.schema)
			testutil.OK(t, err)

			// Unmarshal back
			schema2, err := Unmarshal(jsonData)
			testutil.OK(t, err)

			// Marshal again
			jsonData2, err := Marshal(schema2)
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
			wantErr: "Set type missing element",
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
			wantErr: "Set type missing element",
		},
		{
			name:    "nested attr error",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"nested": {"type": "Record", "attributes": {"bad": {"type": "Set"}}}}}}}, "actions": {}}}`,
			wantErr: "Set type missing element",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := Unmarshal([]byte(tt.json))
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
			&ast.EntityNode{Name: "User"},
			&ast.EnumNode{Name: "Status"},
			&ast.NamespaceNode{
				Name: "App",
				Declarations: []ast.IsDeclaration{
					&ast.EntityNode{Name: "Admin"},
					&ast.EnumNode{Name: "Role"},
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
			&ast.EntityNode{Name: "User"},
		},
	}

	jsonData, err := Marshal(schema)
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
					&ast.EntityNode{
						Name: "User",
						ShapeVal: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: tt.attrType},
							},
						},
					},
				},
			}

			jsonData, err := Marshal(schema)
			testutil.OK(t, err)

			schema2, err := Unmarshal(jsonData)
			testutil.OK(t, err)

			jsonData2, err := Marshal(schema2)
			testutil.OK(t, err)

			testutil.Equals(t, string(jsonData), string(jsonData2))
		})
	}
}

func TestUnknownAttrType(t *testing.T) {
	t.Parallel()

	// Test attr with unknown type - should return empty string for type
	json := `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"field": {"type": ""}}}}}, "actions": {}}}`
	_, err := Unmarshal([]byte(json))
	if err == nil {
		t.Fatal("expected error for unknown attr type")
	}
}
