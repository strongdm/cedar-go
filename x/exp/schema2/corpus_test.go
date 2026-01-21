package schema2_test

import (
	"encoding/json"
	"os"
	"os/exec"
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
				var schema schema2.Schema
				err = schema.UnmarshalCedar(data)
				if err != nil {
					t.Fatalf("failed to parse valid schema %s: %v", name, err)
				}

				// Round-trip Cedar -> Cedar
				marshaled, err := schema.MarshalCedar()
				testutil.OK(t, err)
				var schema2Parsed schema2.Schema
				err = schema2Parsed.UnmarshalCedar(marshaled)
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

				var schema schema2.Schema
				err = schema.UnmarshalCedar(data)
				if err == nil {
					t.Fatalf("expected parse error for invalid schema %s", name)
				}
			})
		}
	})
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
			var schema schema2.Schema
			_ = schema.UnmarshalCedar([]byte(input))
		})
	}
}

// TestJSONCorpus tests JSON marshalling/unmarshalling with corpus files.
// For each valid Cedar schema, we test Cedar -> JSON -> Cedar round-trip.
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

			var schema schema2.Schema
			err = schema.UnmarshalCedar(cedarData)
			if err != nil {
				t.Fatalf("failed to parse Cedar schema: %v", err)
			}

			// Cedar -> JSON
			jsonData, err := schema.MarshalJSON()
			if err != nil {
				t.Fatalf("failed to marshal to JSON: %v", err)
			}

			// JSON -> Cedar (parse JSON back)
			var schema2Parsed schema2.Schema
			err = schema2Parsed.UnmarshalJSON(jsonData)
			if err != nil {
				t.Fatalf("failed to parse JSON schema: %v\nJSON:\n%s", err, string(jsonData))
			}

			// Verify JSON validates with reference implementation
			verifyJSONWithCedarCLI(t, string(jsonData))

			// JSON -> Cedar -> JSON round-trip stability
			jsonData2, err := schema2Parsed.MarshalJSON()
			if err != nil {
				t.Fatalf("failed to re-marshal to JSON: %v", err)
			}

			// Verify JSON stability
			if normalizeJSON(t, string(jsonData)) != normalizeJSON(t, string(jsonData2)) {
				t.Fatalf("JSON round-trip unstable:\nfirst:\n%s\nsecond:\n%s", jsonData, jsonData2)
			}
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
			var schema schema2.Schema
			err = schema.UnmarshalJSON(jsonData)
			if err != nil {
				t.Fatalf("failed to parse JSON schema: %v", err)
			}

			// Schema -> Cedar (to verify we can marshal to Cedar format)
			cedarData, err := schema.MarshalCedar()
			if err != nil {
				t.Fatalf("failed to marshal schema to Cedar format: %v", err)
			}
			if len(cedarData) == 0 {
				t.Fatalf("marshaled Cedar data is empty")
			}

			// Verify Cedar output with reference implementation
			verifyWithCedarCLI(t, string(cedarData))

			// Schema -> JSON
			jsonData2, err := schema.MarshalJSON()
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
			name:        "required attribute has no required field",
			cedar:       `entity User { name: String };`,
			mustNotHave: []string{`"required"`},
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
			expectedFields: []string{`"commonTypes"`, `"Name"`, `"type": "EntityOrCommon"`, `"name": "String"`},
		},
		{
			name:           "empty Record has attributes field",
			cedar:          `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc, context: {} };`,
			expectedFields: []string{`"context"`, `"type": "Record"`, `"attributes": {}`},
		},
		{
			name:        "action without appliesTo has no appliesTo field",
			cedar:       `action read;`,
			mustNotHave: []string{`"appliesTo"`},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var schema schema2.Schema
			err := schema.UnmarshalCedar([]byte(tt.cedar))
			testutil.OK(t, err)

			jsonData, err := schema.MarshalJSON()
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
			var schema schema2.Schema
			err = schema.UnmarshalCedar(cedarData)
			testutil.OK(t, err)

			actualJSON, err := schema.MarshalJSON()
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

// TestCorpusFilesRustSemanticParity tests all corpus files for semantic equivalence with Rust.
func TestCorpusFilesRustSemanticParity(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("testdata/corpus/valid/*.cedarschema")
	testutil.OK(t, err)

	for _, file := range files {
		file := file
		name := filepath.Base(file)
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cedarData, err := os.ReadFile(file)
			testutil.OK(t, err)

			// Parse with our implementation and convert to JSON
			var schema schema2.Schema
			err = schema.UnmarshalCedar(cedarData)
			testutil.OK(t, err)

			goJSON, err := schema.MarshalJSON()
			testutil.OK(t, err)

			// Verify Rust accepts our JSON
			verifyJSONWithCedarCLI(t, string(goJSON))

			// Verify Rust can convert our JSON to Cedar
			rustCedar := translateJSONToCedarWithRust(t, string(goJSON))
			verifyWithCedarCLI(t, rustCedar)

			// Verify we can parse Rust's JSON output (which uses EntityOrCommon format)
			rustJSON := translateCedarToJSONWithRust(t, string(cedarData))
			var schema2Parsed schema2.Schema
			err = schema2Parsed.UnmarshalJSON([]byte(rustJSON))
			if err != nil {
				t.Errorf("Failed to parse Rust's JSON for %s: %v", name, err)
				return
			}

			// And marshal it back
			goJSON2, err := schema2Parsed.MarshalJSON()
			testutil.OK(t, err)

			// And Rust accepts that
			verifyJSONWithCedarCLI(t, string(goJSON2))
		})
	}
}

// Helper functions for JSON/Rust parity tests

// translateCedarToJSONWithRust uses the Rust CLI to convert Cedar schema to JSON.
func translateCedarToJSONWithRust(t *testing.T, cedarContent string) string {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available")
	}

	tmpDir := t.TempDir()
	cedarFile := filepath.Join(tmpDir, "schema.cedarschema")
	if err := os.WriteFile(cedarFile, []byte(cedarContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	cmd := exec.Command(cli, "translate-schema", "--direction", "cedar-to-json", "--schema", cedarFile)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("cedar translate-schema failed: %v\nstderr: %s", err, exitErr.Stderr)
		}
		t.Fatalf("cedar translate-schema failed: %v", err)
	}
	return string(output)
}

// translateJSONToCedarWithRust uses the Rust CLI to convert JSON schema to Cedar.
func translateJSONToCedarWithRust(t *testing.T, jsonContent string) string {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available")
	}

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "schema.cedarschema.json")
	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	cmd := exec.Command(cli, "translate-schema", "--direction", "json-to-cedar", "--schema", jsonFile)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("cedar translate-schema failed: %v\nstderr: %s", err, exitErr.Stderr)
		}
		t.Fatalf("cedar translate-schema failed: %v", err)
	}
	return string(output)
}

// normalizeJSON parses and re-marshals JSON for consistent comparison.
func normalizeJSON(t *testing.T, jsonStr string) string {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		t.Fatalf("failed to parse JSON: %v\nJSON: %s", err, jsonStr)
	}
	normalized, err := json.MarshalIndent(m, "", "    ")
	if err != nil {
		t.Fatalf("failed to marshal JSON: %v", err)
	}
	return string(normalized)
}

// TestComprehensiveCorpus tests comprehensive corpus files with all conversion paths.
// For each .cedarschema file in testdata/corpus/comprehensive/, we expect:
//   - filename.cedarschema: input schema
//   - filename.out.cedarschema: expected Cedar output (normalized)
//   - filename.json: expected JSON output
//   - filename.resolved.cedarschema: expected resolved Cedar output
//   - filename.policy: policy set for validation (optional)
//
// Tests verify:
//  1. Cedar → MarshalCedar matches .out.cedarschema
//  2. Cedar → MarshalJSON matches .json
//  3. JSON → MarshalCedar matches .out.cedarschema
//  4. Cedar → Resolve → MarshalCedar matches .resolved.cedarschema
//  5. JSON → Resolve → MarshalCedar matches .resolved.cedarschema
//  6. Policies validate against original schema (if .policy exists)
//  7. Policies validate against .out schema (if .policy exists)
func TestComprehensiveCorpus(t *testing.T) {
	t.Parallel()

	// Find all .cedarschema files (excluding .out.cedarschema and .resolved.cedarschema)
	allFiles, err := filepath.Glob("testdata/corpus/comprehensive/*.cedarschema")
	testutil.OK(t, err)

	var inputFiles []string
	for _, f := range allFiles {
		if !strings.HasSuffix(f, ".out.cedarschema") && !strings.HasSuffix(f, ".resolved.cedarschema") {
			inputFiles = append(inputFiles, f)
		}
	}

	for _, inputFile := range inputFiles {
		inputFile := inputFile
		baseName := strings.TrimSuffix(filepath.Base(inputFile), ".cedarschema")

		t.Run(baseName, func(t *testing.T) {
			t.Parallel()

			dir := filepath.Dir(inputFile)
			policyFile := filepath.Join(dir, baseName+".policy")

			// Read input Cedar schema
			cedarData, err := os.ReadFile(inputFile)
			testutil.OK(t, err)

			// Parse Cedar schema
			var schema schema2.Schema
			err = schema.UnmarshalCedar(cedarData)
			if err != nil {
				t.Fatalf("failed to parse Cedar schema: %v", err)
			}

			// Test 1: Cedar → MarshalCedar is valid and stable
			actualOut, err := schema.MarshalCedar()
			testutil.OK(t, err)

			// Verify with reference implementation
			verifyWithCedarCLI(t, string(actualOut))

			// Verify round-trip stability
			var schema2Parsed schema2.Schema
			err = schema2Parsed.UnmarshalCedar(actualOut)
			testutil.OK(t, err)

			actualOut2, err := schema2Parsed.MarshalCedar()
			testutil.OK(t, err)

			if string(actualOut) != string(actualOut2) {
				t.Errorf("Cedar round-trip unstable:\nfirst:\n%s\n\nsecond:\n%s",
					string(actualOut), string(actualOut2))
			}

			// Test 2: Cedar → MarshalJSON is valid and stable
			actualJSON, err := schema.MarshalJSON()
			testutil.OK(t, err)

			// Verify with reference implementation
			verifyJSONWithCedarCLI(t, string(actualJSON))

			// Verify round-trip stability
			var schemaFromJSON schema2.Schema
			err = schemaFromJSON.UnmarshalJSON(actualJSON)
			testutil.OK(t, err)

			actualJSON2, err := schemaFromJSON.MarshalJSON()
			testutil.OK(t, err)

			if normalizeJSON(t, string(actualJSON)) != normalizeJSON(t, string(actualJSON2)) {
				t.Errorf("JSON round-trip unstable:\nfirst:\n%s\n\nsecond:\n%s",
					string(actualJSON), string(actualJSON2))
			}

			// Test 3: JSON → Cedar is valid
			cedarFromJSON, err := schemaFromJSON.MarshalCedar()
			testutil.OK(t, err)

			// Verify with reference implementation
			verifyWithCedarCLI(t, string(cedarFromJSON))

			// Test 4: Cedar → Resolve → MarshalCedar produces valid output
			resolved, err := schema.Resolve()
			testutil.OK(t, err)

			actualResolved, err := resolved.TestMarshalCedar()
			testutil.OK(t, err)

			// Verify resolved schema is valid Cedar
			verifyWithCedarCLI(t, string(actualResolved))

			// Verify that resolving again gives the same result
			resolved2, err := schema.Resolve()
			testutil.OK(t, err)

			actualResolved2, err := resolved2.TestMarshalCedar()
			testutil.OK(t, err)

			if string(actualResolved) != string(actualResolved2) {
				t.Errorf("Resolved output not deterministic:\nfirst:\n%s\n\nsecond:\n%s",
					string(actualResolved), string(actualResolved2))
			}

			// Test 5: JSON → Resolve → MarshalCedar produces valid output
			resolvedFromJSON, err := schemaFromJSON.Resolve()
			testutil.OK(t, err)

			actualResolvedFromJSON, err := resolvedFromJSON.TestMarshalCedar()
			testutil.OK(t, err)

			// Verify resolved schema from JSON is valid Cedar
			verifyWithCedarCLI(t, string(actualResolvedFromJSON))

			// Test 6: Validate policies against resolved schema (if .policy exists)
			if _, err := os.Stat(policyFile); err == nil {
				cli := cedarCLI()
				if cli != "" {
					// Create temp file for resolved schema
					tmpDir := t.TempDir()
					resolvedFile := filepath.Join(tmpDir, "resolved.cedarschema")
					err := os.WriteFile(resolvedFile, actualResolved, 0o644)
					testutil.OK(t, err)

					// Validate policies against resolved schema
					cmd := exec.Command(cli, "validate",
						"--schema", resolvedFile,
						"--schema-format", "cedar",
						"--policies", policyFile)
					if output, err := cmd.CombinedOutput(); err != nil {
						t.Errorf("Policy validation failed against resolved schema:\n%s", string(output))
					}
				}
			}

			// Test 7: Validate policies against original schema (if .policy exists)
			if _, err := os.Stat(policyFile); err == nil {
				cli := cedarCLI()
				if cli != "" {
					cmd := exec.Command(cli, "validate",
						"--schema", inputFile,
						"--schema-format", "cedar",
						"--policies", policyFile)
					if output, err := cmd.CombinedOutput(); err != nil {
						t.Errorf("Policy validation failed against original schema:\n%s", string(output))
					}
				}
			}
		})
	}
}
