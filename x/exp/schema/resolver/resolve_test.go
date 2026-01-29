package resolver_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolver"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    *resolver.Schema
		errTest func(testutil.TB, error)
	}{
		{
			name: "simple entity",
			in:   `entity User;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {Name: "User"},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity in namespace",
			in:   `namespace MyApp { entity User; }`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"MyApp": {Name: "MyApp"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"MyApp::User": {Name: "MyApp::User"},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity in nested namespace",
			in:   `namespace MyApp::Models { entity User; }`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"MyApp::Models": {Name: "MyApp::Models"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"MyApp::Models::User": {Name: "MyApp::Models::User"},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "simple enum",
			in:   `entity Status enum ["active", "inactive", "pending"];`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums: map[types.EntityType]resolver.Enum{
					"Status": {
						Name:   "Status",
						Values: []types.String{"active", "inactive", "pending"},
					},
				},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "enum in namespace",
			in:   `namespace MyApp { entity Status enum ["active", "inactive"]; }`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"MyApp": {Name: "MyApp"},
				},
				Entities: map[types.EntityType]resolver.Entity{},
				Enums: map[types.EntityType]resolver.Enum{
					"MyApp::Status": {
						Name:   "MyApp::Status",
						Values: []types.String{"active", "inactive"},
					},
				},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "simple action",
			in:   `action view;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{
					types.NewEntityUID("Action", "view"): {Name: "view"},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "action in namespace",
			in:   `namespace Bananas { action view; }`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"Bananas": {Name: "Bananas"},
				},
				Entities: map[types.EntityType]resolver.Entity{},
				Enums:    map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{
					types.NewEntityUID("Bananas::Action", "view"): {Name: "view"},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with shape attributes",
			in: `entity User = {
				"age": Long,
				"active": Bool,
				"name": String
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"age":    resolver.Attribute{Type: resolver.LongType{}},
								"active": resolver.Attribute{Type: resolver.BoolType{}},
								"name":   resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with extension types",
			in: `entity User = {
				"ip": __cedar::ipaddr,
				"amount": __cedar::decimal
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"ip":     resolver.Attribute{Type: resolver.ExtensionType{Name: "ipaddr"}},
								"amount": resolver.Attribute{Type: resolver.ExtensionType{Name: "decimal"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with set type",
			in: `entity User = {
				"tags": Set<String>
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"tags": resolver.Attribute{Type: resolver.SetType{Element: resolver.StringType{}}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with nested record",
			in: `entity User = {
				"address": {
					"street": String,
					"city": String
				}
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"address": resolver.Attribute{
									Type: resolver.RecordType{
										Attributes: resolver.Attributes{
											"street": resolver.Attribute{Type: resolver.StringType{}},
											"city":   resolver.Attribute{Type: resolver.StringType{}},
										},
									},
								},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with tags",
			in: `entity Document = {
				"title": String
			} tags {
				"classification": String
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"Document": {
						Name: "Document",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"title": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
						Tags: resolver.RecordType{
							Attributes: resolver.Attributes{
								"classification": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with memberOf",
			in:   `entity Group; entity User in [Group];`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"Group": {Name: "Group"},
					"User": {
						Name:     "User",
						MemberOf: []types.EntityType{"Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "action with appliesTo",
			in: `entity User;
			entity Document;
			action view appliesTo {
				principal: [User],
				resource: [Document],
				context: {
					"ip": String
				}
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User":     {Name: "User"},
					"Document": {Name: "Document"},
				},
				Enums: map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{
					types.NewEntityUID("Action", "view"): {
						Name: "view",
						AppliesTo: &resolver.AppliesTo{
							Principals: []types.EntityType{"User"},
							Resources:  []types.EntityType{"Document"},
							Context: resolver.RecordType{
								Attributes: resolver.Attributes{
									"ip": resolver.Attribute{Type: resolver.StringType{}},
								},
							},
						},
					},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "action with memberOf",
			in:   `action parent; action child in [parent];`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{
					types.NewEntityUID("Action", "parent"): {Name: "parent"},
					types.NewEntityUID("Action", "child"): {
						Name:     "child",
						MemberOf: []types.EntityUID{types.NewEntityUID("", "parent")},
					},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type reference",
			in: `type Address = {
				"street": String,
				"city": String
			};
			entity User = {
				"address": Address
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"address": resolver.Attribute{
									Type: resolver.RecordType{
										Attributes: resolver.Attributes{
											"street": resolver.Attribute{Type: resolver.StringType{}},
											"city":   resolver.Attribute{Type: resolver.StringType{}},
										},
									},
								},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type reference in namespace",
			in: `namespace App {
				type MyType = String;
				entity User = {
					"field1": MyType,
					"field2": MyType
				};
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {
						Name: "App::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field1": resolver.Attribute{Type: resolver.StringType{}},
								"field2": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type reused (caching test)",
			in: `type MyType = String;
			entity User = {
				"field1": MyType,
				"field2": MyType
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field1": resolver.Attribute{Type: resolver.StringType{}},
								"field2": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "deeply nested type references",
			in: `type Type1 = String;
			type Type2 = Type1;
			type Type3 = Type2;
			entity User = {
				"field1": Type3,
				"field2": Type2,
				"field3": Type1
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field1": resolver.Attribute{Type: resolver.StringType{}},
								"field2": resolver.Attribute{Type: resolver.StringType{}},
								"field3": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "annotations on namespace",
			in: `@doc("My application")
			namespace App {
				entity User;
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {
						Name: "App",
						Annotations: resolver.Annotations{
							"doc": "My application",
						},
					},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {Name: "App::User"},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "annotations on entity",
			in: `@doc("User entity")
			entity User;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Annotations: resolver.Annotations{
							"doc": "User entity",
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "empty namespace",
			in:   `namespace Empty { }`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"Empty": {Name: "Empty"},
				},
				Entities: map[types.EntityType]resolver.Entity{},
				Enums:    map[types.EntityType]resolver.Enum{},
				Actions:  map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "qualified entity reference in same namespace",
			in: `namespace App {
				entity Group;
				entity User in [App::Group];
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::Group": {Name: "App::Group"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"App::Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "unqualified entity reference in same namespace",
			in: `namespace App {
				entity Group;
				entity User in [Group];
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::Group": {Name: "App::Group"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"App::Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "namespace entity references global entity",
			in: `entity GlobalGroup;
			namespace App {
				entity User in [GlobalGroup];
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"GlobalGroup": {Name: "GlobalGroup"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"GlobalGroup"},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "namespace entity references global enum",
			in: `entity Status enum ["active", "inactive"];
			namespace App {
				entity Document in [Status];
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::Document": {
						Name:     "App::Document",
						MemberOf: []types.EntityType{"Status"},
					},
				},
				Enums: map[types.EntityType]resolver.Enum{
					"Status": {
						Name:   "Status",
						Values: []types.String{"active", "inactive"},
					},
				},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},

		// Cases with unresolved type references (treated as EntityTypeRef)
		{
			name: "undefined type reference",
			in:   `entity User = { "field": NonExistent };`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "undefined type reference in tags",
			in:   `entity User tags NonExistent;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Tags: resolver.EntityTypeRef{Name: "NonExistent"},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "undefined type reference in action context",
			in: `action view appliesTo {
				principal: [],
				resource: [],
				context: NonExistent
			};`,
			errTest: testutil.Error, // Context must be a record type
		},
		{
			name: "common type with undefined nested type",
			in:   `type MyType = NonExistent;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions:    map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type with undefined nested type in record",
			in:   `type MyType = { "field": NonExistent };`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions:    map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type with undefined nested type in set",
			in:   `type MyType = Set<NonExistent>;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions:    map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "undefined type in lazy resolution",
			in: `namespace App {
				type MyType = NonExistent;
				entity User = {
					"field": MyType
				};
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {
						Name: "App::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "namespace common type used multiple times triggers cache",
			in: `namespace App {
				type MyType = String;
				entity User = { "field": MyType };
				entity Group = { "field": MyType };
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {
						Name: "App::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
					"App::Group": {
						Name: "App::Group",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "error resolving namespace common type during lazy resolution",
			in: `namespace App {
				entity User = { "f1": BadType, "f2": BadType };
				type BadType = NonExistent;
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {
						Name: "App::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"f1": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "NonExistent"}},
								"f2": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "nested common types in namespace with cross reference triggers schema cache search",
			in: `namespace App {
				type Type1 = String;
				type Type2 = Type1;
			}
			namespace Other {
				entity User = { "field": Type2 };
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App":   {Name: "App"},
					"Other": {Name: "Other"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"Other::User": {
						Name: "Other::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "Type2"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK, // Type2 is not accessible from Other namespace, becomes EntityTypeRef
		},
		{
			name: "common type aliasing entity type is invalid",
			in: `entity User;
			type UserAlias = User;
			entity Group = { "owner": UserAlias };`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {Name: "User"},
					"Group": {
						Name: "Group",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"owner": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "User"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK, // User is treated as EntityTypeRef when used as type
		},
		{
			name: "namespace can reference global common types",
			in: `type GlobalType = String;
			namespace App {
				type LocalType = GlobalType;
				entity User = { "field": LocalType };
			}`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.Entity{
					"App::User": {
						Name: "App::User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK, // Global types ARE accessible from namespaces
		},
		{
			name: "error resolving global common type during lazy resolution",
			in: `entity User = { "field": BadType };
			type BadType = NonExistent;`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities: map[types.EntityType]resolver.Entity{
					"User": {
						Name: "User",
						Shape: &resolver.RecordType{
							Attributes: resolver.Attributes{
								"field": resolver.Attribute{Type: resolver.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{},
			},
			errTest: testutil.OK,
		},
		{
			name: "action context with unqualified extension type",
			in: `action view appliesTo {
				principal: [],
				resource: [],
				context: {
					"timestamp": datetime,
					"duration": duration
				}
			};`,
			want: &resolver.Schema{
				Namespaces: map[types.Path]resolver.Namespace{},
				Entities:   map[types.EntityType]resolver.Entity{},
				Enums:      map[types.EntityType]resolver.Enum{},
				Actions: map[types.EntityUID]resolver.Action{
					types.NewEntityUID("Action", "view"): {
						Name: "view",
						AppliesTo: &resolver.AppliesTo{
							Context: resolver.RecordType{
								Attributes: resolver.Attributes{
									"timestamp": resolver.Attribute{Type: resolver.ExtensionType{Name: "datetime"}},
									"duration":  resolver.Attribute{Type: resolver.ExtensionType{Name: "duration"}},
								},
							},
						},
					},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "action context resolves to non-record type",
			in: `type NotARecord = String;
			action view appliesTo {
				principal: [],
				resource: [],
				context: NotARecord
			};`,
			errTest: testutil.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s schema.Schema
			err := s.UnmarshalCedar([]byte(tt.in))
			testutil.OK(t, err)

			resolved, err := s.Resolve()
			tt.errTest(t, err)

			if err == nil && tt.want != nil {
				// Compare namespaces
				testutil.Equals(t, resolved.Namespaces, tt.want.Namespaces)
				testutil.Equals(t, resolved.Enums, tt.want.Enums)
				testutil.Equals(t, resolved.Actions, tt.want.Actions)

				// Compare entities, but handle Shape pointers specially
				testutil.Equals(t, len(resolved.Entities), len(tt.want.Entities))
				for k, wantEntity := range tt.want.Entities {
					gotEntity, ok := resolved.Entities[k]
					testutil.FatalIf(t, !ok, "missing entity %v", k)
					testutil.Equals(t, gotEntity.Name, wantEntity.Name)
					testutil.Equals(t, gotEntity.Annotations, wantEntity.Annotations)
					testutil.Equals(t, gotEntity.MemberOf, wantEntity.MemberOf)
					testutil.Equals(t, gotEntity.Tags, wantEntity.Tags)
					// Compare Shape by dereferencing if both are non-nil
					if gotEntity.Shape != nil && wantEntity.Shape != nil {
						testutil.Equals(t, *gotEntity.Shape, *wantEntity.Shape)
					} else {
						testutil.Equals(t, gotEntity.Shape, wantEntity.Shape)
					}
				}
			}
		})
	}
}

// TestEntityTypeRefInSet tests that EntityTypeRef can be resolved when nested in a Set.
// This is a valid AST construction that can come from JSON schemas.
// The Cedar text parser treats Set<User> as Set<TypeRef>, so we must construct the AST directly.
func TestEntityTypeRefInSet(t *testing.T) {
	t.Parallel()

	// Construct schema with Set<EntityTypeRef> directly in AST
	schema := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{},
			"Group": ast.Entity{
				Shape: &ast.RecordType{
					Attributes: ast.Attributes{
						"members": ast.Attribute{Type: ast.Set(ast.EntityType("User"))},
					},
				},
			},
		},
		Enums:       ast.Enums{},
		Actions:     ast.Actions{},
		CommonTypes: ast.CommonTypes{},
		Namespaces:  ast.Namespaces{},
	}

	resolved, err := resolver.Resolve(schema)
	testutil.OK(t, err)

	// Verify Group entity has correct shape with resolved User reference
	group := resolved.Entities["Group"]
	testutil.Equals(t, group.Name, types.EntityType("Group"))
	testutil.Equals(t, len(group.Shape.Attributes), 1)
	membersAttr, exists := group.Shape.Attributes["members"]
	testutil.Equals(t, exists, true)

	// The Set element should still be EntityTypeRef, just with name resolved
	setType, ok := membersAttr.Type.(resolver.SetType)
	testutil.Equals(t, ok, true)
	entityRef, ok := setType.Element.(resolver.EntityTypeRef)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, entityRef.Name, types.EntityType("User"))
}

// TestEnumEntityUIDs tests the EntityUIDs iterator method.
func TestEnumEntityUIDs(t *testing.T) {
	t.Parallel()

	schema := &ast.Schema{
		Entities:    ast.Entities{},
		Enums:       ast.Enums{"Status": ast.Enum{Values: []types.String{"active", "inactive", "pending"}}},
		Actions:     ast.Actions{},
		CommonTypes: ast.CommonTypes{},
		Namespaces:  ast.Namespaces{},
	}
	resolved, err := resolver.Resolve(schema)
	testutil.OK(t, err)

	resolvedEnum := resolved.Enums["Status"]
	var uids []types.EntityUID
	for uid := range resolvedEnum.EntityUIDs() {
		uids = append(uids, uid)
	}

	wantUIDs := []types.EntityUID{
		types.NewEntityUID("Status", "active"),
		types.NewEntityUID("Status", "inactive"),
		types.NewEntityUID("Status", "pending"),
	}
	testutil.Equals(t, uids, wantUIDs)

	uids = nil
	for uid := range resolvedEnum.EntityUIDs() {
		if uid.ID == "inactive" {
			break
		}
		uids = append(uids, uid)
	}
	wantUIDs = []types.EntityUID{
		types.NewEntityUID("Status", "active"),
	}
	testutil.Equals(t, uids, wantUIDs)
}

// TestEdgeCases tests edge cases and validates error handling.
func TestEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		// Valid edge cases
		{
			name: "entity with empty MemberOf",
			in:   `entity User in [];`,
		},
		{
			name: "entity in namespace with empty MemberOf",
			in: `namespace App {
				entity User in [];
			}`,
		},
		{
			name: "entity with namespace but empty namespace name",
			in: `entity User;
			entity Group in [User];`,
		},
		{
			name: "action with empty MemberOf",
			in:   `action view in [];`,
		},
		{
			name: "action with appliesTo but empty principal and resource",
			in: `action view appliesTo {
				principal: [],
				resource: []
			};`,
		},
		{
			name: "action in namespace with appliesTo but empty lists",
			in: `namespace App {
				action view appliesTo {
					principal: [],
					resource: []
				};
			}`,
		},
		{
			name: "entity with optional attributes in record",
			in: `entity User = {
				"required": String,
				"optional"?: Long
			};`,
		},
		{
			name: "entity with annotations on record attributes",
			in: `entity User = {
				@doc("User name")
				"name": String
			};`,
		},
		{
			name: "common type with annotations",
			in: `@doc("Address type")
			type Address = {
				"street": String
			};
			entity User = {
				"address": Address
			};`,
		},
		{
			name: "enum with annotations",
			in: `@doc("Status enum")
			entity Status enum ["active", "inactive"];`,
		},
		{
			name: "action with annotations",
			in: `@doc("View action")
			action view;`,
		},
		{
			name: "entity with empty shape",
			in:   `entity User = {};`,
		},
		{
			name: "action with context but no principal or resource types",
			in: `action view appliesTo {
				context: {}
			};`,
		},
		// Error cases - action context must be record type
		{
			name: "error in resolveTypeRef lazy resolution of namespace common type",
			in: `namespace App {
				type BadContext = String;
				action view appliesTo {
					principal: [],
					resource: [],
					context: BadContext
				};
			}`,
			wantErr: true,
		},
		{
			name: "error in resolveTypeRef lazy resolution of global common type",
			in: `type BadContext = Long;
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadContext
			};`,
			wantErr: true,
		},
		{
			name: "error in nested common type resolution through namespace cache",
			in: `namespace App {
				type BadContext = String;
				type AliasedBadContext = BadContext;
				action view appliesTo {
					principal: [],
					resource: [],
					context: AliasedBadContext
				};
			}`,
			wantErr: true,
		},
		{
			name: "error in nested common type resolution through schema cache",
			in: `type BadContext = Bool;
			type AliasedBadContext = BadContext;
			action view appliesTo {
				principal: [],
				resource: [],
				context: AliasedBadContext
			};`,
			wantErr: true,
		},
		{
			name: "error in resolveSet with action context",
			in: `type BadContext = Set<String>;
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadContext
			};`,
			wantErr: true,
		},
		{
			name: "entity shape error in nested record",
			in: `type BadType = Bool;
			entity User = {
				"nested": {
					"field": Set<BadType>
				}
			};
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadType
			};`,
			wantErr: true,
		},
		{
			name: "namespace entity shape error",
			in: `namespace App {
				type BadType = String;
				entity User = {
					"field": BadType
				};
				action view appliesTo {
					principal: [],
					resource: [],
					context: BadType
				};
			}`,
			wantErr: true,
		},
		{
			name: "namespace entity tags error",
			in: `namespace App {
				type BadTags = Long;
				entity User tags BadTags;
				action view appliesTo {
					principal: [],
					resource: [],
					context: BadTags
				};
			}`,
			wantErr: true,
		},
		{
			name: "namespace action context error with nested set",
			in: `namespace App {
				type BadContext = Set<Set<String>>;
				action view appliesTo {
					principal: [],
					resource: [],
					context: BadContext
				};
			}`,
			wantErr: true,
		},
		{
			name: "top-level common type error in entity",
			in: `type BadType = Set<String>;
			entity User = { "field": BadType };
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadType
			};`,
			wantErr: true,
		},
		{
			name: "error in resolveTypeRef schema cache with qualified namespace reference",
			in: `namespace App {
				type BadContext = String;
			}
			namespace Other {
				action view appliesTo {
					principal: [],
					resource: [],
					context: App::BadContext
				};
			}`,
			wantErr: true,
		},
		{
			name: "error in resolveRecord for entity shape",
			in: `type BadType = String;
			entity User = {
				"field": Set<BadType>
			};
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadType
			};`,
			wantErr: true,
		},
		{
			name: "error in resolveSet nested in entity shape",
			in: `type BadType = Long;
			entity User = {
				"field": Set<Set<BadType>>
			};
			action view appliesTo {
				principal: [],
				resource: [],
				context: BadType
			};`,
			wantErr: true,
		},
		{
			name: "action context resolves to non-record",
			in: `type NotARecord = String;
			action view appliesTo {
				principal: [],
				resource: [],
				context: NotARecord
			};`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var s schema.Schema
			err := s.UnmarshalCedar([]byte(tt.in))
			// Some tests may fail parsing, which is OK for edge cases
			if err != nil {
				return
			}

			_, err = s.Resolve()
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

// TestDirectAST tests scenarios that use AST construction directly to create
// schemas that the parser wouldn't allow, ensuring proper error handling.
func TestDirectAST(t *testing.T) {
	t.Parallel()

	t.Run("extension types work", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities: ast.Entities{
				"Thing": ast.Entity{
					Shape: &ast.RecordType{
						Attributes: ast.Attributes{
							"ext": ast.Attribute{Type: ast.IPAddr()},
						},
					},
				},
			},
		}

		out, err := resolver.Resolve(schema)
		testutil.OK(t, err)
		want := resolver.ExtensionType{Name: "ipaddr"}
		testutil.Equals(t, out.Entities["Thing"].Shape.Attributes["ext"].Type, resolver.IsType(want))
	})

	t.Run("panic on nil type", func(t *testing.T) {
		t.Parallel()
		testutil.Panic(t, func() {
			schema := &ast.Schema{
				Entities: ast.Entities{
					"Thing": ast.Entity{
						Shape: &ast.RecordType{
							Attributes: ast.Attributes{
								"ext": ast.Attribute{Type: nil},
							},
						},
					},
				},
			}
			_, _ = resolver.Resolve(schema)
		})
	})

	t.Run("top-level entity and enum with same name", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities: ast.Entities{
				"Thing": ast.Entity{},
			},
			Enums: ast.Enums{
				"Thing": ast.Enum{Values: []types.String{"value"}},
			},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces:  ast.Namespaces{},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("namespace entity and enum with same name", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					Entities: ast.Entities{
						"Thing": ast.Entity{},
					},
					Enums: ast.Enums{
						"Thing": ast.Enum{Values: []types.String{"active"}},
					},
					Actions:     ast.Actions{},
					CommonTypes: ast.CommonTypes{},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("duplicate entity across top-level and namespace", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities: ast.Entities{
				"App::User": ast.Entity{},
			},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					Entities: ast.Entities{
						"User": ast.Entity{},
					},
					Enums:       ast.Enums{},
					CommonTypes: ast.CommonTypes{},
					Actions:     ast.Actions{},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("duplicate enum across top-level and namespace", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities: ast.Entities{},
			Enums: ast.Enums{
				"App::Status": ast.Enum{Values: []types.String{"active"}},
			},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					Entities: ast.Entities{},
					Enums: ast.Enums{
						"Status": ast.Enum{Values: []types.String{"inactive"}},
					},
					CommonTypes: ast.CommonTypes{},
					Actions:     ast.Actions{},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("entity/enum conflict across top-level and namespace", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities: ast.Entities{
				"App::Thing": ast.Entity{},
			},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					Entities: ast.Entities{},
					Enums: ast.Enums{
						"Thing": ast.Enum{Values: []types.String{"value"}},
					},
					CommonTypes: ast.CommonTypes{},
					Actions:     ast.Actions{},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("action context is Set type", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			CommonTypes: ast.CommonTypes{},
			Actions: ast.Actions{
				"view": ast.Action{
					AppliesTo: &ast.AppliesTo{
						Context: ast.SetType{Element: ast.StringType{}},
					},
				},
			},
			Namespaces: ast.Namespaces{},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("namespace action context is non-record type", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					Entities:    ast.Entities{},
					Enums:       ast.Enums{},
					CommonTypes: ast.CommonTypes{},
					Actions: ast.Actions{
						"view": ast.Action{
							AppliesTo: &ast.AppliesTo{
								Context: ast.StringType{},
							},
						},
					},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("common type resolves to Set used as action context", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{
				"BadContext": ast.CommonType{
					Type: ast.SetType{Element: ast.StringType{}},
				},
			},
			Entities: ast.Entities{},
			Enums:    ast.Enums{},
			Actions: ast.Actions{
				"view": ast.Action{
					AppliesTo: &ast.AppliesTo{
						Context: ast.TypeRef{Name: "BadContext"},
					},
				},
			},
			Namespaces: ast.Namespaces{},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})

	t.Run("namespace common type resolves to Set used as action context", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			CommonTypes: ast.CommonTypes{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					CommonTypes: ast.CommonTypes{
						"BadContext": ast.CommonType{
							Type: ast.SetType{Element: ast.StringType{}},
						},
					},
					Entities: ast.Entities{},
					Enums:    ast.Enums{},
					Actions: ast.Actions{
						"view": ast.Action{
							AppliesTo: &ast.AppliesTo{
								Context: ast.TypeRef{Name: "BadContext"},
							},
						},
					},
				},
			},
		}

		_, err := resolver.Resolve(schema)
		testutil.Error(t, err)
	})
}
