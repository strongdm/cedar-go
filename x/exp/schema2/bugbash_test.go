package schema2_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

// TestCorpus runs all schema files in testdata/corpus against the parser.
// Files in corpus/valid/ must parse successfully and round-trip.
// Files in corpus/invalid/ must fail to parse.
func TestCorpus(t *testing.T) {
	t.Parallel()

	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		files, err := filepath.Glob("testdata/corpus/valid/*.cedarschema")
		testutil.OK(t, err)

		for _, file := range files {
			file := file
			name := filepath.Base(file)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				data, err := os.ReadFile(file)
				testutil.OK(t, err)

				// Parse
				schema, err := schema2.UnmarshalCedar(data)
				if err != nil {
					t.Fatalf("failed to parse valid schema %s: %v", name, err)
				}

				// Round-trip
				marshaled := schema.MarshalCedar()
				_, err = schema2.UnmarshalCedar(marshaled)
				if err != nil {
					t.Fatalf("round-trip failed for %s: %v\nmarshaled:\n%s", name, err, string(marshaled))
				}
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		files, err := filepath.Glob("testdata/corpus/invalid/*.cedarschema")
		testutil.OK(t, err)

		for _, file := range files {
			file := file
			name := filepath.Base(file)
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				data, err := os.ReadFile(file)
				testutil.OK(t, err)

				_, err = schema2.UnmarshalCedar(data)
				if err == nil {
					t.Fatalf("expected parse error for invalid schema %s", name)
				}
			})
		}
	})
}

// TestRoundTripStability ensures parse->marshal->parse->marshal is stable.
func TestRoundTripStability(t *testing.T) {
	t.Parallel()

	inputs := []string{
		`entity User;`,
		`entity User in [Group]; entity Group;`,
		`entity User { name: String, email?: String };`,
		`entity Status enum ["active", "inactive"];`,
		`type Name = String;`,
		`action view appliesTo { principal: [User], resource: [Doc] }; entity User; entity Doc;`,
		`namespace App { entity User; action view; }`,
		`@doc("test") entity User;`,
		`entity User, Admin, Guest;`,
	}

	for _, input := range inputs {
		input := input
		name := input
		if len(name) > 40 {
			name = name[:40] + "..."
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			schema1, err := schema2.UnmarshalCedar([]byte(input))
			testutil.OK(t, err)

			out1 := schema1.MarshalCedar()
			schema2Parsed, err := schema2.UnmarshalCedar(out1)
			testutil.OK(t, err)

			out2 := schema2Parsed.MarshalCedar()
			if string(out1) != string(out2) {
				t.Errorf("round trip unstable:\nfirst:\n%s\nsecond:\n%s", out1, out2)
			}
		})
	}
}

// TestParserDoesNotPanic ensures malformed inputs don't cause panics.
func TestParserDoesNotPanic(t *testing.T) {
	t.Parallel()

	badInputs := []string{
		"{{{{", "}}}}", ";;;;", "@@@", ":::", `"""`,
		"entity", "entity User {", "entity User { name",
		"namespace Foo {", "@doc(", string([]byte{0x00}),
		strings.Repeat("entity User;", 1000),
	}

	for i, input := range badInputs {
		input := input
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("panic on malformed input: %v", r)
				}
			}()
			_, _ = schema2.UnmarshalCedar([]byte(input))
		})
	}
}
