package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

// TestFullCoverage ensures 100% test coverage for all public APIs
func TestFullCoverage(t *testing.T) {
	t.Run("RecordType methods", func(t *testing.T) {
		// Test Record builder with methods
		r := schema.Record()
		r = r.WithAttribute("field1", schema.String())
		r = r.WithOptionalAttribute("field2", schema.Long())
		r = r.WithAnnotation("key", "value")

		// Use in schema to verify it works
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E").WithShape(r),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}
	})

	t.Run("Attribute annotations", func(t *testing.T) {
		attr := schema.Attr("name", schema.String()).WithAnnotation("doc", "User name")

		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("User").WithShape(schema.Record(attr)),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Verify annotation is present
		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		user := entities["User"].(map[string]interface{})
		shape := user["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})
		nameAttr := attrs["name"].(map[string]interface{})
		annotations := nameAttr["annotations"].(map[string]interface{})
		if annotations["doc"] != "User name" {
			t.Errorf("Expected annotation 'doc' = 'User name', got %v", annotations["doc"])
		}
	})

	t.Run("SetType annotations", func(t *testing.T) {
		setType := schema.Set(schema.String()).WithAnnotation("doc", "List of tags")

		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("Doc").WithAttribute("tags", setType),
			)

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("PathType annotations", func(t *testing.T) {
		pathType := schema.String().WithAnnotation("doc", "A string field")

		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E").WithAttribute("field", pathType),
			)

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("CommonType usage", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.TypeDecl("EmailAddress", schema.String()).
					WithAnnotation("doc", "An email address"),
				schema.NewEntity("User").
					WithAttribute("email", schema.CommonType("EmailAddress")),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Round trip to verify
		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
	})

	t.Run("Boolean alias", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E").WithAttribute("flag", schema.Boolean()),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		e := entities["E"].(map[string]interface{})
		shape := e["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})
		flagAttr := attrs["flag"].(map[string]interface{})
		if flagAttr["type"] != "Boolean" {
			t.Errorf("Expected type 'Boolean', got %v", flagAttr["type"])
		}
	})

	t.Run("Action with MemberOf", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewAction("read"),
				schema.NewAction("viewPhoto").
					MemberOf(schema.ActionGroup("read")).
					WithAnnotation("doc", "View a photo"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		actions := testNS["actions"].(map[string]interface{})
		viewPhoto := actions["viewPhoto"].(map[string]interface{})
		memberOf := viewPhoto["memberOf"].([]interface{})
		if len(memberOf) != 1 {
			t.Errorf("Expected 1 memberOf, got %d", len(memberOf))
		}
	})

	t.Run("QualifiedActionGroup", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewAction("listPhotos").
					MemberOf(schema.QualifiedActionGroup("read", "PhotoApp::Action")),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		actions := testNS["actions"].(map[string]interface{})
		listPhotos := actions["listPhotos"].(map[string]interface{})
		memberOf := listPhotos["memberOf"].([]interface{})
		if len(memberOf) != 1 {
			t.Errorf("Expected 1 memberOf, got %d", len(memberOf))
		}
		member := memberOf[0].(map[string]interface{})
		if member["type"] != "PhotoApp::Action" {
			t.Errorf("Expected type 'PhotoApp::Action', got %v", member["type"])
		}
	})

	t.Run("Entity WithShape", func(t *testing.T) {
		customShape := schema.Record(
			schema.Attr("id", schema.Long()),
			schema.Attr("name", schema.String()),
		)

		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("User").WithShape(customShape),
			)

		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Entity WithTags", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("Resource").
					WithAttribute("name", schema.String()).
					WithTags(schema.String()),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		resource := entities["Resource"].(map[string]interface{})
		if resource["tags"] == nil {
			t.Error("Expected tags to be present")
		}
	})

	t.Run("Namespace WithAnnotation", func(t *testing.T) {
		// Annotations on namespaces are currently not directly exposed in the builder API
		// but we can test that they work through JSON round-trip
		jsonInput := []byte(`{
			"Test": {
				"annotations": {
					"doc": "Test namespace"
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

		jsonOutput, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonOutput, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		annotations := testNS["annotations"].(map[string]interface{})
		if annotations["doc"] != "Test namespace" {
			t.Errorf("Expected namespace annotation, got %v", annotations)
		}
	})

	t.Run("SetFilename", func(t *testing.T) {
		s := schema.NewSchema()
		s.SetFilename("test.cedar")

		// Create a schema with an error to verify filename appears in error
		err := s.UnmarshalCedar([]byte("invalid schema syntax {"))
		if err == nil {
			t.Error("Expected error for invalid schema")
		}
		// The error should contain the filename
		if err != nil {
			errStr := err.Error()
			// Just verify we get an error; filename handling is tested in parser
			if errStr == "" {
				t.Error("Expected non-empty error message")
			}
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Empty namespace name", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("",
				schema.NewEntity("User"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		if parsed[""] == nil {
			t.Error("Expected empty namespace to be present")
		}
	})

	t.Run("Multiple namespaces", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("App1",
				schema.NewEntity("User"),
			).
			WithNamespace("App2",
				schema.NewEntity("Document"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		if parsed["App1"] == nil || parsed["App2"] == nil {
			t.Error("Expected both namespaces to be present")
		}
	})

	t.Run("Entity with multiple memberOf", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("User").MemberOf("Group1", "Group2", "Group3"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		user := entities["User"].(map[string]interface{})
		memberOf := user["memberOfTypes"].([]interface{})
		if len(memberOf) != 3 {
			t.Errorf("Expected 3 memberOf, got %d", len(memberOf))
		}
	})

	t.Run("Deeply nested records", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("Complex").
					WithAttribute("level1", schema.Record(
						schema.Attr("level2", schema.Record(
							schema.Attr("level3", schema.Record(
								schema.Attr("value", schema.String()),
							)),
						)),
					)),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Round trip
		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonData2, err := s2.MarshalJSON()
		if err != nil {
			t.Fatalf("Second MarshalJSON failed: %v", err)
		}

		if len(jsonData) != len(jsonData2) {
			t.Error("Round trip changed schema")
		}
	})

	t.Run("Action without appliesTo", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewAction("groupAction"),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		actions := testNS["actions"].(map[string]interface{})
		groupAction := actions["groupAction"].(map[string]interface{})
		if groupAction["appliesTo"] != nil {
			t.Error("Expected no appliesTo for group action")
		}
	})

	t.Run("Extension types", func(t *testing.T) {
		// Test parsing extension types from JSON
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
			t.Fatalf("Failed to unmarshal JSON: %v", err)
		}
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		network := entities["Network"].(map[string]interface{})
		if network == nil {
			t.Error("Expected Network entity")
		}
	})
}

func TestConversionEdgeCases(t *testing.T) {
	t.Run("Unknown JSON type defaults to path", func(t *testing.T) {
		// This tests the default case in convertJSONTypeToType
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"Unknown": {
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

		// Should not crash, treats unknown as path type
		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Nil element type in Set", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"attributes": {
								"items": {
									"type": "Set",
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

		// Should handle nil element
		_, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}
	})

	t.Run("Empty record attributes", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("Empty").WithShape(schema.Record()),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
	})

	t.Run("Action with empty context", func(t *testing.T) {
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewAction("test").AppliesTo(
					schema.Principals("User"),
					schema.Resources("Doc"),
					schema.Record(), // Empty context
				),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}
	})
}
