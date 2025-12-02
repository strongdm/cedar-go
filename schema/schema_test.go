package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

func TestSchemaJSONRoundTrip(t *testing.T) {
	// Create a schema programmatically
	s := schema.NewSchema().
		WithNamespace("PhotoApp",
			schema.NewEntity("User").
				WithAttribute("name", schema.String()).
				WithAttribute("age", schema.Long()).
				MemberOf("UserGroup"),
			schema.NewEntity("UserGroup"),
			schema.NewEntity("Photo").
				WithAttribute("owner", schema.EntityType("User")).
				WithOptionalAttribute("private", schema.Bool()),
			schema.NewAction("viewPhoto").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Photo"),
					schema.Record(
						schema.Attr("authenticated", schema.Bool()),
					),
				),
		)

	// Marshal to JSON
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var rawJSON map[string]interface{}
	if err := json.Unmarshal(jsonData, &rawJSON); err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	// Unmarshal back
	var s2 schema.Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	// Marshal again
	jsonData2, err := s2.MarshalJSON()
	if err != nil {
		t.Fatalf("Second MarshalJSON failed: %v", err)
	}

	// Compare JSON output (should be identical)
	var json1, json2 interface{}
	if err := json.Unmarshal(jsonData, &json1); err != nil {
		t.Fatalf("Failed to unmarshal first JSON: %v", err)
	}
	if err := json.Unmarshal(jsonData2, &json2); err != nil {
		t.Fatalf("Failed to unmarshal second JSON: %v", err)
	}

	if !jsonEqual(t, json1, json2) {
		t.Errorf("JSON round-trip produced different results:\nFirst:  %s\nSecond: %s", jsonData, jsonData2)
	}
}

func TestSchemaCedarRoundTrip(t *testing.T) {
	cedarInput := []byte(`
namespace PhotoApp {
  entity User in [UserGroup] = {
    "name": String,
    "age": Long,
  };
  entity UserGroup;
  entity Photo = {
    "owner": User,
    "private"?: Bool,
  };
  action viewPhoto appliesTo {
    principal: [User],
    resource: [Photo],
    context: {
      "authenticated": Bool,
    }
  };
}
`)

	// Parse Cedar
	var s schema.Schema
	if err := s.UnmarshalCedar(cedarInput); err != nil {
		t.Fatalf("UnmarshalCedar failed: %v", err)
	}

	// Marshal back to Cedar
	cedarOutput, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar failed: %v", err)
	}

	// Parse again
	var s2 schema.Schema
	if err := s2.UnmarshalCedar(cedarOutput); err != nil {
		t.Fatalf("Second UnmarshalCedar failed: %v\nOutput was: %s", err, cedarOutput)
	}

	// Marshal to JSON for comparison (easier than comparing Cedar text)
	json1, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("First MarshalJSON failed: %v", err)
	}

	json2, err := s2.MarshalJSON()
	if err != nil {
		t.Fatalf("Second MarshalJSON failed: %v", err)
	}

	var j1, j2 interface{}
	if err := json.Unmarshal(json1, &j1); err != nil {
		t.Fatalf("Failed to unmarshal first JSON: %v", err)
	}
	if err := json.Unmarshal(json2, &j2); err != nil {
		t.Fatalf("Failed to unmarshal second JSON: %v", err)
	}

	if !jsonEqual(t, j1, j2) {
		t.Errorf("Cedar round-trip produced different results")
	}
}

func TestSchemaCrossFormatConversion(t *testing.T) {
	t.Run("JSON to Cedar", func(t *testing.T) {
		jsonInput := []byte(`{
			"PhotoApp": {
				"entityTypes": {
					"User": {
						"shape": {
							"type": "Record",
							"attributes": {
								"name": {
									"type": "String",
									"required": true
								}
							}
						}
					}
				},
				"actions": {
					"view": {
						"appliesTo": {
							"principalTypes": ["User"],
							"resourceTypes": ["Photo"]
						}
					}
				}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		cedarOutput, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarOutput) == 0 {
			t.Error("MarshalCedar produced empty output")
		}
	})

	t.Run("Cedar to JSON", func(t *testing.T) {
		cedarInput := []byte(`
namespace Test {
  entity User;
  action view appliesTo {
    principal: [User],
    resource: [User]
  };
}
`)

		var s schema.Schema
		if err := s.UnmarshalCedar(cedarInput); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Verify it's valid JSON
		var raw interface{}
		if err := json.Unmarshal(jsonOutput, &raw); err != nil {
			t.Errorf("MarshalJSON produced invalid JSON: %v", err)
		}
	})
}

func TestSchemaBuilder(t *testing.T) {
	s := schema.NewSchema().
		WithNamespace("App",
			schema.NewEntity("User").
				WithAttribute("email", schema.String()).
				WithAttribute("age", schema.Long()).
				WithOptionalAttribute("phone", schema.String()),
			schema.NewEntity("Group"),
			schema.NewEntity("Document").
				WithAttribute("owner", schema.EntityType("User")).
				WithAttribute("readers", schema.Set(schema.EntityType("User"))),
			schema.NewAction("read").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Document"),
					nil,
				),
		)

	// Should be able to marshal
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Should have the expected structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Check namespace exists
	if _, ok := parsed["App"]; !ok {
		t.Error("Namespace 'App' not found in JSON output")
	}
}

func TestComplexTypes(t *testing.T) {
	s := schema.NewSchema().
		WithNamespace("",
			schema.NewEntity("User").
				WithAttribute("profile", schema.Record(
					schema.Attr("firstName", schema.String()),
					schema.Attr("lastName", schema.String()),
					schema.OptionalAttr("middleName", schema.String()),
					schema.Attr("addresses", schema.Set(
						schema.Record(
							schema.Attr("street", schema.String()),
							schema.Attr("city", schema.String()),
						),
					)),
				)),
		)

	// Marshal and verify
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Round-trip test
	var s2 schema.Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}

	jsonData2, err := s2.MarshalJSON()
	if err != nil {
		t.Fatalf("Second MarshalJSON failed: %v", err)
	}

	var j1, j2 interface{}
	if err := json.Unmarshal(jsonData, &j1); err != nil {
		t.Fatalf("Failed to unmarshal first JSON: %v", err)
	}
	if err := json.Unmarshal(jsonData2, &j2); err != nil {
		t.Fatalf("Failed to unmarshal second JSON: %v", err)
	}

	if !jsonEqual(t, j1, j2) {
		t.Error("Complex type round-trip failed")
	}
}

func TestEntityEnum(t *testing.T) {
	s := schema.NewSchema().
		WithNamespace("App",
			schema.NewEntity("PhotoFormat").AsEnum("jpg", "png", "gif"),
		)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify enum structure
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Round-trip
	var s2 schema.Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
}

func TestEmptySchema(t *testing.T) {
	s := schema.NewSchema()

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Should produce empty object
	if string(jsonData) != "{}" {
		t.Errorf("Empty schema should produce '{}', got: %s", jsonData)
	}
}

// jsonEqual performs a deep comparison of JSON values for testing
func jsonEqual(t *testing.T, a, b interface{}) bool {
	t.Helper()

	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)

	var aVal, bVal interface{}
	if err := json.Unmarshal(aJSON, &aVal); err != nil {
		t.Errorf("Failed to unmarshal aJSON: %v", err)
		return false
	}
	if err := json.Unmarshal(bJSON, &bVal); err != nil {
		t.Errorf("Failed to unmarshal bJSON: %v", err)
		return false
	}

	aStr, _ := json.MarshalIndent(aVal, "", "  ")
	bStr, _ := json.MarshalIndent(bVal, "", "  ")

	return string(aStr) == string(bStr)
}
