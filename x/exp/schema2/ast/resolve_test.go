package ast

import (
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// TestResolvedConversionCoverage tests the conversion functions to ensure full coverage
func TestResolvedConversionCoverage(t *testing.T) {
	t.Parallel()

	t.Run("convertResolvedEntityToNode with MemberOf", func(t *testing.T) {
		// Create a resolved entity with MemberOf
		resolved := ResolvedEntity{
			Name:        "App::User",
			Annotations: []Annotation{{Key: "test", Value: "value"}},
			MemberOf:    []types.EntityType{"App::Group", "App::Team"},
			Shape: &RecordType{
				Pairs: []Pair{
					{Key: "name", Type: StringType{}},
				},
			},
			Tags: StringType{},
		}

		// Convert to node
		node := convertResolvedEntityToNode(resolved, "User")

		// Verify conversion
		if node.Name != "User" {
			t.Errorf("expected name 'User', got %v", node.Name)
		}
		if len(node.MemberOfVal) != 2 {
			t.Fatalf("expected 2 MemberOf entries, got %d", len(node.MemberOfVal))
		}
		if node.MemberOfVal[0].Name != "App::Group" {
			t.Errorf("expected MemberOf[0] to be 'App::Group', got %v", node.MemberOfVal[0].Name)
		}
		if node.MemberOfVal[1].Name != "App::Team" {
			t.Errorf("expected MemberOf[1] to be 'App::Team', got %v", node.MemberOfVal[1].Name)
		}
		if node.ShapeVal == nil {
			t.Error("expected Shape to be set")
		}
		if node.TagsVal == nil {
			t.Error("expected Tags to be set")
		}
	})

	t.Run("convertResolvedActionToNode with all fields", func(t *testing.T) {
		// Create a resolved action with all fields populated
		resolved := ResolvedAction{
			Name:        "view",
			Annotations: []Annotation{{Key: "test", Value: "value"}},
			MemberOf: []types.EntityUID{
				types.NewEntityUID("Action", "read"),
				types.NewEntityUID("Action", "write"),
			},
			AppliesTo: &ResolvedAppliesTo{
				PrincipalTypes: []types.EntityType{"User", "Admin"},
				ResourceTypes:  []types.EntityType{"Document", "File"},
				Context:        StringType{},
			},
		}

		// Convert to node
		node := convertResolvedActionToNode(resolved)

		// Verify conversion
		if node.Name != "view" {
			t.Errorf("expected name 'view', got %v", node.Name)
		}
		if len(node.MemberOfVal) != 2 {
			t.Fatalf("expected 2 MemberOf entries, got %d", len(node.MemberOfVal))
		}
		if node.MemberOfVal[0].Type.Name != "Action" || node.MemberOfVal[0].ID != "read" {
			t.Errorf("expected MemberOf[0] to be Action::\"read\", got %v::%v",
				node.MemberOfVal[0].Type.Name, node.MemberOfVal[0].ID)
		}
		if node.AppliesToVal == nil {
			t.Fatal("expected AppliesTo to be set")
		}
		if len(node.AppliesToVal.PrincipalTypes) != 2 {
			t.Errorf("expected 2 PrincipalTypes, got %d", len(node.AppliesToVal.PrincipalTypes))
		}
		if len(node.AppliesToVal.ResourceTypes) != 2 {
			t.Errorf("expected 2 ResourceTypes, got %d", len(node.AppliesToVal.ResourceTypes))
		}
		if node.AppliesToVal.Context == nil {
			t.Error("expected Context to be set")
		}
	})

	t.Run("convertResolvedActionToNode with empty fields", func(t *testing.T) {
		// Create a minimal resolved action with no MemberOf or AppliesTo
		resolved := ResolvedAction{
			Name:        "basic",
			Annotations: []Annotation{},
			MemberOf:    nil,
			AppliesTo:   nil,
		}

		// Convert to node
		node := convertResolvedActionToNode(resolved)

		// Verify conversion
		if node.Name != "basic" {
			t.Errorf("expected name 'basic', got %v", node.Name)
		}
		if len(node.MemberOfVal) != 0 {
			t.Errorf("expected 0 MemberOf entries, got %d", len(node.MemberOfVal))
		}
		if node.AppliesToVal != nil {
			t.Error("expected AppliesTo to be nil")
		}
	})

	t.Run("ResolvedSchema.Schema with entities having MemberOf", func(t *testing.T) {
		// Create a schema with an entity that has MemberOf
		schema := NewSchema(
			Entity("Group"),
			Entity("User").MemberOf(Ref("Group")),
		)

		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		// Convert back to schema
		backToSchema := resolved.Schema()
		if backToSchema == nil {
			t.Fatal("Schema() returned nil")
		}

		// Verify that User entity has MemberOf
		userFound := false
		for _, node := range backToSchema.Nodes {
			if entity, ok := node.(EntityNode); ok && entity.Name == "User" {
				userFound = true
				if len(entity.MemberOfVal) == 0 {
					t.Error("expected User to have MemberOf")
				}
			}
		}
		if !userFound {
			t.Error("User entity not found in converted schema")
		}
	})
}

// TestEnumNodeEntityUIDsEarlyTermination tests early termination of EntityUIDs iterator
func TestEnumNodeEntityUIDsEarlyTermination(t *testing.T) {
	t.Parallel()

	t.Run("EnumNode.EntityUIDs early break", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")

		count := 0
		for range enum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}

		if count != 1 {
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})

	t.Run("ResolvedEnum.EntityUIDs early break", func(t *testing.T) {
		schema := NewSchema(Enum("Status", "active", "inactive", "pending"))
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		resolvedEnum := resolved.Enums["Status"]
		count := 0
		for range resolvedEnum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}

		if count != 1 {
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})
}

func TestResolvedSchemaToSchema(t *testing.T) {
	t.Parallel()

	t.Run("converts resolved schema with mixed types and sorting", func(t *testing.T) {
		// Create a schema with mixed ordering to test sorting
		input := NewSchema(
			// Top-level: enum, entity, action (will be sorted)
			Enum("AStatus", "active", "inactive"),
			Entity("XPlace"),
			Action("BAction"),

			// Namespaces in reverse alphabetical order (will be sorted)
			Namespace("ZNamespace",
				Entity("ZEntity"),
				Action("ZAction"),
			),
			Namespace("ANamespace",
				Entity("AEntity"),
				Enum("AEnum", "one", "two"),
			),
		)

		// Resolve it
		resolved, err := input.Resolve()
		testutil.OK(t, err)

		// Convert back to Schema
		schema := resolved.Schema()
		if schema == nil {
			t.Fatal("Schema() returned nil")
		}
		if len(schema.Nodes) == 0 {
			t.Fatal("Schema() returned empty nodes")
		}

		// Marshal to verify it's valid and sorted
		cedarBytes := schema.MarshalCedar()
		if len(cedarBytes) == 0 {
			t.Fatal("MarshalCedar() returned empty bytes")
		}

		cedarStr := string(cedarBytes)
		if len(cedarStr) == 0 {
			t.Fatal("MarshalCedar() returned empty string")
		}

		// Verify sorting: top-level entities should come before enums, which come before actions
		// And within each type, they should be alphabetically sorted
		xPlaceIdx := strings.Index(cedarStr, "entity XPlace")
		aStatusIdx := strings.Index(cedarStr, "entity AStatus")
		bActionIdx := strings.Index(cedarStr, "action BAction")

		if xPlaceIdx == -1 || aStatusIdx == -1 || bActionIdx == -1 {
			t.Fatal("Expected entities, enums, and actions not found in output")
		}

		// Entities before enums before actions
		if xPlaceIdx >= aStatusIdx || aStatusIdx >= bActionIdx {
			t.Errorf("Top-level nodes not sorted correctly: entity at %d, enum at %d, action at %d",
				xPlaceIdx, aStatusIdx, bActionIdx)
		}

		// Verify namespace order (ANamespace before ZNamespace)
		aNamespaceIdx := strings.Index(cedarStr, "namespace ANamespace")
		zNamespaceIdx := strings.Index(cedarStr, "namespace ZNamespace")

		if aNamespaceIdx == -1 || zNamespaceIdx == -1 {
			t.Fatal("Expected namespaces not found in output")
		}

		if aNamespaceIdx >= zNamespaceIdx {
			t.Errorf("Namespaces not sorted correctly: ANamespace at %d, ZNamespace at %d",
				aNamespaceIdx, zNamespaceIdx)
		}
	})

	t.Run("handles optional attributes in marshaling", func(t *testing.T) {
		// Create a schema with:
		// - An entity with optional attributes (uses compact marshaling)
		// - An action with optional context fields (uses non-compact marshaling)
		input := NewSchema(
			Entity("User").Shape(
				Attribute("name", String()),
				Optional("email", String()),
			),
			Entity("Resource"),
			Action("TestAction").
				Principal(EntityType("User")).
				Resource(EntityType("Resource")).
				Context(Record(
					Attribute("required", String()),
					Optional("optionalField", Long()),
				)),
		)

		// Resolve and convert back
		resolved, err := input.Resolve()
		testutil.OK(t, err)

		schema := resolved.Schema()
		cedarBytes := schema.MarshalCedar()
		cedarStr := string(cedarBytes)

		// Verify optional attributes in entity shape (compact format: "email?": String)
		if !strings.Contains(cedarStr, `"email?":`) {
			t.Errorf("Expected optional email field with '?', got: %s", cedarStr)
		}

		// Verify optional attributes in action context (non-compact format: "optionalField"?: Long)
		if !strings.Contains(cedarStr, `"optionalField"?:`) {
			t.Errorf("Expected optional context field with '?:', got: %s", cedarStr)
		}
	})
}
