package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
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
			schema:   &ast.Schema{},
			expected: "",
		},
		{
			name: "simple entity",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{},
				},
			},
			expected: "entity User;\n",
		},
		{
			name: "entity with shape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						ShapeVal: &ast.RecordType{
							Attributes: ast.Attributes{
								"email": ast.Attribute{Type: ast.String(), Optional: true},
								"name":  ast.Attribute{Type: ast.String(), Optional: false},
							},
						},
					},
				},
			},
			expected: `entity User = {
  "email"?: String,
  "name": String,
};
`,
		},
		{
			name: "entity with memberOf",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOfVal: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Group"))},
					},
				},
			},
			expected: "entity User in [Group];\n",
		},
		{
			name: "entity with multiple parents",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOfVal: []ast.EntityTypeRef{
							ast.EntityType(types.EntityType("Group")),
							ast.EntityType(types.EntityType("Team")),
						},
					},
				},
			},
			expected: "entity User in [Group, Team];\n",
		},
		{
			name: "entity with tags",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						TagsVal: ast.String(),
					},
				},
			},
			expected: "entity Document tags String;\n",
		},
		{
			name: "enum entity",
			schema: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{
						Values: []types.String{"active", "inactive"},
					},
				},
			},
			expected: `entity Status enum ["active", "inactive"];
`,
		},
		{
			name: "common type",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"): ast.CommonType{
						Type: ast.String(),
					},
				},
			},
			expected: "type Name = String;\n",
		},
		{
			name: "common type with record",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Address"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"city":   ast.Attribute{Type: ast.String()},
							"street": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
			expected: `type Address = {
  "city": String,
  "street": String,
};
`,
		},
		{
			name: "action",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{},
				},
			},
			expected: "action view;\n",
		},
		{
			name: "action with appliesTo",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesToVal: &ast.AppliesTo{
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
			name: "action with memberOf",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOfVal: []ast.EntityRef{ast.EntityRefFromID("readActions")},
					},
				},
			},
			expected: `action view in "readActions";
`,
		},
		{
			name: "action with quoted name",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view document"): ast.Action{},
				},
			},
			expected: `action "view document";
`,
		},
		{
			name: "namespace",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Entities: ast.Entities{
							types.EntityType("Document"): ast.Entity{},
							types.EntityType("User"):     ast.Entity{},
						},
					},
				},
			},
			expected: `namespace MyApp {
  entity Document;

  entity User;
}
`,
		},
		{
			name: "annotation",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("A user entity"),
						},
					},
				},
			},
			expected: `@doc("A user entity")
entity User;
`,
		},
		{
			name: "annotation without value",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("deprecated"): types.String(""),
						},
					},
				},
			},
			expected: `@deprecated
entity User;
`,
		},
		{
			name: "set type",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Tags"): ast.CommonType{
						Type: ast.Set(ast.String()),
					},
				},
			},
			expected: "type Tags = Set<String>;\n",
		},
		{
			name: "extension types",
			schema: &ast.Schema{
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
			schema: &ast.Schema{
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
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("AliasedName"): ast.CommonType{
						Type: ast.Type("Name"),
					},
				},
			},
			expected: "type AliasedName = Name;\n",
		},
		{
			name: "long type",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Count"): ast.CommonType{
						Type: ast.Long(),
					},
				},
			},
			expected: "type Count = Long;\n",
		},
		{
			name: "bool type",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Flag"): ast.CommonType{
						Type: ast.Bool(),
					},
				},
			},
			expected: "type Flag = Bool;\n",
		},
		{
			name: "action with explicit entity ref",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOfVal: []ast.EntityRef{ast.NewEntityRef(types.EntityType("MyApp::Action"), "allActions")},
					},
				},
			},
			expected: `action view in MyApp::Action::"allActions";
`,
		},
		{
			name: "record with quoted keys",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"special-key": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
			expected: `type Data = {
  "special-key": String,
};
`,
		},
		{
			name: "empty record",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Empty"): ast.CommonType{
						Type: ast.Record(ast.Attributes{}),
					},
				},
			},
			expected: "type Empty = {};\n",
		},
		{
			name: "nested empty record",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Outer"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"inner": ast.Attribute{Type: ast.Record(ast.Attributes{})},
						}),
					},
				},
			},
			expected: `type Outer = {
  "inner": {},
};
`,
		},
		{
			name: "nested record with optional field",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Outer"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"inner": ast.Attribute{
								Type: ast.Record(ast.Attributes{
									"optField": ast.Attribute{Type: ast.String(), Optional: true},
								}),
							},
						}),
					},
				},
			},
			expected: `type Outer = {
  "inner": {
    "optField"?: String,
  },
};
`,
		},
		{
			name: "action with multiple principals",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesToVal: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{
								ast.EntityType(types.EntityType("Admin")),
								ast.EntityType(types.EntityType("User")),
							},
							ResourceTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
						},
					},
				},
			},
			expected: `action view appliesTo {
  principal: [Admin, User],
  resource: [Document],
};
`,
		},
		{
			name: "action with multiple memberOf",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOfVal: []ast.EntityRef{
							ast.EntityRefFromID("readActions"),
							ast.EntityRefFromID("viewActions"),
						},
					},
				},
			},
			expected: `action view in ["readActions", "viewActions"];
`,
		},
		{
			name: "multiple top-level entities",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{},
					types.EntityType("User"):     ast.Entity{},
				},
			},
			expected: `entity Document;

entity User;
`,
		},
		{
			name: "multiple top-level mixed nodes",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{},
				},
				Actions: ast.Actions{
					types.String("view"): ast.Action{},
				},
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"): ast.CommonType{
						Type: ast.String(),
					},
				},
			},
			expected: `type Name = String;

entity User;

action view;
`,
		},
		{
			name: "action with empty name",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String(""): ast.Action{},
				},
			},
			expected: `action "";
`,
		},
		{
			name: "action name starting with digit",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("123action"): ast.Action{},
				},
			},
			expected: `action "123action";
`,
		},
		{
			name: "record key starting with digit",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"123key": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
			expected: `type Data = {
  "123key": String,
};
`,
		},
		{
			name: "record key with empty name",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
			expected: `type Data = {
  "": String,
};
`,
		},
		{
			name: "action name with reserved keyword 'in'",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("in"): ast.Action{},
				},
			},
			expected: `action "in";
`,
		},
		{
			name: "string with newline escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Line1\nLine2"),
						},
					},
				},
			},
			expected: `@doc("Line1\nLine2")
entity User;
`,
		},
		{
			name: "string with carriage return escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Line1\rLine2"),
						},
					},
				},
			},
			expected: `@doc("Line1\rLine2")
entity User;
`,
		},
		{
			name: "string with tab escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Col1\tCol2"),
						},
					},
				},
			},
			expected: `@doc("Col1\tCol2")
entity User;
`,
		},
		{
			name: "string with backslash escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("path\\to\\file"),
						},
					},
				},
			},
			expected: `@doc("path\\to\\file")
entity User;
`,
		},
		{
			name: "string with null character escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("null\x00char"),
						},
					},
				},
			},
			expected: `@doc("null\0char")
entity User;
`,
		},
		{
			name: "string with single quote escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("It's fine"),
						},
					},
				},
			},
			expected: "@doc(\"It\\'s fine\")\nentity User;\n",
		},
		{
			name: "string with control character",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("test\x01control"),
						},
					},
				},
			},
			expected: "@doc(\"test\\x01control\")\nentity User;\n",
		},
		{
			name: "string with DEL character",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("test\x7Fdel"),
						},
					},
				},
			},
			expected: "@doc(\"test\\x7fdel\")\nentity User;\n",
		},
		{
			name: "string with extended control character",
			schema: &ast.Schema{
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
		{
			name: "namespace with annotations",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("My application namespace"),
						},
						Entities: ast.Entities{
							types.EntityType("User"): ast.Entity{},
						},
					},
				},
			},
			expected: `@doc("My application namespace")
namespace MyApp {
  entity User;
}
`,
		},
		{
			name: "enum with annotations",
			schema: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Status values"),
						},
						Values: []types.String{"active", "inactive"},
					},
				},
			},
			expected: `@doc("Status values")
entity Status enum ["active", "inactive"];
`,
		},
		{
			name: "common type with annotations",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"): ast.CommonType{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("A name type"),
						},
						Type: ast.String(),
					},
				},
			},
			expected: `@doc("A name type")
type Name = String;
`,
		},
		{
			name: "action with annotations",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("View action"),
						},
					},
				},
			},
			expected: `@doc("View action")
action view;
`,
		},
		{
			name: "record with annotated attribute",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("User"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"name": ast.Attribute{
								Type: ast.String(),
								Annotations: ast.Annotations{
									types.Ident("doc"): types.String("User's name"),
								},
							},
						}),
					},
				},
			},
			expected: "type User = {\n  @doc(\"User\\'s name\")\n  \"name\": String,\n};\n",
		},
		{
			name: "entity ref without type",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOfVal: []ast.EntityRef{ast.NewEntityRef("", "someAction")},
					},
				},
			},
			expected: `action view in "someAction";
`,
		},
		{
			name: "action with only context in appliesTo",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesToVal: &ast.AppliesTo{
							Context: ast.Record(ast.Attributes{
								"timestamp": ast.Attribute{Type: ast.Long()},
							}),
						},
					},
				},
			},
			expected: `action view appliesTo {
  context: {
    "timestamp": Long,
  }
};
`,
		},
		{
			name: "action with only principal in appliesTo",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesToVal: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("User"))},
						},
					},
				},
			},
			expected: `action view appliesTo {
  principal: [User],
};
`,
		},
		{
			name: "action with only resource in appliesTo",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesToVal: &ast.AppliesTo{
							ResourceTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
						},
					},
				},
			},
			expected: `action view appliesTo {
  resource: [Document],
};
`,
		},
		{
			name: "entity with shape and memberOf",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOfVal: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Group"))},
						ShapeVal: &ast.RecordType{
							Attributes: ast.Attributes{
								"name": ast.Attribute{Type: ast.String()},
							},
						},
					},
				},
			},
			expected: `entity User in [Group] = {
  "name": String,
};
`,
		},
		{
			name: "entity with shape and tags",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						ShapeVal: &ast.RecordType{
							Attributes: ast.Attributes{
								"title": ast.Attribute{Type: ast.String()},
							},
						},
						TagsVal: ast.String(),
					},
				},
			},
			expected: `entity Document = {
  "title": String,
} tags String;
`,
		},
		{
			name: "entity with memberOf and tags",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						MemberOfVal: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Folder"))},
						TagsVal:     ast.String(),
					},
				},
			},
			expected: `entity Document in [Folder] tags String;
`,
		},
		{
			name: "entity with all features",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						MemberOfVal: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Folder"))},
						ShapeVal: &ast.RecordType{
							Attributes: ast.Attributes{
								"title": ast.Attribute{Type: ast.String()},
							},
						},
						TagsVal: ast.String(),
					},
				},
			},
			expected: `entity Document in [Folder] = {
  "title": String,
} tags String;
`,
		},
		{
			name: "decimal type",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Price"): ast.CommonType{
						Type: ast.Decimal(),
					},
				},
			},
			expected: "type Price = __cedar::decimal;\n",
		},
		{
			name: "string with double quote escape",
			schema: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String(`He said "hello"`),
						},
					},
				},
			},
			expected: "@doc(\"He said \\\"hello\\\"\")\nentity User;\n",
		},
		{
			name: "multiple common types",
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"):     ast.CommonType{Type: ast.String()},
					types.Ident("Age"):      ast.CommonType{Type: ast.Long()},
					types.Ident("IsActive"): ast.CommonType{Type: ast.Bool()},
				},
			},
			expected: `type Age = Long;

type IsActive = Bool;

type Name = String;
`,
		},
		{
			name: "multiple enums",
			schema: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{Values: []types.String{"active", "inactive"}},
					types.EntityType("Role"):   ast.Enum{Values: []types.String{"admin", "user"}},
				},
			},
			expected: `entity Role enum ["admin", "user"];

entity Status enum ["active", "inactive"];
`,
		},
		{
			name: "multiple actions",
			schema: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"):   ast.Action{},
					types.String("edit"):   ast.Action{},
					types.String("delete"): ast.Action{},
				},
			},
			expected: `action delete;

action edit;

action view;
`,
		},
		{
			name: "namespace with multiple common types",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						CommonTypes: ast.CommonTypes{
							types.Ident("Name"): ast.CommonType{Type: ast.String()},
							types.Ident("Age"):  ast.CommonType{Type: ast.Long()},
						},
					},
				},
			},
			expected: `namespace MyApp {
  type Age = Long;

  type Name = String;
}
`,
		},
		{
			name: "namespace with multiple entities",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Entities: ast.Entities{
							types.EntityType("User"):     ast.Entity{},
							types.EntityType("Document"): ast.Entity{},
							types.EntityType("Group"):    ast.Entity{},
						},
					},
				},
			},
			expected: `namespace MyApp {
  entity Document;

  entity Group;

  entity User;
}
`,
		},
		{
			name: "namespace with multiple enums",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Enums: ast.Enums{
							types.EntityType("Status"): ast.Enum{Values: []types.String{"active", "inactive"}},
							types.EntityType("Role"):   ast.Enum{Values: []types.String{"admin", "user"}},
						},
					},
				},
			},
			expected: `namespace MyApp {
  entity Role enum ["admin", "user"];

  entity Status enum ["active", "inactive"];
}
`,
		},
		{
			name: "namespace with multiple actions",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Actions: ast.Actions{
							types.String("view"):   ast.Action{},
							types.String("edit"):   ast.Action{},
							types.String("delete"): ast.Action{},
						},
					},
				},
			},
			expected: `namespace MyApp {
  action delete;

  action edit;

  action view;
}
`,
		},
		{
			name: "namespace with mixed types with blank lines",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						CommonTypes: ast.CommonTypes{
							types.Ident("Name"): ast.CommonType{Type: ast.String()},
						},
						Entities: ast.Entities{
							types.EntityType("User"): ast.Entity{},
						},
						Enums: ast.Enums{
							types.EntityType("Status"): ast.Enum{Values: []types.String{"active"}},
						},
						Actions: ast.Actions{
							types.String("view"): ast.Action{},
						},
					},
				},
			},
			expected: `namespace MyApp {
  type Name = String;

  entity User;

  entity Status enum ["active"];

  action view;
}
`,
		},
		{
			name: "multiple namespaces with blank lines",
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("AppA"): ast.Namespace{
						Entities: ast.Entities{
							types.EntityType("User"): ast.Entity{},
						},
					},
					types.Path("AppB"): ast.Namespace{
						Entities: ast.Entities{
							types.EntityType("Document"): ast.Entity{},
						},
					},
				},
			},
			expected: `namespace AppA {
  entity User;
}

namespace AppB {
  entity Document;
}
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
