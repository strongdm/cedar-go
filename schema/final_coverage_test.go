package schema

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
)

// TestFinalCoverageGaps targets the last remaining uncovered lines for 100% coverage
func TestFinalCoverageGaps(t *testing.T) {
	t.Run("convertJSONTypeToType with nil input", func(t *testing.T) {
		// Test nil check in convertJSONTypeToType (line ~185)
		result := convertJSONTypeToType(nil)
		if result != nil {
			t.Error("Expected nil result for nil input")
		}
	})

	t.Run("convertJSONTypeToType with Boolean type (already Boolean)", func(t *testing.T) {
		// Test the case where type is already "Boolean" not "Bool" (line ~193)
		jsonType := &ast.JSONType{
			Type: "Boolean",
		}
		result := convertJSONTypeToType(jsonType)
		pathType, ok := result.(*PathType)
		if !ok {
			t.Fatal("Expected PathType")
		}
		// Should be normalized to "Bool" internally
		if pathType.path != "Bool" {
			t.Errorf("Expected path 'Bool', got %v", pathType.path)
		}
	})

	t.Run("convertTypeToJSONType with nil input", func(t *testing.T) {
		// Test nil check
		result := convertTypeToJSONType(nil)
		if result != nil {
			t.Error("Expected nil result for nil input")
		}
	})

	t.Run("convertTypeToJSONType default case", func(t *testing.T) {
		// Test the default case when Type is not one of the known types
		// Since Type interface has private method isType(), we can't create external implementations
		// But we can test by passing an invalid/corrupted type through reflection or other means
		// For now, this case is unreachable in normal operation
		// The default case returns nil, which we've tested above with nil input
	})

	t.Run("PathType with Bool normalization in convertTypeToJSONType", func(t *testing.T) {
		// Test the "Bool" -> "Boolean" normalization (line ~332)
		pathType := &PathType{path: "Bool"}
		jsonType := convertTypeToJSONType(pathType)

		if jsonType.Type != "Boolean" {
			t.Errorf("Expected type 'Boolean', got %v", jsonType.Type)
		}
	})

	t.Run("UnmarshalJSON with entity tags", func(t *testing.T) {
		// Test the jsonEntity.Tags != nil branch (line ~686)
		jsonInput := []byte(`{
			"Test": {
				"entityTypes": {
					"Resource": {
						"shape": {
							"type": "Record",
							"attributes": {
								"name": {
									"type": "String",
									"required": true
								}
							}
						},
						"tags": {
							"type": "Set",
							"element": {
								"type": "String"
							}
						}
					}
				},
				"actions": {}
			}
		}`)

		var s Schema
		if err := s.UnmarshalJSON(jsonInput); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		// Verify tags were parsed
		ns := s.namespaces["Test"]
		if ns == nil {
			t.Fatal("Expected Test namespace")
		}
		resource := ns.entities["Resource"]
		if resource == nil {
			t.Fatal("Expected Resource entity")
		}
		if resource.tags == nil {
			t.Error("Expected tags to be set")
		}
	})

	t.Run("UnmarshalCedar error path - invalid Cedar", func(t *testing.T) {
		// Test ParseFile error path (line ~835)
		var s Schema
		s.SetFilename("test.cedar")

		err := s.UnmarshalCedar([]byte("this is not valid cedar syntax {{{"))
		if err == nil {
			t.Error("Expected error for invalid Cedar syntax")
		}
	})

	t.Run("MarshalCedar internal error paths", func(t *testing.T) {
		// These error paths are very difficult to trigger because they require
		// internal Go json.Marshal/Unmarshal to fail on valid data structures
		// The error on line 847 (json.Marshal in UnmarshalCedar) would require
		// ast.ConvertHuman2JSON to return a structure that json.Marshal can't handle
		// The error on line 853 (json.Unmarshal in MarshalCedar) would require
		// s.MarshalJSON() to return invalid JSON
		// The error on line 862 (ast.Format) would require humanSchema to be invalid

		// These are essentially impossible to trigger with valid code
		// Testing would require mocking or corrupting internal structures

		// For coverage purposes, we document that these are defensive error checks
		// that protect against internal bugs or memory corruption, but are not
		// reachable through normal API usage
	})
}

// TestConversionEdgeCases tests additional edge cases in type conversion
func TestConversionEdgeCases(t *testing.T) {
	t.Run("All primitive types through conversion", func(t *testing.T) {
		// Test String
		stringType := &PathType{path: "String"}
		jsonString := convertTypeToJSONType(stringType)
		if jsonString.Type != "String" {
			t.Errorf("Expected 'String', got %v", jsonString.Type)
		}

		// Test Long
		longType := &PathType{path: "Long"}
		jsonLong := convertTypeToJSONType(longType)
		if jsonLong.Type != "Long" {
			t.Errorf("Expected 'Long', got %v", jsonLong.Type)
		}

		// Test Bool (should normalize to Boolean)
		boolType := &PathType{path: "Bool"}
		jsonBool := convertTypeToJSONType(boolType)
		if jsonBool.Type != "Boolean" {
			t.Errorf("Expected 'Boolean', got %v", jsonBool.Type)
		}

		// Test Boolean (should stay Boolean)
		booleanType := &PathType{path: "Boolean"}
		jsonBoolean := convertTypeToJSONType(booleanType)
		if jsonBoolean.Type != "Boolean" {
			t.Errorf("Expected 'Boolean', got %v", jsonBoolean.Type)
		}
	})

	t.Run("Entity type conversion", func(t *testing.T) {
		// Test entity type path
		entityType := &PathType{path: "User"}
		jsonEntity := convertTypeToJSONType(entityType)
		if jsonEntity.Type != "EntityOrCommon" {
			t.Errorf("Expected 'EntityOrCommon', got %v", jsonEntity.Type)
		}
		if jsonEntity.Name != "User" {
			t.Errorf("Expected name 'User', got %v", jsonEntity.Name)
		}
	})

	t.Run("Set with nil element", func(t *testing.T) {
		// Test set with nil element
		setType := &SetType{element: nil}
		jsonSet := convertTypeToJSONType(setType)
		if jsonSet.Type != "Set" {
			t.Errorf("Expected 'Set', got %v", jsonSet.Type)
		}
		if jsonSet.Element != nil {
			t.Error("Expected nil element")
		}
	})

	t.Run("Record with empty attributes", func(t *testing.T) {
		// Test empty record
		recordType := &RecordType{attributes: make(map[string]*Attribute)}
		jsonRecord := convertTypeToJSONType(recordType)
		if jsonRecord.Type != "Record" {
			t.Errorf("Expected 'Record', got %v", jsonRecord.Type)
		}
		if len(jsonRecord.Attributes) != 0 {
			t.Error("Expected empty attributes")
		}
	})
}

// TestSchemaInternalStructures tests internal schema structures
func TestSchemaInternalStructures(t *testing.T) {
	t.Run("Direct namespace manipulation", func(t *testing.T) {
		s := &Schema{
			namespaces: make(map[string]*Namespace),
		}

		ns := &Namespace{
			name:        "Test",
			schema:      s,
			entities:    make(map[string]*Entity),
			actions:     make(map[string]*Action),
			commonTypes: make(map[string]Type),
		}

		s.namespaces["Test"] = ns

		// Add entity with tags
		entity := &Entity{
			name:      "Resource",
			namespace: ns,
			tags:      &SetType{element: &PathType{path: "String"}},
		}
		ns.entities["Resource"] = entity

		// Marshal to JSON to exercise the tags != nil branch
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
		resource := entities["Resource"].(map[string]interface{})

		if resource["tags"] == nil {
			t.Error("Expected tags field in JSON output")
		}
	})

	t.Run("Entity with enum values", func(t *testing.T) {
		s := &Schema{
			namespaces: make(map[string]*Namespace),
		}

		ns := &Namespace{
			name:        "Test",
			schema:      s,
			entities:    make(map[string]*Entity),
			actions:     make(map[string]*Action),
			commonTypes: make(map[string]Type),
		}

		s.namespaces["Test"] = ns

		// Add entity with enum
		entity := &Entity{
			name:      "Status",
			namespace: ns,
			enum:      []string{"active", "inactive", "pending"},
		}
		ns.entities["Status"] = entity

		// Marshal and verify
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
		status := entities["Status"].(map[string]interface{})

		if status["enum"] == nil {
			t.Error("Expected enum field in JSON output")
		}
	})
}
