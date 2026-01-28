package parser

import (
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// TestParseNormalizations tests that the parser accepts alternative syntaxes
// that get normalized during marshaling (e.g., "in Group" becomes "in [Group]")
func TestParseNormalizations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, schema *ast.Schema)
	}{
		{
			name:  "entity with single memberOf without brackets",
			input: `entity User in Group;`,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity := schema.Entities[types.EntityType("User")]
				testutil.Equals(t, len(entity.MemberOfVal), 1)
			},
		},
		{
			name: "entity with equals sign (normalized to without)",
			input: `entity User = {
				name: String,
			};`,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity := schema.Entities[types.EntityType("User")]
				testutil.Equals(t, entity.ShapeVal != nil, true)
			},
		},
		{
			name: "action with appliesTo principal without brackets",
			input: `action view appliesTo {
				principal: User,
				resource: Document,
			};`,
			validate: func(t *testing.T, schema *ast.Schema) {
				action := schema.Actions[types.String("view")]
				testutil.Equals(t, action.AppliesToVal != nil, true)
				testutil.Equals(t, len(action.AppliesToVal.PrincipalTypes), 1)
				testutil.Equals(t, len(action.AppliesToVal.ResourceTypes), 1)
			},
		},
		{
			name:  "action with single memberOf (without brackets)",
			input: `action view in "readActions";`,
			validate: func(t *testing.T, schema *ast.Schema) {
				action := schema.Actions[types.String("view")]
				testutil.Equals(t, len(action.MemberOfVal), 1)
				testutil.Equals(t, action.MemberOfVal[0].ID, types.String("readActions"))
			},
		},
		{
			name: "action edit with single parent in memberOf",
			input: `action edit in "view" appliesTo {
				principal: User,
				resource: Document,
			};`,
			validate: func(t *testing.T, schema *ast.Schema) {
				action := schema.Actions[types.String("edit")]
				testutil.Equals(t, len(action.MemberOfVal), 1)
				testutil.Equals(t, action.MemberOfVal[0].ID, types.String("view"))
			},
		},
		{
			name: "namespace with path",
			input: `namespace MyApp::Core {
				entity User;
			}`,
			validate: func(t *testing.T, schema *ast.Schema) {
				_, ok := schema.Namespaces[types.Path("MyApp::Core")]
				testutil.Equals(t, ok, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema, err := ParseSchema("", []byte(tt.input))
			testutil.OK(t, err)
			if tt.validate != nil {
				tt.validate(t, schema)
			}
		})
	}
}

func TestParseFromReader(t *testing.T) {
	t.Parallel()

	t.Run("NewFromReader nil", func(t *testing.T) {
		t.Parallel()
		_, err := NewFromReader("", nil)
		testutil.Equals(t, err != nil, true)
	})

	t.Run("NewFromReader valid", func(t *testing.T) {
		t.Parallel()
		p, err := NewFromReader("", strings.NewReader("entity User;"))
		testutil.OK(t, err)
		schema, err := p.Parse()
		testutil.OK(t, err)
		testutil.Equals(t, len(schema.Entities), 1)
	})
}
