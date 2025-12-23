package schema2_test

import (
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

// TestBugBash_EdgeCases tests various edge cases to find bugs
func TestBugBash_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		// Empty and whitespace
		{"empty input", "", false},
		{"just whitespace", "   \t\n\r  ", false},
		{"just comments", "// comment\n/* block */", false},

		// Trailing commas - BUG: our parser accepts these but reference rejects
		{"entity trailing comma in memberOf", "entity User in [Group,];", true},
		{"action trailing comma in appliesTo principals", "entity User; entity Doc; action view appliesTo { principal: [User,], resource: Doc };", true},
		{"record trailing comma allowed", "entity User { name: String, };", false},

		// Multiple/leading commas
		{"entity double comma", "entity User,, Admin;", true},
		{"entity leading comma", "entity , User;", true},
		{"action double comma", "action read,, write;", true},

		// Empty declarations
		{"empty namespace", "namespace Empty {}", false},
		{"entity with empty shape", "entity User {};", false},
		{"entity with empty memberOf list", "entity User in [];", false}, // valid in reference

		// Reserved words as identifiers
		{"entity named entity", "entity entity;", false},       // context-sensitive, valid
		{"entity named action", "entity action;", false},       // context-sensitive, valid
		{"entity named namespace", "entity namespace;", false}, // context-sensitive, valid
		{"entity named type", "entity type;", false},           // context-sensitive, valid
		{"entity named in", "entity in;", true},                // BUG: 'in' is a reserved identifier
		{"entity named enum", "entity enum;", false},           // context-sensitive, valid
		{"entity named appliesTo", "entity appliesTo;", false}, // context-sensitive, valid
		{"entity named tags", "entity tags;", false},           // context-sensitive, valid

		// Self-referential and circular
		{"self-referential entity", "entity Group in [Group];", false},
		{"mutual reference", "entity A in [B]; entity B in [A];", false},

		// Very long inputs
		{"very long entity name", "entity " + strings.Repeat("A", 10000) + ";", false},
		{"very long attribute name", "entity User { " + strings.Repeat("x", 10000) + ": String };", false},
		{"very long string value", `entity User { name: String }; // ` + strings.Repeat("x", 10000), false},

		// Deeply nested structures
		{"deeply nested record", "type Deep = { a: { b: { c: { d: { e: { f: String } } } } } };", false},
		{"deeply nested set", "type Deep = Set<Set<Set<Set<Set<String>>>>>;", false},

		// Unicode
		{"unicode entity name", "entity 用户;", true}, // probably not valid
		{"unicode in attribute name", `entity User { "名前": String };`, false},
		{"unicode in string", `type Name = String; // 日本語コメント`, false},
		{"emoji in comment", `entity User; // 🎉`, false},
		{"emoji in string key", `entity User { "🔑": String };`, false},

		// Escape sequences in strings
		{"newline in string", `entity User { "multi\nline": String };`, false},
		{"tab in string", `entity User { "with\ttab": String };`, false},
		{"quote in string", `entity User { "with\"quote": String };`, false},
		{"backslash in string", `entity User { "with\\slash": String };`, false},

		// Numbers in various positions
		{"number as entity name", "entity 123;", true},
		{"number in path", "entity Ns::123::User;", true},
		{"negative number anywhere", "type Neg = -123;", true},

		// Extension types
		{"all extension types", `type A = __cedar::ipaddr; type B = __cedar::decimal; type C = __cedar::datetime; type D = __cedar::duration;`, false},
		{"unknown extension", "type X = __cedar::unknown;", false}, // parsing succeeds, resolution would fail
		{"extension without cedar prefix", "type X = __foo::bar;", false}, // parsing succeeds, resolution would fail

		// Actions edge cases - semantic validation, not parsing (we don't validate principal/resource requirements)
		{"action with only context", "entity U; entity D; action view appliesTo { context: { ip: String } };", false}, // semantic error
		{"action with only principal", "entity User; action view appliesTo { principal: User };", false},              // semantic error
		{"action with only resource", "entity Doc; action view appliesTo { resource: Doc };", false},                  // semantic error
		{"action memberOf self", `action view in ["view"];`, false},
		{"action with entity ref in memberOf", `action view in Action::"parent";`, false},

		// Annotations
		{"annotation with empty string", `@doc("") entity User;`, false},
		{"annotation with special chars", `@doc("special: <>\"'") entity User;`, false},
		{"multiple same annotations", `@doc("one") @doc("two") entity User;`, false},
		{"annotation on everything", `@a namespace N { @b type T = String; @c entity E; @d action act; }`, false},

		// Paths - entity names cannot contain paths
		{"entity with path name", "entity A::B::C;", true}, // entity names don't support paths
		{"path ending with double colon", "entity A::B::;", true},
		{"path starting with double colon", "entity ::A::B;", true},
		{"path with triple colon", "entity A:::B;", true},

		// Comments
		{"nested block comments", "/* outer /* inner */ still outer */ entity User;", true}, // or should work?
		{"unclosed block comment", "entity User; /* unclosed", true},
		{"comment inside string", `entity User { "/* not a comment */": String };`, false},

		// Mixed declarations
		{"interleaved types and entities", "type A = String; entity B; type C = Long; entity D;", false},
		{"all declaration types", "namespace N { type T = String; entity E; entity F enum [\"a\"]; action A; }", false},

		// Enum edge cases
		{"enum with one value", `entity Status enum ["single"];`, false},
		{"enum with empty string", `entity Status enum [""];`, false},
		{"enum with duplicate values", `entity Status enum ["a", "a"];`, false}, // should this error?
		{"enum with special chars", `entity Status enum ["with space", "with\ttab"];`, false},

		// Optional in various places
		{"optional only attribute", "entity User { name?: String };", false},
		{"all optional attributes", "entity User { a?: String, b?: Long, c?: Bool };", false},
		{"optional in nested record", "entity User { config: { inner?: String } };", false},

		// Type references
		{"forward type reference", "entity User { data: NotYetDefined }; type NotYetDefined = String;", false},
		{"circular type reference", "type A = { b: B }; type B = { a: A };", false},
		{"self-referential type", "type Node = { children: Set<Node> };", false},

		// Empty things
		{"empty type record", "type Empty = {};", false},
		{"action with empty context", "entity U; entity D; action view appliesTo { principal: U, resource: D, context: {} };", false},

		// Semicolon edge cases
		{"double semicolon", "entity User;; entity Admin;", true},
		{"semicolon in namespace", "namespace N { ; }", true},

		// Brace matching
		{"extra closing brace", "entity User { name: String }};", true},
		{"extra opening brace", "entity User {{ name: String };", true},
		{"mismatched braces", "entity User { name: String ];", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := schema2.UnmarshalCedar([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for input: %s", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v for input: %s", err, tt.input)
				}
			}
		})
	}
}

// TestBugBash_RoundTrip tests that parse->marshal->parse produces equivalent results
func TestBugBash_RoundTrip(t *testing.T) {
	t.Parallel()

	inputs := []string{
		`entity User;`,
		`entity User in [Group]; entity Group;`,
		`entity User { name: String, age?: Long };`,
		`entity Status enum ["active", "inactive"];`,
		`type Name = String;`,
		`type Complex = { nested: { deep: Set<String> } };`,
		`action view appliesTo { principal: [User], resource: [Doc] }; entity User; entity Doc;`,
		`action read, write, delete appliesTo { principal: User, resource: Doc }; entity User; entity Doc;`,
		`namespace App { entity User; action view; }`,
		`@doc("test") entity User;`,
		`entity User, Admin, Guest;`,
		`entity User in [Group] { name: String } tags String; entity Group;`,
	}

	for _, input := range inputs {
		input := input
		t.Run(input[:min(30, len(input))], func(t *testing.T) {
			t.Parallel()

			// Parse
			schema1, err := schema2.UnmarshalCedar([]byte(input))
			if err != nil {
				t.Fatalf("initial parse failed: %v", err)
			}

			// Marshal
			output1 := schema1.MarshalCedar()

			// Parse again
			schema2Parsed, err := schema2.UnmarshalCedar(output1)
			if err != nil {
				t.Fatalf("second parse failed: %v\nmarshaled output:\n%s", err, string(output1))
			}

			// Marshal again
			output2 := schema2Parsed.MarshalCedar()

			// The two marshaled outputs should be identical
			if string(output1) != string(output2) {
				t.Errorf("round trip not stable:\nfirst marshal:\n%s\nsecond marshal:\n%s", string(output1), string(output2))
			}
		})
	}
}

// TestBugBash_FuzzLike tests random-ish malformed inputs
func TestBugBash_FuzzLike(t *testing.T) {
	t.Parallel()

	// These should all fail gracefully (not panic)
	badInputs := []string{
		"{{{{",
		"}}}}",
		"[[[[",
		"]]]]",
		";;;;",
		"@@@",
		":::",
		"<<<",
		">>>",
		`"""`,
		"entity",
		"action",
		"type",
		"namespace",
		"in",
		"appliesTo",
		"entity User in",
		"entity User {",
		"entity User { name",
		"entity User { name:",
		"entity User { name: }",
		"action view appliesTo",
		"action view appliesTo {",
		"action view appliesTo { principal",
		"action view appliesTo { principal:",
		"namespace Foo",
		"namespace Foo {",
		"namespace { }",
		"type =",
		"type Foo",
		"type Foo =",
		"@",
		"@doc",
		"@doc(",
		"@doc()",
		"@doc(\"",
		string([]byte{0x00}),                  // null byte
		string([]byte{0xff, 0xfe}),            // invalid utf8
		string([]byte{0x80, 0x81, 0x82}),      // invalid utf8
		strings.Repeat("entity User;", 10000), // lots of entities
		"entity " + strings.Repeat("(", 1000), // stack overflow attempt
		"entity User { " + strings.Repeat("x", 100000) + " }",
	}

	for i, input := range badInputs {
		input := input
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Parallel()
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on input %q: %v", input[:min(50, len(input))], r)
				}
			}()
			schema2.UnmarshalCedar([]byte(input))
		})
	}
}

// TestBugBash_MarshalSpecialChars tests marshaling handles special characters correctly
func TestBugBash_MarshalSpecialChars(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name   string
		input  string
		expect string // substring that must be in output
	}{
		{
			name:   "action name with space needs quoting",
			input:  `action "view doc";`,
			expect: `"view doc"`,
		},
		{
			name:   "attribute with dash needs quoting",
			input:  `entity User { "my-attr": String };`,
			expect: `"my-attr"`,
		},
		{
			name:   "attribute starting with number needs quoting",
			input:  `entity User { "123key": String };`,
			expect: `"123key"`,
		},
		{
			name:   "attribute with space needs quoting",
			input:  `entity User { "my attr": String };`,
			expect: `"my attr"`,
		},
		{
			name:   "normal identifier no quoting",
			input:  `entity User { myAttr: String };`,
			expect: `myAttr:`,
		},
		{
			name:   "underscore start no quoting",
			input:  `entity User { _private: String };`,
			expect: `_private:`,
		},
	}

	for _, tt := range inputs {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.input))
			testutil.OK(t, err)

			output := string(schema.MarshalCedar())
			if !strings.Contains(output, tt.expect) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.expect, output)
			}
		})
	}
}

// TestBugBash_CommaEdgeCases specifically tests comma-separated declaration edge cases
func TestBugBash_CommaEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
		count   int // expected number of nodes
	}{
		{"single entity", "entity User;", false, 1},
		{"two entities", "entity User, Admin;", false, 2},
		{"three entities", "entity User, Admin, Guest;", false, 3},
		{"entities with shared in", "entity Group; entity User, Admin in [Group];", false, 3},
		{"entities with shared shape", "entity User, Admin { name: String };", false, 2},
		{"entities with shared tags", "entity User, Admin tags String;", false, 2},
		{"entities with all shared", "entity Group; entity User, Admin in [Group] { name: String } tags Long;", false, 3},

		{"single action", "action view;", false, 1},
		{"two actions", "action view, edit;", false, 2},
		{"three actions", "action view, edit, delete;", false, 3},
		{"actions with shared memberOf", `action parent; action view, edit in ["parent"];`, false, 3},
		{"actions with shared appliesTo", "entity U; entity D; action view, edit appliesTo { principal: U, resource: D };", false, 4},

		// Error cases
		{"comma at end of entity", "entity User,;", true, 0},
		{"comma at start of entity", "entity , User;", true, 0},
		{"double comma entity", "entity User,, Admin;", true, 0},
		{"comma at end of action", "action view,;", true, 0},
		{"comma at start of action", "action , view;", true, 0},
		{"double comma action", "action view,, edit;", true, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got success")
				}
				return
			}

			testutil.OK(t, err)
			if len(schema.Nodes) != tt.count {
				t.Errorf("expected %d nodes, got %d", tt.count, len(schema.Nodes))
			}
		})
	}
}

// TestBugBash_ReferenceComparison tests our output against what we know the reference accepts
func TestBugBash_ReferenceComparison(t *testing.T) {
	t.Parallel()

	// These are patterns from the reference implementation that should work
	validPatterns := []string{
		// From reference tests
		`entity Video in [Account, Album];`,
		`entity User in [UserGroup] { department: String, jobLevel: Long };`,
		`entity Account in [AccountGroup] { owner?: User };`,
		`action view, delete, edit appliesTo { principal: [User], resource: [Photo, Album], context: { source_ip: __cedar::ipaddr } }; entity User; entity Photo; entity Album;`,
		`type Task = { "id": Long, "name": String, "state": String };`,
		`type Tasks = Set<Task>;`,
		`entity Application enum ["TinyTodo"];`,

		// Tricky patterns
		`entity Team, UserGroup in [UserGroup];`,
		`entity Issue = { "repo": Repository, "reporter": User }; entity Repository; entity User;`,
		`action "Role-A Actions";`,
		`action UpdateWorkspace in ["Role-A Actions"] appliesTo { principal: User, resource: Workspace }; entity User; entity Workspace;`,
	}

	for _, input := range validPatterns {
		input := input
		t.Run(input[:min(40, len(input))], func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(input))
			if err != nil {
				t.Fatalf("failed to parse valid pattern: %v\ninput: %s", err, input)
			}

			// Round trip should work
			output := schema.MarshalCedar()
			_, err = schema2.UnmarshalCedar(output)
			if err != nil {
				t.Fatalf("failed to parse marshaled output: %v\noutput:\n%s", err, string(output))
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
