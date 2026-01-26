package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/parser"
)

func TestSchemaMarshalCedar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   *ast.Schema
		expected string
	}{
		{
			name:     "empty schema",
			schema:   ast.NewSchema(),
			expected: "",
		},
		{
			name:     "simple entity",
			schema:   ast.NewSchema(ast.Entity("User")),
			expected: "entity User;\n",
		},
		{
			name: "entity with shape",
			schema: ast.NewSchema(
				ast.Entity("User").Shape(
					ast.Attribute("name", ast.String()),
					ast.Optional("email", ast.String()),
				),
			),
			expected: `entity User = {"name": String, "email?": String};
`,
		},
		{
			name: "entity with memberOf",
			schema: ast.NewSchema(
				ast.Entity("User").MemberOf(ast.EntityType("Group")),
			),
			expected: "entity User in Group;\n",
		},
		{
			name: "entity with multiple parents",
			schema: ast.NewSchema(
				ast.Entity("User").MemberOf(
					ast.EntityType("Group"),
					ast.EntityType("Team"),
				),
			),
			expected: "entity User in [Group, Team];\n",
		},
		{
			name: "entity with tags",
			schema: ast.NewSchema(
				ast.Entity("Document").Tags(ast.String()),
			),
			expected: "entity Document tags String;\n",
		},
		{
			name: "enum entity",
			schema: ast.NewSchema(
				ast.Enum("Status", "active", "inactive"),
			),
			expected: `entity Status enum ["active", "inactive"];
`,
		},
		{
			name: "common type",
			schema: ast.NewSchema(
				ast.CommonType("Name", ast.String()),
			),
			expected: "type Name = String;\n",
		},
		{
			name: "common type with record",
			schema: ast.NewSchema(
				ast.CommonType("Address", ast.Record(
					ast.Attribute("street", ast.String()),
					ast.Attribute("city", ast.String()),
				)),
			),
			expected: `type Address = {
  "street": String,
  "city": String,
};
`,
		},
		{
			name: "action",
			schema: ast.NewSchema(
				ast.Action("view"),
			),
			expected: "action view;\n",
		},
		{
			name: "action with appliesTo",
			schema: ast.NewSchema(
				ast.Action("view").
					Principal(ast.EntityType("User")).
					Resource(ast.EntityType("Document")).
					Context(ast.Record(
						ast.Attribute("ip", ast.IPAddr()),
					)),
			),
			expected: `action view appliesTo {
  principal: User,
  resource: Document,
  context: {
    "ip": __cedar::ipaddr,
  },
};
`,
		},
		{
			name: "action with memberOf",
			schema: ast.NewSchema(
				ast.Action("view").MemberOf(ast.UID("readActions")),
			),
			expected: `action view in "readActions";
`,
		},
		{
			name: "action with quoted name",
			schema: ast.NewSchema(
				ast.Action("view document"),
			),
			expected: `action "view document";
`,
		},
		{
			name: "namespace",
			schema: ast.NewSchema(
				ast.Namespace("MyApp",
					ast.Entity("User"),
					ast.Entity("Document"),
				),
			),
			expected: `namespace MyApp {
  entity User;

  entity Document;
}
`,
		},
		{
			name: "annotation",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "A user entity"),
			),
			expected: `@doc("A user entity")
entity User;
`,
		},
		{
			name: "annotation without value",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("deprecated", ""),
			),
			expected: `@deprecated
entity User;
`,
		},
		{
			name: "set type",
			schema: ast.NewSchema(
				ast.CommonType("Tags", ast.Set(ast.String())),
			),
			expected: "type Tags = Set<String>;\n",
		},
		{
			name: "extension types",
			schema: ast.NewSchema(
				ast.CommonType("IP", ast.IPAddr()),
			),
			expected: "type IP = __cedar::ipaddr;\n",
		},
		{
			name: "entity type reference",
			schema: ast.NewSchema(
				ast.CommonType("UserRef", ast.EntityType("MyApp::User")),
			),
			expected: "type UserRef = MyApp::User;\n",
		},
		{
			name: "type reference",
			schema: ast.NewSchema(
				ast.CommonType("AliasedName", ast.Type("Name")),
			),
			expected: "type AliasedName = Name;\n",
		},
		{
			name: "long type",
			schema: ast.NewSchema(
				ast.CommonType("Count", ast.Long()),
			),
			expected: "type Count = Long;\n",
		},
		{
			name: "bool type",
			schema: ast.NewSchema(
				ast.CommonType("Flag", ast.Bool()),
			),
			expected: "type Flag = Bool;\n",
		},
		{
			name: "action with explicit entity ref",
			schema: ast.NewSchema(
				ast.Action("view").MemberOf(ast.EntityUID("MyApp::Action", "allActions")),
			),
			expected: `action view in MyApp::Action::"allActions";
`,
		},
		{
			name: "record with quoted keys",
			schema: ast.NewSchema(
				ast.CommonType("Data", ast.Record(
					ast.Attribute("special-key", ast.String()),
				)),
			),
			expected: `type Data = {
  "special-key": String,
};
`,
		},
		{
			name: "empty record",
			schema: ast.NewSchema(
				ast.CommonType("Empty", ast.Record()),
			),
			expected: "type Empty = {};\n",
		},
		{
			name: "nested empty record",
			schema: ast.NewSchema(
				ast.CommonType("Outer", ast.Record(
					ast.Attribute("inner", ast.Record()),
				)),
			),
			expected: `type Outer = {
  "inner": {},
};
`,
		},
		{
			name: "nested record with optional field",
			schema: ast.NewSchema(
				ast.CommonType("Outer", ast.Record(
					ast.Attribute("inner", ast.Record(
						ast.Optional("optField", ast.String()),
					)),
				)),
			),
			expected: `type Outer = {
  "inner": {
    "optField"?: String,
  },
};
`,
		},
		{
			name: "action with multiple principals",
			schema: ast.NewSchema(
				ast.Action("view").
					Principal(ast.EntityType("User"), ast.EntityType("Admin")).
					Resource(ast.EntityType("Document")),
			),
			expected: `action view appliesTo {
  principal: [User, Admin],
  resource: Document,
  context: {},
};
`,
		},
		{
			name: "action with multiple memberOf",
			schema: ast.NewSchema(
				ast.Action("view").MemberOf(
					ast.UID("readActions"),
					ast.UID("viewActions"),
				),
			),
			expected: `action view in ["readActions", "viewActions"];
`,
		},
		{
			name: "multiple top-level entities",
			schema: ast.NewSchema(
				ast.Entity("User"),
				ast.Entity("Document"),
			),
			expected: `entity User;

entity Document;
`,
		},
		{
			name: "multiple top-level mixed nodes",
			schema: ast.NewSchema(
				ast.Entity("User"),
				ast.Action("view"),
				ast.CommonType("Name", ast.String()),
			),
			expected: `entity User;

action view;

type Name = String;
`,
		},
		{
			name: "action with empty name",
			schema: ast.NewSchema(
				ast.Action(""),
			),
			expected: `action "";
`,
		},
		{
			name: "action name starting with digit",
			schema: ast.NewSchema(
				ast.Action("123action"),
			),
			expected: `action "123action";
`,
		},
		{
			name: "record key starting with digit",
			schema: ast.NewSchema(
				ast.CommonType("Data", ast.Record(
					ast.Attribute("123key", ast.String()),
				)),
			),
			expected: `type Data = {
  "123key": String,
};
`,
		},
		{
			name: "record key with empty name",
			schema: ast.NewSchema(
				ast.CommonType("Data", ast.Record(
					ast.Attribute("", ast.String()),
				)),
			),
			expected: `type Data = {
  "": String,
};
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := string(parser.MarshalSchema(tt.schema))
			testutil.Equals(t, result, tt.expected)
		})
	}
}
