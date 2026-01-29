package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parser"
)

func TestSchemaRoundtrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cedar string
		ast   *ast.Schema
	}{
		{
			name:  "empty schema",
			cedar: "",
			ast:   &ast.Schema{},
		},
		{
			name:  "simple entity",
			cedar: "entity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{},
				},
			},
		},
		{
			name: "entity with shape",
			cedar: `entity User = {
  "email"?: String,
  "name": String,
};
`,
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Shape: &ast.RecordType{
							Attributes: ast.Attributes{
								"email": ast.Attribute{Type: ast.String(), Optional: true},
								"name":  ast.Attribute{Type: ast.String(), Optional: false},
							},
						},
					},
				},
			},
		},
		{
			name:  "entity with memberOf",
			cedar: "entity User in [Group];\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOf: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Group"))},
					},
				},
			},
		},
		{
			name:  "entity with multiple parents",
			cedar: "entity User in [Group, Team];\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOf: []ast.EntityTypeRef{
							ast.EntityType(types.EntityType("Group")),
							ast.EntityType(types.EntityType("Team")),
						},
					},
				},
			},
		},
		{
			name:  "entity with tags",
			cedar: "entity Document tags String;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						Tags: ast.String(),
					},
				},
			},
		},
		{
			name:  "enum entity",
			cedar: `entity Status enum ["active", "inactive"];` + "\n",
			ast: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{
						Values: []types.String{"active", "inactive"},
					},
				},
			},
		},
		{
			name:  "common type",
			cedar: "type Name = String;\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"): ast.CommonType{
						Type: ast.String(),
					},
				},
			},
		},
		{
			name: "common type with record",
			cedar: `type Address = {
  "city": String,
  "street": String,
};
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Address"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"city":   ast.Attribute{Type: ast.String()},
							"street": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
		},
		{
			name:  "action",
			cedar: "action view;\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{},
				},
			},
		},
		{
			name: "action with appliesTo and context",
			cedar: `action view appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "timestamp": Long,
  }
};
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("User"))},
							ResourceTypes:  []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
							Context: ast.Record(ast.Attributes{
								"timestamp": ast.Attribute{Type: ast.Long()},
							}),
						},
					},
				},
			},
		},
		{
			name:  "action with memberOf",
			cedar: `action view in "readActions";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOf: []ast.EntityRef{ast.EntityRefFromID("readActions")},
					},
				},
			},
		},
		{
			name:  "action with quoted name",
			cedar: `action "view document";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view document"): ast.Action{},
				},
			},
		},
		{
			name: "namespace",
			cedar: `namespace MyApp {
  entity Document;

  entity User;
}
`,
			ast: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Entities: ast.Entities{
							types.EntityType("Document"): ast.Entity{},
							types.EntityType("User"):     ast.Entity{},
						},
					},
				},
			},
		},
		{
			name:  "annotation",
			cedar: "@doc(\"A user entity\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("A user entity"),
						},
					},
				},
			},
		},
		{
			name:  "annotation without value",
			cedar: "@deprecated\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("deprecated"): types.String(""),
						},
					},
				},
			},
		},
		{
			name:  "set type",
			cedar: "type Tags = Set<String>;\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Tags"): ast.CommonType{
						Type: ast.Set(ast.String()),
					},
				},
			},
		},
		{
			name:  "long type",
			cedar: "type Count = Long;\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Count"): ast.CommonType{
						Type: ast.Long(),
					},
				},
			},
		},
		{
			name:  "bool type",
			cedar: "type Flag = Bool;\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Flag"): ast.CommonType{
						Type: ast.Bool(),
					},
				},
			},
		},
		{
			name:  "action with explicit entity ref",
			cedar: `action view in MyApp::Action::"allActions";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOf: []ast.EntityRef{ast.NewEntityRef(types.EntityType("MyApp::Action"), "allActions")},
					},
				},
			},
		},
		{
			name: "record with quoted keys",
			cedar: `type Data = {
  "special-key": String,
};
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"special-key": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
		},
		{
			name:  "empty record",
			cedar: "type Empty = {};\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Empty"): ast.CommonType{
						Type: ast.Record(ast.Attributes{}),
					},
				},
			},
		},
		{
			name: "nested empty record",
			cedar: `type Outer = {
  "inner": {},
};
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Outer"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"inner": ast.Attribute{Type: ast.Record(ast.Attributes{})},
						}),
					},
				},
			},
		},
		{
			name: "nested record with optional field",
			cedar: `type Outer = {
  "inner": {
    "optField"?: String,
  },
};
`,
			ast: &ast.Schema{
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
		},
		{
			name: "action with multiple principals",
			cedar: `action view appliesTo {
  principal: [Admin, User],
  resource: [Document],
};
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{
								ast.EntityType(types.EntityType("Admin")),
								ast.EntityType(types.EntityType("User")),
							},
							ResourceTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
						},
					},
				},
			},
		},
		{
			name: "action with multiple memberOf",
			cedar: `action view in ["readActions", "viewActions"];
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOf: []ast.EntityRef{
							ast.EntityRefFromID("readActions"),
							ast.EntityRefFromID("viewActions"),
						},
					},
				},
			},
		},
		{
			name: "multiple top-level entities",
			cedar: `entity Document;

entity User;
`,
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{},
					types.EntityType("User"):     ast.Entity{},
				},
			},
		},
		{
			name: "multiple top-level mixed nodes",
			cedar: `type Name = String;

entity User;

action view;
`,
			ast: &ast.Schema{
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
		},
		{
			name:  "action with empty name",
			cedar: `action "";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String(""): ast.Action{},
				},
			},
		},
		{
			name:  "action name starting with digit",
			cedar: `action "123action";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("123action"): ast.Action{},
				},
			},
		},
		{
			name: "record key starting with digit",
			cedar: `type Data = {
  "123key": String,
};
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"123key": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
		},
		{
			name: "record key with empty name",
			cedar: `type Data = {
  "": String,
};
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Data"): ast.CommonType{
						Type: ast.Record(ast.Attributes{
							"": ast.Attribute{Type: ast.String()},
						}),
					},
				},
			},
		},
		{
			name:  "action name with reserved keyword 'in'",
			cedar: `action "in";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("in"): ast.Action{},
				},
			},
		},
		{
			name:  "string with newline escape",
			cedar: "@doc(\"Line1\\nLine2\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Line1\nLine2"),
						},
					},
				},
			},
		},
		{
			name:  "string with carriage return escape",
			cedar: "@doc(\"Line1\\rLine2\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Line1\rLine2"),
						},
					},
				},
			},
		},
		{
			name:  "string with tab escape",
			cedar: "@doc(\"Col1\\tCol2\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Col1\tCol2"),
						},
					},
				},
			},
		},
		{
			name:  "string with backslash escape",
			cedar: "@doc(\"path\\\\to\\\\file\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("path\\to\\file"),
						},
					},
				},
			},
		},
		{
			name:  "string with null character escape",
			cedar: "@doc(\"null\\0char\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("null\x00char"),
						},
					},
				},
			},
		},
		{
			name:  "string with single quote escape",
			cedar: "@doc(\"It\\'s fine\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("It's fine"),
						},
					},
				},
			},
		},
		{
			name:  "string with control character",
			cedar: "@doc(\"test\\x01control\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("test\x01control"),
						},
					},
				},
			},
		},
		{
			name:  "string with DEL character",
			cedar: "@doc(\"test\\x7fdel\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("test\x7Fdel"),
						},
					},
				},
			},
		},
		{
			name: "namespace with annotations",
			cedar: `@doc("My application namespace")
namespace MyApp {
  entity User;
}
`,
			ast: &ast.Schema{
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
		},
		{
			name:  "enum with annotations",
			cedar: "@doc(\"Status values\")\nentity Status enum [\"active\", \"inactive\"];\n",
			ast: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("Status values"),
						},
						Values: []types.String{"active", "inactive"},
					},
				},
			},
		},
		{
			name:  "common type with annotations",
			cedar: "@doc(\"A name type\")\ntype Name = String;\n",
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"): ast.CommonType{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("A name type"),
						},
						Type: ast.String(),
					},
				},
			},
		},
		{
			name:  "action with annotations",
			cedar: "@doc(\"View action\")\naction view;\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String("View action"),
						},
					},
				},
			},
		},
		{
			name:  "record with annotated attribute",
			cedar: "type User = {\n  @doc(\"User\\'s name\")\n  \"name\": String,\n};\n",
			ast: &ast.Schema{
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
		},
		{
			name:  "entity ref without type",
			cedar: `action view in "someAction";` + "\n",
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						MemberOf: []ast.EntityRef{ast.NewEntityRef("", "someAction")},
					},
				},
			},
		},
		{
			name: "action with only context in appliesTo",
			cedar: `action view appliesTo {
  context: {
    "timestamp": Long,
  }
};
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							Context: ast.Record(ast.Attributes{
								"timestamp": ast.Attribute{Type: ast.Long()},
							}),
						},
					},
				},
			},
		},
		{
			name: "action with only principal in appliesTo",
			cedar: `action view appliesTo {
  principal: [User],
};
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							PrincipalTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("User"))},
						},
					},
				},
			},
		},
		{
			name: "action with only resource in appliesTo",
			cedar: `action view appliesTo {
  resource: [Document],
};
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"): ast.Action{
						AppliesTo: &ast.AppliesTo{
							ResourceTypes: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Document"))},
						},
					},
				},
			},
		},
		{
			name: "entity with shape and memberOf",
			cedar: `entity User in [Group] = {
  "name": String,
};
`,
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						MemberOf: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Group"))},
						Shape: &ast.RecordType{
							Attributes: ast.Attributes{
								"name": ast.Attribute{Type: ast.String()},
							},
						},
					},
				},
			},
		},
		{
			name: "entity with shape and tags",
			cedar: `entity Document = {
  "title": String,
} tags String;
`,
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						Shape: &ast.RecordType{
							Attributes: ast.Attributes{
								"title": ast.Attribute{Type: ast.String()},
							},
						},
						Tags: ast.String(),
					},
				},
			},
		},
		{
			name:  "entity with memberOf and tags",
			cedar: `entity Document in [Folder] tags String;` + "\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						MemberOf: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Folder"))},
						Tags:     ast.String(),
					},
				},
			},
		},
		{
			name: "entity with all features",
			cedar: `entity Document in [Folder] = {
  "title": String,
} tags String;
`,
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("Document"): ast.Entity{
						MemberOf: []ast.EntityTypeRef{ast.EntityType(types.EntityType("Folder"))},
						Shape: &ast.RecordType{
							Attributes: ast.Attributes{
								"title": ast.Attribute{Type: ast.String()},
							},
						},
						Tags: ast.String(),
					},
				},
			},
		},
		{
			name:  "string with double quote escape",
			cedar: "@doc(\"He said \\\"hello\\\"\")\nentity User;\n",
			ast: &ast.Schema{
				Entities: ast.Entities{
					types.EntityType("User"): ast.Entity{
						Annotations: ast.Annotations{
							types.Ident("doc"): types.String(`He said "hello"`),
						},
					},
				},
			},
		},
		{
			name: "multiple common types",
			cedar: `type Age = Long;

type IsActive = Bool;

type Name = String;
`,
			ast: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					types.Ident("Name"):     ast.CommonType{Type: ast.String()},
					types.Ident("Age"):      ast.CommonType{Type: ast.Long()},
					types.Ident("IsActive"): ast.CommonType{Type: ast.Bool()},
				},
			},
		},
		{
			name: "multiple enums",
			cedar: `entity Role enum ["admin", "user"];

entity Status enum ["active", "inactive"];
`,
			ast: &ast.Schema{
				Enums: ast.Enums{
					types.EntityType("Status"): ast.Enum{Values: []types.String{"active", "inactive"}},
					types.EntityType("Role"):   ast.Enum{Values: []types.String{"admin", "user"}},
				},
			},
		},
		{
			name: "multiple actions",
			cedar: `action delete;

action edit;

action view;
`,
			ast: &ast.Schema{
				Actions: ast.Actions{
					types.String("view"):   ast.Action{},
					types.String("edit"):   ast.Action{},
					types.String("delete"): ast.Action{},
				},
			},
		},
		{
			name: "namespace with multiple common types",
			cedar: `namespace MyApp {
  type Age = Long;

  type Name = String;
}
`,
			ast: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						CommonTypes: ast.CommonTypes{
							types.Ident("Name"): ast.CommonType{Type: ast.String()},
							types.Ident("Age"):  ast.CommonType{Type: ast.Long()},
						},
					},
				},
			},
		},
		{
			name: "namespace with multiple entities",
			cedar: `namespace MyApp {
  entity Document;

  entity Group;

  entity User;
}
`,
			ast: &ast.Schema{
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
		},
		{
			name: "namespace with multiple enums",
			cedar: `namespace MyApp {
  entity Role enum ["admin", "user"];

  entity Status enum ["active", "inactive"];
}
`,
			ast: &ast.Schema{
				Namespaces: ast.Namespaces{
					types.Path("MyApp"): ast.Namespace{
						Enums: ast.Enums{
							types.EntityType("Status"): ast.Enum{Values: []types.String{"active", "inactive"}},
							types.EntityType("Role"):   ast.Enum{Values: []types.String{"admin", "user"}},
						},
					},
				},
			},
		},
		{
			name: "namespace with multiple actions",
			cedar: `namespace MyApp {
  action delete;

  action edit;

  action view;
}
`,
			ast: &ast.Schema{
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
		},
		{
			name: "namespace with mixed types with blank lines",
			cedar: `namespace MyApp {
  type Name = String;

  entity User;

  entity Status enum ["active"];

  action view;
}
`,
			ast: &ast.Schema{
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
		},
		{
			name: "multiple namespaces with blank lines",
			cedar: `namespace AppA {
  entity User;
}

namespace AppB {
  entity Document;
}
`,
			ast: &ast.Schema{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name+" cedar->ast", func(t *testing.T) {
			t.Parallel()
			schema, err := parser.ParseSchema("", []byte(tt.cedar))
			testutil.OK(t, err)
			testutil.Equals(t, schema, tt.ast)
		})

		t.Run(tt.name+" ast->cedar", func(t *testing.T) {
			t.Parallel()
			result := string(parser.MarshalSchema(tt.ast))
			testutil.Equals(t, result, tt.cedar)
		})
	}
}
