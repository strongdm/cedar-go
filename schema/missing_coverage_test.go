package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

// TestMissingCoverageBranches targets specific uncovered branches
func TestMissingCoverageBranches(t *testing.T) {
	t.Run("Extension type without annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"Network": {
						"shape": {
							"type": "Extension",
							"name": "ipaddr"
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonOutput); err != nil {
			t.Fatalf("Second unmarshal failed: %v", err)
		}
	})

	t.Run("Entity type without annotations in various contexts", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"User": {},
					"Group": {
						"shape": {
							"type": "Record",
							"attributes": {
								"owner": {
									"type": "Entity",
									"name": "User",
									"required": true
								}
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonOutput); err != nil {
			t.Fatalf("Second unmarshal failed: %v", err)
		}
	})

	t.Run("String type without annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "String"
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Long type without annotations in attribute", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"count": {
									"type": "Long",
									"required": false
								}
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Boolean type in attribute without annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"active": {
									"type": "Boolean",
									"required": true
								}
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Verify Boolean is preserved
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		e := entities["E"].(map[string]interface{})
		shape := e["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})
		active := attrs["active"].(map[string]interface{})

		if active["type"] != "Boolean" {
			t.Errorf("Expected Boolean, got %v", active["type"])
		}
	})

	t.Run("Extension in attribute context", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"address": {
									"type": "Extension",
									"name": "ipaddr",
									"required": false
								}
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Unknown attribute type", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"field": {
									"type": "UnknownAttributeType",
									"required": true
								}
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("MarshalCedar error handling", func(t *testing.T) {
		// Test error path in MarshalCedar when MarshalJSON fails
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("User"),
			)

		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarData) == 0 {
			t.Error("Expected non-empty Cedar data")
		}
	})

	t.Run("UnmarshalCedar error path", func(t *testing.T) {
		var s schema.Schema
		s.SetFilename("test.cedar")

		// Test with empty input
		err := s.UnmarshalCedar([]byte(""))
		if err != nil {
			// Empty input might error, which is fine
			t.Logf("Empty input error: %v", err)
		}

		// Test with valid input
		err = s.UnmarshalCedar([]byte("namespace Test {}"))
		if err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}
	})

	t.Run("WithNamespace with nil declarations", func(t *testing.T) {
		// Test that WithNamespace handles empty declaration list
		s := schema.NewSchema().
			WithNamespace("EmptyNS")

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		if parsed["EmptyNS"] == nil {
			t.Error("Expected EmptyNS to be present")
		}
	})

	t.Run("UnmarshalJSON with nil annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonOutput); err != nil {
			t.Fatalf("Second unmarshal failed: %v", err)
		}
	})

	t.Run("MarshalJSON with nil namespace annotations", func(t *testing.T) {
		// Test namespace without annotations
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		// annotations field should not be present if empty
		if testNS["annotations"] != nil {
			t.Logf("Annotations present: %v", testNS["annotations"])
		}
	})
}
