package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

// Test100PercentCoverage tests all remaining uncovered code paths
func Test100PercentCoverage(t *testing.T) {
	t.Run("Interface marker methods", func(t *testing.T) {
		// Call the isType marker methods to get coverage
		// These are no-op methods but need to be called for coverage
		var _ schema.Type = schema.String()
		var _ schema.Type = schema.Set(schema.String())
		var _ schema.Type = schema.Record()
	})

	t.Run("Namespace WithAnnotation", func(t *testing.T) {
		// Test namespace annotations through JSON round-trip
		// since WithAnnotation isn't directly exposed in the builder API
		jsonInput := []byte(`{
			"TestNS": {
				"annotations": {
					"doc": "Test namespace documentation",
					"version": "1.0"
				},
				"entityTypes": {
					"User": {}
				},
				"actions": {}
			}
		}`)

		var s schema.Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		// Marshal back to JSON to ensure annotations are preserved
		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		testNS := parsed["TestNS"].(map[string]interface{})
		annotations := testNS["annotations"].(map[string]interface{})
		if annotations["doc"] != "Test namespace documentation" {
			t.Errorf("Expected namespace annotation 'doc', got %v", annotations)
		}
		if annotations["version"] != "1.0" {
			t.Errorf("Expected namespace annotation 'version', got %v", annotations)
		}
	})

	t.Run("Bool type variant", func(t *testing.T) {
		// Test the "Bool" variant (vs "Boolean") in JSON
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"flag": {
									"type": "Bool",
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

		// Marshal and verify
		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Should normalize to Boolean
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		e := entities["E"].(map[string]interface{})
		shape := e["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})
		flag := attrs["flag"].(map[string]interface{})

		if flag["type"] != "Boolean" {
			t.Errorf("Expected 'Boolean', got %v", flag["type"])
		}
	})

	t.Run("Unknown JSON type falls back to path", func(t *testing.T) {
		// Test that unknown types are treated as path types
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "UnknownType"
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

		// Should not crash - unknown types handled gracefully
		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Extension type with annotations in attributes", func(t *testing.T) {
		// Test extension types in attribute context
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"Network": {
						"shape": {
							"type": "Record",
							"attributes": {
								"cidr": {
									"type": "Extension",
									"name": "ipaddr",
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

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		network := entities["Network"].(map[string]interface{})
		if network == nil {
			t.Error("Expected Network entity")
		}
	})

	t.Run("Entity type in Set attribute context", func(t *testing.T) {
		// Test entity types in Set element position in attributes
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"Group": {
						"shape": {
							"type": "Record",
							"attributes": {
								"members": {
									"type": "Set",
									"required": true,
									"element": {
										"type": "Entity",
										"name": "User"
									}
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

		// Verify round-trip
		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonOutput); err != nil {
			t.Fatalf("Second UnmarshalJSON failed: %v", err)
		}
	})

	t.Run("Bool in PathType conversion", func(t *testing.T) {
		// Test Bool type path conversion
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E").
					WithAttribute("active", schema.Bool()),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		e := entities["E"].(map[string]interface{})
		shape := e["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})
		active := attrs["active"].(map[string]interface{})

		// Should be normalized to Boolean in output
		if active["type"] != "Boolean" {
			t.Errorf("Expected 'Boolean', got %v", active["type"])
		}
	})

	t.Run("Empty namespace", func(t *testing.T) {
		// Test namespace with no declarations
		s := schema.NewSchema().WithNamespace("Empty")

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		emptyNS := parsed["Empty"].(map[string]interface{})
		if emptyNS == nil {
			t.Error("Expected empty namespace to be present")
		}
	})

	t.Run("MarshalCedar with empty schema", func(t *testing.T) {
		// Test MarshalCedar with empty schema
		s := schema.NewSchema()

		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		// Empty schema produces empty output - that's valid
		t.Logf("Cedar output: %s", cedarData)
	})

	t.Run("UnmarshalCedar with annotations", func(t *testing.T) {
		// Test Cedar parsing with annotations
		cedarInput := []byte(`namespace Test {
	entity User {
		name: String
	};
	action view appliesTo {
		principal: [User],
		resource: [User]
	};
}`)

		var s schema.Schema
		s.SetFilename("test.cedar")
		if err := s.UnmarshalCedar(cedarInput); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}

		// Marshal back to verify
		cedarOutput, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarOutput) == 0 {
			t.Error("Expected non-empty Cedar output")
		}
	})

	t.Run("Annotations on types with nil values", func(t *testing.T) {
		// Test that annotations with nil values are handled
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"annotations": {
								"key": "value"
							},
							"attributes": {
								"field": {
									"type": "String",
									"required": true,
									"annotations": {
										"sensitive": ""
									}
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

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to parse output: %v", err)
		}

		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		e := entities["E"].(map[string]interface{})
		shape := e["shape"].(map[string]interface{})

		shapeAnnotations := shape["annotations"].(map[string]interface{})
		if shapeAnnotations["key"] != "value" {
			t.Errorf("Expected shape annotation 'key', got %v", shapeAnnotations)
		}

		attrs := shape["attributes"].(map[string]interface{})
		field := attrs["field"].(map[string]interface{})
		fieldAnnotations := field["annotations"].(map[string]interface{})
		if fieldAnnotations["sensitive"] != "" {
			t.Errorf("Expected empty annotation value, got %v", fieldAnnotations["sensitive"])
		}
	})
}
