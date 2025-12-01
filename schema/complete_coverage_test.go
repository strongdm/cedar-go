package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

// TestCompleteCoverage adds tests for remaining uncovered code paths
func TestCompleteCoverage(t *testing.T) {
	t.Run("MarshalCedar error cases", func(t *testing.T) {
		// Test error when converting JSON with issues
		s := schema.NewSchema()

		// Empty schema should return empty result
		result, err := s.MarshalCedar()
		if err != nil {
			t.Logf("MarshalCedar on empty schema: %v", err)
		}
		if len(result) > 0 {
			t.Logf("Got result: %s", result)
		}
	})

	t.Run("JSON unmarshaling edge cases", func(t *testing.T) {
		// Test various JSON structures to cover convertJSONTypeToType branches
		testCases := []struct {
			name  string
			input string
		}{
			{
				name: "Entity with Boolean type (capitalized)",
				input: `{
					"Test": {
						"entityTypes": {
							"E": {
								"shape": {
									"type": "Record",
									"attributes": {
										"flag": {
											"type": "Boolean",
											"required": true
										}
									}
								}
							}
						},
						"actions": {}
					}
				}`,
			},
			{
				name: "Entity with Set of primitives",
				input: `{
					"Test": {
						"entityTypes": {
							"E": {
								"shape": {
									"type": "Record",
									"attributes": {
										"tags": {
											"type": "Set",
											"required": true,
											"element": {
												"type": "String"
											}
										}
									}
								}
							}
						},
						"actions": {}
					}
				}`,
			},
			{
				name: "Entity with nested Set in attribute",
				input: `{
					"Test": {
						"entityTypes": {
							"E": {
								"shape": {
									"type": "Record",
									"attributes": {
										"items": {
											"type": "Set",
											"required": false,
											"element": {
												"type": "EntityOrCommon",
												"name": "Item"
											}
										}
									}
								}
							}
						},
						"actions": {}
					}
				}`,
			},
			{
				name: "Entity with nested Record in attribute",
				input: `{
					"Test": {
						"entityTypes": {
							"E": {
								"shape": {
									"type": "Record",
									"attributes": {
										"metadata": {
											"type": "Record",
											"required": true,
											"attributes": {
												"created": {
													"type": "Long",
													"required": true
												}
											}
										}
									}
								}
							}
						},
						"actions": {}
					}
				}`,
			},
			{
				name: "Action with empty principal and resource types",
				input: `{
					"Test": {
						"entityTypes": {},
						"actions": {
							"groupAction": {
								"appliesTo": {
									"principalTypes": [],
									"resourceTypes": []
								}
							}
						}
					}
				}`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var s schema.Schema
				if err := s.UnmarshalJSON([]byte(tc.input)); err != nil {
					t.Fatalf("UnmarshalJSON failed: %v", err)
				}

				// Marshal back
				jsonData, err := s.MarshalJSON()
				if err != nil {
					t.Fatalf("MarshalJSON failed: %v", err)
				}

				// Unmarshal again to verify round-trip
				var s2 schema.Schema
				if err := s2.UnmarshalJSON(jsonData); err != nil {
					t.Fatalf("Second UnmarshalJSON failed: %v", err)
				}

				// Try Cedar format too
				cedarData, err := s.MarshalCedar()
				if err != nil {
					t.Fatalf("MarshalCedar failed: %v", err)
				}

				if len(cedarData) == 0 {
					t.Error("Expected non-empty Cedar output")
				}
			})
		}
	})

	t.Run("WithNamespace edge cases", func(t *testing.T) {
		// Test adding multiple declarations to namespace
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("E1"),
				schema.NewEntity("E2"),
				schema.NewAction("a1"),
				schema.NewAction("a2"),
				schema.TypeDecl("T1", schema.String()),
			)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(jsonData, &parsed)
		testNS := parsed["Test"].(map[string]interface{})

		entities := testNS["entityTypes"].(map[string]interface{})
		if len(entities) != 2 {
			t.Errorf("Expected 2 entities, got %d", len(entities))
		}

		actions := testNS["actions"].(map[string]interface{})
		if len(actions) != 2 {
			t.Errorf("Expected 2 actions, got %d", len(actions))
		}

		commonTypes := testNS["commonTypes"].(map[string]interface{})
		if len(commonTypes) != 1 {
			t.Errorf("Expected 1 common type, got %d", len(commonTypes))
		}
	})

	t.Run("UnmarshalCedar with invalid syntax", func(t *testing.T) {
		testCases := []string{
			"invalid syntax",
			"namespace Test { invalid }",
			"",
		}

		for i, tc := range testCases {
			var s schema.Schema
			s.SetFilename("test.cedar")
			err := s.UnmarshalCedar([]byte(tc))
			// Empty string should not error
			if i == 2 && err != nil {
				t.Errorf("Case %d: Unexpected error for empty input: %v", i, err)
			}
			// Invalid syntax should error
			if i < 2 && err == nil {
				t.Errorf("Case %d: Expected error for invalid syntax", i)
			}
		}
	})

	t.Run("Entity WithOptionalAttribute when shape exists", func(t *testing.T) {
		// First add a required attribute, then add an optional one
		entity := schema.NewEntity("User").
			WithAttribute("name", schema.String()).
			WithOptionalAttribute("email", schema.String()).
			WithOptionalAttribute("phone", schema.String())

		s := schema.NewSchema().
			WithNamespace("Test", entity)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		json.Unmarshal(jsonData, &parsed)
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		user := entities["User"].(map[string]interface{})
		shape := user["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})

		nameAttr := attrs["name"].(map[string]interface{})
		if !nameAttr["required"].(bool) {
			t.Error("Expected name to be required")
		}

		emailAttr := attrs["email"].(map[string]interface{})
		if emailAttr["required"].(bool) {
			t.Error("Expected email to be optional")
		}
	})

	t.Run("Conversion of all primitive types", func(t *testing.T) {
		// Create schema with all primitive types to ensure conversion coverage
		s := schema.NewSchema().
			WithNamespace("Test",
				schema.NewEntity("AllTypes").
					WithAttribute("stringField", schema.String()).
					WithAttribute("longField", schema.Long()).
					WithAttribute("boolField", schema.Bool()),
			)

		// Convert to JSON
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Parse JSON to verify all types are correct
		var parsed map[string]interface{}
		json.Unmarshal(jsonData, &parsed)
		testNS := parsed["Test"].(map[string]interface{})
		entities := testNS["entityTypes"].(map[string]interface{})
		allTypes := entities["AllTypes"].(map[string]interface{})
		shape := allTypes["shape"].(map[string]interface{})
		attrs := shape["attributes"].(map[string]interface{})

		stringField := attrs["stringField"].(map[string]interface{})
		if stringField["type"] != "String" {
			t.Errorf("Expected String type, got %v", stringField["type"])
		}

		longField := attrs["longField"].(map[string]interface{})
		if longField["type"] != "Long" {
			t.Errorf("Expected Long type, got %v", longField["type"])
		}

		boolField := attrs["boolField"].(map[string]interface{})
		if boolField["type"] != "Boolean" {
			t.Errorf("Expected Boolean type, got %v", boolField["type"])
		}

		// Round-trip through Cedar format
		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalCedar(cedarData); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}
	})

	t.Run("MarshalJSON with uninitialized namespaces", func(t *testing.T) {
		s := schema.NewSchema()
		// Should return empty JSON object
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		if string(jsonData) != "{}" {
			t.Errorf("Expected '{}', got %s", jsonData)
		}
	})

	t.Run("Complex schema from corpus", func(t *testing.T) {
		// Create a complex schema similar to what might be in corpus tests
		s := schema.NewSchema().
			WithNamespace("PhotoFlash",
				schema.NewEntity("User").
					WithAttribute("department", schema.String()).
					WithAttribute("jobLevel", schema.Long()).
					MemberOf("UserGroup").
					WithAnnotation("doc", "A user in the system"),
				schema.NewEntity("UserGroup"),
				schema.NewEntity("Photo").
					WithAttribute("account", schema.EntityType("Account")).
					WithOptionalAttribute("private", schema.Bool()).
					MemberOf("Album"),
				schema.NewEntity("Album").
					MemberOf("Album"),
				schema.NewEntity("Account").
					WithOptionalAttribute("owner", schema.EntityType("User")).
					WithOptionalAttribute("admins", schema.Set(schema.EntityType("User"))),
				schema.NewAction("viewPhoto").
					AppliesTo(
						schema.Principals("User"),
						schema.Resources("Photo"),
						schema.Record(
							schema.OptionalAttr("authenticated", schema.Bool()),
						),
					).
					MemberOf(schema.ActionGroup("read")),
				schema.NewAction("read"),
			)

		// Test all operations
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var s2 schema.Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if !strings.Contains(string(cedarData), "PhotoFlash") {
			t.Error("Expected Cedar output to contain namespace")
		}

		var s3 schema.Schema
		s3.SetFilename("test.cedar")
		if err := s3.UnmarshalCedar(cedarData); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}
	})
}
