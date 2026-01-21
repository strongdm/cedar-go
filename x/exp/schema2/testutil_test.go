package schema2_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

// cedarCLI returns the path to the cedar-policy CLI for verification.
// Set CEDAR_CLI env var to override, or it will look in common locations.
func cedarCLI() string {
	if cli := os.Getenv("CEDAR_CLI"); cli != "" {
		return cli
	}
	// Check common locations
	paths := []string{
		"/home/user/cedar-rust/target/release/cedar",
		"/usr/local/bin/cedar",
		"cedar",
	}
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return ""
}

// verifyWithCedarCLI checks that a schema parses correctly using the reference implementation.
func verifyWithCedarCLI(t *testing.T, schemaContent string) {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available, skipping verification")
	}

	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.cedarschema")
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	cmd := exec.Command(cli, "check-parse", "--schema", schemaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cedar CLI rejected schema:\n%s\nOutput: %s", schemaContent, string(output))
	}
}

// verifyJSONWithCedarCLI checks that a JSON schema parses correctly using the reference implementation.
func verifyJSONWithCedarCLI(t *testing.T, jsonContent string) {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available, skipping verification")
	}

	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.cedarschema.json")
	if err := os.WriteFile(schemaFile, []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	cmd := exec.Command(cli, "check-parse", "--schema", schemaFile, "--schema-format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cedar CLI rejected JSON schema:\n%s\nOutput: %s", jsonContent, string(output))
	}
}

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

// roundTripCedar tests that a schema can be marshaled and unmarshaled without loss.
func roundTripCedar(t *testing.T, schema *schema2.Schema) []byte {
	t.Helper()
	marshaled, err := schema.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	var reparsed schema2.Schema
	if err := reparsed.UnmarshalCedar(marshaled); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	marshaled2, err := reparsed.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() second pass error = %v", err)
	}

	if string(marshaled) != string(marshaled2) {
		t.Fatalf("Cedar round-trip unstable:\nfirst:\n%s\n\nsecond:\n%s", marshaled, marshaled2)
	}

	return marshaled
}

// roundTripJSON tests that a schema can be marshaled and unmarshaled to JSON without loss.
func roundTripJSON(t *testing.T, schema *schema2.Schema) []byte {
	t.Helper()
	marshaled, err := schema.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var reparsed schema2.Schema
	if err := reparsed.UnmarshalJSON(marshaled); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	marshaled2, err := reparsed.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() second pass error = %v", err)
	}

	if normalizeJSON(t, string(marshaled)) != normalizeJSON(t, string(marshaled2)) {
		t.Fatalf("JSON round-trip unstable:\nfirst:\n%s\n\nsecond:\n%s", marshaled, marshaled2)
	}

	return marshaled
}
