package schema2

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestSchemaCedarMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid schema",
			input: `namespace foo {
				action Bar appliesTo {
					principal: String,
					resource: String
				};
			}`,
			wantErr: false,
		},
		{
			name:    "empty schema",
			input:   "",
			wantErr: false,
		},
		{
			name: "invalid schema",
			input: `namespace foo {
				action Bar = {
					invalid syntax here
				};
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Schema
			s.SetFilename("test.cedar")
			err := s.UnmarshalCedar([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalCedar() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Test marshaling
			out, err := s.MarshalCedar()
			if err != nil {
				t.Errorf("MarshalCedar() error = %v", err)
				return
			}

			// For valid schemas, unmarshaling and marshaling should preserve content
			if !tt.wantErr {
				var s2 Schema
				s2.SetFilename("test.cedar")
				err = s2.UnmarshalCedar(out)
				if err != nil {
					t.Errorf("UnmarshalCedar() second pass error = %v", err)
					return
				}

				out2, err := s2.MarshalCedar()
				if err != nil {
					t.Errorf("MarshalCedar() second pass error = %v", err)
					return
				}

				if !reflect.DeepEqual(out, out2) {
					t.Errorf("Marshal/Unmarshal cycle produced different results:\nFirst: %s\nSecond: %s", out, out2)
				}
			}
		})
	}
}

func TestSchemaCedarMarshalEmpty(t *testing.T) {
	var s Schema
	s.SetFilename("test.cedar")
	out, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
		return
	}
	if len(out) != 0 {
		t.Errorf("MarshalCedar() produced non-empty output for empty schema")
	}
}

func TestSchemaJSONMarshalEmpty(t *testing.T) {
	var s Schema
	s.SetFilename("test.json")
	out, err := s.MarshalJSON()
	if err != nil {
		t.Errorf("MarshalJSON() error = %v", err)
		return
	}
	// Empty schema marshals to "{}"
	if string(out) != "{}" {
		t.Errorf("MarshalJSON() = %s, want {}", string(out))
	}
}

func TestSchemaJSONMarshalUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name: "valid JSON schema",
			input: `{
				"entityTypes": {
					"User": {
						"shape": {
							"type": "Record",
							"attributes": {
								"name": {"type": "String"}
							}
						}
					}
				}
			}`,
			wantErr: false,
		},
		{
			name:    "empty JSON",
			input:   "{}",
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			input:   "{invalid json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Schema
			s.SetFilename("test.json")
			err := s.UnmarshalJSON([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Test marshaling
			out, err := s.MarshalJSON()
			if err != nil {
				t.Errorf("MarshalJSON() error = %v", err)
				return
			}

			// Verify JSON validity
			var raw interface{}
			if err := json.Unmarshal(out, &raw); err != nil {
				t.Errorf("MarshalJSON() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestSchemaCrossFormatMarshaling(t *testing.T) {
	t.Run("JSON to Cedar Marshalling", func(t *testing.T) {
		var s Schema
		err := s.UnmarshalJSON([]byte(`{}`))
		if err != nil {
			t.Fatalf("UnmarshalJSON() error = %v", err)
		}

		_, err = s.MarshalCedar()
		if err != nil {
			t.Error("MarshalCedar() should not return error after UnmarshalJSON")
		}
	})

	t.Run("Cedar to JSON marshaling allowed", func(t *testing.T) {
		var s Schema
		s.SetFilename("test.cedar")
		err := s.UnmarshalCedar([]byte(`namespace test {}`))
		if err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		_, err = s.MarshalJSON()
		if err != nil {
			t.Errorf("MarshalJSON() error = %v", err)
		}
	})
}

func TestResolvedSchemaMarshalCedar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           string
		expectedEntries []string // Entries that should appear in marshaled output
	}{
		{
			name: "simple entity schema",
			input: `namespace MyApp {
	entity User;
	entity Group;
}`,
			expectedEntries: []string{"namespace MyApp", "entity User", "entity Group"},
		},
		{
			name: "entity with common types resolved inline",
			input: `namespace MyApp {
	type Address = {
		street: String,
		city: String,
	};
	entity User {
		address: Address,
	};
}`,
			expectedEntries: []string{"namespace MyApp", "entity User", "street", "city"},
		},
		{
			name: "entity with enum",
			input: `namespace MyApp {
	entity Status enum ["active", "inactive"];
}`,
			expectedEntries: []string{"namespace MyApp", "entity Status enum", "active", "inactive"},
		},
		{
			name: "action with appliesTo",
			input: `namespace MyApp {
	entity User;
	entity Document;
	action View appliesTo {
		principal: User,
		resource: Document,
	};
}`,
			expectedEntries: []string{"namespace MyApp", "entity User", "entity Document", "action View", "principal", "resource"},
		},
		{
			name: "multiple namespaces",
			input: `namespace App1 {
	entity User;
}
namespace App2 {
	entity Admin;
}`,
			expectedEntries: []string{"namespace App1", "entity User", "namespace App2", "entity Admin"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse the original schema
			var s Schema
			err := s.UnmarshalCedar([]byte(tt.input))
			if err != nil {
				t.Fatalf("UnmarshalCedar() error = %v", err)
			}

			// Resolve it
			resolved, err := s.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			// Convert to Schema and marshal to Cedar
			// Use marshalCedar directly
			cedarBytes, err := resolved.marshalCedar()
			if err != nil {
				t.Fatalf("marshalCedar() error = %v", err)
			}

			cedarStr := string(cedarBytes)

			// Verify expected entries are present in the marshaled output
			for _, entry := range tt.expectedEntries {
				if !strings.Contains(cedarStr, entry) {
					t.Errorf("Expected entry %q not found in marshaled Cedar:\n%s", entry, cedarStr)
				}
			}

			// Verify the output is non-empty
			if len(cedarBytes) == 0 {
				t.Error("marshalCedar() produced empty output")
			}

			// Marshal again to verify stability
			// Use marshalCedar directly
			cedarBytes2, err := resolved.marshalCedar()
			if err != nil {
				t.Fatalf("Second marshalCedar() error = %v", err)
			}

			if !reflect.DeepEqual(cedarBytes, cedarBytes2) {
				t.Errorf("marshalCedar() is not stable:\nFirst:\n%s\nSecond:\n%s", cedarBytes, cedarBytes2)
			}
		})
	}
}

func TestResolvedSchemaMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "simple entity schema",
			input: `namespace MyApp {
	entity User;
}`,
		},
		{
			name: "entity with attributes",
			input: `namespace MyApp {
	entity User {
		name: String,
		age: Long,
	};
}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse and resolve
			var s Schema
			err := s.UnmarshalCedar([]byte(tt.input))
			if err != nil {
				t.Fatalf("UnmarshalCedar() error = %v", err)
			}

			resolved, err := s.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			// Convert to Schema and marshal to JSON
			// Use marshalJSON directly
			jsonBytes, err := resolved.marshalJSON()
			if err != nil {
				t.Fatalf("marshalJSON() error = %v", err)
			}

			// Verify it's valid JSON
			var raw interface{}
			if err := json.Unmarshal(jsonBytes, &raw); err != nil {
				t.Errorf("MarshalJSON() produced invalid JSON: %v\nJSON:\n%s", err, string(jsonBytes))
			}

			// Parse the JSON back
			var s2 Schema
			err = s2.UnmarshalJSON(jsonBytes)
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v\nJSON:\n%s", err, string(jsonBytes))
			}

			// Resolve again
			resolved2, err := s2.Resolve()
			if err != nil {
				t.Fatalf("Resolve() on JSON schema error = %v", err)
			}

			// Compare
			if len(resolved.Entities) != len(resolved2.Entities) {
				t.Errorf("Entity count mismatch: %d != %d", len(resolved.Entities), len(resolved2.Entities))
			}
		})
	}
}

func TestResolvedSchemaRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
	}{
		{
			name: "comprehensive schema with all features",
			input: `namespace MyApp {
	entity User;
	entity Group;
	entity Document;

	entity Status enum ["active", "inactive", "pending"];

	action View appliesTo {
		principal: User,
		resource: Document,
	};
}`,
		},
		{
			name: "top-level and namespaced declarations",
			input: `entity TopLevelUser;

namespace App {
	entity User;
	action Read;
}`,
		},
		{
			name: "cross-namespace references",
			input: `namespace Core {
	entity BaseUser;
}

namespace App {
	entity User in [Core::BaseUser];
}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Step 1: Parse original schema
			var original Schema
			if err := original.UnmarshalCedar([]byte(tt.input)); err != nil {
				t.Fatalf("UnmarshalCedar() error = %v", err)
			}

			// Step 2: Resolve
			resolved, err := original.Resolve()
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			// Step 3: Convert to Schema and marshal to Cedar
			// Use marshalCedar directly
			cedarBytes, err := resolved.marshalCedar()
			if err != nil {
				t.Fatalf("marshalCedar() error = %v", err)
			}

			// Verify output is non-empty
			if len(cedarBytes) == 0 {
				t.Fatal("marshalCedar() produced empty output")
			}

			// Step 4: Marshal resolved schema to JSON
			jsonBytes, err := resolved.marshalJSON()
			if err != nil {
				t.Fatalf("marshalJSON() error = %v", err)
			}

			// Verify JSON is valid
			var raw interface{}
			if err := json.Unmarshal(jsonBytes, &raw); err != nil {
				t.Fatalf("marshalJSON() produced invalid JSON: %v", err)
			}

			// Step 5: Verify stability - marshaling again produces same output
			// Use marshalCedar directly
			cedarBytes2, err := resolved.marshalCedar()
			if err != nil {
				t.Fatalf("Second marshalCedar() error = %v", err)
			}
			if !reflect.DeepEqual(cedarBytes, cedarBytes2) {
				t.Errorf("marshalCedar() not stable")
			}

			jsonBytes2, err := resolved.marshalJSON()
			if err != nil {
				t.Fatalf("Second marshalJSON() error = %v", err)
			}
			if !reflect.DeepEqual(jsonBytes, jsonBytes2) {
				t.Errorf("marshalJSON() not stable")
			}

			// Step 6: Verify counts match original
			if len(resolved.Entities) == 0 && len(resolved.Enums) == 0 && len(resolved.Actions) == 0 {
				t.Error("Resolved schema is empty but input was not")
			}
		})
	}
}

func TestResolvedSchemaSortingCoverage(t *testing.T) {
	t.Parallel()

	t.Run("multiple enums sorted", func(t *testing.T) {
		t.Parallel()
		input := `namespace MyApp {
	entity Status enum ["active", "inactive"];
	entity Priority enum ["high", "low"];
	entity Category enum ["a", "b"];
}`
		var s Schema
		if err := s.UnmarshalCedar([]byte(input)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		// Test that we can marshal (this exercises the enum sorting)
		// Use marshalCedar directly
		cedarBytes, err := resolved.marshalCedar()
		if err != nil {
			t.Fatalf("marshalCedar() error = %v", err)
		}

		// Verify enums are present
		cedarStr := string(cedarBytes)
		if !strings.Contains(cedarStr, "Status") {
			t.Error("Expected Status enum in output")
		}
		if !strings.Contains(cedarStr, "Priority") {
			t.Error("Expected Priority enum in output")
		}
		if !strings.Contains(cedarStr, "Category") {
			t.Error("Expected Category enum in output")
		}
	})

	t.Run("multiple actions sorted by type and ID", func(t *testing.T) {
		t.Parallel()
		input := `namespace App {
	entity User;
	entity Doc;
	action Read;
	action Write;
	action Delete;
}
namespace Other {
	entity Resource;
	action View;
}`
		var s Schema
		if err := s.UnmarshalCedar([]byte(input)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		// Test that we can marshal (this exercises the action sorting with different types)
		// Use marshalCedar directly
		cedarBytes, err := resolved.marshalCedar()
		if err != nil {
			t.Fatalf("marshalCedar() error = %v", err)
		}

		// Verify actions are present
		cedarStr := string(cedarBytes)
		if !strings.Contains(cedarStr, "Read") {
			t.Error("Expected Read action in output")
		}
		if !strings.Contains(cedarStr, "Write") {
			t.Error("Expected Write action in output")
		}
		if !strings.Contains(cedarStr, "Delete") {
			t.Error("Expected Delete action in output")
		}
		if !strings.Contains(cedarStr, "View") {
			t.Error("Expected View action in output")
		}
	})
}

func TestSchemaResolve(t *testing.T) {
	t.Parallel()

	t.Run("resolve empty schema", func(t *testing.T) {
		t.Parallel()
		var s Schema
		resolved, err := s.Resolve()
		if err != nil {
			t.Errorf("Resolve() error = %v", err)
		}
		if len(resolved.Entities) != 0 {
			t.Errorf("Resolve() entities count = %d, want 0", len(resolved.Entities))
		}
		if len(resolved.Enums) != 0 {
			t.Errorf("Resolve() enums count = %d, want 0", len(resolved.Enums))
		}
		if len(resolved.Actions) != 0 {
			t.Errorf("Resolve() actions count = %d, want 0", len(resolved.Actions))
		}
	})

	t.Run("resolve valid schema with entities", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
namespace MyApp {
	entity User;
	entity Group;
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Entities) != 2 {
			t.Errorf("Resolve() entities count = %d, want 2", len(resolved.Entities))
		}

		if _, found := resolved.Entities["MyApp::User"]; !found {
			t.Error("Resolve() should contain MyApp::User entity")
		}
		if _, found := resolved.Entities["MyApp::Group"]; !found {
			t.Error("Resolve() should contain MyApp::Group entity")
		}
	})

	t.Run("resolve schema with common types", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
namespace MyApp {
	type Address = {
		street: String,
		city: String
	};

	entity User {
		address: Address
	};
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Entities) != 1 {
			t.Errorf("Resolve() entities count = %d, want 1", len(resolved.Entities))
		}

		if _, found := resolved.Entities["MyApp::User"]; !found {
			t.Error("Resolve() should contain MyApp::User entity")
		}
	})

	t.Run("resolve schema with enums", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
namespace MyApp {
	entity Status enum ["active", "inactive", "pending"];
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Enums) != 1 {
			t.Errorf("Resolve() enums count = %d, want 1", len(resolved.Enums))
		}

		if _, found := resolved.Enums["MyApp::Status"]; !found {
			t.Error("Resolve() should contain MyApp::Status enum")
		}
	})

	t.Run("resolve schema with actions", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
namespace MyApp {
	action View appliesTo {
		principal: User,
		resource: Document
	};

	entity User;
	entity Document;
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Actions) != 1 {
			t.Errorf("Resolve() actions count = %d, want 1", len(resolved.Actions))
		}

		// Actions should be indexed by EntityUID
		found := false
		for uid := range resolved.Actions {
			if uid.Type == "MyApp::Action" && uid.ID == "View" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Resolve() should contain MyApp::Action::View action")
		}
	})

	t.Run("resolve schema with undefined type error", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
namespace MyApp {
	entity User {
		address: NonExistentType
	};
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		_, err := s.Resolve()
		if err == nil {
			t.Error("Resolve() should return error for undefined type")
		}

		expectedMsg := `type "MyApp::NonExistentType" not found`
		if err.Error() != expectedMsg {
			t.Errorf("Resolve() error = %v, want %v", err.Error(), expectedMsg)
		}
	})

	t.Run("resolve top-level entities", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
entity User;
entity Group;
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Entities) != 2 {
			t.Errorf("Resolve() entities count = %d, want 2", len(resolved.Entities))
		}

		if _, found := resolved.Entities["User"]; !found {
			t.Error("Resolve() should contain User entity")
		}
		if _, found := resolved.Entities["Group"]; !found {
			t.Error("Resolve() should contain Group entity")
		}
	})

	t.Run("resolve mixed top-level and namespaced declarations", func(t *testing.T) {
		t.Parallel()
		cedarSchema := `
entity TopLevel;

namespace MyApp {
	entity Namespaced;
}
`
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
			t.Fatalf("UnmarshalCedar() error = %v", err)
		}

		resolved, err := s.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		if len(resolved.Entities) != 2 {
			t.Errorf("Resolve() entities count = %d, want 2", len(resolved.Entities))
		}

		if _, found := resolved.Entities["TopLevel"]; !found {
			t.Error("Resolve() should contain TopLevel entity")
		}
		if _, found := resolved.Entities["MyApp::Namespaced"]; !found {
			t.Error("Resolve() should contain MyApp::Namespaced entity")
		}
	})
}
