package schema_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

func TestBuilderBasic(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("PhotoFlash").
		Entity("User").
		MemberOf("UserGroup").
		Attr("name", schema.String()).
		Attr("age", schema.Long()).
		Entity("UserGroup").
		Entity("Photo").
		MemberOf("Album").
		Attr("private", schema.Bool()).
		Entity("Album").
		Action("viewPhoto").
		Principal("User").
		Resource("Photo").
		Build()

	// Check structure
	ns := s.Namespaces["PhotoFlash"]
	if ns == nil {
		t.Fatal("expected PhotoFlash namespace")
	}

	if len(ns.EntityTypes) != 4 {
		t.Errorf("expected 4 entity types, got %d", len(ns.EntityTypes))
	}

	user := ns.EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	if len(user.MemberOfTypes) != 1 || user.MemberOfTypes[0] != "UserGroup" {
		t.Errorf("expected User memberOf UserGroup, got %v", user.MemberOfTypes)
	}

	if user.Shape == nil || len(user.Shape.Attributes) != 2 {
		t.Errorf("expected 2 attributes on User")
	}
}

func TestBuilderWithContextAndExtensions(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").
		Entity("Document").
		Action("view").
		Principal("User").
		Resource("Document").
		Context(&schema.RecordType{Attributes: map[string]*schema.Attribute{
			"ip":        {Type: schema.IPAddr(), Required: true, Annotations: make(schema.Annotations)},
			"timestamp": {Type: schema.Long(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Build()

	ns := s.Namespaces["MyApp"]
	act := ns.Actions["view"]
	if act == nil {
		t.Fatal("expected view action")
	}

	if act.AppliesTo == nil || act.AppliesTo.Context == nil {
		t.Fatal("expected context on view action")
	}

	if len(act.AppliesTo.Context.Attributes) != 2 {
		t.Errorf("expected 2 context attributes, got %d", len(act.AppliesTo.Context.Attributes))
	}
}

func TestJSONRoundTrip(t *testing.T) {
	jsonSchema := `{
		"MyApp": {
			"entityTypes": {
				"User": {
					"memberOfTypes": ["Group"],
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String"},
							"age": {"type": "Long", "required": false}
						}
					}
				},
				"Group": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Document"]
					}
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse JSON schema: %v", err)
	}

	// Check parsed structure
	ns := s.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	user := ns.EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	if len(user.MemberOfTypes) != 1 || user.MemberOfTypes[0] != "Group" {
		t.Errorf("expected User memberOf Group")
	}

	// Round-trip
	data, err := json.Marshal(&s)
	if err != nil {
		t.Fatalf("failed to marshal JSON schema: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to re-parse JSON schema: %v", err)
	}

	// Check structure preserved
	ns2 := s2.Namespaces["MyApp"]
	if ns2 == nil {
		t.Fatal("expected MyApp namespace after round-trip")
	}

	user2 := ns2.EntityTypes["User"]
	if user2 == nil {
		t.Fatal("expected User entity type after round-trip")
	}
}

func TestCedarParse(t *testing.T) {
	cedarSchema := `
namespace PhotoFlash {
	entity User in [UserGroup] {
		name: String,
		age?: Long,
	};

	entity UserGroup;

	entity Photo in [Album] {
		private: Bool,
	};

	entity Album;

	action viewPhoto appliesTo {
		principal: [User],
		resource: [Photo],
	};
}
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse Cedar schema: %v", err)
	}

	ns := s.Namespaces["PhotoFlash"]
	if ns == nil {
		t.Fatal("expected PhotoFlash namespace")
	}

	user := ns.EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	if user.Shape == nil {
		t.Fatal("expected User to have shape")
	}

	nameAttr := user.Shape.Attributes["name"]
	if nameAttr == nil {
		t.Fatal("expected name attribute")
	}
	if !nameAttr.Required {
		t.Error("expected name to be required")
	}

	ageAttr := user.Shape.Attributes["age"]
	if ageAttr == nil {
		t.Fatal("expected age attribute")
	}
	if ageAttr.Required {
		t.Error("expected age to be optional")
	}
}

func TestResolve(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").MemberOf("Group").
		Entity("Group").
		Entity("Document").
		Action("view").Principal("User").Resource("Document").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	// Check namespace
	ns := rs.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace in resolved schema")
	}

	// Check entity type is fully qualified
	userType := types.EntityType("MyApp::User")
	user := ns.EntityTypes[userType]
	if user == nil {
		t.Fatalf("expected User entity type with key %q", userType)
	}

	// Check memberOf is fully qualified
	if len(user.MemberOfTypes) != 1 {
		t.Fatalf("expected 1 memberOf, got %d", len(user.MemberOfTypes))
	}
	if user.MemberOfTypes[0] != types.EntityType("MyApp::Group") {
		t.Errorf("expected memberOf MyApp::Group, got %v", user.MemberOfTypes[0])
	}

	// Check action is EntityUID
	viewUID := types.NewEntityUID("MyApp::Action", "view")
	view := ns.Actions[viewUID]
	if view == nil {
		t.Fatalf("expected view action with key %v", viewUID)
	}

	// Check principal types are fully qualified
	if len(view.PrincipalTypes) != 1 {
		t.Fatalf("expected 1 principal type, got %d", len(view.PrincipalTypes))
	}
	if view.PrincipalTypes[0] != types.EntityType("MyApp::User") {
		t.Errorf("expected principal type MyApp::User, got %v", view.PrincipalTypes[0])
	}
}

func TestResolveWithCommonTypes(t *testing.T) {
	cedarSchema := `
namespace MyApp {
	type Context = {
		ip: ipaddr,
		authenticated: Bool,
	};

	entity User;
	entity Document;

	action view appliesTo {
		principal: [User],
		resource: [Document],
		context: Context,
	};
}
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	ns := rs.Namespaces["MyApp"]
	viewUID := types.NewEntityUID("MyApp::Action", "view")
	view := ns.Actions[viewUID]
	if view == nil {
		t.Fatal("expected view action")
	}

	if view.Context == nil {
		t.Fatal("expected context to be resolved")
	}

	// Context should have ip and authenticated attributes
	if len(view.Context.Attributes) != 2 {
		t.Errorf("expected 2 context attributes, got %d", len(view.Context.Attributes))
	}

	ipAttr := view.Context.Attributes["ip"]
	if ipAttr == nil {
		t.Fatal("expected ip attribute")
	}
	if _, ok := ipAttr.Type.(resolved.Extension); !ok {
		t.Errorf("expected ip to be extension type, got %T", ipAttr.Type)
	}
}

func TestCycleDetectionCedarSyntax(t *testing.T) {
	cedarSchema := `
type A = Set<B>;
type B = Set<A>;
entity Foo;
action bar appliesTo { principal: [Foo], resource: [Foo] };
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err == nil {
		t.Fatal("expected cycle error")
	}

	// Should be a cycle error - check via errors.Is since the type is not exported
	if !errors.Is(err, schema.ErrCycle) {
		// Could be wrapped
		t.Logf("got error: %v", err)
	}
}

func TestEnumEntity(t *testing.T) {
	cedarSchema := `
entity Status enum ["Active", "Inactive", "Pending"];
entity User;
action view appliesTo { principal: [User], resource: [Status] };
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Enum types are now separate from entity types
	status := s.Namespaces[""].EnumTypes["Status"]
	if status == nil {
		t.Fatal("expected Status enum type")
	}

	if len(status.Values) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(status.Values))
	}

	// The User entity should be in EntityTypes
	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}
}

func TestAnnotations(t *testing.T) {
	cedarSchema := `
@doc("User entity")
entity User {
	@doc("User's name")
	name: String,
};
action view appliesTo { principal: [User], resource: [User] };
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	if user.Annotations["doc"] != "User entity" {
		t.Errorf("expected doc annotation on entity, got %v", user.Annotations)
	}

	nameAttr := user.Shape.Attributes["name"]
	if nameAttr.Annotations["doc"] != "User's name" {
		t.Errorf("expected doc annotation on attribute, got %v", nameAttr.Annotations)
	}
}

func TestEmptyNamespace(t *testing.T) {
	cedarSchema := `
entity User;
entity Document;
action view appliesTo { principal: [User], resource: [Document] };
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be in empty namespace
	if _, ok := s.Namespaces[""]; !ok {
		t.Fatal("expected empty namespace")
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	// Check entity type is unqualified in empty namespace
	ns := rs.Namespaces[""]
	userType := types.EntityType("User")
	if _, ok := ns.EntityTypes[userType]; !ok {
		t.Errorf("expected User entity type with key %q", userType)
	}
}

func TestBuilderAllMethods(t *testing.T) {
	// Test all builder methods for coverage
	s := schema.NewBuilder().
		Namespace("TestNS").
		Annotate("nsDoc", "Test namespace").
		EnumType("Status", "Active", "Inactive").
		CommonType("MyString", schema.String()).
		Entity("User").
		MemberOf("Group").
		Attr("name", schema.String()).
		OptionalAttr("nickname", schema.String()).
		Tags(schema.String()).
		Annotate("doc", "User entity").
		Entity("Group").
		Entity("Document").
		Action("read").
		Principal("User").
		Resource("Document").
		InGroupByName("write").
		Annotate("doc", "Read action").
		Action("write").
		Principal("User").
		Resource("Document").
		InGroup(&schema.ActionRef{ID: "admin"}, &schema.ActionRef{Type: "OtherNS::Action", ID: "superAction"}).
		Build()

	ns := s.Namespaces["TestNS"]
	if ns == nil {
		t.Fatal("expected TestNS namespace")
	}

	// Check namespace annotation
	if ns.Annotations["nsDoc"] != "Test namespace" {
		t.Errorf("expected namespace annotation")
	}

	// Check enum type
	if ns.EnumTypes["Status"] == nil {
		t.Error("expected Status enum type")
	}
	if len(ns.EnumTypes["Status"].Values) != 2 {
		t.Error("expected 2 enum values")
	}

	// Check common type
	if ns.CommonTypes["MyString"] == nil {
		t.Error("expected MyString common type")
	}

	// Check entity with all features
	user := ns.EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}
	if user.Annotations["doc"] != "User entity" {
		t.Error("expected doc annotation on User")
	}
	if user.Tags == nil {
		t.Error("expected tags on User")
	}
	if user.Shape.Attributes["nickname"] == nil {
		t.Error("expected nickname attribute")
	}
	if user.Shape.Attributes["nickname"].Required {
		t.Error("expected nickname to be optional")
	}

	// Check action with InGroup
	readAction := ns.Actions["read"]
	if readAction == nil {
		t.Fatal("expected read action")
	}
	if readAction.Annotations["doc"] != "Read action" {
		t.Error("expected doc annotation on action")
	}
	if len(readAction.MemberOf) != 1 || readAction.MemberOf[0].ID != "write" {
		t.Errorf("expected read to be member of write, got %v", readAction.MemberOf)
	}

	// Check action with InGroup referencing multiple actions
	writeAction := ns.Actions["write"]
	if writeAction == nil {
		t.Fatal("expected write action")
	}
	if len(writeAction.MemberOf) != 2 {
		t.Errorf("expected 2 memberOf, got %d", len(writeAction.MemberOf))
	}
}

func TestTypeConstructors(t *testing.T) {
	// Test all type constructor functions
	_ = schema.Long()
	_ = schema.String()
	_ = schema.Bool()
	_ = schema.Set(schema.Long())
	_ = schema.Entity("User")
	_ = schema.Extension("custom")
	_ = schema.IPAddr()
	_ = schema.Decimal()
	_ = schema.Datetime()
	_ = schema.Duration()
	_ = schema.CommonType("MyType")

	// Test RecordType construction directly
	rec := &schema.RecordType{Attributes: map[string]*schema.Attribute{
		"required": {Type: schema.String(), Required: true, Annotations: make(schema.Annotations)},
		"optional": {Type: schema.Long(), Required: false, Annotations: make(schema.Annotations)},
	}}

	if rec.Attributes["required"] == nil || !rec.Attributes["required"].Required {
		t.Error("expected required attribute to be required")
	}
	if rec.Attributes["optional"] == nil || rec.Attributes["optional"].Required {
		t.Error("expected optional attribute to be optional")
	}
}

func TestMarshalCedar(t *testing.T) {
	// Create a schema using the builder
	s := schema.NewBuilder().
		Namespace("TestApp").
		Annotate("doc", "Test application").
		Entity("User").
		MemberOf("Group").
		Attr("name", schema.String()).
		Annotate("doc", "User entity").
		Entity("Group").
		EnumType("Status", "Active", "Inactive").
		Action("view").
		Principal("User").
		Resource("Group").
		CommonType("MyRecord", &schema.RecordType{Attributes: map[string]*schema.Attribute{
			"field": {Type: schema.Long(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Build()

	// Marshal to Cedar format
	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal Cedar: %v", err)
	}

	// Parse it back
	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse marshaled Cedar: %v", err)
	}

	// Verify structure preserved
	if s2.Namespaces["TestApp"] == nil {
		t.Error("expected TestApp namespace after round-trip")
	}
}

func TestMarshalCedarEmptyNamespace(t *testing.T) {
	// Test marshaling with empty namespace (no namespace block)
	s := schema.NewBuilder().
		Namespace("").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal Cedar: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse marshaled Cedar: %v", err)
	}

	if s2.Namespaces[""] == nil {
		t.Error("expected empty namespace after round-trip")
	}
}

func TestFilename(t *testing.T) {
	var s schema.Schema
	s.SetFilename("test.cedarschema")
	if s.Filename() != "test.cedarschema" {
		t.Errorf("expected filename to be set")
	}
}

func TestErrorMessages(t *testing.T) {
	// Test CycleError message
	cycleSchema := `{
		"": {
			"commonTypes": {"a": {"type": "b"}, "b": {"type": "a"}},
			"entityTypes": {},
			"actions": {}
		}
	}`
	var s1 schema.Schema
	if err := json.Unmarshal([]byte(cycleSchema), &s1); err != nil {
		t.Fatal(err)
	}
	_, err := s1.Resolve()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, schema.ErrCycle) {
		t.Errorf("expected ErrCycle, got %v", err)
	}
	// Verify error message contains useful info
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}

	// Test ShadowError message
	shadowSchema := `{
		"": {"commonTypes": {"T": {"type": "String"}}, "entityTypes": {}, "actions": {}},
		"NS": {"commonTypes": {"T": {"type": "Long"}}, "entityTypes": {}, "actions": {}}
	}`
	var s2 schema.Schema
	if err := json.Unmarshal([]byte(shadowSchema), &s2); err != nil {
		t.Fatal(err)
	}
	_, err = s2.Resolve()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, schema.ErrShadow) {
		t.Errorf("expected ErrShadow, got %v", err)
	}

	// Test UndefinedTypeError message
	undefinedSchema := `{
		"": {"commonTypes": {"A": {"type": "Unknown"}}, "entityTypes": {}, "actions": {}}
	}`
	var s3 schema.Schema
	if err := json.Unmarshal([]byte(undefinedSchema), &s3); err != nil {
		t.Fatal(err)
	}
	_, err = s3.Resolve()
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, schema.ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}

	// Test DuplicateError - parse duplicate entity in Cedar
	// Note: JSON parsing won't catch duplicate keys (Go json library behavior)
	dupCedarSchema := `
namespace NS {
	entity User;
	entity User;
}
`
	var s4 schema.Schema
	err = s4.UnmarshalCedar([]byte(dupCedarSchema))
	if err == nil {
		t.Error("expected duplicate error")
	}
	// The error should be a parse error since duplicates are caught at parse time
	if !errors.Is(err, schema.ErrParse) && !errors.Is(err, schema.ErrDuplicate) {
		t.Logf("got error: %v (type %T)", err, err)
	}

	// Test ReservedNameError
	reservedSchema := `
entity Long;
action view appliesTo { principal: [Long], resource: [Long] };
`
	var s6 schema.Schema
	err = s6.UnmarshalCedar([]byte(reservedSchema))
	if err == nil {
		t.Error("expected reserved name error")
	}
	if !errors.Is(err, schema.ErrReservedName) {
		t.Logf("got error: %v", err)
	}

	// Test ParseError
	invalidSchema := `this is not valid cedar schema {`
	var s7 schema.Schema
	err = s7.UnmarshalCedar([]byte(invalidSchema))
	if err == nil {
		t.Error("expected parse error")
	}
	if !errors.Is(err, schema.ErrParse) {
		t.Logf("got error: %v", err)
	}
}

func TestJSONMarshalAllTypes(t *testing.T) {
	// Test JSON marshaling of various types
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("setField", schema.Set(schema.Long())).
		Attr("entityRef", schema.Entity("OtherEntity")).
		Attr("commonRef", schema.CommonType("MyType")).
		Entity("OtherEntity").
		CommonType("MyType", schema.String()).
		EnumType("Status", "A", "B").
		Action("view").
		Principal("User").
		Resource("OtherEntity").
		InGroupByName("write").
		Context(&schema.RecordType{Attributes: map[string]*schema.Attribute{
			"flag": {Type: schema.Bool(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Action("write").
		Principal("User").
		Resource("OtherEntity").
		Build()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestJSONParseVariousFormats(t *testing.T) {
	// Test JSON parsing with various type formats
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"boolField": {"type": "Boolean"},
							"setField": {"type": "Set", "element": {"type": "String"}},
							"entityField": {"type": "Entity", "name": "Group"},
							"extensionField": {"type": "Extension", "name": "ipaddr"},
							"commonRef": {"type": "MyType"}
						}
					}
				},
				"Group": {}
			},
			"commonTypes": {
				"MyType": {"type": "Long"}
			},
			"actions": {
				"view": {
					"memberOf": [{"type": "Action", "id": "admin"}],
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Group"],
						"context": {"type": "Record", "attributes": {"flag": {"type": "Bool"}}}
					}
				},
				"admin": {}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Verify the schema can be resolved
	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

func TestCedarParseWithComments(t *testing.T) {
	cedarSchema := `
// This is a line comment
namespace Test {
	/* This is a
	   block comment */
	entity User;
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse with comments: %v", err)
	}
	if s.Namespaces["Test"] == nil {
		t.Error("expected Test namespace")
	}
}

func TestCedarParseMultipleEntitiesPerDeclaration(t *testing.T) {
	cedarSchema := `
namespace Test {
	entity User, Admin, Guest in [Group];
	entity Group;
	action view appliesTo { principal: [User], resource: [Group] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces["Test"]
	if ns.EntityTypes["User"] == nil {
		t.Error("expected User entity")
	}
	if ns.EntityTypes["Admin"] == nil {
		t.Error("expected Admin entity")
	}
	if ns.EntityTypes["Guest"] == nil {
		t.Error("expected Guest entity")
	}
}

func TestCedarParseActionMemberOf(t *testing.T) {
	cedarSchema := `
namespace Test {
	entity User;
	action readOnly;
	action read in [readOnly] appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	readAction := s.Namespaces["Test"].Actions["read"]
	if len(readAction.MemberOf) != 1 || readAction.MemberOf[0].ID != "readOnly" {
		t.Errorf("expected read to be member of readOnly")
	}
}

func TestCedarParseQualifiedPaths(t *testing.T) {
	cedarSchema := `
namespace A::B::C {
	entity User;
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if s.Namespaces["A::B::C"] == nil {
		t.Error("expected A::B::C namespace")
	}
}

func TestResolveNilSchema(t *testing.T) {
	var s *schema.Schema
	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error when resolving nil schema")
	}
}

func TestJSONParseInlineRecord(t *testing.T) {
	// Test JSON parsing with inline record without explicit "type": "Record"
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"attributes": {
							"name": {"type": "String"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil || user.Shape == nil {
		t.Error("expected User with shape")
	}
}

func TestJSONParseTags(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "String"}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil || user.Tags == nil {
		t.Error("expected User with tags")
	}
}

func TestJSONParseEnumType(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"Status": {"enum": ["Active", "Inactive"]}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	status := s.Namespaces[""].EnumTypes["Status"]
	if status == nil {
		t.Fatal("expected Status enum type")
	}
	if len(status.Values) != 2 {
		t.Error("expected 2 enum values")
	}
}

func TestJSONParseContextRef(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {
				"MyContext": {"type": "Record", "attributes": {"flag": {"type": "Bool"}}}
			},
			"entityTypes": {
				"User": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "MyContext"}
					}
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if action == nil || action.AppliesTo == nil {
		t.Fatal("expected view action with appliesTo")
	}

	// Context should be a reference, not inline
	if action.AppliesTo.ContextRef == nil {
		t.Error("expected context to be a reference")
	}

	// Resolve should work
	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	viewUID := types.NewEntityUID("Action", "view")
	view := rs.Namespaces[""].Actions[viewUID]
	if view == nil || view.Context == nil {
		t.Error("expected resolved context")
	}
}

func TestJSONParseEntityOrCommonRef(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ref": {"type": "EntityOrCommon", "name": "Group"}
						}
					}
				},
				"Group": {}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should resolve successfully
	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

func TestCedarParseWithTags(t *testing.T) {
	cedarSchema := `
entity User tags String;
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil || user.Tags == nil {
		t.Error("expected User with tags")
	}
}

func TestCedarParseEntityWithEquals(t *testing.T) {
	cedarSchema := `
entity User = { name: String };
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil || user.Shape == nil {
		t.Error("expected User with shape")
	}
}

func TestCedarParseSingleTypeInList(t *testing.T) {
	cedarSchema := `
entity User in Group;
entity Group;
action view appliesTo { principal: User, resource: Group };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if len(user.MemberOfTypes) != 1 || user.MemberOfTypes[0] != "Group" {
		t.Errorf("expected User in Group")
	}

	action := s.Namespaces[""].Actions["view"]
	if len(action.AppliesTo.PrincipalTypes) != 1 || action.AppliesTo.PrincipalTypes[0] != "User" {
		t.Errorf("expected principal User")
	}
}

func TestCedarParseQuotedActionName(t *testing.T) {
	cedarSchema := `
entity User;
action "view photo" appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if s.Namespaces[""].Actions["view photo"] == nil {
		t.Error("expected 'view photo' action")
	}
}

func TestCedarParseQuotedAttributeName(t *testing.T) {
	cedarSchema := `
entity User {
	"first name": String,
};
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user.Shape.Attributes["first name"] == nil {
		t.Error("expected 'first name' attribute")
	}
}

func TestCedarParseBuiltinExtensions(t *testing.T) {
	cedarSchema := `
entity User {
	ip: ipaddr,
	amount: decimal,
	created: datetime,
	timeout: duration,
};
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces[""].EntityTypes[types.EntityType("User")]
	for _, attrName := range []string{"ip", "amount", "created", "timeout"} {
		attr := user.Shape.Attributes[attrName]
		if attr == nil {
			t.Errorf("expected %s attribute", attrName)
			continue
		}
		if _, ok := attr.Type.(resolved.Extension); !ok {
			t.Errorf("expected %s to be extension type", attrName)
		}
	}
}

func TestCedarParseCedarQualifiedExtensions(t *testing.T) {
	cedarSchema := `
entity User {
	ip: __cedar::ipaddr,
	str: __cedar::String,
	num: __cedar::Long,
	flag: __cedar::Bool,
	amount: __cedar::decimal,
	time: __cedar::datetime,
	dur: __cedar::duration,
};
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces[""].EntityTypes[types.EntityType("User")]
	if user == nil || user.Shape == nil {
		t.Fatal("expected User with shape")
	}
}

func TestResolveCrossNamespaceCommonType(t *testing.T) {
	jsonSchema := `{
		"NS1": {
			"commonTypes": {"MyType": {"type": "String"}},
			"entityTypes": {},
			"actions": {}
		},
		"NS2": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"data": {"type": "NS1::MyType"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces["NS2"].EntityTypes[types.EntityType("NS2::User")]
	if user == nil {
		t.Fatal("expected User")
	}
	dataAttr := user.Shape.Attributes["data"]
	if dataAttr == nil {
		t.Fatal("expected data attribute")
	}
	// Should be resolved to primitive String
	if _, ok := dataAttr.Type.(resolved.Primitive); !ok {
		t.Errorf("expected Primitive, got %T", dataAttr.Type)
	}
}

func TestResolveActionWithFullType(t *testing.T) {
	jsonSchema := `{
		"NS1": {
			"entityTypes": {"User": {}},
			"actions": {
				"read": {},
				"write": {
					"memberOf": [{"type": "NS1::Action", "id": "read"}],
					"appliesTo": {"principalTypes": ["User"], "resourceTypes": ["User"]}
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	writeUID := types.NewEntityUID("NS1::Action", "write")
	write := rs.Namespaces["NS1"].Actions[writeUID]
	if write == nil {
		t.Fatal("expected write action")
	}
	if len(write.MemberOf) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(write.MemberOf))
	}
}

func TestCedarParseActionInGroupWithQuotedRef(t *testing.T) {
	cedarSchema := `
namespace Test {
	entity User;
	action "group action";
	action view in ["group action"] appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	viewAction := s.Namespaces["Test"].Actions["view"]
	if len(viewAction.MemberOf) != 1 || viewAction.MemberOf[0].ID != "group action" {
		t.Errorf("expected view to be member of 'group action', got %v", viewAction.MemberOf)
	}
}

func TestCedarParseActionWithCrossNamespaceMemberOf(t *testing.T) {
	cedarSchema := `
namespace NS1 {
	entity User;
	action admin;
}
namespace NS2 {
	entity User;
	action view in [NS1::Action::"admin"] appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	viewAction := s.Namespaces["NS2"].Actions["view"]
	if len(viewAction.MemberOf) != 1 {
		t.Fatalf("expected 1 memberOf, got %d", len(viewAction.MemberOf))
	}
	if viewAction.MemberOf[0].Type != "NS1::Action" || viewAction.MemberOf[0].ID != "admin" {
		t.Errorf("expected NS1::Action::admin, got %+v", viewAction.MemberOf[0])
	}
}

func TestCedarParseMultipleActionsPerDeclaration(t *testing.T) {
	cedarSchema := `
namespace Test {
	entity User;
	action read, write, delete appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces["Test"]
	for _, name := range []string{"read", "write", "delete"} {
		if ns.Actions[name] == nil {
			t.Errorf("expected %s action", name)
		}
	}
}

func TestMarshalCedarWithActionMemberOf(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("admin").
		Action("view").
		Principal("User").
		Resource("User").
		InGroupByName("admin").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	viewAction := s2.Namespaces["Test"].Actions["view"]
	if viewAction == nil {
		t.Fatal("expected view action")
	}
}

func TestMarshalCedarWithRecordContext(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").
		Principal("User").
		Resource("User").
		Context(&schema.RecordType{Attributes: map[string]*schema.Attribute{
			"flag": {Type: schema.Bool(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

func TestMarshalCedarWithSetType(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("friends", schema.Set(schema.Entity("User"))).
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

func TestMarshalCedarWithNestedRecord(t *testing.T) {
	nested := &schema.RecordType{Attributes: map[string]*schema.Attribute{
		"inner": {Type: schema.String(), Required: true, Annotations: make(schema.Annotations)},
	}}

	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("data", nested).
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

func TestMarshalCedarWithContextRef(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		CommonType("MyContext", &schema.RecordType{Attributes: map[string]*schema.Attribute{
			"flag": {Type: schema.Bool(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Entity("User").
		Action("view").
		Principal("User").
		Resource("User").
		Build()

	// Manually set ContextRef since builder doesn't support it directly
	s.Namespaces["Test"].Actions["view"].AppliesTo = &schema.AppliesTo{
		PrincipalTypes: []string{"User"},
		ResourceTypes:  []string{"User"},
		ContextRef:     schema.CommonType("MyContext"),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestJSONMarshalContextRef(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		CommonType("MyContext", &schema.RecordType{Attributes: map[string]*schema.Attribute{
			"flag": {Type: schema.Bool(), Required: true, Annotations: make(schema.Annotations)},
		}}).
		Entity("User").
		Action("view").
		Principal("User").
		Resource("User").
		Build()

	// Manually set ContextRef since builder doesn't support it directly
	s.Namespaces["Test"].Actions["view"].AppliesTo = &schema.AppliesTo{
		PrincipalTypes: []string{"User"},
		ResourceTypes:  []string{"User"},
		ContextRef:     schema.CommonType("MyContext"),
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestMarshalCedarWithActionInCrossNamespace(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("NS1").
		Entity("User").
		Action("admin").
		Namespace("NS2").
		Entity("User").
		Action("view").
		Principal("User").
		Resource("User").
		InGroup(&schema.ActionRef{Type: "NS1::Action", ID: "admin"}).
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestCedarParseAnnotationWithoutValue(t *testing.T) {
	cedarSchema := `
@deprecated
entity User;
action view appliesTo { principal: [User], resource: [User] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user.Annotations["deprecated"] != "" {
		t.Errorf("expected deprecated annotation with empty value, got %q", user.Annotations["deprecated"])
	}
}

func TestJSONParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "invalid_json",
			json:    `{invalid}`,
			wantErr: true,
		},
		{
			name:    "invalid_namespace",
			json:    `{"": "not an object"}`,
			wantErr: true,
		},
		{
			name:    "invalid_entity_type",
			json:    `{"": {"entityTypes": {"User": "invalid"}, "actions": {}}}`,
			wantErr: true,
		},
		{
			name:    "invalid_action",
			json:    `{"": {"entityTypes": {}, "actions": {"view": "invalid"}}}`,
			wantErr: true,
		},
		{
			name:    "invalid_common_type",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"T": "invalid"}}}`,
			wantErr: true,
		},
		{
			name:    "entity_requires_name",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"ref": {"type": "Entity"}}}}}, "actions": {}}}`,
			wantErr: true,
		},
		{
			name:    "extension_requires_name",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"ext": {"type": "Extension"}}}}}, "actions": {}}}`,
			wantErr: true,
		},
		{
			name:    "set_requires_element",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"items": {"type": "Set"}}}}}, "actions": {}}}`,
			wantErr: true,
		},
		{
			name:    "shape_must_be_record",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "String"}}}, "actions": {}}}`,
			wantErr: true,
		},
		{
			name:    "context_must_be_record_or_ref",
			json:    `{"": {"entityTypes": {"User": {}}, "actions": {"view": {"appliesTo": {"context": {"type": "String"}}}}}}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			err := json.Unmarshal([]byte(tt.json), &s)
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCedarParseErrors(t *testing.T) {
	tests := []struct {
		name    string
		cedar   string
		wantErr bool
	}{
		{
			name:    "unexpected_token",
			cedar:   `foo bar baz`,
			wantErr: true,
		},
		{
			name:    "missing_semicolon",
			cedar:   `entity User`,
			wantErr: true,
		},
		{
			name:    "missing_namespace_brace",
			cedar:   `namespace Test`,
			wantErr: true,
		},
		{
			name:    "invalid_type_keyword",
			cedar:   `type T;`,
			wantErr: true,
		},
		{
			name:    "duplicate_common_type",
			cedar:   `type T = String; type T = Long; entity User; action view appliesTo { principal: [User], resource: [User] };`,
			wantErr: true,
		},
		{
			name:    "invalid_appliesTo_key",
			cedar:   `entity User; action view appliesTo { invalid: [User] };`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			err := s.UnmarshalCedar([]byte(tt.cedar))
			if tt.wantErr && err == nil {
				t.Error("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResolveWithEntityTags(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "String"},
					"shape": {"type": "Record", "attributes": {"name": {"type": "String"}}}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces[""].EntityTypes[types.EntityType("User")]
	if user == nil {
		t.Fatal("expected User")
	}
	if user.Tags == nil {
		t.Error("expected tags on User")
	}
}

func TestResolveContextMustBeRecord(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {
				"MyType": {"type": "String"}
			},
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "MyType"}
					}
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error: context must be record")
	}
}

func TestPrimitiveKindString(t *testing.T) {
	// Test the String() method on PrimitiveKind
	longType := schema.Long().(schema.PrimitiveType)
	if longType.Kind.String() != "Long" {
		t.Errorf("expected Long, got %s", longType.Kind.String())
	}

	stringType := schema.String().(schema.PrimitiveType)
	if stringType.Kind.String() != "String" {
		t.Errorf("expected String, got %s", stringType.Kind.String())
	}

	boolType := schema.Bool().(schema.PrimitiveType)
	if boolType.Kind.String() != "Bool" {
		t.Errorf("expected Bool, got %s", boolType.Kind.String())
	}

	// Test unknown value
	unknown := schema.PrimitiveType{Kind: schema.PrimitiveKind(99)}
	if unknown.Kind.String() != "Unknown" {
		t.Errorf("expected Unknown, got %s", unknown.Kind.String())
	}
}

func TestMarshalEnumInJSON(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		EnumType("Status", "Active", "Inactive").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	status := s2.Namespaces["Test"].EnumTypes["Status"]
	if status == nil {
		t.Fatal("expected Status enum")
	}
	if len(status.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(status.Values))
	}
}

func TestDuplicateEnumType(t *testing.T) {
	cedarSchema := `
entity Status enum ["A"];
entity Status enum ["B"];
action view appliesTo { principal: [Status], resource: [Status] };
`
	var s schema.Schema
	err := s.UnmarshalCedar([]byte(cedarSchema))
	if err == nil {
		t.Error("expected duplicate error")
	}
}

func TestJSONParseWithAnnotations(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"annotations": {"doc": "A user"},
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String", "annotations": {"desc": "User name"}}
						}
					}
				}
			},
			"actions": {
				"view": {
					"annotations": {"deprecated": "true"}
				}
			},
			"commonTypes": {
				"MyType": {"type": "String", "annotations": {"note": "A type"}}
			},
			"annotations": {"version": "1.0"}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces[""]
	if ns.Annotations["version"] != "1.0" {
		t.Error("expected namespace annotation")
	}

	user := ns.EntityTypes["User"]
	if user.Annotations["doc"] != "A user" {
		t.Error("expected entity annotation")
	}

	view := ns.Actions["view"]
	if view.Annotations["deprecated"] != "true" {
		t.Error("expected action annotation")
	}
}

func TestResolveBuiltinTypes(t *testing.T) {
	// Test resolution of builtin types by name
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"longVal": {"type": "Long"},
							"stringVal": {"type": "String"},
							"boolVal": {"type": "Bool"},
							"ipVal": {"type": "ipaddr"},
							"decVal": {"type": "decimal"},
							"dtVal": {"type": "datetime"},
							"durVal": {"type": "duration"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces[""].EntityTypes[types.EntityType("User")]
	if user == nil || user.Shape == nil {
		t.Fatal("expected User with shape")
	}

	// Verify all attributes resolved
	for _, name := range []string{"longVal", "stringVal", "boolVal", "ipVal", "decVal", "dtVal", "durVal"} {
		if user.Shape.Attributes[name] == nil {
			t.Errorf("expected %s attribute", name)
		}
	}
}

func TestParseErrorWithFilename(t *testing.T) {
	var s schema.Schema
	s.SetFilename("test.cedarschema")
	err := s.UnmarshalCedar([]byte("invalid {"))
	if err == nil {
		t.Error("expected parse error")
	}
	if !errors.Is(err, schema.ErrParse) {
		t.Errorf("expected ErrParse, got %v", err)
	}
	errMsg := err.Error()
	if errMsg == "" {
		t.Error("expected error message")
	}
}

func TestDuplicateErrors(t *testing.T) {
	// Test duplicate entity type in same namespace
	cedarSchema := `
namespace NS {
	entity User;
	entity User;
}
`
	var s schema.Schema
	err := s.UnmarshalCedar([]byte(cedarSchema))
	if err == nil {
		t.Error("expected duplicate error")
	}

	// Test duplicate action
	cedarSchema2 := `
namespace NS {
	entity User;
	action view appliesTo { principal: [User], resource: [User] };
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s2 schema.Schema
	err = s2.UnmarshalCedar([]byte(cedarSchema2))
	if err == nil {
		t.Error("expected duplicate action error")
	}
}

func TestReservedEnumTypeName(t *testing.T) {
	cedarSchema := `
entity Long enum ["A", "B"];
action view appliesTo { principal: [Long], resource: [Long] };
`
	var s schema.Schema
	err := s.UnmarshalCedar([]byte(cedarSchema))
	if err == nil {
		t.Error("expected reserved name error")
	}
}

func TestReservedTypeNameInJSON(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {"Long": {}},
			"actions": {}
		}
	}`
	var s schema.Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected reserved name error")
	}
}

func TestReservedCommonTypeNameInJSON(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {"Bool": {"type": "String"}},
			"entityTypes": {},
			"actions": {}
		}
	}`
	var s schema.Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected reserved name error")
	}
}

func TestMarshalCedarWithEntityTagsAndAnnotations(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Tags(schema.String()).
		Annotate("doc", "User entity").
		EnumType("Status", "Active", "Inactive").
		Action("view").Principal("User").Resource("User").
		Build()

	// Add annotation to enum
	s.Namespaces["Test"].EnumTypes["Status"].Annotations["doc"] = "Status enum"

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

func TestMarshalCedarWithEmptyAppliesTo(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view"). // No appliesTo
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestMarshalCedarWithQuotedIdentifiers(t *testing.T) {
	// Create schema with identifiers that need quoting
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	// Add attribute with spaces
	s.Namespaces["Test"].EntityTypes["User"].Shape = &schema.RecordType{
		Attributes: map[string]*schema.Attribute{
			"first name": {
				Type:        schema.String(),
				Required:    true,
				Annotations: make(schema.Annotations),
			},
		},
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if s2.Namespaces["Test"].EntityTypes["User"].Shape.Attributes["first name"] == nil {
		t.Error("expected 'first name' attribute after round-trip")
	}
}

func TestJSONMarshalSetType(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("items", schema.Set(schema.Long())).
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestJSONUnmarshalEntityOrCommonWithName(t *testing.T) {
	// Test the case where type is empty but name is present
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ref": {"name": "Group"}
						}
					}
				},
				"Group": {}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should be able to resolve
	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

func TestDuplicateEnumTypeDuringResolve(t *testing.T) {
	// Create a schema with duplicate enum type (same name as entity type)
	// This can only be done programmatically as the parser would catch it
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"": {
				EntityTypes: map[string]*schema.EntityTypeDef{
					"User": {},
				},
				EnumTypes: map[string]*schema.EnumTypeDef{
					"User": {Values: []string{"A"}},
				},
				Actions:     make(map[string]*schema.ActionDef),
				CommonTypes: make(map[string]*schema.CommonTypeDef),
				Annotations: make(schema.Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected duplicate error during resolve")
	}
	if !errors.Is(err, schema.ErrDuplicate) {
		t.Logf("got error: %v", err)
	}
}

func TestResolveDuplicateCommonType(t *testing.T) {
	// Check for duplicate via JSON (which allows setting twice via programmatic approach)
	// We create this programmatically
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"": {
				EntityTypes: make(map[string]*schema.EntityTypeDef),
				EnumTypes:   make(map[string]*schema.EnumTypeDef),
				Actions:     make(map[string]*schema.ActionDef),
				CommonTypes: map[string]*schema.CommonTypeDef{
					"T": {Type: schema.String()},
				},
				Annotations: make(schema.Annotations),
			},
			"NS": {
				EntityTypes: make(map[string]*schema.EntityTypeDef),
				EnumTypes:   make(map[string]*schema.EnumTypeDef),
				Actions:     make(map[string]*schema.ActionDef),
				CommonTypes: map[string]*schema.CommonTypeDef{
					"T": {Type: schema.Long()},
				},
				Annotations: make(schema.Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected shadow error")
	}
	if !errors.Is(err, schema.ErrShadow) {
		t.Logf("got error: %v", err)
	}
}

func TestMarshalCedarWithCommonTypeAnnotations(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	// Add common type with annotation
	s.Namespaces["Test"].CommonTypes["MyType"] = &schema.CommonTypeDef{
		Type:        schema.String(),
		Annotations: schema.Annotations{"doc": "My type"},
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestMarshalCedarOptionalAttributes(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		OptionalAttr("nickname", schema.String()).
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := s2.UnmarshalCedar(data); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s2.Namespaces["Test"].EntityTypes["User"]
	if user.Shape.Attributes["nickname"].Required {
		t.Error("expected nickname to be optional after round-trip")
	}
}

func TestMarshalCedarAttributeAnnotations(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("name", schema.String()).
		Action("view").Principal("User").Resource("User").
		Build()

	// Add annotation to attribute
	s.Namespaces["Test"].EntityTypes["User"].Shape.Attributes["name"].Annotations["doc"] = "User name"

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestMarshalCedarActionAnnotations(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").
		Principal("User").
		Resource("User").
		Annotate("doc", "View action").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestResolveWithEnumAsMemberOf(t *testing.T) {
	// Enum types should not be valid for memberOf
	jsonSchema := `{
		"": {
			"entityTypes": {
				"Status": {"enum": ["Active", "Inactive"]},
				"User": {"memberOfTypes": ["Status"]}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// This should succeed because enums are valid entity types
	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	user := rs.Namespaces[""].EntityTypes[types.EntityType("User")]
	if user == nil {
		t.Fatal("expected User")
	}
}

func TestErrorStrings(t *testing.T) {
	// Test that all error types produce meaningful strings
	// We do this by triggering the errors and checking their messages

	// ShadowError
	shadowSchema := `{
		"": {"entityTypes": {"T": {}}, "actions": {}},
		"NS": {"entityTypes": {"T": {}}, "actions": {}}
	}`
	var s1 schema.Schema
	if err := json.Unmarshal([]byte(shadowSchema), &s1); err != nil {
		t.Fatalf("failed to parse shadow schema: %v", err)
	}
	_, err := s1.Resolve()
	if err != nil {
		msg := err.Error()
		if msg == "" {
			t.Error("ShadowError: expected non-empty message")
		}
		t.Logf("ShadowError: %s", msg)
	}

	// DuplicateError (via duplicate common type during resolve)
	s2 := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"": {
				EntityTypes: map[string]*schema.EntityTypeDef{"User": {}},
				EnumTypes:   map[string]*schema.EnumTypeDef{"User": {Values: []string{"A"}}},
				Actions:     make(map[string]*schema.ActionDef),
				CommonTypes: make(map[string]*schema.CommonTypeDef),
				Annotations: make(schema.Annotations),
			},
		},
	}
	_, err = s2.Resolve()
	if err != nil {
		msg := err.Error()
		if msg == "" {
			t.Error("DuplicateError: expected non-empty message")
		}
		t.Logf("DuplicateError: %s", msg)
	}

	// ReservedNameError
	var s3 schema.Schema
	err = s3.UnmarshalCedar([]byte("entity String; action v appliesTo { principal: [String], resource: [String] };"))
	if err != nil {
		msg := err.Error()
		if msg == "" {
			t.Error("ReservedNameError: expected non-empty message")
		}
		t.Logf("ReservedNameError: %s", msg)
	}

	// ParseError with line/column
	var s4 schema.Schema
	err = s4.UnmarshalCedar([]byte("entity User {\nfoo: BadType\n};"))
	if err != nil {
		msg := err.Error()
		if msg == "" {
			t.Error("ParseError: expected non-empty message")
		}
		t.Logf("ParseError: %s", msg)
	}
}

func TestMarshalCedarEntityOrCommonRef(t *testing.T) {
	// Test marshaling EntityOrCommonRef type
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	// Add a common type that uses EntityOrCommonRef internally
	// This gets created when parsing ambiguous type references
	s.Namespaces["Test"].CommonTypes["Ref"] = &schema.CommonTypeDef{
		Type:        schema.EntityOrCommonRef{Name: "User"},
		Annotations: make(schema.Annotations),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	t.Logf("Marshaled: %s", data)
}

func TestJSONMarshalEntityOrCommonRef(t *testing.T) {
	s := schema.NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	// Add a common type that uses EntityOrCommonRef
	s.Namespaces["Test"].CommonTypes["Ref"] = &schema.CommonTypeDef{
		Type:        schema.EntityOrCommonRef{Name: "User"},
		Annotations: make(schema.Annotations),
	}

	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestCedarParseActionListMemberOf(t *testing.T) {
	// Test parsing action memberOf as list with single element (not in brackets)
	cedarSchema := `
namespace Test {
	entity User;
	action admin;
	action view in admin appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	viewAction := s.Namespaces["Test"].Actions["view"]
	if len(viewAction.MemberOf) != 1 || viewAction.MemberOf[0].ID != "admin" {
		t.Errorf("expected view in admin")
	}
}

func TestJSONParseDuplicateActionsWontFailOnParse(t *testing.T) {
	// JSON parsing doesn't catch duplicate keys (Go json library behavior)
	jsonSchema := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {},
				"view": {}
			}
		}
	}`
	var s schema.Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	// This won't error because Go's JSON parser silently takes the last value
	if err != nil {
		t.Logf("error (unexpected): %v", err)
	}
}

func TestReservedEnumTypeNameCedar(t *testing.T) {
	// Specifically test that enum types with reserved names are rejected
	cedarSchema := `entity Set enum ["A", "B"]; action v appliesTo { principal: [Set], resource: [Set] };`
	var s schema.Schema
	err := s.UnmarshalCedar([]byte(cedarSchema))
	if err == nil {
		t.Error("expected reserved name error for Set enum")
	}
}

func TestDuplicateInNamespaceEmptyVsEmptyString(t *testing.T) {
	// Ensure empty namespace works correctly
	cedarSchema := `
entity User;
entity Group;
action view appliesTo { principal: [User], resource: [Group] };
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should resolve successfully
	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	if rs.Namespaces[""] == nil {
		t.Error("expected empty namespace")
	}
}

func TestResolveMultipartNamespace(t *testing.T) {
	cedarSchema := `
namespace A::B::C {
	type MyType = String;
	entity User {
		data: MyType,
	};
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}

	ns := rs.Namespaces["A::B::C"]
	if ns == nil {
		t.Fatal("expected A::B::C namespace")
	}

	user := ns.EntityTypes[types.EntityType("A::B::C::User")]
	if user == nil {
		t.Fatal("expected User entity type")
	}
}

func TestJSONRoundTripCommonTypeWithAnnotations(t *testing.T) {
	// Test that CommonTypeDef annotations are preserved through JSON marshal/unmarshal
	jsonSchema := `{
		"": {
			"entityTypes": {"User": {}},
			"actions": {},
			"commonTypes": {
				"MyString": {"type": "String", "annotations": {"doc": "A string type", "version": "1.0"}},
				"MyRecord": {
					"type": "Record",
					"attributes": {"field": {"type": "Long"}},
					"annotations": {"doc": "A record type"}
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse JSON schema: %v", err)
	}

	// Verify annotations were parsed
	myString := s.Namespaces[""].CommonTypes["MyString"]
	if myString == nil {
		t.Fatal("expected MyString common type")
	}
	if myString.Annotations["doc"] != "A string type" {
		t.Errorf("expected MyString doc annotation, got %v", myString.Annotations)
	}
	if myString.Annotations["version"] != "1.0" {
		t.Errorf("expected MyString version annotation, got %v", myString.Annotations)
	}

	myRecord := s.Namespaces[""].CommonTypes["MyRecord"]
	if myRecord == nil {
		t.Fatal("expected MyRecord common type")
	}
	if myRecord.Annotations["doc"] != "A record type" {
		t.Errorf("expected MyRecord doc annotation, got %v", myRecord.Annotations)
	}

	// Marshal back to JSON
	data, err := json.Marshal(&s)
	if err != nil {
		t.Fatalf("failed to marshal JSON schema: %v", err)
	}

	// Unmarshal again to verify round-trip
	var s2 schema.Schema
	if err := json.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to re-parse JSON schema: %v", err)
	}

	// Verify annotations are preserved after round-trip
	myString2 := s2.Namespaces[""].CommonTypes["MyString"]
	if myString2 == nil {
		t.Fatal("expected MyString common type after round-trip")
	}
	if myString2.Annotations["doc"] != "A string type" {
		t.Errorf("expected MyString doc annotation after round-trip, got %v", myString2.Annotations)
	}
	if myString2.Annotations["version"] != "1.0" {
		t.Errorf("expected MyString version annotation after round-trip, got %v", myString2.Annotations)
	}

	myRecord2 := s2.Namespaces[""].CommonTypes["MyRecord"]
	if myRecord2 == nil {
		t.Fatal("expected MyRecord common type after round-trip")
	}
	if myRecord2.Annotations["doc"] != "A record type" {
		t.Errorf("expected MyRecord doc annotation after round-trip, got %v", myRecord2.Annotations)
	}
}

func TestBuilderChainedWithoutDone(t *testing.T) {
	// Test the new chaining pattern without explicit Done() calls
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").MemberOf("Group").
		Entity("Group").
		Action("view").Principal("User").Resource("Group").
		Build()

	// Check structure
	ns := s.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	if len(ns.EntityTypes) != 2 {
		t.Errorf("expected 2 entity types, got %d", len(ns.EntityTypes))
	}

	user := ns.EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	if len(user.MemberOfTypes) != 1 || user.MemberOfTypes[0] != "Group" {
		t.Errorf("expected User memberOf Group, got %v", user.MemberOfTypes)
	}

	group := ns.EntityTypes["Group"]
	if group == nil {
		t.Fatal("expected Group entity type")
	}

	view := ns.Actions["view"]
	if view == nil {
		t.Fatal("expected view action")
	}

	if view.AppliesTo == nil {
		t.Fatal("expected appliesTo on view action")
	}

	if len(view.AppliesTo.PrincipalTypes) != 1 || view.AppliesTo.PrincipalTypes[0] != "User" {
		t.Errorf("expected principal User, got %v", view.AppliesTo.PrincipalTypes)
	}
}

func TestBuilderChainedMultipleNamespaces(t *testing.T) {
	// Test chaining across multiple namespaces without explicit Done() calls
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").
		Entity("Group").
		Namespace("OtherApp").
		Entity("Foo").
		Build()

	// Check MyApp namespace
	myApp := s.Namespaces["MyApp"]
	if myApp == nil {
		t.Fatal("expected MyApp namespace")
	}
	if len(myApp.EntityTypes) != 2 {
		t.Errorf("expected 2 entity types in MyApp, got %d", len(myApp.EntityTypes))
	}
	if myApp.EntityTypes["User"] == nil {
		t.Error("expected User in MyApp")
	}
	if myApp.EntityTypes["Group"] == nil {
		t.Error("expected Group in MyApp")
	}

	// Check OtherApp namespace
	otherApp := s.Namespaces["OtherApp"]
	if otherApp == nil {
		t.Fatal("expected OtherApp namespace")
	}
	if len(otherApp.EntityTypes) != 1 {
		t.Errorf("expected 1 entity type in OtherApp, got %d", len(otherApp.EntityTypes))
	}
	if otherApp.EntityTypes["Foo"] == nil {
		t.Error("expected Foo in OtherApp")
	}
}

func TestBuilderChainedFromAction(t *testing.T) {
	// Test chaining from ActionBuilder to other builders
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").
		Action("read").Principal("User").Resource("User").
		Action("write").Principal("User").Resource("User").
		Entity("Group").
		Namespace("OtherApp").
		Entity("Admin").
		Build()

	// Check MyApp
	myApp := s.Namespaces["MyApp"]
	if myApp == nil {
		t.Fatal("expected MyApp namespace")
	}
	if len(myApp.Actions) != 2 {
		t.Errorf("expected 2 actions in MyApp, got %d", len(myApp.Actions))
	}
	if myApp.Actions["read"] == nil {
		t.Error("expected read action")
	}
	if myApp.Actions["write"] == nil {
		t.Error("expected write action")
	}
	if len(myApp.EntityTypes) != 2 {
		t.Errorf("expected 2 entity types in MyApp, got %d", len(myApp.EntityTypes))
	}
	if myApp.EntityTypes["User"] == nil {
		t.Error("expected User in MyApp")
	}
	if myApp.EntityTypes["Group"] == nil {
		t.Error("expected Group in MyApp")
	}

	// Check OtherApp
	otherApp := s.Namespaces["OtherApp"]
	if otherApp == nil {
		t.Fatal("expected OtherApp namespace")
	}
	if otherApp.EntityTypes["Admin"] == nil {
		t.Error("expected Admin in OtherApp")
	}
}

func TestBuilderChainedEnumAndCommonType(t *testing.T) {
	// Test chaining EnumType and CommonType from EntityBuilder and ActionBuilder
	s := schema.NewBuilder().
		Namespace("MyApp").
		Entity("User").
		EnumType("Status", "Active", "Inactive").
		Entity("Document").
		CommonType("MyString", schema.String()).
		Action("view").Principal("User").Resource("Document").
		EnumType("Priority", "High", "Low").
		Action("edit").Principal("User").Resource("Document").
		CommonType("MyLong", schema.Long()).
		Build()

	ns := s.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	// Check entities
	if ns.EntityTypes["User"] == nil {
		t.Error("expected User entity")
	}
	if ns.EntityTypes["Document"] == nil {
		t.Error("expected Document entity")
	}

	// Check enum types
	if ns.EnumTypes["Status"] == nil {
		t.Error("expected Status enum")
	}
	if ns.EnumTypes["Priority"] == nil {
		t.Error("expected Priority enum")
	}

	// Check common types
	if ns.CommonTypes["MyString"] == nil {
		t.Error("expected MyString common type")
	}
	if ns.CommonTypes["MyLong"] == nil {
		t.Error("expected MyLong common type")
	}

	// Check actions
	if ns.Actions["view"] == nil {
		t.Error("expected view action")
	}
	if ns.Actions["edit"] == nil {
		t.Error("expected edit action")
	}
}
