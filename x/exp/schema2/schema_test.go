package schema2_test

import (
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		err  bool
	}{
		// Success cases - basic declarations
		{"empty", ``, false},
		{"simpleEntity", `entity User;`, false},
		{"entityWithShape", `entity User {
			name: String,
			age: Long,
		};`, false},
		{"entityWithOptional", `entity User {
			name: String,
			email?: String,
		};`, false},
		{"entityWithMemberOf", `entity User in Group;`, false},
		{"entityWithMemberOfMultiple", `entity User in [Group, Team];`, false},
		{"entityWithTags", `entity Document tags String;`, false},
		{"entityWithEquals", `entity User = { name: String };`, false},
		{"entityWithEverything", `entity User in [Group, Team] {
			name: String,
			email?: String,
		} tags Long;`, false},

		// Enum entities
		{"enumEntity", `entity Status enum ["active", "inactive", "pending"];`, false},
		{"enumEntitySingle", `entity Status enum ["active"];`, false},

		// Common types
		{"commonTypeString", `type Name = String;`, false},
		{"commonTypeLong", `type Count = Long;`, false},
		{"commonTypeBool", `type Flag = Bool;`, false},
		{"commonTypeRecord", `type Address = {
			street: String,
			city: String,
		};`, false},
		{"commonTypeSet", `type Tags = Set<String>;`, false},
		{"commonTypeNestedSet", `type Matrix = Set<Set<Long>>;`, false},
		{"commonTypeRef", `type AliasedName = Name;`, false},
		{"commonTypeEntityRef", `type UserRef = MyApp::User;`, false},
		{"commonTypeExtension", `type IP = __cedar::ipaddr;`, false},
		{"commonTypeDecimal", `type Price = __cedar::decimal;`, false},
		{"commonTypeDatetime", `type Timestamp = __cedar::datetime;`, false},
		{"commonTypeDuration", `type Timeout = __cedar::duration;`, false},

		// Actions
		{"simpleAction", `action view;`, false},
		{"actionQuotedName", `action "view document";`, false},
		{"actionWithPrincipal", `action view appliesTo { principal: User };`, false},
		{"actionWithResource", `action view appliesTo { resource: Document };`, false},
		{"actionWithContext", `action view appliesTo { context: { ip: __cedar::ipaddr } };`, false},
		{"actionWithContextTypeRef", `action view appliesTo { context: RequestContext };`, false},
		{"actionWithMultiplePrincipals", `action view appliesTo { principal: [User, Admin, ServiceAccount] };`, false},
		{"actionWithMemberOf", `action view in "readActions";`, false},
		{"actionWithMemberOfList", `action view in ["readActions", "allActions"];`, false},
		{"actionWithExplicitEntityRef", `action view in Action::"allActions";`, false},
		{"actionWithFullEntityRef", `action view in MyApp::Action::"allActions";`, false},
		{"actionWithAppliesTo", `action view appliesTo {
			principal: User,
			resource: Document,
			context: { authenticated: Bool },
		};`, false},

		// Namespaces
		{"namespace", `namespace MyApp { entity User; }`, false},
		{"namespaceWithPath", `namespace MyApp::Core { entity User; }`, false},
		{"namespaceWithMultiple", `namespace MyApp {
			type Name = String;
			entity User;
			entity Group;
			action view;
		}`, false},

		// Annotations
		{"annotation", `@doc("A user entity") entity User;`, false},
		{"annotationNoValue", `@deprecated entity User;`, false},
		{"multipleAnnotations", `@doc("A user") @deprecated entity User;`, false},
		{"namespaceAnnotation", `@doc("My namespace") namespace MyApp { entity User; }`, false},

		// Comments (should be ignored)
		{"lineComment", `// This is a comment
entity User;`, false},
		{"blockComment", `/* Block comment */ entity User;`, false},

		// Complex real-world examples
		{"photoFlash", `namespace PhotoFlash {
			entity User in UserGroup = {
				"department": String,
				"jobLevel": Long,
			} tags String;
			entity UserGroup;
			entity Album in Album = {
				"account": Account,
				"private": Bool,
			};
			entity Account {
				"admins"?: Set<User>,
				"owner": User,
			};
			entity Photo in [Album, PhotoFlash::Other::Album] = {
				"account": Account,
				"private": Bool,
			};
			entity PhotoFormat enum ["jpg", "gif", "png"];
			type LongAlias = Long;
			action "uploadPhoto" appliesTo {
				principal: User,
				resource: [Album, Photo],
				context: {
					"authenticated": Bool,
					"photo": {
						"file_size": LongAlias,
						"file_type": String,
					},
				},
			};
			action "read";
			action "viewPhoto" in ["read"] appliesTo {
				principal: User,
				resource: Photo,
				context: { "authenticated": Bool },
			};
			action "listAlbums" in "read" appliesTo {
				principal: User,
				resource: Account,
				context: { "authenticated": Bool },
			};
		}`, false},

		// Edge cases
		{"recordWithQuotedKeys", `type Data = { "special-key": String };`, false},
		{"emptyRecord", `type Empty = {};`, false},
		{"deeplyNestedRecord", `type Nested = {
			level1: {
				level2: {
					level3: String,
				},
			},
		};`, false},
		{"setOfEntityRefs", `type UserSet = Set<MyApp::User>;`, false},

		// Error cases
		{"missingEntityName", `entity;`, true},
		{"missingSemicolon", `entity User`, true},
		{"missingBrace", `namespace Foo { entity User;`, true},
		{"invalidToken", `foo bar;`, true},
		{"invalidType", `type Foo = 123;`, true},
		{"invalidNamespaceDecl", `namespace Foo { invalid; }`, true},
		{"missingAnnotationParen", `@doc("test" entity User;`, true},
		{"missingEnumBracket", `entity Status enum "a", "b";`, true},
		{"invalidAppliesToContent", `action view appliesTo { invalid: User };`, true},
		{"missingSetElement", `type Tags = Set;`, true},
		{"missingCommonTypeEquals", `type Foo String;`, true},
		{"missingRecordColon", `type Data = { key String };`, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.in))
			if tt.err {
				testutil.Error(t, err)
				return
			}
			testutil.OK(t, err)

			// Round-trip test: marshal and unmarshal again
			if schema != nil && len(schema.Nodes) > 0 {
				cedarBytes := schema.MarshalCedar()
				schema2, err := schema2.UnmarshalCedar(cedarBytes)
				testutil.OK(t, err)
				testutil.Equals(t, len(schema2.Nodes), len(schema.Nodes))
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	// Test that we can parse -> marshal -> parse and get equivalent results
	tests := []struct {
		name string
		in   string
	}{
		{"entity", `entity User;`},
		{"entityWithShape", `entity User {
	name: String,
	email?: String,
};`},
		{"namespace", `namespace MyApp {
	entity User;
	entity Document;
}
`},
		{"action", `action view appliesTo {
	principal: User,
	resource: Document,
};`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema1, err := schema2.UnmarshalCedar([]byte(tt.in))
			testutil.OK(t, err)

			cedarBytes := schema1.MarshalCedar()
			schema2, err := schema2.UnmarshalCedar(cedarBytes)
			testutil.OK(t, err)

			testutil.Equals(t, len(schema2.Nodes), len(schema1.Nodes))
		})
	}
}

func TestUnmarshalCedarLarge(t *testing.T) {
	t.Parallel()

	// Test with a large schema that exercises many features
	src := `
// PhotoFlash schema - a comprehensive example
@doc("PhotoFlash namespace containing all entities and actions")
namespace PhotoFlash {
	// Type aliases
	type EmailAddress = String;
	type UserName = String;
	type IPAddress = __cedar::ipaddr;

	// Base entities
	@doc("A user in the system")
	entity User in [UserGroup, Team] = {
		"email": EmailAddress,
		"name": UserName,
		"department"?: String,
		"jobLevel"?: Long,
		"active": Bool,
	} tags String;

	entity UserGroup;
	entity Team;

	// Resources
	entity Album in Album = {
		"account": Account,
		"private": Bool,
		"createdAt": __cedar::datetime,
	};

	entity Account {
		"admins"?: Set<User>,
		"owner": User,
	};

	entity Photo in [Album] = {
		"account": Account,
		"private": Bool,
		"format": PhotoFormat,
	};

	// Enum entity
	entity PhotoFormat enum ["jpg", "gif", "png", "webp"];

	// Context type
	type RequestContext = {
		"authenticated": Bool,
		"ip"?: IPAddress,
		"timestamp": __cedar::datetime,
	};

	// Actions
	action "uploadPhoto" appliesTo {
		principal: User,
		resource: Album,
		context: RequestContext,
	};

	action "viewPhoto" in ["read"] appliesTo {
		principal: [User, UserGroup],
		resource: Photo,
		context: RequestContext,
	};

	action "deletePhoto" in ["write"] appliesTo {
		principal: User,
		resource: Photo,
		context: RequestContext,
	};

	action "read";
	action "write";
	action "admin" in ["read", "write"];
}
`
	schema, err := schema2.UnmarshalCedar([]byte(src))
	testutil.OK(t, err)

	ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, ns.Name, "PhotoFlash")
	testutil.Equals(t, len(ns.Annotations), 1)

	// Verify we got all declarations
	typeCount := 0
	entityCount := 0
	enumCount := 0
	actionCount := 0
	for _, decl := range ns.Declarations {
		switch decl.(type) {
		case *ast.CommonTypeNode:
			typeCount++
		case *ast.EntityNode:
			entityCount++
		case *ast.EnumNode:
			enumCount++
		case *ast.ActionNode:
			actionCount++
		}
	}
	testutil.Equals(t, typeCount, 4)   // EmailAddress, UserName, IPAddress, RequestContext
	testutil.Equals(t, entityCount, 6) // User, UserGroup, Team, Album, Account, Photo
	testutil.Equals(t, enumCount, 1)   // PhotoFormat
	testutil.Equals(t, actionCount, 6) // uploadPhoto, viewPhoto, deletePhoto, read, write, admin

	// Round-trip test
	cedarBytes := schema.MarshalCedar()
	testutil.Equals(t, len(cedarBytes) > 0, true)

	schema2, err := schema2.UnmarshalCedar(cedarBytes)
	testutil.OK(t, err)
	testutil.Equals(t, len(schema2.Nodes), 1)
}

func TestMarshalCedarFormats(t *testing.T) {
	t.Parallel()

	t.Run("marshal preserves structure", func(t *testing.T) {
		t.Parallel()
		schema := ast.NewSchema(
			ast.Entity("User").
				MemberOf(ast.EntityType("Group")).
				Shape(
					ast.Attribute("name", ast.String()),
					ast.Optional("email", ast.String()),
				).
				Annotate("doc", "A user entity"),
		)

		cedarBytes := schema.MarshalCedar()
		result := string(cedarBytes)

		// Verify key elements are present
		testutil.Equals(t, strings.Contains(result, "entity User"), true)
		testutil.Equals(t, strings.Contains(result, "in Group"), true)
		testutil.Equals(t, strings.Contains(result, "name: String"), true)
		testutil.Equals(t, strings.Contains(result, "email?: String"), true)
		testutil.Equals(t, strings.Contains(result, "@doc"), true)
	})
}
