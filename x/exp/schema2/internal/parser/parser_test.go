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

	t.Run("empty schema", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(""))
		testutil.OK(t, err)
		testutil.Equals(t, len(schema.Nodes), 0)
	})

	t.Run("simple entity", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User;`))
		testutil.OK(t, err)
		testutil.Equals(t, len(schema.Nodes), 1)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, entity.Name, types.EntityType("User"))
	})

	t.Run("entity with shape", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User {
			name: String,
			age: Long,
		};`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, entity.ShapeVal != nil, true)
		testutil.Equals(t, len(entity.ShapeVal.Pairs), 2)
	})

	t.Run("entity with optional attributes", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User {
			name: String,
			email?: String,
		};`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, entity.ShapeVal.Pairs[0].Optional, false)
		testutil.Equals(t, entity.ShapeVal.Pairs[1].Optional, true)
	})

	t.Run("entity with memberOf", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User in [Group, Team];`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(entity.MemberOfVal), 2)
		testutil.Equals(t, entity.MemberOfVal[0].Name, types.EntityType("Group"))
		testutil.Equals(t, entity.MemberOfVal[1].Name, types.EntityType("Team"))
	})

	t.Run("entity with single memberOf", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User in Group;`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(entity.MemberOfVal), 1)
	})

	t.Run("entity with tags", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity Document tags String;`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		_, ok = entity.TagsVal.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("entity with equals sign", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity User = {
			name: String,
		};`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, entity.ShapeVal != nil, true)
	})

	t.Run("enum entity", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`entity Status enum ["active", "inactive", "pending"];`))
		testutil.OK(t, err)
		enum, ok := schema.Nodes[0].(*ast.EnumNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, enum.Name, types.EntityType("Status"))
		testutil.Equals(t, len(enum.Values), 3)
		testutil.Equals(t, enum.Values[0], types.String("active"))
	})

	t.Run("simple action", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view;`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, action.Name, types.String("view"))
	})

	t.Run("action with quoted name", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action "view document";`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, action.Name, types.String("view document"))
	})

	t.Run("action with appliesTo", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view appliesTo {
			principal: User,
			resource: Document,
		};`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, action.AppliesToVal != nil, true)
		testutil.Equals(t, len(action.AppliesToVal.PrincipalTypes), 1)
		testutil.Equals(t, len(action.AppliesToVal.ResourceTypes), 1)
	})

	t.Run("action with context", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view appliesTo {
			principal: User,
			resource: Document,
			context: { ip: __cedar::ipaddr },
		};`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, action.AppliesToVal.Context != nil, true)
	})

	t.Run("action with memberOf", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view in "readActions";`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(action.MemberOfVal), 1)
		testutil.Equals(t, action.MemberOfVal[0].ID, types.String("readActions"))
	})

	t.Run("action with memberOf list", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view in ["readActions", "viewActions"];`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(action.MemberOfVal), 2)
	})

	t.Run("common type", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Name = String;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ct.Name, types.Ident("Name"))
		_, ok = ct.Type.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("common type record", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Address = {
			street: String,
			city: String,
		};`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		rt, ok := ct.Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rt.Pairs), 2)
	})

	t.Run("namespace", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`namespace MyApp {
			entity User;
			entity Document;
		}`))
		testutil.OK(t, err)
		ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ns.Name, types.Path("MyApp"))
		testutil.Equals(t, len(ns.Declarations), 2)
	})

	t.Run("namespace with path", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`namespace MyApp::Core {
			entity User;
		}`))
		testutil.OK(t, err)
		ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ns.Name, types.Path("MyApp::Core"))
	})

	t.Run("annotation", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`@doc("A user entity")
entity User;`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(entity.Annotations), 1)
		testutil.Equals(t, entity.Annotations[0].Key, types.Ident("doc"))
		testutil.Equals(t, entity.Annotations[0].Value, types.String("A user entity"))
	})

	t.Run("annotation without value", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`@deprecated
entity User;`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(entity.Annotations), 1)
		testutil.Equals(t, entity.Annotations[0].Key, types.Ident("deprecated"))
		testutil.Equals(t, entity.Annotations[0].Value, types.String(""))
	})

	t.Run("multiple annotations", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`@doc("A user")
@deprecated
entity User;`))
		testutil.OK(t, err)
		entity, ok := schema.Nodes[0].(*ast.EntityNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(entity.Annotations), 2)
	})

	t.Run("set type", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Tags = Set<String>;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		st, ok := ct.Type.(ast.SetType)
		testutil.Equals(t, ok, true)
		_, ok = st.Element.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("extension types", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type IP = __cedar::ipaddr;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		et, ok := ct.Type.(ast.ExtensionType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, et.Name, types.Ident("ipaddr"))
	})

	t.Run("entity type reference", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type UserRef = MyApp::User;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		ref, ok := ct.Type.(ast.EntityTypeRef)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ref.Name, types.EntityType("MyApp::User"))
	})

	t.Run("common type reference", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type AliasedName = Name;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		ref, ok := ct.Type.(ast.TypeRef)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ref.Name, types.Path("Name"))
	})

	t.Run("nested record", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Contact = {
			address: {
				street: String,
				city: String,
			},
		};`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		rt, ok := ct.Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		inner, ok := rt.Pairs[0].Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(inner.Pairs), 2)
	})

	t.Run("complex schema", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`
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
`))
		testutil.OK(t, err)
		ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ns.Name, types.Path("MyApp"))
		testutil.Equals(t, len(ns.Declarations), 7) // type + 4 entities + 2 actions
	})
}

func TestParserErrors(t *testing.T) {
	t.Parallel()

	t.Run("unexpected token", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`foo bar;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing semicolon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace Foo { entity User;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Foo = 123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid namespace declaration", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace Foo { invalid; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing annotation paren", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`@doc("test" entity User;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing entity name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing enum bracket", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity Status enum "a", "b";`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid appliesTo content", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { invalid: User };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing Set element type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Tags = Set;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing common type equals", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Foo String;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing annotation identifier", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`@123 entity User;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid annotation string", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`@doc(123) entity User;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing namespace identifier", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace 123 { entity User; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing namespace open brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace Foo entity User; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing enum closing bracket", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity Status enum ["a", "b";`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid enum value", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity Status enum [123];`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing entity memberOf identifier", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User in 123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing entity memberOf list closing bracket", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User in [Group, Team;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing shape open brace pair", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User { name String };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid tags type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity Document tags 123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing action name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action 123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing action memberOf ref", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view in 123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing action memberOf list closing bracket", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view in ["read", "write";`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing appliesTo open brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo principal: User };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing principal colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { principal User };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing resource colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { resource User };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing context colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { context { ip: String } };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid principal type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { principal: 123 };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid resource type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { resource: 123 };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid context type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { context: 123 };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing appliesTo closing brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view appliesTo { principal: User;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing common type name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type = String;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid common type value", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Foo = ;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing set closing angle bracket", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Tags = Set<String;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid set element type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Tags = Set<123>;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing extension double colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type IP = __cedar ipaddr;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing extension name", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type IP = __cedar::123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing record closing brace", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Data = { name: String;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing record pair colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Data = { name String };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("invalid record pair type", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Data = { name: 123 };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing path after double colon", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Ref = MyApp::123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("missing entity ref path component", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view in MyApp::123;`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("namespace annotation error in declaration", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace MyApp { @123 entity User; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action with full entity ref in memberOf", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view in MyApp::Action::"allActions";`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, action.MemberOfVal[0].Type.Name, types.EntityType("MyApp::Action"))
	})

	t.Run("namespace type declaration error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace MyApp { type = String; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("namespace action error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`namespace MyApp { action 123; }`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("entity memberOf list parse error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User in [123];`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("entity shape parse error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User { 123: String };`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("action entity refs list error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`action view in [123];`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("common type semicolon error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`type Foo = String`))
		testutil.Equals(t, err != nil, true)
	})

	t.Run("record pair in shape error", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSchema([]byte(`entity User { ?: String };`))
		testutil.Equals(t, err != nil, true)
	})
}

func TestParseFromReader(t *testing.T) {
	t.Parallel()

	t.Run("NewFromReader nil", func(t *testing.T) {
		t.Parallel()
		_, err := NewFromReader(nil)
		// Should error with nil reader
		testutil.Equals(t, err != nil, true)
	})

	t.Run("NewFromReader valid", func(t *testing.T) {
		t.Parallel()
		p, err := NewFromReader(strings.NewReader("entity User;"))
		testutil.OK(t, err)
		schema, err := p.Parse()
		testutil.OK(t, err)
		testutil.Equals(t, len(schema.Nodes), 1)
	})
}

func TestParseAdditionalCases(t *testing.T) {
	t.Parallel()

	t.Run("action with explicit entity ref", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view in Action::"allActions";`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(action.MemberOfVal), 1)
		testutil.Equals(t, action.MemberOfVal[0].Type.Name, types.EntityType("Action"))
		testutil.Equals(t, action.MemberOfVal[0].ID, types.String("allActions"))
	})

	t.Run("record with quoted keys", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Data = {
			"special-key": String,
		};`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		rt, ok := ct.Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, rt.Pairs[0].Key, types.String("special-key"))
	})

	t.Run("bool type", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type Flag = Bool;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		_, ok = ct.Type.(ast.BoolType)
		testutil.Equals(t, ok, true)
	})

	t.Run("namespace annotation", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`@doc("My namespace")
namespace MyApp {
	entity User;
}`))
		testutil.OK(t, err)
		ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(ns.Annotations), 1)
	})

	t.Run("action with context type ref", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view appliesTo {
			principal: User,
			resource: Document,
			context: RequestContext,
		};`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		_, ok = action.AppliesToVal.Context.(ast.TypeRef)
		testutil.Equals(t, ok, true)
	})

	t.Run("action with multiple principals", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`action view appliesTo {
			principal: [User, Admin, ServiceAccount],
			resource: Document,
		};`))
		testutil.OK(t, err)
		action, ok := schema.Nodes[0].(*ast.ActionNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(action.AppliesToVal.PrincipalTypes), 3)
	})

	t.Run("set of entity refs", func(t *testing.T) {
		t.Parallel()
		schema, err := ParseSchema([]byte(`type UserSet = Set<MyApp::User>;`))
		testutil.OK(t, err)
		ct, ok := schema.Nodes[0].(*ast.CommonTypeNode)
		testutil.Equals(t, ok, true)
		st, ok := ct.Type.(ast.SetType)
		testutil.Equals(t, ok, true)
		_, ok = st.Element.(ast.EntityTypeRef)
		testutil.Equals(t, ok, true)
	})
}
