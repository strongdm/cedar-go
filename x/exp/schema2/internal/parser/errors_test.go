package parser_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/parser"
)

func TestParserErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		// Syntax errors
		{"unexpected token", `foo bar;`},
		{"missing semicolon", `entity User`},
		{"missing brace", `namespace Foo { entity User;`},
		{"invalid type", `type Foo = 123;`},
		{"invalid namespace declaration", `namespace Foo { invalid; }`},
		{"missing annotation paren", `@doc("test" entity User;`},
		{"missing entity name", `entity;`},
		{"missing enum bracket", `entity Status enum "a", "b";`},
		{"invalid appliesTo content", `action view appliesTo { invalid: User };`},
		{"missing Set element type", `type Tags = Set;`},
		{"missing common type equals", `type Foo String;`},
		{"missing annotation identifier", `@123 entity User;`},
		{"invalid annotation string", `@doc(123) entity User;`},
		{"missing namespace identifier", `namespace 123 { entity User; }`},
		{"missing namespace open brace", `namespace Foo entity User; }`},
		{"missing enum closing bracket", `entity Status enum ["a", "b";`},
		{"invalid enum value", `entity Status enum [123];`},
		{"missing entity memberOf identifier", `entity User in 123;`},
		{"missing entity memberOf list closing bracket", `entity User in [Group, Team;`},
		{"missing shape open brace pair", `entity User { name String };`},
		{"invalid tags type", `entity Document tags 123;`},
		{"missing action name", `action 123;`},
		{"missing action memberOf ref", `action view in 123;`},
		{"missing action memberOf list closing bracket", `action view in ["read", "write";`},
		{"missing appliesTo open brace", `action view appliesTo principal: User };`},
		{"missing principal colon", `action view appliesTo { principal User };`},
		{"missing resource colon", `action view appliesTo { resource User };`},
		{"missing context colon", `action view appliesTo { context { ip: String } };`},
		{"invalid principal type", `action view appliesTo { principal: 123 };`},
		{"invalid resource type", `action view appliesTo { resource: 123 };`},
		{"invalid context type", `action view appliesTo { context: 123 };`},
		{"missing appliesTo closing brace", `action view appliesTo { principal: User;`},
		{"missing common type name", `type = String;`},
		{"invalid common type value", `type Foo = ;`},
		{"missing set closing angle bracket", `type Tags = Set<String;`},
		{"invalid set element type", `type Tags = Set<123>;`},
		{"missing extension double colon", `type IP = __cedar ipaddr;`},
		{"missing extension name", `type IP = __cedar::123;`},
		{"missing record closing brace", `type Data = { name: String;`},
		{"missing record pair colon", `type Data = { name String };`},
		{"invalid record pair type", `type Data = { name: 123 };`},
		{"missing path after double colon", `type Ref = MyApp::123;`},
		{"missing entity ref path component", `action view in MyApp::123;`},
		{"namespace annotation error in declaration", `namespace MyApp { @123 entity User; }`},
		{"namespace type declaration error", `namespace MyApp { type = String; }`},
		{"namespace action error", `namespace MyApp { action 123; }`},
		{"entity memberOf list parse error", `entity User in [123];`},
		{"entity shape parse error", `entity User { 123: String };`},
		{"action entity refs list error", `action view in [123];`},
		{"common type semicolon error", `type Foo = String`},
		{"record pair in shape error", `entity User { ?: String };`},
		{"record pair annotation error", `type Data = { @123 name: String };`},

		// Duplicate key errors
		{"duplicate annotation", `@doc("first") @doc("second") entity User;`},
		{"duplicate entity", `entity User; entity User;`},
		{"duplicate action", `action view; action view;`},
		{"duplicate common type", `type Name = String; type Name = Long;`},
		{"duplicate namespace", `namespace App {} namespace App {}`},
		{"duplicate enum", `entity Status enum ["a"]; entity Status enum ["b"];`},
		{"entity conflicts with enum", `entity Status enum ["a"]; entity Status;`},
		{"enum conflicts with entity", `entity Status; entity Status enum ["a"];`},
		{"duplicate attribute in record", `type User = { name: String, name: Long };`},
		{"duplicate attribute in entity shape", `entity User { name: String, name: Long };`},
		{"duplicate entity in namespace", `namespace App { entity User; entity User; }`},
		{"duplicate action in namespace", `namespace App { action view; action view; }`},
		{"duplicate common type in namespace", `namespace App { type Name = String; type Name = Long; }`},
		{"duplicate enum in namespace", `namespace App { entity Status enum ["a"]; entity Status enum ["b"]; }`},
		{"entity conflicts with enum in namespace", `namespace App { entity Status enum ["a"]; entity Status; }`},
		{"enum conflicts with entity in namespace", `namespace App { entity Status; entity Status enum ["a"]; }`},
		{"duplicate annotation on attribute", `type User = { @doc("a") @doc("b") name: String };`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := parser.ParseSchema("", []byte(tt.input))
			testutil.Equals(t, err != nil, true)
		})
	}
}
