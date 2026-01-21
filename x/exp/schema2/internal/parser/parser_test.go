package parser

import (
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestParseSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantNodes int
		validate  func(t *testing.T, schema *ast.Schema)
	}{
		{
			name:      "empty schema",
			input:     "",
			wantNodes: 0,
		},
		{
			name:      "simple entity",
			input:     `entity User;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, entity.Name, types.EntityType("User"))
			},
		},
		{
			name: "entity with shape",
			input: `entity User {
				name: String,
				age: Long,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, entity.ShapeVal != nil, true)
				testutil.Equals(t, len(entity.ShapeVal.Pairs), 2)
			},
		},
		{
			name: "entity with optional attributes",
			input: `entity User {
				name: String,
				email?: String,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, entity.ShapeVal.Pairs[0].Optional, false)
				testutil.Equals(t, entity.ShapeVal.Pairs[1].Optional, true)
			},
		},
		{
			name:      "entity with memberOf",
			input:     `entity User in [Group, Team];`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(entity.MemberOfVal), 2)
				testutil.Equals(t, entity.MemberOfVal[0].Name, types.EntityType("Group"))
				testutil.Equals(t, entity.MemberOfVal[1].Name, types.EntityType("Team"))
			},
		},
		{
			name:      "entity with single memberOf",
			input:     `entity User in Group;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(entity.MemberOfVal), 1)
			},
		},
		{
			name:      "entity with tags",
			input:     `entity Document tags String;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				_, ok = entity.TagsVal.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name: "entity with equals sign",
			input: `entity User = {
				name: String,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, entity.ShapeVal != nil, true)
			},
		},
		{
			name:      "enum entity",
			input:     `entity Status enum ["active", "inactive", "pending"];`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				enum, ok := schema.Nodes[0].(ast.EnumNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, enum.Name, types.EntityType("Status"))
				testutil.Equals(t, len(enum.Values), 3)
				testutil.Equals(t, enum.Values[0], types.String("active"))
			},
		},
		{
			name:      "simple action",
			input:     `action view;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, action.Name, types.String("view"))
			},
		},
		{
			name:      "action with quoted name",
			input:     `action "view document";`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, action.Name, types.String("view document"))
			},
		},
		{
			name: "action with appliesTo",
			input: `action view appliesTo {
				principal: User,
				resource: Document,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, action.AppliesToVal != nil, true)
				testutil.Equals(t, len(action.AppliesToVal.PrincipalTypes), 1)
				testutil.Equals(t, len(action.AppliesToVal.ResourceTypes), 1)
			},
		},
		{
			name: "action with context",
			input: `action view appliesTo {
				principal: User,
				resource: Document,
				context: { ip: __cedar::ipaddr },
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, action.AppliesToVal.Context != nil, true)
			},
		},
		{
			name:      "action with memberOf",
			input:     `action view in "readActions";`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(action.MemberOfVal), 1)
				testutil.Equals(t, action.MemberOfVal[0].ID, types.String("readActions"))
			},
		},
		{
			name:      "action with memberOf list",
			input:     `action view in ["readActions", "viewActions"];`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(action.MemberOfVal), 2)
			},
		},
		{
			name:      "common type",
			input:     `type Name = String;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ct.Name, types.Ident("Name"))
				_, ok = ct.Type.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name: "common type record",
			input: `type Address = {
				street: String,
				city: String,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				rt, ok := ct.Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(rt.Pairs), 2)
			},
		},
		{
			name: "namespace",
			input: `namespace MyApp {
				entity User;
				entity Document;
			}`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ns, ok := schema.Nodes[0].(ast.NamespaceNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ns.Name, types.Path("MyApp"))
				testutil.Equals(t, len(ns.Declarations), 2)
			},
		},
		{
			name: "namespace with path",
			input: `namespace MyApp::Core {
				entity User;
			}`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ns, ok := schema.Nodes[0].(ast.NamespaceNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ns.Name, types.Path("MyApp::Core"))
			},
		},
		{
			name: "annotation",
			input: `@doc("A user entity")
entity User;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(entity.Annotations), 1)
				testutil.Equals(t, entity.Annotations[0].Key, types.Ident("doc"))
				testutil.Equals(t, entity.Annotations[0].Value, types.String("A user entity"))
			},
		},
		{
			name: "annotation without value",
			input: `@deprecated
entity User;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(entity.Annotations), 1)
				testutil.Equals(t, entity.Annotations[0].Key, types.Ident("deprecated"))
				testutil.Equals(t, entity.Annotations[0].Value, types.String(""))
			},
		},
		{
			name: "multiple annotations",
			input: `@doc("A user")
@deprecated
entity User;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				entity, ok := schema.Nodes[0].(ast.EntityNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(entity.Annotations), 2)
			},
		},
		{
			name:      "set type",
			input:     `type Tags = Set<String>;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				st, ok := ct.Type.(ast.SetType)
				testutil.Equals(t, ok, true)
				_, ok = st.Element.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name:      "extension types",
			input:     `type IP = __cedar::ipaddr;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				et, ok := ct.Type.(ast.ExtensionType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, et.Name, types.Ident("ipaddr"))
			},
		},
		{
			name:      "entity type reference",
			input:     `type UserRef = MyApp::User;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				ref, ok := ct.Type.(ast.EntityTypeRef)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ref.Name, types.EntityType("MyApp::User"))
			},
		},
		{
			name:      "common type reference",
			input:     `type AliasedName = Name;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				ref, ok := ct.Type.(ast.TypeRef)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ref.Name, types.Path("Name"))
			},
		},
		{
			name: "nested record",
			input: `type Contact = {
				address: {
					street: String,
					city: String,
				},
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				rt, ok := ct.Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				inner, ok := rt.Pairs[0].Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(inner.Pairs), 2)
			},
		},
		{
			name: "bool type",
			input: `type Flag = Bool;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				_, ok = ct.Type.(ast.BoolType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name: "namespace annotation",
			input: `@doc("My namespace")
namespace MyApp {
	entity User;
}`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ns, ok := schema.Nodes[0].(ast.NamespaceNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(ns.Annotations), 1)
			},
		},
		{
			name: "action with context type ref",
			input: `action view appliesTo {
				principal: User,
				resource: Document,
				context: RequestContext,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				_, ok = action.AppliesToVal.Context.(ast.TypeRef)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name: "action with multiple principals",
			input: `action view appliesTo {
				principal: [User, Admin, ServiceAccount],
				resource: Document,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(action.AppliesToVal.PrincipalTypes), 3)
			},
		},
		{
			name:      "set of entity refs",
			input:     `type UserSet = Set<MyApp::User>;`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				st, ok := ct.Type.(ast.SetType)
				testutil.Equals(t, ok, true)
				_, ok = st.Element.(ast.EntityTypeRef)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name:      "action with explicit entity ref",
			input:     `action view in Action::"allActions";`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(action.MemberOfVal), 1)
				testutil.Equals(t, action.MemberOfVal[0].Type.Name, types.EntityType("Action"))
				testutil.Equals(t, action.MemberOfVal[0].ID, types.String("allActions"))
			},
		},
		{
			name:      "action with full entity ref in memberOf",
			input:     `action view in MyApp::Action::"allActions";`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				action, ok := schema.Nodes[0].(ast.ActionNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, action.MemberOfVal[0].Type.Name, types.EntityType("MyApp::Action"))
			},
		},
		{
			name: "record with quoted keys",
			input: `type Data = {
				"special-key": String,
			};`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ct, ok := schema.Nodes[0].(ast.CommonTypeNode)
				testutil.Equals(t, ok, true)
				rt, ok := ct.Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, rt.Pairs[0].Key, types.String("special-key"))
			},
		},
		{
			name: "complex schema",
			input: `
namespace MyApp {
	type Name = String;

	@doc("A user in the system")
	entity User in [Group, Team] {
		name: Name,
		email?: String,
	};

	entity Group;

	entity Team;

	entity Document {
		title: String,
		owner: MyApp::User,
	} tags String;

	action view appliesTo {
		principal: [User, Group],
		resource: Document,
		context: { ip?: __cedar::ipaddr },
	};

	action edit in "view" appliesTo {
		principal: User,
		resource: Document,
	};
}
`,
			wantNodes: 1,
			validate: func(t *testing.T, schema *ast.Schema) {
				ns, ok := schema.Nodes[0].(ast.NamespaceNode)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ns.Name, types.Path("MyApp"))
				testutil.Equals(t, len(ns.Declarations), 7)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			schema, err := ParseSchema("", []byte(tt.input))
			testutil.OK(t, err)
			testutil.Equals(t, len(schema.Nodes), tt.wantNodes)
			if tt.validate != nil {
				tt.validate(t, schema)
			}
		})
	}
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{"unexpected token", `foo bar;`},
		{"missing semicolon", `entity User`},
		{"missing brace", `namespace Foo { entity User;`},
		{"invalid type", `type Foo = 123;`},
		{"invalid namespace declaration", `namespace Foo { invalid; }`},
		{"missing annotation paren", `@doc("test" entity User;`},
		{"missing entity name", `entity;`},
		{"missing enum bracket", `entity Status enum "a", "b";`},
		{"invalid appliesTo content", `action view appliesTo { invalid: User };`},
		{"missing Set element type", `type Tags = Set;`},
		{"missing common type equals", `type Foo String;`},
		{"missing annotation identifier", `@123 entity User;`},
		{"invalid annotation string", `@doc(123) entity User;`},
		{"missing namespace identifier", `namespace 123 { entity User; }`},
		{"missing namespace open brace", `namespace Foo entity User; }`},
		{"missing enum closing bracket", `entity Status enum ["a", "b";`},
		{"invalid enum value", `entity Status enum [123];`},
		{"missing entity memberOf identifier", `entity User in 123;`},
		{"missing entity memberOf list closing bracket", `entity User in [Group, Team;`},
		{"missing shape open brace pair", `entity User { name String };`},
		{"invalid tags type", `entity Document tags 123;`},
		{"missing action name", `action 123;`},
		{"missing action memberOf ref", `action view in 123;`},
		{"missing action memberOf list closing bracket", `action view in ["read", "write";`},
		{"missing appliesTo open brace", `action view appliesTo principal: User };`},
		{"missing principal colon", `action view appliesTo { principal User };`},
		{"missing resource colon", `action view appliesTo { resource User };`},
		{"missing context colon", `action view appliesTo { context { ip: String } };`},
		{"invalid principal type", `action view appliesTo { principal: 123 };`},
		{"invalid resource type", `action view appliesTo { resource: 123 };`},
		{"invalid context type", `action view appliesTo { context: 123 };`},
		{"missing appliesTo closing brace", `action view appliesTo { principal: User;`},
		{"missing common type name", `type = String;`},
		{"invalid common type value", `type Foo = ;`},
		{"missing set closing angle bracket", `type Tags = Set<String;`},
		{"invalid set element type", `type Tags = Set<123>;`},
		{"missing extension double colon", `type IP = __cedar ipaddr;`},
		{"missing extension name", `type IP = __cedar::123;`},
		{"missing record closing brace", `type Data = { name: String;`},
		{"missing record pair colon", `type Data = { name String };`},
		{"invalid record pair type", `type Data = { name: 123 };`},
		{"missing path after double colon", `type Ref = MyApp::123;`},
		{"missing entity ref path component", `action view in MyApp::123;`},
		{"namespace annotation error in declaration", `namespace MyApp { @123 entity User; }`},
		{"namespace type declaration error", `namespace MyApp { type = String; }`},
		{"namespace action error", `namespace MyApp { action 123; }`},
		{"entity memberOf list parse error", `entity User in [123];`},
		{"entity shape parse error", `entity User { 123: String };`},
		{"action entity refs list error", `action view in [123];`},
		{"common type semicolon error", `type Foo = String`},
		{"record pair in shape error", `entity User { ?: String };`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseSchema("", []byte(tt.input))
			testutil.Equals(t, err != nil, true)
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
		testutil.Equals(t, len(schema.Nodes), 1)
	})
}
