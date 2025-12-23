package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestSchemaMarshalCedar(t *testing.T) {
	t.Parallel()

	t.Run("empty schema", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema()
		result := s.MarshalCedar()
		testutil.Equals(t, string(result), "")
	})

	t.Run("simple entity", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(ast.Entity("User"))
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "entity User;\n")
	})

	t.Run("entity with shape", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User").Shape(
				ast.Attribute("name", ast.String()),
				ast.Optional("email", ast.String()),
			),
		)
		result := string(s.MarshalCedar())
		expected := `entity User {
	name: String,
	email?: String,
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("entity with memberOf", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User").MemberOf(ast.EntityType("Group")),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "entity User in Group;\n")
	})

	t.Run("entity with multiple parents", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User").MemberOf(
				ast.EntityType("Group"),
				ast.EntityType("Team"),
			),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "entity User in [Group, Team];\n")
	})

	t.Run("entity with tags", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("Document").Tags(ast.String()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "entity Document tags String;\n")
	})

	t.Run("enum entity", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Enum("Status", "active", "inactive"),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `entity Status enum ["active", "inactive"];
`)
	})

	t.Run("common type", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Name", ast.String()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type Name = String;\n")
	})

	t.Run("common type with record", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Address", ast.Record(
				ast.Attribute("street", ast.String()),
				ast.Attribute("city", ast.String()),
			)),
		)
		result := string(s.MarshalCedar())
		expected := `type Address = {
	street: String,
	city: String,
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("action", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view"),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "action view;\n")
	})

	t.Run("action with appliesTo", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view").
				Principal(ast.EntityType("User")).
				Resource(ast.EntityType("Document")).
				Context(ast.Record(
					ast.Attribute("ip", ast.IPAddr()),
				)),
		)
		result := string(s.MarshalCedar())
		expected := `action view appliesTo {
	principal: User,
	resource: Document,
	context: {
		ip: __cedar::ipaddr,
	},
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("action with memberOf", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view").MemberOf(ast.UID("readActions")),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action view in "readActions";
`)
	})

	t.Run("action with quoted name", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view document"),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action "view document";
`)
	})

	t.Run("namespace", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Namespace("MyApp",
				ast.Entity("User"),
				ast.Entity("Document"),
			),
		)
		result := string(s.MarshalCedar())
		expected := `namespace MyApp {
	entity User;

	entity Document;
}
`
		testutil.Equals(t, result, expected)
	})

	t.Run("annotation", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User").Annotate("doc", "A user entity"),
		)
		result := string(s.MarshalCedar())
		expected := `@doc("A user entity")
entity User;
`
		testutil.Equals(t, result, expected)
	})

	t.Run("annotation without value", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User").Annotate("deprecated", ""),
		)
		result := string(s.MarshalCedar())
		expected := `@deprecated
entity User;
`
		testutil.Equals(t, result, expected)
	})

	t.Run("set type", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Tags", ast.Set(ast.String())),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type Tags = Set<String>;\n")
	})

	t.Run("extension types", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("IP", ast.IPAddr()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type IP = __cedar::ipaddr;\n")
	})

	t.Run("entity type reference", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("UserRef", ast.EntityType("MyApp::User")),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type UserRef = MyApp::User;\n")
	})

	t.Run("type reference", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("AliasedName", ast.Type("Name")),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type AliasedName = Name;\n")
	})

	t.Run("long type", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Count", ast.Long()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type Count = Long;\n")
	})

	t.Run("bool type", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Flag", ast.Bool()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type Flag = Bool;\n")
	})

	t.Run("action with explicit entity ref", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view").MemberOf(ast.EntityUID("MyApp::Action", "allActions")),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action view in MyApp::Action::"allActions";
`)
	})

	t.Run("record with quoted keys", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Data", ast.Record(
				ast.Attribute("special-key", ast.String()),
			)),
		)
		result := string(s.MarshalCedar())
		expected := `type Data = {
	"special-key": String,
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("empty record", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Empty", ast.Record()),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, "type Empty = {};\n")
	})

	t.Run("action with multiple principals", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view").
				Principal(ast.EntityType("User"), ast.EntityType("Admin")).
				Resource(ast.EntityType("Document")),
		)
		result := string(s.MarshalCedar())
		expected := `action view appliesTo {
	principal: [User, Admin],
	resource: Document,
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("action with multiple memberOf", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("view").MemberOf(
				ast.UID("readActions"),
				ast.UID("viewActions"),
			),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action view in ["readActions", "viewActions"];
`)
	})
}

func TestResolvedSchemaMarshalCedar(t *testing.T) {
	t.Parallel()

	t.Run("resolved schema marshal", func(t *testing.T) {
		t.Parallel()
		// ResolvedSchema has the same structure, so marshaling should work the same
		rs := &ast.ResolvedSchema{
			Nodes: []ast.IsNode{ast.Entity("User")},
		}
		result := string(rs.MarshalCedar())
		testutil.Equals(t, result, "entity User;\n")
	})

	t.Run("resolved schema with multiple nodes", func(t *testing.T) {
		t.Parallel()
		rs := &ast.ResolvedSchema{
			Nodes: []ast.IsNode{
				ast.Entity("User"),
				ast.Entity("Document"),
			},
		}
		result := string(rs.MarshalCedar())
		expected := `entity User;

entity Document;
`
		testutil.Equals(t, result, expected)
	})
}

func TestSchemaMultipleTopLevelNodes(t *testing.T) {
	t.Parallel()

	t.Run("multiple top-level entities", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User"),
			ast.Entity("Document"),
		)
		result := string(s.MarshalCedar())
		expected := `entity User;

entity Document;
`
		testutil.Equals(t, result, expected)
	})

	t.Run("multiple top-level mixed nodes", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Entity("User"),
			ast.Action("view"),
			ast.CommonType("Name", ast.String()),
		)
		result := string(s.MarshalCedar())
		expected := `entity User;

action view;

type Name = String;
`
		testutil.Equals(t, result, expected)
	})
}

func TestNeedsQuotingEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("action with empty name", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action(""),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action "";
`)
	})

	t.Run("action name starting with digit", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.Action("123action"),
		)
		result := string(s.MarshalCedar())
		testutil.Equals(t, result, `action "123action";
`)
	})

	t.Run("record key starting with digit", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Data", ast.Record(
				ast.Attribute("123key", ast.String()),
			)),
		)
		result := string(s.MarshalCedar())
		expected := `type Data = {
	"123key": String,
};
`
		testutil.Equals(t, result, expected)
	})

	t.Run("record key with empty name", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema(
			ast.CommonType("Data", ast.Record(
				ast.Attribute("", ast.String()),
			)),
		)
		result := string(s.MarshalCedar())
		expected := `type Data = {
	"": String,
};
`
		testutil.Equals(t, result, expected)
	})
}
