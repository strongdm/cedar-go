package schema2

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/resolver"
)

func TestSchemaCedarMarshalUnmarshal(t *testing.T) {
	t.Parallel()

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
		},
		{
			name:  "empty schema",
			input: "",
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
			t.Parallel()
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

			// Test marshaling and round-trip stability
			out, err := s.MarshalCedar()
			if err != nil {
				t.Fatalf("MarshalCedar() error = %v", err)
			}

			var s2 Schema
			s2.SetFilename("test.cedar")
			if err = s2.UnmarshalCedar(out); err != nil {
				t.Fatalf("UnmarshalCedar() second pass error = %v", err)
			}

			out2, err := s2.MarshalCedar()
			if err != nil {
				t.Fatalf("MarshalCedar() second pass error = %v", err)
			}

			if string(out) != string(out2) {
				t.Errorf("Marshal/Unmarshal cycle produced different results:\nFirst: %s\nSecond: %s", out, out2)
			}
		})
	}
}

func TestSchemaJSONMarshalUnmarshal(t *testing.T) {
	t.Parallel()

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
		},
		{
			name:  "empty JSON",
			input: "{}",
		},
		{
			name:    "invalid JSON",
			input:   "{invalid json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
				t.Fatalf("MarshalJSON() error = %v", err)
			}

			// Verify JSON validity
			var raw any
			if err := json.Unmarshal(out, &raw); err != nil {
				t.Errorf("MarshalJSON() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestSchemaEmptyMarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		format     string
		wantOutput string
	}{
		{
			name:       "empty Cedar schema",
			format:     "cedar",
			wantOutput: "",
		},
		{
			name:       "empty JSON schema",
			format:     "json",
			wantOutput: "{}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			s.SetFilename("test." + tt.format)

			var out []byte
			var err error
			if tt.format == "cedar" {
				out, err = s.MarshalCedar()
			} else {
				out, err = s.MarshalJSON()
			}

			if err != nil {
				t.Fatalf("Marshal%s() error = %v", tt.format, err)
			}
			if string(out) != tt.wantOutput {
				t.Errorf("Marshal%s() = %q, want %q", tt.format, string(out), tt.wantOutput)
			}
		})
	}
}

func TestSchemaCrossFormatMarshaling(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		inputFmt   string
		outputFmt  string
		shouldFail bool
	}{
		{
			name:      "JSON to Cedar",
			input:     `{}`,
			inputFmt:  "json",
			outputFmt: "cedar",
		},
		{
			name:      "Cedar to JSON",
			input:     `namespace test {}`,
			inputFmt:  "cedar",
			outputFmt: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			s.SetFilename("test." + tt.inputFmt)

			// Unmarshal input
			var err error
			if tt.inputFmt == "cedar" {
				err = s.UnmarshalCedar([]byte(tt.input))
			} else {
				err = s.UnmarshalJSON([]byte(tt.input))
			}
			if err != nil {
				t.Fatalf("Unmarshal%s() error = %v", tt.inputFmt, err)
			}

			// Marshal to output format
			if tt.outputFmt == "cedar" {
				_, err = s.MarshalCedar()
			} else {
				_, err = s.MarshalJSON()
			}

			if tt.shouldFail && err == nil {
				t.Error("expected error but got none")
			} else if !tt.shouldFail && err != nil {
				t.Errorf("Marshal%s() error = %v", tt.outputFmt, err)
			}
		})
	}
}

func TestSchemaResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantErr     bool
		errContains string
		checkFn     func(*testing.T, *resolver.ResolvedSchema)
	}{
		{
			name:  "empty schema",
			input: "",
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 0 || len(rs.Enums) != 0 || len(rs.Actions) != 0 {
					t.Errorf("Resolve() counts = entities:%d enums:%d actions:%d, want all 0",
						len(rs.Entities), len(rs.Enums), len(rs.Actions))
				}
			},
		},
		{
			name: "valid schema with entities",
			input: `
namespace MyApp {
	entity User;
	entity Group;
}`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 2 {
					t.Errorf("Resolve() entities count = %d, want 2", len(rs.Entities))
				}
				for _, name := range []types.EntityType{"MyApp::User", "MyApp::Group"} {
					if _, found := rs.Entities[name]; !found {
						t.Errorf("Resolve() should contain %s entity", name)
					}
				}
			},
		},
		{
			name: "schema with common types",
			input: `
namespace MyApp {
	type Address = {
		street: String,
		city: String
	};

	entity User {
		address: Address
	};
}`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 1 {
					t.Errorf("Resolve() entities count = %d, want 1", len(rs.Entities))
				}
				if _, found := rs.Entities["MyApp::User"]; !found {
					t.Error("Resolve() should contain MyApp::User entity")
				}
			},
		},
		{
			name: "schema with enums",
			input: `
namespace MyApp {
	entity Status enum ["active", "inactive", "pending"];
}`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Enums) != 1 {
					t.Errorf("Resolve() enums count = %d, want 1", len(rs.Enums))
				}
				if _, found := rs.Enums["MyApp::Status"]; !found {
					t.Error("Resolve() should contain MyApp::Status enum")
				}
			},
		},
		{
			name: "schema with actions",
			input: `
namespace MyApp {
	action View appliesTo {
		principal: User,
		resource: Document
	};

	entity User;
	entity Document;
}`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Actions) != 1 {
					t.Errorf("Resolve() actions count = %d, want 1", len(rs.Actions))
				}
				found := false
				for uid := range rs.Actions {
					if uid.Type == "MyApp::Action" && uid.ID == "View" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Resolve() should contain MyApp::Action::View action")
				}
			},
		},
		{
			name: "undefined type error",
			input: `
namespace MyApp {
	entity User {
		address: NonExistentType
	};
}`,
			wantErr:     true,
			errContains: `type "MyApp::NonExistentType" not found`,
		},
		{
			name: "top-level entities",
			input: `
entity User;
entity Group;`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 2 {
					t.Errorf("Resolve() entities count = %d, want 2", len(rs.Entities))
				}
				for _, name := range []types.EntityType{"User", "Group"} {
					if _, found := rs.Entities[name]; !found {
						t.Errorf("Resolve() should contain %s entity", name)
					}
				}
			},
		},
		{
			name: "mixed top-level and namespaced declarations",
			input: `
entity TopLevel;

namespace MyApp {
	entity Namespaced;
}`,
			checkFn: func(t *testing.T, rs *resolver.ResolvedSchema) {
				if len(rs.Entities) != 2 {
					t.Errorf("Resolve() entities count = %d, want 2", len(rs.Entities))
				}
				for _, name := range []types.EntityType{"TopLevel", "MyApp::Namespaced"} {
					if _, found := rs.Entities[name]; !found {
						t.Errorf("Resolve() should contain %s entity", name)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			if err := s.UnmarshalCedar([]byte(tt.input)); err != nil {
				t.Fatalf("UnmarshalCedar() error = %v", err)
			}

			resolved, err := s.Resolve()
			if tt.wantErr {
				if err == nil {
					t.Fatal("Resolve() expected error but got none")
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("Resolve() error = %v, want %v", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}

			if tt.checkFn != nil {
				tt.checkFn(t, resolved)
			}
		})
	}
}

func TestSchemaAST(t *testing.T) {
	t.Parallel()

	cedarSchema := `namespace MyApp {
	entity User;
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ast := s.AST()
	if ast == nil {
		t.Fatal("AST() returned nil")
	}

	if len(ast.Nodes) == 0 {
		t.Error("AST() returned schema with no nodes")
	}
}
