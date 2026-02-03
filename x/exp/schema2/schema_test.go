package schema2

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
)

func TestNewSchema(t *testing.T) {
	s := NewSchema()
	if s == nil {
		t.Fatal("NewSchema returned nil")
	}
	if s.ast == nil {
		t.Fatal("Schema.ast is nil")
	}
}

func TestSchemaBuilder(t *testing.T) {
	s := NewSchema().
		Namespace("MyApp").
		Entity("User").In("Group").Attributes(
			Attr("name", String()),
			Attr("email", String()),
			OptionalAttr("age", Long()),
		).
		Entity("Group").
		Entity("Document").In("Folder").
		Entity("Folder").
		Action("read").Principals("User", "Group").Resources("Document", "Folder").
		Action("write").Principals("User").Resources("Document")

	// Verify we can resolve it
	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Check entity types
	userType := resolved.EntityType(types.EntityType("MyApp::User"))
	if userType == nil {
		t.Fatal("User entity type not found")
	}
	if userType.Name() != "MyApp::User" {
		t.Errorf("expected name MyApp::User, got %s", userType.Name())
	}

	// Check User has Group as parent (so Group should have User as descendant)
	groupType := resolved.EntityType(types.EntityType("MyApp::Group"))
	if groupType == nil {
		t.Fatal("Group entity type not found")
	}
	if !groupType.HasDescendant(types.EntityType("MyApp::User")) {
		t.Error("Group should have User as descendant")
	}

	// Check attributes
	if userType.Attributes() == nil {
		t.Fatal("User should have attributes")
	}
	nameAttr := userType.Attributes().Attribute("name")
	if nameAttr == nil {
		t.Fatal("User should have name attribute")
	}
	if !nameAttr.Required() {
		t.Error("name attribute should be required")
	}
	ageAttr := userType.Attributes().Attribute("age")
	if ageAttr == nil {
		t.Fatal("User should have age attribute")
	}
	if ageAttr.Required() {
		t.Error("age attribute should be optional")
	}

	// Check actions
	readAction := resolved.Action(types.NewEntityUID("MyApp::Action", "read"))
	if readAction == nil {
		t.Fatal("read action not found")
	}
	if readAction.AppliesTo() == nil {
		t.Fatal("read action should have appliesTo")
	}
	principals := readAction.AppliesTo().Principals()
	if len(principals) != 2 {
		t.Errorf("expected 2 principals, got %d", len(principals))
	}
}

func TestEnumEntityType(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("Status").Enum("Active", "Inactive", "Pending")

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	statusType := resolved.EntityType(types.EntityType("App::Status"))
	if statusType == nil {
		t.Fatal("Status entity type not found")
	}

	enumKind, ok := statusType.Kind().(EnumEntityType)
	if !ok {
		t.Fatal("Status should be an enum type")
	}

	values := enumKind.Values()
	if len(values) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(values))
	}

	// Check that values are EntityUIDs with the entity type
	for _, v := range values {
		if v.Type != "App::Status" {
			t.Errorf("expected type App::Status, got %s", v.Type)
		}
	}
}

func TestCommonTypes(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		CommonType("Address", Record(
			Attr("street", String()),
			Attr("city", String()),
			Attr("zip", String()),
		)).
		Entity("User").Attributes(
			Attr("homeAddress", EntityRef("Address")),
			Attr("workAddress", EntityRef("Address")),
		)

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	userType := resolved.EntityType(types.EntityType("App::User"))
	if userType == nil {
		t.Fatal("User entity type not found")
	}
	if userType.Attributes() == nil {
		t.Fatal("User should have attributes")
	}
}

func TestMarshalJSON(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").In("Group").Attributes(
			Attr("name", String()),
		).
		Entity("Group").
		Action("read").Principals("User").Resources("User")

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var js map[string]interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		t.Fatalf("MarshalJSON produced invalid JSON: %v", err)
	}

	// Verify namespace exists
	if _, ok := js["App"]; !ok {
		t.Error("expected App namespace in JSON")
	}
}

func TestParseJSON(t *testing.T) {
	jsonData := `{
		"App": {
			"entityTypes": {
				"User": {
					"memberOfTypes": ["Group"],
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String", "required": true}
						}
					}
				},
				"Group": {}
			},
			"actions": {
				"read": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"]
					}
				}
			}
		}
	}`

	s, err := ParseJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	userType := resolved.EntityType(types.EntityType("App::User"))
	if userType == nil {
		t.Fatal("User entity type not found")
	}

	groupType := resolved.EntityType(types.EntityType("App::Group"))
	if groupType == nil {
		t.Fatal("Group entity type not found")
	}

	if !groupType.HasDescendant(types.EntityType("App::User")) {
		t.Error("Group should have User as descendant")
	}
}

func TestRoundTrip(t *testing.T) {
	// Build a schema
	original := NewSchema().
		Namespace("App").
		Entity("User").In("Group").Attributes(
			Attr("name", String()),
			OptionalAttr("email", String()),
		).
		Entity("Group").
		Action("read").Principals("User").Resources("User")

	// Marshal to JSON
	jsonData, err := original.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}

	// Parse back
	parsed, err := ParseJSON(jsonData)
	if err != nil {
		t.Fatalf("ParseJSON failed: %v", err)
	}

	// Resolve both and compare
	originalResolved, err := original.Resolve()
	if err != nil {
		t.Fatalf("Resolve original failed: %v", err)
	}

	parsedResolved, err := parsed.Resolve()
	if err != nil {
		t.Fatalf("Resolve parsed failed: %v", err)
	}

	// Check same entity types exist
	originalUser := originalResolved.EntityType(types.EntityType("App::User"))
	parsedUser := parsedResolved.EntityType(types.EntityType("App::User"))

	if originalUser == nil || parsedUser == nil {
		t.Fatal("User entity type missing after round trip")
	}

	if originalUser.Name() != parsedUser.Name() {
		t.Error("User names don't match after round trip")
	}
}

func TestValidationErrors(t *testing.T) {
	// Test undefined type reference
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("NonExistent").Resources("User")

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for undefined type reference")
	}
}
