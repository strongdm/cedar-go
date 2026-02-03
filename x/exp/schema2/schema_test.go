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

func TestUnmarshalJSON(t *testing.T) {
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

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonData)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
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

	// Parse back using UnmarshalJSON
	var parsed Schema
	if err := parsed.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
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

func TestMustResolve(t *testing.T) {
	// Test successful resolve
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("User").Resources("User")

	resolved := s.MustResolve()
	if resolved == nil {
		t.Fatal("MustResolve returned nil")
	}

	// Test panic on error
	badSchema := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("NonExistent").Resources("User")

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustResolve should panic on error")
		}
	}()
	_ = badSchema.MustResolve()
}

func TestLookupEntityType(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User")

	resolved := s.MustResolve()

	// Test found case
	userType, found := resolved.LookupEntityType(types.EntityType("App::User"))
	if !found {
		t.Error("LookupEntityType should find App::User")
	}
	if userType == nil {
		t.Error("LookupEntityType should return non-nil entity type")
	}

	// Test not found case
	_, found = resolved.LookupEntityType(types.EntityType("App::NonExistent"))
	if found {
		t.Error("LookupEntityType should not find App::NonExistent")
	}
}

func TestLookupAction(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("User").Resources("User")

	resolved := s.MustResolve()

	// Test found case
	readAction, found := resolved.LookupAction(types.NewEntityUID("App::Action", "read"))
	if !found {
		t.Error("LookupAction should find read action")
	}
	if readAction == nil {
		t.Error("LookupAction should return non-nil action")
	}

	// Test not found case
	_, found = resolved.LookupAction(types.NewEntityUID("App::Action", "nonexistent"))
	if found {
		t.Error("LookupAction should not find nonexistent action")
	}
}

func TestIsEnumAndAsEnum(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Entity("Status").Enum("Active", "Inactive", "Pending")

	resolved := s.MustResolve()

	// Test standard entity type
	userType := resolved.EntityType(types.EntityType("App::User"))
	if userType.IsEnum() {
		t.Error("User should not be an enum type")
	}
	_, ok := userType.AsEnum()
	if ok {
		t.Error("AsEnum should return false for non-enum type")
	}

	// Test enum entity type
	statusType := resolved.EntityType(types.EntityType("App::Status"))
	if !statusType.IsEnum() {
		t.Error("Status should be an enum type")
	}
	enumKind, ok := statusType.AsEnum()
	if !ok {
		t.Error("AsEnum should return true for enum type")
	}
	if len(enumKind.Values()) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(enumKind.Values()))
	}
}

func TestParseCedarWithOptions(t *testing.T) {
	cedarSchema := `namespace App {
		entity User;
		action read appliesTo { principal: User, resource: User };
	}`

	// Test without options
	s, err := ParseCedar([]byte(cedarSchema))
	if err != nil {
		t.Fatalf("ParseCedar failed: %v", err)
	}
	if s == nil {
		t.Fatal("ParseCedar returned nil schema")
	}

	// Test with WithFilename option
	s, err = ParseCedar([]byte(cedarSchema), WithFilename("test.cedarschema"))
	if err != nil {
		t.Fatalf("ParseCedar with WithFilename failed: %v", err)
	}
	if s == nil {
		t.Fatal("ParseCedar with WithFilename returned nil schema")
	}

	// Verify backward compatibility with deprecated function
	s, err = ParseCedarWithFilename("test.cedarschema", []byte(cedarSchema))
	if err != nil {
		t.Fatalf("ParseCedarWithFilename failed: %v", err)
	}
	if s == nil {
		t.Fatal("ParseCedarWithFilename returned nil schema")
	}
}

func TestBuilderTypeSafety(t *testing.T) {
	// This test verifies the builder pattern provides compile-time safety
	// by ensuring EntityBuilder and ActionBuilder have the correct methods

	s := NewSchema()

	// Entity returns EntityBuilder
	eb := s.Namespace("App").Entity("User")

	// EntityBuilder methods
	eb = eb.In("Group")
	eb = eb.Attributes(Attr("name", String()))
	eb = eb.Tags(String())

	// EntityBuilder can chain to new Entity
	eb2 := eb.Entity("Group")
	_ = eb2

	// EntityBuilder can chain to Action
	ab := eb.Entity("Document").Action("read")

	// ActionBuilder methods
	ab = ab.In("allActions")
	ab = ab.Principals("User")
	ab = ab.Resources("Document")
	ab = ab.Context(Record(Attr("reason", String())))

	// ActionBuilder can chain to new Action
	ab2 := ab.Action("write")
	_ = ab2

	// ActionBuilder can chain to Entity
	eb3 := ab.Entity("Folder")
	_ = eb3

	// Both can get back to Schema
	schema := ab.Schema()
	if schema == nil {
		t.Error("Schema() should return non-nil schema")
	}

	// Both builders have Resolve convenience method
	_, err := ab.Resolve()
	if err != nil {
		t.Fatalf("Resolve via ActionBuilder failed: %v", err)
	}
}

func TestEntityBuilderCommonType(t *testing.T) {
	// Test that EntityBuilder.CommonType works
	s := NewSchema().
		Namespace("App").
		Entity("User").
		CommonType("Address", Record(
			Attr("street", String()),
		)).
		Entity("Company").Attributes(
			Attr("address", EntityRef("Address")),
		)

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	companyType := resolved.EntityType(types.EntityType("App::Company"))
	if companyType == nil {
		t.Fatal("Company entity type not found")
	}
}

func TestActionBuilderCommonType(t *testing.T) {
	// Test that ActionBuilder.CommonType works
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("User").Resources("User").
		CommonType("Metadata", Record(
			Attr("timestamp", Long()),
		)).
		Action("write").Principals("User").Resources("User")

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	writeAction := resolved.Action(types.NewEntityUID("App::Action", "write"))
	if writeAction == nil {
		t.Fatal("write action not found")
	}
}

func TestSortedEntityTypes(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("Zebra").
		Entity("Apple").
		Entity("Mango")

	resolved := s.MustResolve()

	sorted := resolved.SortedEntityTypes()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 entity types, got %d", len(sorted))
	}

	// Should be sorted alphabetically
	expected := []types.EntityType{"App::Apple", "App::Mango", "App::Zebra"}
	for i, name := range sorted {
		if name != expected[i] {
			t.Errorf("sorted[%d] = %s, expected %s", i, name, expected[i])
		}
	}
}

func TestSortedActions(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("write").Principals("User").Resources("User").
		Action("delete").Principals("User").Resources("User").
		Action("read").Principals("User").Resources("User")

	resolved := s.MustResolve()

	sorted := resolved.SortedActions()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(sorted))
	}

	// Should be sorted by ID (all have same type App::Action)
	expectedIDs := []string{"delete", "read", "write"}
	for i, uid := range sorted {
		if string(uid.ID) != expectedIDs[i] {
			t.Errorf("sorted[%d].ID = %s, expected %s", i, uid.ID, expectedIDs[i])
		}
	}
}

func TestSortedAttributeNames(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").Attributes(
			Attr("zebra", String()),
			Attr("apple", String()),
			Attr("mango", Long()),
		)

	resolved := s.MustResolve()
	userType := resolved.EntityType(types.EntityType("App::User"))
	attrs := userType.Attributes()

	sorted := attrs.SortedAttributeNames()
	if len(sorted) != 3 {
		t.Fatalf("expected 3 attributes, got %d", len(sorted))
	}

	expected := []string{"apple", "mango", "zebra"}
	for i, name := range sorted {
		if name != expected[i] {
			t.Errorf("sorted[%d] = %s, expected %s", i, name, expected[i])
		}
	}
}

func TestAsTypeHelpers(t *testing.T) {
	s := NewSchema().
		Namespace("App").
		Entity("User").Attributes(
			Attr("name", String()),
			Attr("tags", Set(String())),
			Attr("owner", EntityRef("User")),
			Attr("metadata", Record(
				Attr("created", Long()),
			)),
			Attr("ip", IPAddr()),
		)

	resolved := s.MustResolve()
	userType := resolved.EntityType(types.EntityType("App::User"))
	attrs := userType.Attributes()

	// Test AsPrimitive
	nameAttr := attrs.Attribute("name")
	if p, ok := AsPrimitive(nameAttr.Type()); !ok {
		t.Error("name should be a primitive type")
	} else if p.Name() != "String" {
		t.Errorf("name type = %s, expected String", p.Name())
	}

	// Test AsSet
	tagsAttr := attrs.Attribute("tags")
	if s, ok := AsSet(tagsAttr.Type()); !ok {
		t.Error("tags should be a set type")
	} else {
		if elem, ok := AsPrimitive(s.Element()); !ok || elem.Name() != "String" {
			t.Error("tags element should be String")
		}
	}

	// Test AsEntityRef
	ownerAttr := attrs.Attribute("owner")
	if e, ok := AsEntityRef(ownerAttr.Type()); !ok {
		t.Error("owner should be an entity ref type")
	} else if e.Name() != "App::User" {
		t.Errorf("owner ref = %s, expected App::User", e.Name())
	}

	// Test AsRecord
	metaAttr := attrs.Attribute("metadata")
	if r, ok := AsRecord(metaAttr.Type()); !ok {
		t.Error("metadata should be a record type")
	} else {
		if r.Attribute("created") == nil {
			t.Error("metadata should have created attribute")
		}
	}

	// Test AsExtension
	ipAttr := attrs.Attribute("ip")
	if e, ok := AsExtension(ipAttr.Type()); !ok {
		t.Error("ip should be an extension type")
	} else if e.Name() != "ipaddr" {
		t.Errorf("ip extension = %s, expected ipaddr", e.Name())
	}
}

func TestContextRequiresRecord(t *testing.T) {
	// This test verifies that Context() only accepts RecordType at compile time.
	// The fact that this compiles proves the type constraint works.
	s := NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").
			Principals("User").
			Resources("User").
			Context(Record(
				Attr("reason", String()),
			))

	resolved := s.MustResolve()
	action := resolved.Action(types.NewEntityUID("App::Action", "read"))
	if action.Context() == nil {
		t.Error("action should have context")
	}
}
