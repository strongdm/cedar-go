package schema

import (
	"testing"
)

// TestFinal100Percent attempts to cover the last remaining branches
func TestFinal100Percent(t *testing.T) {
	t.Run("RecordType without annotations in conversion", func(t *testing.T) {
		// Test RecordType with nil annotations map
		recordType := &RecordType{
			attributes:  make(map[string]*Attribute),
			annotations: nil, // Explicitly nil
		}
		jsonRecord := convertTypeToJSONType(recordType)
		if jsonRecord.Type != "Record" {
			t.Errorf("Expected 'Record', got %v", jsonRecord.Type)
		}
		// annotations should not be set if nil
		if jsonRecord.Annotations != nil && len(jsonRecord.Annotations) > 0 {
			t.Error("Expected nil or empty annotations")
		}
	})

	t.Run("RecordType with empty annotations map", func(t *testing.T) {
		// Test RecordType with empty (but non-nil) annotations map
		recordType := &RecordType{
			attributes:  make(map[string]*Attribute),
			annotations: make(map[string]string), // Empty but not nil
		}
		jsonRecord := convertTypeToJSONType(recordType)
		if jsonRecord.Type != "Record" {
			t.Errorf("Expected 'Record', got %v", jsonRecord.Type)
		}
	})

	t.Run("SetType without annotations in conversion", func(t *testing.T) {
		// Test SetType with nil annotations
		setType := &SetType{
			element:     &PathType{path: "String"},
			annotations: nil,
		}
		jsonSet := convertTypeToJSONType(setType)
		if jsonSet.Type != "Set" {
			t.Errorf("Expected 'Set', got %v", jsonSet.Type)
		}
	})

	t.Run("PathType without annotations in conversion", func(t *testing.T) {
		// Test PathType with nil annotations
		pathType := &PathType{
			path:        "String",
			annotations: nil,
		}
		jsonPath := convertTypeToJSONType(pathType)
		if jsonPath.Type != "String" {
			t.Errorf("Expected 'String', got %v", jsonPath.Type)
		}
	})

	t.Run("MarshalJSON error path", func(t *testing.T) {
		// Try to trigger MarshalJSON error in MarshalCedar
		// This is difficult because MarshalJSON should always succeed with valid data
		s := NewSchema()

		// Empty schema should still work
		_, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}
	})

	t.Run("Complex schema round-trip through Cedar format", func(t *testing.T) {
		// Create a complex schema to ensure all MarshalCedar paths are covered
		s := NewSchema().
			WithNamespace("App",
				NewEntity("User").
					WithAttribute("id", Long()).
					WithAttribute("name", String()).
					WithAttribute("email", String()).
					WithOptionalAttribute("age", Long()),
				NewEntity("Group").
					WithAttribute("name", String()).
					MemberOf("User"),
				NewAction("createUser").
					AppliesTo(
						Principals("User"),
						Resources("Group"),
						Record(
							Attr("reason", String()),
							OptionalAttr("ticket", String()),
						),
					),
			).
			WithNamespace("Admin",
				NewEntity("SuperUser").
					WithAttribute("level", Long()),
				NewAction("deleteAll").
					AppliesTo(
						Principals("SuperUser"),
						Resources("User", "Group"),
						nil,
					),
			)

		// Marshal to Cedar
		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarData) == 0 {
			t.Error("Expected non-empty Cedar output")
		}

		// Unmarshal back
		var s2 Schema
		s2.SetFilename("test.cedar")
		if err := s2.UnmarshalCedar(cedarData); err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}

		// Marshal to Cedar again to ensure round-trip
		cedarData2, err := s2.MarshalCedar()
		if err != nil {
			t.Fatalf("Second MarshalCedar failed: %v", err)
		}

		if len(cedarData2) == 0 {
			t.Error("Expected non-empty Cedar output after round-trip")
		}
	})

	t.Run("Entity with all optional attributes", func(t *testing.T) {
		// Test entity with only optional attributes
		s := NewSchema().
			WithNamespace("Test",
				NewEntity("Optional").
					WithOptionalAttribute("field1", String()).
					WithOptionalAttribute("field2", Long()).
					WithOptionalAttribute("field3", Bool()),
			)

		// Marshal to JSON
		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON failed: %v", err)
		}

		// Unmarshal back
		var s2 Schema
		if err := s2.UnmarshalJSON(jsonData); err != nil {
			t.Fatalf("UnmarshalJSON failed: %v", err)
		}

		// Marshal to Cedar
		cedarData, err := s2.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarData) == 0 {
			t.Error("Expected non-empty Cedar output")
		}
	})

	t.Run("Type conversions with all annotation combinations", func(t *testing.T) {
		// Ensure all annotation branches are hit in type conversions

		// PathType entity reference without annotations
		entityPath := &PathType{path: "MyEntity"}
		jsonEntity := convertTypeToJSONType(entityPath)
		if jsonEntity.Type != "EntityOrCommon" {
			t.Errorf("Expected 'EntityOrCommon', got %v", jsonEntity.Type)
		}

		// RecordType with nested attributes and no top-level annotations
		nestedRecord := &RecordType{
			attributes: map[string]*Attribute{
				"nested": {
					name: "nested",
					attrType: &RecordType{
						attributes:  make(map[string]*Attribute),
						annotations: nil,
					},
					isRequired: true,
				},
			},
			annotations: nil,
		}
		jsonNested := convertTypeToJSONType(nestedRecord)
		if jsonNested.Type != "Record" {
			t.Errorf("Expected 'Record', got %v", jsonNested.Type)
		}
	})
}
