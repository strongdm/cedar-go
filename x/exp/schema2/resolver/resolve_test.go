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
								{Key: "field", Type: ast.StringType{}},
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
				"field1": MyType
			};
			entity Admin = {
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
							},
						},
					},
					"Admin": {
						Name: "Admin",
						Shape: &ast.RecordType{
							Pairs: []ast.Pair{
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
				"field": Type3
			};`,
			want: &resolver.ResolvedSchema{
				Namespaces: map[types.Path]resolver.ResolvedNamespace{},
				Entities: map[types.EntityType]resolver.ResolvedEntity{
					"User": {
						Name: "User",
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

		// Error cases
		{
			name:    "undefined type reference",
			in:      `entity User = { "field": NonExistent };`,
			errTest: testutil.Error,
		},
		{
			name:    "undefined type reference in tags",
			in:      `entity User tags NonExistent;`,
			errTest: testutil.Error,
		},
		{
			name: "undefined type reference in action context",
			in: `action view appliesTo {
				principal: [],
				resource: [],
				context: NonExistent
			};`,
			errTest: testutil.Error,
		},
		{
			name:    "common type with undefined nested type",
			in:      `type MyType = NonExistent;`,
			errTest: testutil.Error,
		},
		{
			name:    "common type with undefined nested type in record",
			in:      `type MyType = { "field": NonExistent };`,
			errTest: testutil.Error,
		},
		{
			name:    "common type with undefined nested type in set",
			in:      `type MyType = Set<NonExistent>;`,
			errTest: testutil.Error,
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
