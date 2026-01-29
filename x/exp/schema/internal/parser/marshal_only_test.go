package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parser"
)

// TestMarshalOnlyFeatures tests marshaling features using semantic types like ast.IPAddr()
// These don't roundtrip exactly because parsing treats them as TypeRefs
func TestMarshalOnlyFeatures(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ast      *ast.Schema
		expected string
	}{
		{
			name: "extension types",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("IP"): ast.CommonType{
						Type: ast.IPAddr(),
					},
				},
			},
			expected: "type IP = __cedar::ipaddr;\n",
		},
		{
			name: "entity type reference",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("UserRef"): ast.CommonType{
						Type: ast.EntityType(types.EntityType("MyApp::User")),
					},
				},
			},
			expected: "type UserRef = MyApp::User;\n",
		},
		{
			name: "type reference",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("AliasedName"): ast.CommonType{
						Type: ast.Type("Name"),
					},
				},
			},
			expected: "type AliasedName = Name;\n",
		},
		{
			name: "decimal type",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Price"): ast.CommonType{
						Type: ast.Decimal(),
					},
				},
			},
			expected: "type Price = __cedar::decimal;\n",
		},
		{
			name: "action with appliesTo and ipaddr context",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("User"))},
							ResourceTypes:  []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
							Context: ast.Record(ast.Attributes{
								"ip": ast.Attribute{Type: ast.IPAddr()},
							}),
						},
					},
				},
			},
			expected: `action view appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "ip": __cedar::ipaddr,
  }
};
`,
		},
		{
			name: "string with extended control character",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("test\u0085ext"),
						},
					},
				},
			},
			expected: "@doc(\"test\\x85ext\")\nentity User;\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := string(parser.MarshalSchema(tt.ast))
			testutil.Equals(t, result, tt.expected)
		})
	}
}
