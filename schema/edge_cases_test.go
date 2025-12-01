package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/schema"
)

// TestRemainingEdgeCases tests edge cases and rarely-hit branches
func TestRemainingEdgeCases(t *testing.T) {
	t.Run("Schema without NewSchema initialization", func(t *testing.T) {
		// Create schema without using NewSchema() to hit the nil check in WithNamespace
		var s schema.Schema
		s = *s.WithNamespace("Test",
			schema.NewEntity("User"),
		)

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		if parsed["Test"] == nil {
			t.Error("Expected Test namespace")
		}
	})

	t.Run("Type with annotations in Set", func(t *testing.T) {
		// Test Set with annotated element type
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Set",
							"element": {
								"type": "String",
								"annotations": {
									"doc": "String element"
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

	t.Run("EntityOrCommon with annotations", func(t *testing.T) {
		// Test EntityOrCommon type path with annotations
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "EntityOrCommon",
							"name": "User",
							"annotations": {
								"doc": "User reference"
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

	t.Run("Long type with annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Long",
							"annotations": {
								"range": "0..100"
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

	t.Run("Record type with annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Record",
							"annotations": {
								"doc": "A record type"
							},
							"attributes": {}
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

	t.Run("Set with annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Set",
							"annotations": {
								"doc": "A set type"
							},
							"element": {
								"type": "String"
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

	t.Run("Extension with annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "Extension",
							"name": "ipaddr",
							"annotations": {
								"doc": "IP address extension"
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

	t.Run("Unknown type with annotations", func(t *testing.T) {
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"E": {
						"shape": {
							"type": "SomeUnknownType",
							"annotations": {
								"doc": "Unknown type"
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

	t.Run("UnmarshalCedar with invalid JSON conversion", func(t *testing.T) {
		// Test error handling in UnmarshalCedar if JSON conversion fails
		var s schema.Schema
		s.SetFilename("test.cedar")

		// Test with valid Cedar syntax
		cedarInput := []byte(`namespace App {
	entity User;
}`)

		if err := s.UnmarshalCedar(cedarInput); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}

		// Verify it worked
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			t.Fatalf("Failed to parse: %v", err)
		}

		if parsed["App"] == nil {
			t.Error("Expected App namespace")
		}
	})

	t.Run("MarshalCedar JSON conversion coverage", func(t *testing.T) {
		// Ensure MarshalCedar's internal JSON marshaling is covered
		s := schema.NewSchema().
			WithNamespace("App",
				schema.NewEntity("User").
					WithAttribute("name", schema.String()),
				schema.NewAction("view").
					AppliesTo(
						schema.Principals("User"),
						schema.Resources("User"),
						nil,
					),
			)

		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarData) == 0 {
			t.Error("Expected non-empty Cedar output")
		}

		// Verify it can be parsed back
		var s2 schema.Schema
		s2.SetFilename("test.cedar")
		if err := s2.UnmarshalCedar(cedarData); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}
	})

	t.Run("UnmarshalJSON with invalid JSON", func(t *testing.T) {
		var s schema.Schema
		err := s.UnmarshalJSON([]byte("invalid json"))
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("UnmarshalJSON with complex nested structures", func(t *testing.T) {
		jsonInput := []byte(`{
			"App": {
				"entityTypes": {
					"User": {
						"shape": {
							"type": "Record",
							"attributes": {
								"profile": {
									"type": "Record",
									"required": true,
									"attributes": {
										"settings": {
											"type": "Record",
											"required": false,
											"attributes": {
												"theme": {
													"type": "String",
													"required": true
												}
											}
										}
									}
								}
							}
						}
					}
				},
				"actions": {
					"update": {
						"appliesTo": {
							"principalTypes": ["User"],
							"resourceTypes": ["User"],
							"context": {
								"type": "Record",
								"attributes": {
									"ip": {
										"type": "Extension",
										"name": "ipaddr",
										"required": true
									}
								}
							}
						}
					}
				}
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
			t.Fatalf("Second unmarshal failed: %v", err)
		}
	})
}
