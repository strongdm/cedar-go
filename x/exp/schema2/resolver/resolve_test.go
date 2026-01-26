package resolver_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/resolver"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		in      string
		want    *resolver.ResolvedSchema
		errTest func(testutil.TB, error)
	}{
		{
			name: "simple entity",
			in:   `entity User;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {Name: "User"},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity in namespace",
			in:   `namespace MyApp { entity User; }`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"MyApp": {Name: "MyApp"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"MyApp::User": {Name: "MyApp::User"},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity in nested namespace",
			in:   `namespace MyApp::Models { entity User; }`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"MyApp::Models": {Name: "MyApp::Models"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"MyApp::Models::User": {Name: "MyApp::Models::User"},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "simple enum",
			in:   `entity Status enum ["active", "inactive", "pending"];`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums: map[types.EntityType]resolver.ResolvedEnum{
					"Status": {
						Name:   "Status",
						Values: []types.String{"active", "inactive", "pending"},
					},
				},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "enum in namespace",
			in:   `namespace MyApp { entity Status enum ["active", "inactive"]; }`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"MyApp": {Name: "MyApp"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{},
				Enums: map[types.EntityType]resolver.ResolvedEnum{
					"MyApp::Status": {
						Name:   "MyApp::Status",
						Values: []types.String{"active", "inactive"},
					},
				},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "simple action",
			in:   `action view;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{
					types.NewEntityUID("Action", "view"): {Name: "view"},
				},
			},
			errTest: testutil.OK,
		},
		{
			name: "action in namespace",
			in:   `namespace Bananas { action view; }`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"Bananas": {Name: "Bananas"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{},
				Enums:    map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "age", Type: ast.LongType{}},
								{Key: "active", Type: ast.BoolType{}},
								{Key: "name", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with extension types",
			in: `entity User = {
				"ip": __cedar::ipaddr,
				"amount": __cedar::decimal
			};`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "ip", Type: ast.ExtensionType{Name: "ipaddr"}},
								{Key: "amount", Type: ast.ExtensionType{Name: "decimal"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with set type",
			in: `entity User = {
				"tags": Set<String>
			};`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "tags", Type: ast.SetType{Element: ast.StringType{}}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{
									Key: "address",
									Type: ast.RecordType{
										Pairs: []ast.Pair{
											{Key: "street", Type: ast.StringType{}},
											{Key: "city", Type: ast.StringType{}},
										},
									},
								},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"Document": {
						Name: "Document",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "title", Type: ast.StringType{}},
							},
						},
						Tags: ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "classification", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "entity with memberOf",
			in:   `entity Group; entity User in [Group];`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"Group": {Name: "Group"},
					"User": {
						Name:     "User",
						MemberOf: []types.EntityType{"Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User":     {Name: "User"},
					"Document": {Name: "Document"},
				},
				Enums: map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{
					types.NewEntityUID("Action", "view"): {
						Name: "view",
						AppliesTo: &resolver.ResolvedAppliesTo{
							PrincipalTypes: []types.EntityType{"User"},
							ResourceTypes:  []types.EntityType{"Document"},
							Context: ast.RecordType{
								Pairs: []ast.Pair{
									{Key: "ip", Type: ast.StringType{}},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{
					types.NewEntityUID("Action", "parent"): {Name: "parent"},
					types.NewEntityUID("Action", "child"): {
						Name:     "child",
						MemberOf: []types.EntityUID{types.NewEntityUID("Action", "parent")},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{
									Key: "address",
									Type: ast.RecordType{
										Pairs: []ast.Pair{
											{Key: "street", Type: ast.StringType{}},
											{Key: "city", Type: ast.StringType{}},
										},
									},
								},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field1", Type: ast.StringType{}},
								{Key: "field2", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field1", Type: ast.StringType{}},
								{Key: "field2", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field1", Type: ast.StringType{}},
								{Key: "field2", Type: ast.StringType{}},
								{Key: "field3", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "annotations on namespace",
			in: `@doc("My application")
			namespace App {
				entity User;
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {
						Name: "App",
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "My application"},
						},
					},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {Name: "App::User"},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "annotations on entity",
			in: `@doc("User entity")
			entity User;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Annotations: []ast.Annotation{
							{Key: "doc", Value: "User entity"},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "empty namespace",
			in:   `namespace Empty { }`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"Empty": {Name: "Empty"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{},
				Enums:    map[types.EntityType]resolver.ResolvedEnum{},
				Actions:  map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "qualified entity reference in same namespace",
			in: `namespace App {
				entity Group;
				entity User in [App::Group];
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::Group": {Name: "App::Group"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"App::Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "unqualified entity reference in same namespace",
			in: `namespace App {
				entity Group;
				entity User in [Group];
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::Group": {Name: "App::Group"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"App::Group"},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "namespace entity references global entity",
			in: `entity GlobalGroup;
			namespace App {
				entity User in [GlobalGroup];
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"GlobalGroup": {Name: "GlobalGroup"},
					"App::User": {
						Name:     "App::User",
						MemberOf: []types.EntityType{"GlobalGroup"},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "namespace entity references global enum",
			in: `entity Status enum ["active", "inactive"];
			namespace App {
				entity Document in [Status];
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::Document": {
						Name:     "App::Document",
						MemberOf: []types.EntityType{"Status"},
					},
				},
				Enums: map[types.EntityType]resolver.ResolvedEnum{
					"Status": {
						Name:   "Status",
						Values: []types.String{"active", "inactive"},
					},
				},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},

		// Cases with unresolved type references (treated as EntityTypeRef)
		{
			name: "undefined type reference",
			in:   `entity User = { "field": NonExistent };`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "undefined type reference in tags",
			in:   `entity User tags NonExistent;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Tags: ast.EntityTypeRef{Name: "NonExistent"},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions:    map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type with undefined nested type in record",
			in:   `type MyType = { "field": NonExistent };`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions:    map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "common type with undefined nested type in set",
			in:   `type MyType = Set<NonExistent>;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions:    map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name:    "duplicate entity definition",
			in:      `entity User; entity User;`,
			errTest: testutil.Error,
		},
		{
			name:    "duplicate enum definition",
			in:      `entity Status enum ["active"]; entity Status enum ["inactive"];`,
			errTest: testutil.Error,
		},
		{
			name:    "duplicate action definition",
			in:      `action view; action view;`,
			errTest: testutil.Error,
		},
		{
			name:    "entity and enum with same name",
			in:      `entity Thing; entity Thing enum ["value"];`,
			errTest: testutil.Error,
		},
		{
			name:    "enum and entity with same name",
			in:      `entity Status enum ["active"]; entity Status;`,
			errTest: testutil.Error,
		},
		{
			name:    "duplicate entity in namespace",
			in:      `namespace App { entity User; entity User; }`,
			errTest: testutil.Error,
		},
		{
			name: "nested namespace vs qualified name conflict - entity",
			in: `namespace Goat::Gorilla {
				entity Cows;
			}
			namespace Goat::Gorilla {
				entity Cows;
			}`,
			errTest: testutil.Error,
		},
		{
			name: "nested namespace vs qualified name conflict - enum",
			in: `namespace Goat::Gorilla {
				entity Status enum ["active"];
			}
			namespace Goat::Gorilla {
				entity Status enum ["inactive"];
			}`,
			errTest: testutil.Error,
		},
		{
			name: "action conflict in same namespace",
			in: `namespace Goat::Gorilla {
				action view;
				action view;
			}`,
			errTest: testutil.Error,
		},
		{
			name: "undefined type in lazy resolution",
			in: `namespace App {
				type MyType = NonExistent;
				entity User = {
					"field": MyType
				};
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.String()},
							},
						},
					},
					"App::Group": {
						Name: "App::Group",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.String()},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "error resolving namespace common type during lazy resolution",
			in: `namespace App {
				entity User = { "f1": BadType, "f2": BadType };
				type BadType = NonExistent;
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "f1", Type: ast.EntityTypeRef{Name: "NonExistent"}},
								{Key: "f2", Type: ast.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App":   {Name: "App"},
					"Other": {Name: "Other"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"Other::User": {
						Name: "Other::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.EntityTypeRef{Name: "Type2"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK, // Type2 is not accessible from Other namespace, becomes EntityTypeRef
		},
		{
			name: "common type aliasing entity type is invalid",
			in: `entity User;
			type UserAlias = User;
			entity Group = { "owner": UserAlias };`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {Name: "User"},
					"Group": {
						Name: "Group",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "owner", Type: ast.EntityTypeRef{Name: "User"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.StringType{}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK, // Global types ARE accessible from namespaces
		},
		{
			name: "error resolving global common type during lazy resolution",
			in: `entity User = { "field": BadType };
			type BadType = NonExistent;`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.EntityTypeRef{Name: "NonExistent"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK,
		},
		{
			name: "lazy resolution finds namespaced common type in schema cache",
			in: `namespace App {
				type Type1 = String;
				type Type2 = Type1;
			}
			namespace App {
				entity User = { "field": Type2 };
			}`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{
					"App": {Name: "App"},
				},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"App::User": {
						Name: "App::User",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "field", Type: ast.EntityTypeRef{Name: "Type2"}},
							},
						},
					},
				},
				Enums:   map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{},
			},
			errTest: testutil.OK, // Type2 not found in namespace-local cache, not qualified for schema cache
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
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities:   map[types.EntityType]resolver.ResolvedEntity{},
				Enums:      map[types.EntityType]resolver.ResolvedEnum{},
				Actions: map[types.EntityUID]resolver.ResolvedAction{
					types.NewEntityUID("Action", "view"): {
						Name: "view",
						AppliesTo: &resolver.ResolvedAppliesTo{
							Context: ast.RecordType{
								Pairs: []ast.Pair{
									{Key: "timestamp", Type: ast.ExtensionType{Name: "datetime"}},
									{Key: "duration", Type: ast.ExtensionType{Name: "duration"}},
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

			var s schema2.Schema
			err := s.UnmarshalCedar([]byte(tt.in))
			testutil.OK(t, err)

			resolved, err := s.Resolve()
			tt.errTest(t, err)

			if err == nil && tt.want != nil {
				testutil.Equals(t, resolved, tt.want)
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
	schema := ast.NewSchema(
		ast.Entity("User"),
		ast.Entity("Group").Shape(
			ast.Attribute("members", ast.Set(ast.EntityType("User"))),
		),
	)

	resolved, err := resolver.Resolve(schema)
	testutil.OK(t, err)

	// Verify Group entity has correct shape with resolved User reference
	group := resolved.Entities["Group"]
	testutil.Equals(t, group.Name, types.EntityType("Group"))
	testutil.Equals(t, len(group.Shape.Pairs), 1)
	testutil.Equals(t, group.Shape.Pairs[0].Key, types.String("members"))

	// The Set element should still be EntityTypeRef, just with name resolved
	setType, ok := group.Shape.Pairs[0].Type.(ast.SetType)
	testutil.Equals(t, ok, true)
	entityRef, ok := setType.Element.(ast.EntityTypeRef)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, entityRef.Name, types.EntityType("User"))
}

// TestEnumEntityUIDs tests the EntityUIDs iterator method.
func TestEnumEntityUIDs(t *testing.T) {
	t.Parallel()

	schema := ast.NewSchema(ast.Enum("Status", "active", "inactive", "pending"))
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
