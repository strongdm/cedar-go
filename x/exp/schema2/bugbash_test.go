package schema2_test

import (
	"encoding/json"
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

// TestJSONCorpus tests JSON marshalling/unmarshalling with corpus files.
// For each valid Cedar schema, we test:
// 1. Cedar -> JSON -> Cedar round-trip
// 2. JSON -> Cedar -> JSON round-trip (if .json file exists)
func TestJSONCorpus(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/corpus/valid/*.cedarschema")
	testutil.OK(t, err)

	for _, file := range files {
		file := file
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Parse Cedar schema
			cedarData, err := os.ReadFile(file)
			testutil.OK(t, err)

			schema, err := schema2.UnmarshalCedar(cedarData)
			if err != nil {
				t.Fatalf("failed to parse Cedar schema: %v", err)
			}

			// Cedar -> JSON
			jsonData, err := schema2.MarshalJSON(schema)
			if err != nil {
				t.Fatalf("failed to marshal to JSON: %v", err)
			}

			// JSON -> Cedar (parse JSON back)
			schema2Parsed, err := schema2.UnmarshalJSON(jsonData)
			if err != nil {
				t.Fatalf("failed to parse JSON schema: %v\nJSON:\n%s", err, string(jsonData))
			}

			// Verify JSON validates with reference implementation
			verifyJSONWithCedarCLI(t, string(jsonData))

			// JSON -> Cedar -> JSON round-trip stability
			jsonData2, err := schema2.MarshalJSON(schema2Parsed)
			if err != nil {
				t.Fatalf("failed to re-marshal to JSON: %v", err)
			}

			// Parse the re-marshaled JSON to verify it's valid
			_, err = schema2.UnmarshalJSON(jsonData2)
			if err != nil {
				t.Fatalf("failed to parse re-marshaled JSON: %v\nJSON:\n%s", err, string(jsonData2))
			}
		})
	}
}

// TestJSONRoundTripStability tests JSON round-trip stability.
func TestJSONRoundTripStability(t *testing.T) {
	t.Parallel()

	inputs := []string{
		`entity User;`,
		`entity User in [Group]; entity Group;`,
		`entity User { name: String, email?: String };`,
		`type Name = String;`,
		`action view appliesTo { principal: [User], resource: [Doc] }; entity User; entity Doc;`,
		`namespace App { entity User; action view; }`,
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

			// Parse Cedar
			schema1, err := schema2.UnmarshalCedar([]byte(input))
			testutil.OK(t, err)

			// Cedar -> JSON
			json1, err := schema2.MarshalJSON(schema1)
			testutil.OK(t, err)

			// JSON -> Schema
			schema2Parsed, err := schema2.UnmarshalJSON(json1)
			testutil.OK(t, err)

			// Schema -> JSON again
			json2, err := schema2.MarshalJSON(schema2Parsed)
			testutil.OK(t, err)

			// Verify JSON round-trip stability
			if string(json1) != string(json2) {
				t.Errorf("JSON round trip unstable:\nfirst:\n%s\nsecond:\n%s", json1, json2)
			}

			// Verify with reference
			verifyJSONWithCedarCLI(t, string(json1))
		})
	}
}

// TestJSONCorpusFiles tests JSON files in the corpus directly.
// For each .json file, we test: JSON -> Cedar -> JSON round-trip.
func TestJSONCorpusFiles(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/corpus/valid/*.json")
	testutil.OK(t, err)

	for _, file := range files {
		file := file
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Read JSON file
			jsonData, err := os.ReadFile(file)
			testutil.OK(t, err)

			// JSON -> Schema
			schema, err := schema2.UnmarshalJSON(jsonData)
			if err != nil {
				t.Fatalf("failed to parse JSON schema: %v", err)
			}

			// Schema -> Cedar (to verify we can marshal to Cedar format)
			cedarData := schema.MarshalCedar()
			if len(cedarData) == 0 {
				t.Fatalf("failed to marshal schema to Cedar format")
			}

			// Verify Cedar output with reference implementation
			verifyWithCedarCLI(t, string(cedarData))

			// Schema -> JSON
			jsonData2, err := schema2.MarshalJSON(schema)
			if err != nil {
				t.Fatalf("failed to marshal schema to JSON: %v", err)
			}

			// Verify JSON output with reference implementation
			verifyJSONWithCedarCLI(t, string(jsonData2))
		})
	}
}

// TestJSONSchemaFormatCompliance tests that our JSON output matches the Cedar JSON schema spec.
// This ensures compatibility with the Rust reference implementation.
func TestJSONSchemaFormatCompliance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		cedar          string
		expectedFields []string // Fields that must be present in JSON
		mustNotHave    []string // Fields that must NOT be present
	}{
		{
			name:           "entity type without shape has empty object",
			cedar:          `entity User;`,
			expectedFields: []string{`"entityTypes"`, `"User": {}`},
			mustNotHave:    []string{`"shape"`},
		},
		{
			name:           "entity with shape has Record type",
			cedar:          `entity User { name: String };`,
			expectedFields: []string{`"shape"`, `"type": "Record"`, `"attributes"`},
		},
		{
			name:           "optional attribute has required false",
			cedar:          `entity User { email?: String };`,
			expectedFields: []string{`"required": false`},
		},
		{
			name:           "required attribute has no required field",
			cedar:          `entity User { name: String };`,
			mustNotHave:    []string{`"required"`},
		},
		{
			name:           "entity reference uses Entity type with name",
			cedar:          `entity User; entity Doc { owner: User };`,
			expectedFields: []string{`"type": "Entity"`, `"name": "User"`},
		},
		{
			name:           "Set type uses element field",
			cedar:          `entity User { tags: Set<String> };`,
			expectedFields: []string{`"type": "Set"`, `"element"`},
		},
		{
			name:           "Extension type uses name field",
			cedar:          `entity User { ip: __cedar::ipaddr };`,
			expectedFields: []string{`"type": "Extension"`, `"name": "ipaddr"`},
		},
		{
			name:           "memberOf uses array of entity UIDs",
			cedar:          `entity Group; entity User in [Group];`,
			expectedFields: []string{`"memberOfTypes"`, `"Group"`},
		},
		{
			name:           "action appliesTo has principalTypes and resourceTypes",
			cedar:          `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc };`,
			expectedFields: []string{`"appliesTo"`, `"principalTypes"`, `"resourceTypes"`},
		},
		{
			name:           "action context has Record type",
			cedar:          `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc, context: { ip: __cedar::ipaddr } };`,
			expectedFields: []string{`"context"`, `"type": "Record"`},
		},
		{
			name:           "action memberOf uses type and id",
			cedar:          `action read; action write in [read];`,
			expectedFields: []string{`"memberOf"`, `"type": "Action"`, `"id": "read"`},
		},
		{
			name:           "namespace creates nested JSON object",
			cedar:          `namespace App { entity User; }`,
			expectedFields: []string{`"App"`, `"entityTypes"`, `"User"`},
			mustNotHave:    []string{`"": {`}, // Should not have empty namespace
		},
		{
			name:           "common type defined in commonTypes",
			cedar:          `type Name = String;`,
			expectedFields: []string{`"commonTypes"`, `"Name"`, `"type": "String"`},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.cedar))
			testutil.OK(t, err)

			jsonData, err := schema2.MarshalJSON(schema)
			testutil.OK(t, err)
			jsonStr := string(jsonData)

			for _, field := range tt.expectedFields {
				if !strings.Contains(jsonStr, field) {
					t.Errorf("JSON missing expected field %q:\n%s", field, jsonStr)
				}
			}

			for _, field := range tt.mustNotHave {
				if strings.Contains(jsonStr, field) {
					t.Errorf("JSON should not contain %q:\n%s", field, jsonStr)
				}
			}

			// Verify with reference implementation
			verifyJSONWithCedarCLI(t, jsonStr)
		})
	}
}

// TestJSONReferenceSchemasParsing tests that we can parse JSON schemas from the Rust implementation.
func TestJSONReferenceSchemasParsing(t *testing.T) {
	t.Parallel()

	// These are JSON schemas that match what the Rust implementation produces
	tests := []struct {
		name string
		json string
	}{
		{
			name: "minimal entity",
			json: `{"": {"entityTypes": {"User": {}}, "actions": {}}}`,
		},
		{
			name: "entity with memberOfTypes",
			json: `{"": {"entityTypes": {"User": {"memberOfTypes": ["Group"]}, "Group": {}}, "actions": {}}}`,
		},
		{
			name: "entity with shape",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "String"}}}}}, "actions": {}}}`,
		},
		{
			name: "entity with optional attribute",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"email": {"type": "String", "required": false}}}}}, "actions": {}}}`,
		},
		{
			name: "entity with Set attribute",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"tags": {"type": "Set", "element": {"type": "String"}}}}}}, "actions": {}}}`,
		},
		{
			name: "entity with Entity reference",
			json: `{"": {"entityTypes": {"User": {}, "Doc": {"shape": {"type": "Record", "attributes": {"owner": {"type": "Entity", "name": "User"}}}}}, "actions": {}}}`,
		},
		{
			name: "entity with Extension type",
			json: `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"ip": {"type": "Extension", "name": "ipaddr"}}}}}, "actions": {}}}`,
		},
		{
			name: "action with appliesTo",
			json: `{"": {"entityTypes": {"User": {}, "Doc": {}}, "actions": {"view": {"appliesTo": {"principalTypes": ["User"], "resourceTypes": ["Doc"]}}}}}`,
		},
		{
			name: "action with context",
			json: `{"": {"entityTypes": {"User": {}, "Doc": {}}, "actions": {"view": {"appliesTo": {"principalTypes": ["User"], "resourceTypes": ["Doc"], "context": {"type": "Record", "attributes": {"ip": {"type": "Extension", "name": "ipaddr"}}}}}}}}`,
		},
		{
			name: "action with memberOf",
			json: `{"": {"entityTypes": {}, "actions": {"read": {}, "write": {"memberOf": [{"type": "Action", "id": "read"}]}}}}`,
		},
		{
			name: "common type primitive",
			json: `{"": {"commonTypes": {"Name": {"type": "String"}}, "entityTypes": {}, "actions": {}}}`,
		},
		{
			name: "common type Record",
			json: `{"": {"commonTypes": {"Address": {"type": "Record", "attributes": {"city": {"type": "String"}, "zip": {"type": "String", "required": false}}}}, "entityTypes": {}, "actions": {}}}`,
		},
		{
			name: "common type Set",
			json: `{"": {"commonTypes": {"Tags": {"type": "Set", "element": {"type": "String"}}}, "entityTypes": {}, "actions": {}}}`,
		},
		{
			name: "named namespace",
			json: `{"App": {"entityTypes": {"User": {}}, "actions": {"view": {}}}}`,
		},
		{
			name: "multiple namespaces",
			json: `{"Core": {"entityTypes": {"Base": {}}, "actions": {}}, "App": {"entityTypes": {"User": {}}, "actions": {}}}`,
		},
		{
			name: "nested Record",
			json: `{"": {"entityTypes": {"Config": {"shape": {"type": "Record", "attributes": {"nested": {"type": "Record", "attributes": {"value": {"type": "String"}}}}}}}, "actions": {}}}`,
		},
		{
			name: "Set of Set",
			json: `{"": {"commonTypes": {"Matrix": {"type": "Set", "element": {"type": "Set", "element": {"type": "Long"}}}}, "entityTypes": {}, "actions": {}}}`,
		},
		{
			name: "type reference to common type",
			json: `{"": {"commonTypes": {"Name": {"type": "String"}}, "entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"name": {"type": "Name"}}}}}, "actions": {}}}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse JSON
			schema, err := schema2.UnmarshalJSON([]byte(tt.json))
			if err != nil {
				t.Fatalf("failed to parse JSON: %v\nJSON: %s", err, tt.json)
			}

			// Round-trip: JSON -> Schema -> JSON
			jsonData, err := schema2.MarshalJSON(schema)
			testutil.OK(t, err)

			// Parse again to verify
			_, err = schema2.UnmarshalJSON(jsonData)
			testutil.OK(t, err)

			// Verify with reference implementation
			verifyJSONWithCedarCLI(t, string(jsonData))
		})
	}
}

// TestJSONEdgeCases tests edge cases in JSON marshalling/unmarshalling.
func TestJSONEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cedar string
	}{
		{
			name:  "deeply nested Record",
			cedar: `type Deep = { l1: { l2: { l3: { l4: { value: String } } } } };`,
		},
		{
			name:  "Record with many attributes",
			cedar: `entity User { a: String, b: Long, c: Bool, d: String, e: Long, f: Bool };`,
		},
		{
			name:  "multiple entities with hierarchy",
			cedar: `entity A in [B]; entity B in [C]; entity C in [D]; entity D;`,
		},
		{
			name:  "action with multiple principals and resources",
			cedar: `entity User; entity Admin; entity Doc; entity File; action view appliesTo { principal: [User, Admin], resource: [Doc, File] };`,
		},
		{
			name:  "complex context type",
			cedar: `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc, context: { ip: __cedar::ipaddr, time: __cedar::datetime, count: Long, flag: Bool, tags: Set<String> } };`,
		},
		{
			name:  "Set of Entity",
			cedar: `entity User; entity Group { members: Set<User> };`,
		},
		{
			name:  "all extension types",
			cedar: `type Extensions = { ip: __cedar::ipaddr, time: __cedar::datetime, money: __cedar::decimal, dur: __cedar::duration };`,
		},
		{
			name:  "action group hierarchy",
			cedar: `action "base"; action "read" in ["base"]; action "write" in ["base"]; action "admin" in ["read", "write"];`,
		},
		{
			name:  "quoted action names with spaces",
			cedar: `action "view document"; action "edit document" in ["view document"];`,
		},
		{
			name:  "empty namespace",
			cedar: `namespace Empty { }`,
		},
		{
			name:  "namespace with all declaration types",
			cedar: `namespace Full { type Name = String; entity User { name: Name }; action view appliesTo { principal: User, resource: User }; }`,
		},
		{
			name:  "entity with tags",
			cedar: `entity User tags String;`,
		},
		{
			name:  "entity with enum",
			cedar: `entity Status enum ["active", "pending", "closed"];`,
		},
		{
			name:  "multiple entities on one line",
			cedar: `entity User, Admin, Guest;`,
		},
		{
			name:  "multiple actions on one line",
			cedar: `entity User; entity Doc; action view, edit, delete appliesTo { principal: User, resource: Doc };`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse Cedar
			schema, err := schema2.UnmarshalCedar([]byte(tt.cedar))
			testutil.OK(t, err)

			// Marshal to JSON
			jsonData, err := schema2.MarshalJSON(schema)
			testutil.OK(t, err)

			// Parse JSON back
			schema2Parsed, err := schema2.UnmarshalJSON(jsonData)
			if err != nil {
				t.Fatalf("failed to parse JSON: %v\nJSON: %s", err, string(jsonData))
			}

			// Re-marshal to JSON
			jsonData2, err := schema2.MarshalJSON(schema2Parsed)
			testutil.OK(t, err)

			// Verify JSON stability
			if string(jsonData) != string(jsonData2) {
				t.Errorf("JSON round-trip unstable:\nfirst:\n%s\nsecond:\n%s", jsonData, jsonData2)
			}

			// Verify with reference
			verifyJSONWithCedarCLI(t, string(jsonData))
		})
	}
}

// TestJSONCorpusExactMatch compares our JSON output with the expected JSON files.
func TestJSONCorpusExactMatch(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/corpus/valid/*.cedarschema")
	testutil.OK(t, err)

	for _, cedarFile := range files {
		cedarFile := cedarFile
		baseName := strings.TrimSuffix(filepath.Base(cedarFile), ".cedarschema")
		jsonFile := filepath.Join(filepath.Dir(cedarFile), baseName+".json")

		// Skip if no corresponding JSON file exists
		if _, err := os.Stat(jsonFile); os.IsNotExist(err) {
			continue
		}

		t.Run(baseName, func(t *testing.T) {
			t.Parallel()

			// Read Cedar schema
			cedarData, err := os.ReadFile(cedarFile)
			testutil.OK(t, err)

			// Read expected JSON
			expectedJSON, err := os.ReadFile(jsonFile)
			testutil.OK(t, err)

			// Parse Cedar and marshal to JSON
			schema, err := schema2.UnmarshalCedar(cedarData)
			testutil.OK(t, err)

			actualJSON, err := schema2.MarshalJSON(schema)
			testutil.OK(t, err)

			// Parse both JSONs and re-marshal to normalize formatting
			var expectedMap, actualMap map[string]interface{}
			if err := json.Unmarshal(expectedJSON, &expectedMap); err != nil {
				t.Fatalf("failed to parse expected JSON: %v", err)
			}
			if err := json.Unmarshal(actualJSON, &actualMap); err != nil {
				t.Fatalf("failed to parse actual JSON: %v", err)
			}

			// Re-marshal with consistent formatting
			expectedNorm, _ := json.MarshalIndent(expectedMap, "", "    ")
			actualNorm, _ := json.MarshalIndent(actualMap, "", "    ")

			if string(expectedNorm) != string(actualNorm) {
				t.Errorf("JSON mismatch for %s:\nexpected:\n%s\n\nactual:\n%s",
					baseName, string(expectedNorm), string(actualNorm))
			}
		})
	}
}
