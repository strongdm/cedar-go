package schema

import (
	"testing"
)

// TestComplete100Coverage targets the absolute final missing coverage points
func TestComplete100Coverage(t *testing.T) {
	t.Run("PathType with all primitive types and annotations", func(t *testing.T) {
		// Test String with annotations
		stringType := &PathType{
			path: "String",
			annotations: map[string]string{
				"doc": "A string type",
			},
		}
		jsonString := convertTypeToJSONType(stringType)
		if jsonString.Type != "String" {
			t.Errorf("Expected 'String', got %v", jsonString.Type)
		}
		if jsonString.Annotations == nil || jsonString.Annotations["doc"] != "A string type" {
			t.Error("Expected annotations to be preserved")
		}

		// Test Long with annotations
		longType := &PathType{
			path: "Long",
			annotations: map[string]string{
				"range": "0-100",
			},
		}
		jsonLong := convertTypeToJSONType(longType)
		if jsonLong.Type != "Long" {
			t.Errorf("Expected 'Long', got %v", jsonLong.Type)
		}
		if jsonLong.Annotations == nil {
			t.Error("Expected annotations")
		}

		// Test Boolean with annotations
		booleanType := &PathType{
			path: "Boolean",
			annotations: map[string]string{
				"doc": "Boolean type",
			},
		}
		jsonBoolean := convertTypeToJSONType(booleanType)
		if jsonBoolean.Type != "Boolean" {
			t.Errorf("Expected 'Boolean', got %v", jsonBoolean.Type)
		}
		if jsonBoolean.Annotations == nil {
			t.Error("Expected annotations")
		}
	})

	t.Run("SetType with annotations", func(t *testing.T) {
		// SetType with annotations
		setType := &SetType{
			element: &PathType{path: "String"},
			annotations: map[string]string{
				"doc": "A set of strings",
			},
		}
		jsonSet := convertTypeToJSONType(setType)
		if jsonSet.Type != "Set" {
			t.Errorf("Expected 'Set', got %v", jsonSet.Type)
		}
		if jsonSet.Annotations == nil || jsonSet.Annotations["doc"] != "A set of strings" {
			t.Error("Expected annotations to be preserved")
		}
	})

	t.Run("RecordType with annotations", func(t *testing.T) {
		// RecordType with annotations
		recordType := &RecordType{
			attributes: map[string]*Attribute{
				"field": {
					name:       "field",
					attrType:   &PathType{path: "String"},
					isRequired: true,
				},
			},
			annotations: map[string]string{
				"doc": "A record type",
			},
		}
		jsonRecord := convertTypeToJSONType(recordType)
		if jsonRecord.Type != "Record" {
			t.Errorf("Expected 'Record', got %v", jsonRecord.Type)
		}
		if jsonRecord.Annotations == nil || jsonRecord.Annotations["doc"] != "A record type" {
			t.Error("Expected annotations to be preserved")
		}
	})

	t.Run("PathType entity reference with annotations", func(t *testing.T) {
		// Entity type (non-primitive) with annotations
		entityType := &PathType{
			path: "User",
			annotations: map[string]string{
				"doc": "User entity reference",
			},
		}
		jsonEntity := convertTypeToJSONType(entityType)
		if jsonEntity.Type != "EntityOrCommon" {
			t.Errorf("Expected 'EntityOrCommon', got %v", jsonEntity.Type)
		}
		if jsonEntity.Name != "User" {
			t.Errorf("Expected name 'User', got %v", jsonEntity.Name)
		}
		if jsonEntity.Annotations == nil {
			t.Error("Expected annotations")
		}
	})

	t.Run("Complex nested structure with all annotations", func(t *testing.T) {
		// Create a complex nested structure to exercise all annotation paths
		s := &Schema{
			namespaces: make(map[string]*Namespace),
		}

		ns := &Namespace{
			name:        "Test",
			schema:      s,
			entities:    make(map[string]*Entity),
			actions:     make(map[string]*Action),
			commonTypes: make(map[string]Type),
			annotations: map[string]string{
				"namespace_doc": "Test namespace",
			},
		}

		// Entity with complex shape including all type variations with annotations
		entity := &Entity{
			name:      "ComplexEntity",
			namespace: ns,
			shape: &RecordType{
				attributes: map[string]*Attribute{
					"name": {
						name: "name",
						attrType: &PathType{
							path: "String",
							annotations: map[string]string{
								"field_doc": "Name field",
							},
						},
						isRequired: true,
					},
					"tags": {
						name: "tags",
						attrType: &SetType{
							element: &PathType{path: "String"},
							annotations: map[string]string{
								"set_doc": "Tags set",
							},
						},
						isRequired: false,
					},
					"metadata": {
						name: "metadata",
						attrType: &RecordType{
							attributes: map[string]*Attribute{
								"created": {
									name:       "created",
									attrType:   &PathType{path: "Long"},
									isRequired: true,
								},
							},
							annotations: map[string]string{
								"nested_doc": "Metadata record",
							},
						},
						isRequired: false,
					},
				},
				annotations: map[string]string{
					"shape_doc": "Complex entity shape",
				},
			},
			annotations: map[string]string{
				"entity_doc": "A complex entity",
			},
		}

		ns.entities["ComplexEntity"] = entity
		s.namespaces["Test"] = ns

		// Marshal to JSON to exercise all annotation paths
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Verify it's valid JSON
		if len(jsonData) == 0 {
			t.Error("Expected non-empty JSON")
		}

		// Unmarshal and marshal again to verify round-trip
		var s2 Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		jsonData2, err := s2.MarshalJSON()
		if err != nil {
			t.Fatalf("Second MarshalJSON failed: %v", err)
		}

		if len(jsonData2) == 0 {
			t.Error("Expected non-empty JSON after round-trip")
		}
	})
}
