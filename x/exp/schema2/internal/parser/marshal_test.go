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
			expected: `entity User = {
  "name": String,
  "email"?: String,
};
`,
		},
		{
			name: "entity with memberOf",
			schema: ast.NewSchema(
				ast.Entity("User").MemberOf(ast.EntityType("Group")),
			),
			expected: "entity User in [Group];\n",
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
  principal: [User],
  resource: [Document],
  context: {
    "ip": __cedar::ipaddr,
  }
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
  resource: [Document],
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
		{
			name: "action name with reserved keyword 'in'",
			schema: ast.NewSchema(
				ast.Action("in"),
			),
			expected: `action "in";
`,
		},
		{
			name: "string with newline escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "Line1\nLine2"),
			),
			expected: `@doc("Line1\nLine2")
entity User;
`,
		},
		{
			name: "string with carriage return escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "Line1\rLine2"),
			),
			expected: `@doc("Line1\rLine2")
entity User;
`,
		},
		{
			name: "string with tab escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "Col1\tCol2"),
			),
			expected: `@doc("Col1\tCol2")
entity User;
`,
		},
		{
			name: "string with backslash escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "path\\to\\file"),
			),
			expected: `@doc("path\\to\\file")
entity User;
`,
		},
		{
			name: "string with null character escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "null\x00char"),
			),
			expected: `@doc("null\0char")
entity User;
`,
		},
		{
			name: "string with single quote escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "It's fine"),
			),
			expected: "@doc(\"It\\'s fine\")\nentity User;\n",
		},
		{
			name: "string with control character",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "test\x01control"),
			),
			expected: "@doc(\"test\\x01control\")\nentity User;\n",
		},
		{
			name: "string with DEL character",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "test\x7Fdel"),
			),
			expected: "@doc(\"test\\x7fdel\")\nentity User;\n",
		},
		{
			name: "string with extended control character",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", "test\u0085ext"),
			),
			expected: "@doc(\"test\\x85ext\")\nentity User;\n",
		},
		{
			name: "namespace with annotations",
			schema: ast.NewSchema(
				ast.Namespace("MyApp",
					ast.Entity("User"),
				).Annotate("doc", "My application namespace"),
			),
			expected: `@doc("My application namespace")
namespace MyApp {
  entity User;
}
`,
		},
		{
			name: "enum with annotations",
			schema: ast.NewSchema(
				ast.Enum("Status", "active", "inactive").Annotate("doc", "Status values"),
			),
			expected: `@doc("Status values")
entity Status enum ["active", "inactive"];
`,
		},
		{
			name: "common type with annotations",
			schema: ast.NewSchema(
				ast.CommonType("Name", ast.String()).Annotate("doc", "A name type"),
			),
			expected: `@doc("A name type")
type Name = String;
`,
		},
		{
			name: "action with annotations",
			schema: ast.NewSchema(
				ast.Action("view").Annotate("doc", "View action"),
			),
			expected: `@doc("View action")
action view;
`,
		},
		{
			name: "record with annotated attribute",
			schema: ast.NewSchema(
				ast.CommonType("User", ast.Record(
					ast.Attribute("name", ast.String()).Annotate(ast.Annotation{Key: "doc", Value: "User's name"}),
				)),
			),
			expected: "type User = {\n  @doc(\"User\\'s name\")\n  \"name\": String,\n};\n",
		},
		{
			name: "entity ref without type",
			schema: ast.NewSchema(
				ast.Action("view").MemberOf(ast.EntityUID("", "someAction")),
			),
			expected: `action view in "someAction";
`,
		},
		{
			name: "action with only context in appliesTo",
			schema: ast.NewSchema(
				ast.Action("view").Context(ast.Record(
					ast.Attribute("timestamp", ast.Long()),
				)),
			),
			expected: `action view appliesTo {
  context: {
    "timestamp": Long,
  }
};
`,
		},
		{
			name: "action with only principal in appliesTo",
			schema: ast.NewSchema(
				ast.Action("view").Principal(ast.EntityType("User")),
			),
			expected: `action view appliesTo {
  principal: [User],
};
`,
		},
		{
			name: "action with only resource in appliesTo",
			schema: ast.NewSchema(
				ast.Action("view").Resource(ast.EntityType("Document")),
			),
			expected: `action view appliesTo {
  resource: [Document],
};
`,
		},
		{
			name: "entity with shape and memberOf",
			schema: ast.NewSchema(
				ast.Entity("User").
					MemberOf(ast.EntityType("Group")).
					Shape(ast.Attribute("name", ast.String())),
			),
			expected: `entity User in [Group] = {
  "name": String,
};
`,
		},
		{
			name: "entity with shape and tags",
			schema: ast.NewSchema(
				ast.Entity("Document").
					Shape(ast.Attribute("title", ast.String())).
					Tags(ast.String()),
			),
			expected: `entity Document = {
  "title": String,
} tags String;
`,
		},
		{
			name: "entity with memberOf and tags",
			schema: ast.NewSchema(
				ast.Entity("Document").
					MemberOf(ast.EntityType("Folder")).
					Tags(ast.String()),
			),
			expected: `entity Document in [Folder] tags String;
`,
		},
		{
			name: "entity with all features",
			schema: ast.NewSchema(
				ast.Entity("Document").
					MemberOf(ast.EntityType("Folder")).
					Shape(ast.Attribute("title", ast.String())).
					Tags(ast.String()),
			),
			expected: `entity Document in [Folder] = {
  "title": String,
} tags String;
`,
		},
		{
			name: "decimal type",
			schema: ast.NewSchema(
				ast.CommonType("Price", ast.Decimal()),
			),
			expected: "type Price = __cedar::decimal;\n",
		},
		{
			name: "string with double quote escape",
			schema: ast.NewSchema(
				ast.Entity("User").Annotate("doc", `He said "hello"`),
			),
			expected: "@doc(\"He said \\\"hello\\\"\")\nentity User;\n",
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
