package schema2

import (
	"encoding/json"
	"reflect"
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
