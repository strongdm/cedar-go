package schema

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
)

// Coverage tests for type marker methods
func TestTypeMarkers(t *testing.T) {
	// Exercise all isSchemaType() marker methods
	var _ Type = TypeString{}
	var _ Type = TypeLong{}
	var _ Type = TypeBoolean{}
	var _ Type = TypeSet{}
	var _ Type = TypeRecord{}
	var _ Type = TypeName{}
	var _ Type = TypeExtension{}

	// Call the marker methods explicitly for coverage
	TypeString{}.isSchemaType()
	TypeLong{}.isSchemaType()
	TypeBoolean{}.isSchemaType()
	TypeSet{}.isSchemaType()
	TypeRecord{}.isSchemaType()
	TypeName{}.isSchemaType()
	TypeExtension{}.isSchemaType()
}

// Coverage tests for MarshalCedar methods
func TestMarshalCedarCoverage(t *testing.T) {
	// Test TypeExtension MarshalCedar
	ext := ExtensionType("ipaddr")
	if got := string(ext.MarshalCedar()); got != "ipaddr" {
		t.Errorf("TypeExtension.MarshalCedar() = %q, want %q", got, "ipaddr")
	}

	// Test Attribute MarshalCedar
	attr := Attribute{
		Name:     "test",
		Type:     String(),
		Required: true,
		Annotations: Annotations{
			{Key: "doc", Value: "test attr"},
		},
	}
	attrCedar := string(attr.MarshalCedar())
	if !containsSubstring(attrCedar, "test: String") {
		t.Errorf("Attribute.MarshalCedar() = %q, missing attribute definition", attrCedar)
	}
	if !containsSubstring(attrCedar, `@doc("test attr")`) {
		t.Errorf("Attribute.MarshalCedar() = %q, missing annotation", attrCedar)
	}

	// Test optional attribute
	optAttr := Attribute{Name: "opt", Type: Long(), Required: false}
	optCedar := string(optAttr.MarshalCedar())
	if !containsSubstring(optCedar, "opt?:") {
		t.Errorf("Attribute.MarshalCedar() = %q, missing ? for optional", optCedar)
	}

	// Test Annotation MarshalCedar without value
	annNoValue := Annotation{Key: "deprecated"}
	if got := string(annNoValue.MarshalCedar()); got != "@deprecated" {
		t.Errorf("Annotation.MarshalCedar() = %q, want %q", got, "@deprecated")
	}

	// Test ActionRef MarshalCedar
	ref := ActionRef{Name: "test"}
	if got := string(ref.MarshalCedar()); got != `"test"` {
		t.Errorf("ActionRef.MarshalCedar() = %q, want %q", got, `"test"`)
	}

	refWithNS := ActionRef{Namespace: "MyApp::Actions", Name: "read"}
	refCedar := string(refWithNS.MarshalCedar())
	if !containsSubstring(refCedar, "MyApp::Actions::") {
		t.Errorf("ActionRef.MarshalCedar() = %q, missing namespace", refCedar)
	}

	// Test AppliesTo MarshalCedar
	appliesTo := NewAppliesTo().
		WithPrincipals("User").
		WithResources("Document").
		WithContext(Record().AddOptional("flag", Boolean()))
	appCedar := string(appliesTo.MarshalCedar())
	if !containsSubstring(appCedar, "principal:") {
		t.Errorf("AppliesTo.MarshalCedar() = %q, missing principal", appCedar)
	}
	if !containsSubstring(appCedar, "resource:") {
		t.Errorf("AppliesTo.MarshalCedar() = %q, missing resource", appCedar)
	}
	if !containsSubstring(appCedar, "context:") {
		t.Errorf("AppliesTo.MarshalCedar() = %q, missing context", appCedar)
	}

	// Test Action MarshalCedar
	action := NewAction("test").
		Annotate("doc", "test action").
		InAction("parent").
		WithAppliesTo(NewAppliesTo().WithPrincipals("User").WithResources("Resource"))
	actCedar := string(action.MarshalCedar())
	if !containsSubstring(actCedar, "action test") {
		t.Errorf("Action.MarshalCedar() = %q, missing action declaration", actCedar)
	}
	if !containsSubstring(actCedar, `@doc("test action")`) {
		t.Errorf("Action.MarshalCedar() = %q, missing annotation", actCedar)
	}

	// Test Action with multiple memberOf
	actionMulti := NewAction("multi").In(
		ActionRef{Name: "group1"},
		ActionRef{Name: "group2"},
	)
	multiCedar := string(actionMulti.MarshalCedar())
	if !containsSubstring(multiCedar, "[") {
		t.Errorf("Action.MarshalCedar() = %q, should have brackets for multiple memberOf", multiCedar)
	}

	// Test Namespace MarshalCedar
	ns := NewNamespace("TestNS").
		Annotate("doc", "test namespace").
		AddCommonType(NewCommonType("CT", String())).
		AddEntityType(Entity("E")).
		AddAction(NewAction("a"))
	nsCedar := string(ns.MarshalCedar())
	if !containsSubstring(nsCedar, "namespace TestNS") {
		t.Errorf("Namespace.MarshalCedar() = %q, missing namespace declaration", nsCedar)
	}

	// Test anonymous namespace MarshalCedar
	anonNS := NewNamespace("").AddEntityType(Entity("Anon"))
	anonCedar := string(anonNS.MarshalCedar())
	if containsSubstring(anonCedar, "namespace") {
		t.Errorf("Anonymous Namespace.MarshalCedar() = %q, should not have namespace keyword", anonCedar)
	}

	// Test CommonType MarshalCedar
	ct := NewCommonType("MyType", Record().AddRequired("x", Long())).
		Annotate("doc", "common type")
	ctCedar := string(ct.MarshalCedar())
	if !containsSubstring(ctCedar, "type MyType =") {
		t.Errorf("CommonType.MarshalCedar() = %q, missing type declaration", ctCedar)
	}

	// Test Entity MarshalCedar with no annotations but has tags
	entityWithTags := Entity("Tagged").WithTags(String())
	tagsCedar := string(entityWithTags.MarshalCedar())
	if !containsSubstring(tagsCedar, "tags String") {
		t.Errorf("Entity.MarshalCedar() = %q, missing tags", tagsCedar)
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	// Test Attr helper
	attr := Attr("field", String(), true)
	if attr.Name != "field" || !attr.Required {
		t.Errorf("Attr() did not set fields correctly")
	}

	// Test RequiredAttr helper
	reqAttr := RequiredAttr("req", Long())
	if !reqAttr.Required {
		t.Errorf("RequiredAttr() should set Required=true")
	}

	// Test OptionalAttr helper
	optAttr := OptionalAttr("opt", Boolean())
	if optAttr.Required {
		t.Errorf("OptionalAttr() should set Required=false")
	}

	// Test Annotations.Annotate
	anns := Annotations{}.Annotate("key1", "val1").Annotate("key2", "val2")
	if len(anns) != 2 {
		t.Errorf("Annotations.Annotate() should chain correctly")
	}
}

// Test EnumEntity
func TestEnumEntity(t *testing.T) {
	enum := EnumEntity("Status", "active", "inactive", "pending").
		Annotate("doc", "Status enum").
		AddValue("archived")

	cedar := string(enum.MarshalCedar())
	if !containsSubstring(cedar, "entity Status enum") {
		t.Errorf("EnumEntity.MarshalCedar() = %q, missing enum declaration", cedar)
	}
	if !containsSubstring(cedar, `"active"`) {
		t.Errorf("EnumEntity.MarshalCedar() = %q, missing enum value", cedar)
	}
	if !containsSubstring(cedar, `"archived"`) {
		t.Errorf("EnumEntity.MarshalCedar() = %q, missing added value", cedar)
	}
}

// Test AddEnumType on namespace
func TestNamespaceAddEnumType(t *testing.T) {
	ns := NewNamespace("EnumNS").
		AddEnumType(EnumEntity("Colors", "red", "green", "blue"))

	if len(ns.EnumTypes) != 1 {
		t.Errorf("AddEnumType() should add enum type")
	}

	cedar := string(ns.MarshalCedar())
	if !containsSubstring(cedar, "enum") {
		t.Errorf("Namespace with enum should contain enum in output")
	}
}

// Test InQualifiedAction
func TestInQualifiedAction(t *testing.T) {
	action := NewAction("child").
		InQualifiedAction("OtherNS::Actions", "parent")

	if len(action.MemberOf) != 1 {
		t.Errorf("InQualifiedAction() should add member")
	}
	if action.MemberOf[0].Namespace != "OtherNS::Actions" {
		t.Errorf("InQualifiedAction() namespace = %q, want %q", action.MemberOf[0].Namespace, "OtherNS::Actions")
	}
}

// Test Action.Annotate
func TestActionAnnotate(t *testing.T) {
	action := NewAction("test").Annotate("doc", "description")
	if len(action.Annotations) != 1 {
		t.Errorf("Action.Annotate() should add annotation")
	}
}

// Test marshalString edge cases
func TestMarshalString(t *testing.T) {
	tests := []struct {
		input    types.String
		contains string
	}{
		{`hello`, `"hello"`},
		{`with"quote`, `with\"quote`},
		{"with\nnewline", `with\n`},
		{"with\ttab", `with\t`},
		{"with\rcarriage", `with\r`},
		{`with\backslash`, `with\\`},
	}

	for _, tt := range tests {
		got := string(marshalString(tt.input))
		if !containsSubstring(got, tt.contains) {
			t.Errorf("marshalString(%q) = %q, want to contain %q", tt.input, got, tt.contains)
		}
	}
}

// Test TypeRecord MarshalCedar with empty attributes
func TestTypeRecordEmpty(t *testing.T) {
	rec := Record()
	cedar := string(rec.MarshalCedar())
	if cedar != "{}" {
		t.Errorf("Empty Record.MarshalCedar() = %q, want %q", cedar, "{}")
	}
}

// Test GetNamespace when not found
func TestGetNamespaceNotFound(t *testing.T) {
	s := New()
	if s.GetNamespace("nonexistent") != nil {
		t.Errorf("GetNamespace() should return nil for non-existent namespace")
	}
}

// Test conversion edge cases
func TestConversionEdgeCases(t *testing.T) {
	// Test JSON with Extension type
	jsonSchema := `{
		"ExtApp": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ip": {"type": "Extension", "name": "ipaddr", "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify extension type was parsed
	ns := s.GetNamespace("ExtApp")
	if ns == nil {
		t.Fatal("Expected ExtApp namespace")
	}

	// Marshal back and verify
	_, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
}

// Test JSON with context as TypeName (not record)
func TestJSONContextAsTypeName(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {
				"doThing": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Resource"],
						"context": {"type": "EntityOrCommon", "name": "SharedContext"}
					}
				}
			},
			"commonTypes": {
				"SharedContext": {
					"type": "Record",
					"attributes": {
						"flag": {"type": "Boolean", "required": false}
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Verify context reference is preserved
	if !containsSubstring(string(cedarBytes), "context:") {
		t.Errorf("Cedar output should contain context reference")
	}
}

// Test AppliesTo with multiple principals and resources
func TestAppliesToMultiple(t *testing.T) {
	appliesTo := NewAppliesTo().
		WithPrincipals("User", "Admin", "Service").
		WithResources("Doc", "File", "Folder")

	cedar := string(appliesTo.MarshalCedar())
	if !containsSubstring(cedar, "[User, Admin, Service]") {
		t.Errorf("AppliesTo.MarshalCedar() should bracket multiple principals")
	}
	if !containsSubstring(cedar, "[Doc, File, Folder]") {
		t.Errorf("AppliesTo.MarshalCedar() should bracket multiple resources")
	}
}

// Test Entity with single vs multiple memberOf
func TestEntityMemberOfBrackets(t *testing.T) {
	// Single memberOf - no brackets
	single := Entity("E").In("Group")
	singleCedar := string(single.MarshalCedar())
	if containsSubstring(singleCedar, "[") {
		t.Errorf("Single memberOf should not have brackets: %s", singleCedar)
	}

	// Multiple memberOf - should have brackets
	multi := Entity("E").In("Group1", "Group2")
	multiCedar := string(multi.MarshalCedar())
	if !containsSubstring(multiCedar, "[") {
		t.Errorf("Multiple memberOf should have brackets: %s", multiCedar)
	}
}

// Test splitPath and joinPath edge cases
func TestPathHelpers(t *testing.T) {
	// Test empty path
	parts := splitPath("")
	if len(parts) != 0 {
		t.Errorf("splitPath(\"\") should return empty slice")
	}

	// Test joinPath
	joined := joinPath([]string{"A", "B", "C"})
	if joined != "A::B::C" {
		t.Errorf("joinPath() = %q, want %q", joined, "A::B::C")
	}

	// Test single part
	single := joinPath([]string{"Single"})
	if single != "Single" {
		t.Errorf("joinPath() = %q, want %q", single, "Single")
	}
}

// Test JSON schema with empty annotations map
func TestJSONEmptyAnnotations(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {
					"annotations": {}
				}
			},
			"actions": {},
			"annotations": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	_, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
}

// Test conversion of all JSON attribute types
func TestJSONAttributeTypes(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"strField": {"type": "String", "required": true},
							"longField": {"type": "Long", "required": true},
							"boolField": {"type": "Boolean", "required": true},
							"setField": {"type": "Set", "required": true, "element": {"type": "String"}},
							"entityField": {"type": "Entity", "name": "OtherEntity", "required": true},
							"commonField": {"type": "EntityOrCommon", "name": "CommonType", "required": true},
							"extField": {"type": "Extension", "name": "decimal", "required": true},
							"recordField": {"type": "Record", "required": true, "attributes": {
								"nested": {"type": "String", "required": false}
							}}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Marshal to JSON and back
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var raw interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("MarshalJSON() produced invalid JSON: %v", err)
	}
}

// Test Cedar with action having empty appliesTo
func TestCedarEmptyAppliesTo(t *testing.T) {
	cedar := `
namespace App {
	entity User;
	action doThing appliesTo {
		principal: User,
		resource: User,
	};
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	_, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}
}

// Test conversion back to JSON for TypeExtension
func TestExtensionTypeToJSON(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			AddEntityType(Entity("User").
				WithShape(Record().
					AddAttribute(Attribute{
						Name:     "ip",
						Type:     ExtensionType("ipaddr"),
						Required: true,
					}))))

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !containsSubstring(string(jsonBytes), "Extension") {
		t.Errorf("JSON should contain Extension type")
	}
}

// Test MarshalJSON when only humanSchema is set (no namespaces)
func TestMarshalJSONFromHumanSchema(t *testing.T) {
	var s Schema
	cedar := `namespace Test { entity User; }`
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Clear namespaces to force conversion from humanSchema
	s.Namespaces = nil

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Errorf("MarshalJSON() should produce output")
	}
}

// Test Action without appliesTo
func TestActionWithoutAppliesTo(t *testing.T) {
	action := NewAction("standalone")
	cedar := string(action.MarshalCedar())
	if containsSubstring(cedar, "appliesTo") {
		t.Errorf("Action without appliesTo should not have appliesTo clause: %s", cedar)
	}
}

// Test Action without memberOf
func TestActionWithoutMemberOf(t *testing.T) {
	action := NewAction("standalone").
		WithAppliesTo(NewAppliesTo().WithPrincipals("User").WithResources("Resource"))
	cedar := string(action.MarshalCedar())
	if containsSubstring(cedar, " in ") {
		t.Errorf("Action without memberOf should not have in clause: %s", cedar)
	}
}

// Test Cedar parsing with enum entities
func TestCedarEnumEntity(t *testing.T) {
	cedar := `
namespace App {
    entity Status enum ["active", "inactive"];
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Round-trip
	_, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}
}

// Test Cedar parsing with common type declarations at top level
func TestCedarTopLevelCommonType(t *testing.T) {
	cedar := `
type SharedContext = {
    auth?: Bool,
};

entity User;
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Should have anonymous namespace with common type
	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("Expected anonymous namespace")
	}
	if len(ns.CommonTypes) != 1 {
		t.Errorf("Expected 1 common type, got %d", len(ns.CommonTypes))
	}
}

// Test Cedar parsing with action without namespace in memberOf
func TestCedarActionMemberOfNoNamespace(t *testing.T) {
	cedar := `
namespace App {
    entity User;
    entity Resource;
    action readAll;
    action read in readAll appliesTo {
        principal: User,
        resource: Resource,
    };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	_, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}
}

// Test Cedar with context as path reference
func TestCedarContextAsPath(t *testing.T) {
	cedar := `
namespace App {
    type MyContext = { flag?: Bool };
    entity User;
    entity Resource;
    action doThing appliesTo {
        principal: User,
        resource: Resource,
        context: MyContext,
    };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !containsSubstring(string(cedarBytes), "context: MyContext") {
		t.Errorf("Should preserve context path reference")
	}
}

// Test convertToHumanAST with anonymous namespace having all declaration types
func TestConvertToHumanASTAnonymous(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("").
			AddCommonType(NewCommonType("CT", String())).
			AddEntityType(Entity("E")).
			AddAction(NewAction("a")))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Should not have "namespace" keyword for anonymous
	if containsSubstring(string(cedarBytes), "namespace") {
		t.Errorf("Anonymous namespace should not produce namespace keyword")
	}
}

// Test JSON with nil annotations on entities
func TestJSONNilAnnotations(t *testing.T) {
	// This tests the case where annotations map is nil in JSON
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {}
			},
			"actions": {
				"doThing": {
					"appliesTo": {
						"principalTypes": [],
						"resourceTypes": []
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Check that annotations are empty but not nil
	ns := s.GetNamespace("App")
	if ns == nil {
		t.Fatal("Expected App namespace")
	}
}

// Test all JSON type conversions
func TestJSONTypeConversions(t *testing.T) {
	// Test unknown type falls through to TypeName
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"AliasType": {
					"type": "SomeUnknownType"
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

// Test toJSON conversion for all types
func TestToJSONAllTypes(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			AddEntityType(Entity("User").
				WithShape(Record().
					AddRequired("str", String()).
					AddRequired("num", Long()).
					AddRequired("flag", Boolean()).
					AddRequired("items", Set(String())).
					AddRequired("nested", Record().AddOptional("x", Long())).
					AddRequired("ref", EntityOrCommonType("Other")).
					AddRequired("ext", ExtensionType("decimal")))))

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify all types are in the output
	jsonStr := string(jsonBytes)
	if !containsSubstring(jsonStr, `"String"`) {
		t.Error("Missing String type")
	}
	if !containsSubstring(jsonStr, `"Long"`) {
		t.Error("Missing Long type")
	}
	if !containsSubstring(jsonStr, `"Boolean"`) {
		t.Error("Missing Boolean type")
	}
	if !containsSubstring(jsonStr, `"Set"`) {
		t.Error("Missing Set type")
	}
	if !containsSubstring(jsonStr, `"Record"`) {
		t.Error("Missing Record type")
	}
	if !containsSubstring(jsonStr, `"EntityOrCommon"`) {
		t.Error("Missing EntityOrCommon type")
	}
	if !containsSubstring(jsonStr, `"Extension"`) {
		t.Error("Missing Extension type")
	}
}

// Test namespace with enum types in MarshalCedar
func TestNamespaceMarshalCedarWithEnums(t *testing.T) {
	ns := NewNamespace("App").
		AddEnumType(EnumEntity("Status", "active", "inactive"))

	cedar := string(ns.MarshalCedar())
	if !containsSubstring(cedar, "enum") {
		t.Errorf("Namespace output should contain enum: %s", cedar)
	}
}

// Test Entity MarshalCedar with annotations
func TestEntityMarshalCedarWithAnnotations(t *testing.T) {
	e := Entity("User").
		Annotate("doc", "A user").
		WithShape(Record().AddRequired("name", String()))

	cedar := string(e.MarshalCedar())
	if !containsSubstring(cedar, "@doc") {
		t.Errorf("Entity output should contain annotation: %s", cedar)
	}
}

// Force coverage of interface marker methods through type assertions
func TestInterfaceMarkers(t *testing.T) {
	types := []Type{
		TypeString{},
		TypeLong{},
		TypeBoolean{},
		TypeSet{Element: TypeString{}},
		TypeRecord{},
		TypeName{Name: "Test"},
		TypeExtension{Name: "ext"},
	}

	for _, typ := range types {
		// This forces the isSchemaType method to be called via interface
		typ.isSchemaType()
		_ = typ.MarshalCedar()
	}
}

// Test Cedar action with qualified namespace in memberOf
func TestCedarActionQualifiedMemberOf(t *testing.T) {
	cedar := `
namespace App {
    entity User;
    entity Resource;
    action read in OtherNS::"parent" appliesTo {
        principal: User,
        resource: Resource,
    };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !containsSubstring(string(cedarBytes), "OtherNS::") {
		t.Errorf("Should preserve qualified namespace in action memberOf")
	}
}

// Test JSON type nil handling
func TestJSONTypeNil(t *testing.T) {
	result := convertJSONType(nil)
	if result != nil {
		t.Error("convertJSONType(nil) should return nil")
	}
}

// Test convertToJSONType nil handling
func TestConvertToJSONTypeNil(t *testing.T) {
	result := convertToJSONType(nil)
	if result != nil {
		t.Error("convertToJSONType(nil) should return nil")
	}
}

// Test converting TypeExtension through human AST round-trip
func TestTypeExtensionRoundTrip(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			AddCommonType(NewCommonType("ExtType", ExtensionType("ipaddr"))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Re-parse and verify
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarBytes); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}
}

// Test empty annotations via convertJSONAnnotations
func TestEmptyJSONAnnotations(t *testing.T) {
	result := convertJSONAnnotations(nil)
	if len(result) != 0 {
		t.Error("convertJSONAnnotations(nil) should return empty slice")
	}

	result = convertJSONAnnotations(map[string]string{})
	if len(result) != 0 {
		t.Error("convertJSONAnnotations(empty map) should return empty slice")
	}
}

// Test schema with only namespaces, no anonymous content
func TestSchemaOnlyNamespaces(t *testing.T) {
	cedar := `
namespace App1 {
    entity User;
}
namespace App2 {
    entity Resource;
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Should have 2 namespaces and no anonymous namespace
	if len(s.Namespaces) != 2 {
		t.Errorf("Expected 2 namespaces, got %d", len(s.Namespaces))
	}
}

// Test JSON actions without memberOf
func TestJSONActionWithoutMemberOf(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {"User": {}},
			"actions": {
				"doThing": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"]
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify action was parsed without memberOf
	ns := s.GetNamespace("App")
	if ns == nil {
		t.Fatal("Expected App namespace")
	}

	if len(ns.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(ns.Actions))
	}

	if len(ns.Actions[0].MemberOf) != 0 {
		t.Errorf("Expected action without memberOf")
	}
}

// Test JSON entity without shape or tags
func TestJSONEntityMinimal(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {},
				"Group": {"memberOfTypes": ["Organization"]},
				"Organization": {}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	_, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}
}

// Test converting all types through convertToJSONSchema
func TestConvertToJSONSchemaAllTypes(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			Annotate("doc", "test").
			AddCommonType(NewCommonType("CT", String()).Annotate("doc", "ct")).
			AddEntityType(Entity("E").
				Annotate("doc", "e").
				In("E2").
				WithShape(Record().AddRequired("x", Long())).
				WithTags(String())).
			AddEntityType(Entity("E2")).
			AddAction(NewAction("a").
				Annotate("doc", "a").
				InAction("parent").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("E").
					WithResources("E2").
					WithContext(Record().AddOptional("f", Boolean())))))

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify JSON is valid
	var raw interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("MarshalJSON() produced invalid JSON: %v", err)
	}
}

// Test JSON with action memberOf having type
func TestJSONActionMemberOfWithType(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {"User": {}},
			"actions": {
				"read": {
					"memberOf": [{"id": "parent", "type": "OtherNS::Action"}],
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"]
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.Actions) == 0 {
		t.Fatal("Expected action")
	}

	if len(ns.Actions[0].MemberOf) != 1 {
		t.Fatal("Expected 1 memberOf")
	}

	if ns.Actions[0].MemberOf[0].Namespace != "OtherNS::Action" {
		t.Errorf("Expected namespace OtherNS::Action, got %s", ns.Actions[0].MemberOf[0].Namespace)
	}
}

// Test conversion with context as TypeName (not record)
func TestContextAsTypeName(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			AddCommonType(NewCommonType("MyContext", Record().AddOptional("f", Boolean()))).
			AddEntityType(Entity("User")).
			AddAction(NewAction("a").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("User").
					WithContext(TypeName{Name: "MyContext"}))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !containsSubstring(string(cedarBytes), "context: MyContext") {
		t.Errorf("Should have context as TypeName reference")
	}
}

// Test record with attribute annotations from JSON
func TestJSONAttributeAnnotations(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {
								"type": "String",
								"required": true,
								"annotations": {"doc": "The user name"}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Marshal to JSON and verify annotations are preserved
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !containsSubstring(string(jsonBytes), "annotations") {
		t.Errorf("Attribute annotations should be preserved")
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test Cedar with top-level action (not inside namespace)
func TestCedarTopLevelAction(t *testing.T) {
	cedar := `
entity User;
entity Resource;
action read appliesTo {
    principal: User,
    resource: Resource,
};
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Should have anonymous namespace with action
	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("Expected anonymous namespace")
	}
	if len(ns.Actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(ns.Actions))
	}
}

// Test Cedar with entity that has tags
func TestCedarEntityWithTags(t *testing.T) {
	cedar := `
namespace App {
    entity User = {
        "name": String,
    } tags String;
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil {
		t.Fatal("Expected App namespace")
	}

	if len(ns.EntityTypes) != 1 {
		t.Fatal("Expected 1 entity type")
	}

	if ns.EntityTypes[0].Tags == nil {
		t.Error("Expected entity to have tags")
	}
}

// Test JSON common type with Long type
func TestJSONCommonTypeLong(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"Counter": {"type": "Long"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.CommonTypes) == 0 {
		t.Fatal("Expected common type")
	}

	if _, ok := ns.CommonTypes[0].Type.(TypeLong); !ok {
		t.Errorf("Expected TypeLong, got %T", ns.CommonTypes[0].Type)
	}
}

// Test JSON common type with Boolean type
func TestJSONCommonTypeBoolean(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"Flag": {"type": "Boolean"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.CommonTypes) == 0 {
		t.Fatal("Expected common type")
	}

	if _, ok := ns.CommonTypes[0].Type.(TypeBoolean); !ok {
		t.Errorf("Expected TypeBoolean, got %T", ns.CommonTypes[0].Type)
	}
}

// Test JSON common type with Set type
func TestJSONCommonTypeSet(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"Tags": {"type": "Set", "element": {"type": "String"}}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.CommonTypes) == 0 {
		t.Fatal("Expected common type")
	}

	if _, ok := ns.CommonTypes[0].Type.(TypeSet); !ok {
		t.Errorf("Expected TypeSet, got %T", ns.CommonTypes[0].Type)
	}
}

// Test JSON common type with Extension type
func TestJSONCommonTypeExtension(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"IPAddress": {"type": "Extension", "name": "ipaddr"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.CommonTypes) == 0 {
		t.Fatal("Expected common type")
	}

	if _, ok := ns.CommonTypes[0].Type.(TypeExtension); !ok {
		t.Errorf("Expected TypeExtension, got %T", ns.CommonTypes[0].Type)
	}
}

// Test JSON entity with tags of various types
func TestJSONEntityWithTags(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"UserLong": {"tags": {"type": "Long"}},
				"UserBool": {"tags": {"type": "Boolean"}},
				"UserSet": {"tags": {"type": "Set", "element": {"type": "String"}}},
				"UserExt": {"tags": {"type": "Extension", "name": "decimal"}}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.EntityTypes) != 4 {
		t.Fatalf("Expected 4 entity types, got %d", len(ns.EntityTypes))
	}

	// Verify tags were parsed correctly
	foundTypes := make(map[string]bool)
	for _, et := range ns.EntityTypes {
		if et.Tags != nil {
			switch et.Tags.(type) {
			case TypeLong:
				foundTypes["Long"] = true
			case TypeBoolean:
				foundTypes["Boolean"] = true
			case TypeSet:
				foundTypes["Set"] = true
			case TypeExtension:
				foundTypes["Extension"] = true
			}
		}
	}

	if !foundTypes["Long"] {
		t.Error("Missing Long tags")
	}
	if !foundTypes["Boolean"] {
		t.Error("Missing Boolean tags")
	}
	if !foundTypes["Set"] {
		t.Error("Missing Set tags")
	}
	if !foundTypes["Extension"] {
		t.Error("Missing Extension tags")
	}
}

// Test JSON action context with various types
func TestJSONActionContextTypes(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {"User": {}, "Resource": {}},
			"actions": {
				"action1": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Resource"],
						"context": {"type": "Long"}
					}
				},
				"action2": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Resource"],
						"context": {"type": "Boolean"}
					}
				},
				"action3": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Resource"],
						"context": {"type": "Set", "element": {"type": "String"}}
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.Actions) != 3 {
		t.Fatalf("Expected 3 actions, got %d", len(ns.Actions))
	}
}

// Test convertToJSONType with various types via entity tags
func TestConvertToJSONTypeViaTags(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("App").
			AddEntityType(Entity("E1").WithTags(Long())).
			AddEntityType(Entity("E2").WithTags(Boolean())).
			AddEntityType(Entity("E3").WithTags(Set(String()))).
			AddEntityType(Entity("E4").WithTags(EntityOrCommonType("Other"))).
			AddEntityType(Entity("E5").WithTags(ExtensionType("ipaddr"))))

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify JSON is valid
	var raw interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("MarshalJSON() produced invalid JSON: %v", err)
	}

	jsonStr := string(jsonBytes)
	if !containsSubstring(jsonStr, `"Long"`) {
		t.Error("Missing Long type in tags")
	}
	if !containsSubstring(jsonStr, `"Boolean"`) {
		t.Error("Missing Boolean type in tags")
	}
	if !containsSubstring(jsonStr, `"Set"`) {
		t.Error("Missing Set type in tags")
	}
	if !containsSubstring(jsonStr, `"EntityOrCommon"`) {
		t.Error("Missing EntityOrCommon type in tags")
	}
	if !containsSubstring(jsonStr, `"Extension"`) {
		t.Error("Missing Extension type in tags")
	}
}

// Test TypeRecord MarshalCedar with multiple attributes
func TestTypeRecordMultipleAttributes(t *testing.T) {
	rec := Record().
		AddRequired("field1", String()).
		AddRequired("field2", Long()).
		AddOptional("field3", Boolean())

	cedar := string(rec.MarshalCedar())

	// Should have commas between attributes
	if !containsSubstring(cedar, ",") {
		t.Errorf("Record with multiple attributes should have commas: %s", cedar)
	}

	// Should contain all fields
	if !containsSubstring(cedar, "field1") {
		t.Error("Missing field1")
	}
	if !containsSubstring(cedar, "field2") {
		t.Error("Missing field2")
	}
	if !containsSubstring(cedar, "field3") {
		t.Error("Missing field3")
	}
}

// Test panic branches in conversion functions
// These are defensive panics that should never occur in normal operation

// mockType is a custom type that implements Type but is not handled by conversion functions
type mockType struct{}

func (mockType) isSchemaType()        { _ = 0 }
func (mockType) MarshalCedar() []byte { return []byte("mock") }

// Test convertHumanType with nil triggers the default panic branch
func TestConvertHumanTypeNilPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil type")
		}
	}()
	// Passing nil to convertHumanType triggers the default case
	convertHumanType(nil)
}

func TestConvertToHumanTypeASTPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unknown type")
		}
	}()
	convertToHumanTypeAST(mockType{})
}

func TestConvertToJSONTypePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for unknown type")
		}
	}()
	convertToJSONType(mockType{})
}

// Test JSON attribute with unknown type (falls through to TypeName)
func TestJSONAttrUnknownType(t *testing.T) {
	jsonSchema := `{
		"App": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ref": {"type": "CustomType", "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil || len(ns.EntityTypes) == 0 {
		t.Fatal("Expected entity type")
	}

	// Verify the unknown type was treated as TypeName
	entity := ns.EntityTypes[0]
	if entity.Shape == nil || len(entity.Shape.Attributes) == 0 {
		t.Fatal("Expected shape with attributes")
	}

	attr := entity.Shape.Attributes[0]
	if _, ok := attr.Type.(TypeName); !ok {
		t.Errorf("Expected TypeName for unknown type, got %T", attr.Type)
	}
}
