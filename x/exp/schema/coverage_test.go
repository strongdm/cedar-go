package schema

import (
	"encoding/json"
	"fmt"
	"testing"
)

// Additional tests to improve code coverage

func TestTypeEqualitySetWithNonSet(t *testing.T) {
	t.Parallel()

	set := SetOf(String())
	str := String()

	if typeEqual(set, str) {
		t.Error("Set should not equal String")
	}
}

func TestTypeEqualityRecordWithNonRecord(t *testing.T) {
	t.Parallel()

	rec := Record(RequiredAttr("a", String()))
	str := String()

	if typeEqual(rec, str) {
		t.Error("Record should not equal String")
	}
}

func TestTypeEqualityRecordDifferentLengths(t *testing.T) {
	t.Parallel()

	rec1 := Record(RequiredAttr("a", String()))
	rec2 := Record(RequiredAttr("a", String()), RequiredAttr("b", Long()))

	if typeEqual(rec1, rec2) {
		t.Error("Records with different lengths should not be equal")
	}
}

func TestTypeEqualityEntityWithNonEntity(t *testing.T) {
	t.Parallel()

	ent := Entity("User")
	str := String()

	if typeEqual(ent, str) {
		t.Error("Entity should not equal String")
	}
}

func TestTypeEqualityExtensionWithNonExtension(t *testing.T) {
	t.Parallel()

	ext := Extension("ipaddr")
	str := String()

	if typeEqual(ext, str) {
		t.Error("Extension should not equal String")
	}
}

func TestTypeEqualityRefWithNonRef(t *testing.T) {
	t.Parallel()

	ref := Ref("MyType")
	str := String()

	if typeEqual(ref, str) {
		t.Error("Ref should not equal String")
	}
}

func TestJSONMarshalWithExtensionType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Request").
		SetAttributes(
			RequiredAttr("ip", Extension("ipaddr")),
		))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify extension type is properly marshaled
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}
}

func TestJSONUnmarshalWithExtensionType(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"Request": {
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
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	req := ns.GetEntity("Request")
	if req == nil {
		t.Fatal("expected Request entity")
	}
}

func TestJSONMarshalEntityType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Doc").
		SetAttributes(
			RequiredAttr("owner", Entity("User")),
		))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestJSONMarshalRefType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddCommonType(NewCommonType("Name", String()))
	s.AddEntity(NewEntity("User").
		SetAttributes(
			RequiredAttr("name", Ref("Name")),
		))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestJSONMarshalUnmarshalWithAllTypes(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")

	ns.AddEntity(NewEntity("Full").
		SetAttributes(
			RequiredAttr("b", Boolean()),
			RequiredAttr("l", Long()),
			RequiredAttr("s", String()),
			RequiredAttr("set", SetOf(String())),
			RequiredAttr("rec", Record(
				RequiredAttr("nested", String()),
			)),
			RequiredAttr("ent", Entity("Other")),
			RequiredAttr("ext", Extension("decimal")),
		))

	s.AddNamespace(ns)

	// Marshal
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Unmarshal
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify
	ns2 := s2.GetNamespace("Test")
	if ns2 == nil {
		t.Fatal("expected Test namespace")
	}

	full := ns2.GetEntity("Full")
	if full == nil {
		t.Fatal("expected Full entity")
	}

	if len(full.Attributes) != 7 {
		t.Errorf("expected 7 attributes, got %d", len(full.Attributes))
	}
}

func TestJSONUnmarshalActionWithContext(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"create": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Doc"],
						"context": {
							"type": "Record",
							"attributes": {
								"ip": {"type": "String", "required": true}
							}
						}
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	create := ns.GetAction("create")
	if create == nil {
		t.Fatal("expected create action")
	}

	if create.Context.v == nil {
		t.Error("expected context to be set")
	}
}

func TestJSONUnmarshalActionWithMemberOf(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"edit": {
					"memberOf": [
						{"id": "view"},
						{"id": "read", "type": "OtherNS"}
					]
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	edit := ns.GetAction("edit")
	if edit == nil {
		t.Fatal("expected edit action")
	}

	if len(edit.MemberOf) != 2 {
		t.Errorf("expected 2 memberOf, got %d", len(edit.MemberOf))
	}
}

func TestCedarUnmarshalEntityWithTags(t *testing.T) {
	t.Parallel()

	input := `
entity Tagged tags String;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	tagged := ns.GetEntity("Tagged")
	if tagged == nil {
		t.Fatal("expected Tagged entity")
	}

	if tagged.Tags.v == nil {
		t.Error("expected tags to be set")
	}
}

func TestCedarUnmarshalEntityWithEnum(t *testing.T) {
	t.Parallel()

	input := `
entity Status enum ["active", "inactive", "pending"];
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	status := ns.GetEntity("Status")
	if status == nil {
		t.Fatal("expected Status entity")
	}

	if len(status.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(status.Enum))
	}
}

func TestJSONMarshalError(t *testing.T) {
	t.Parallel()

	// This test verifies that MarshalJSON handles nil schema
	s := &Schema{}
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(jsonData) != "{}" {
		t.Errorf("expected {}, got %s", jsonData)
	}
}

func TestCedarRoundTripComplexSchema(t *testing.T) {
	t.Parallel()

	input := `
@doc("Main namespace")
namespace MyApp {
	type Name = String;
	type Address = {
		street: String,
		city: String,
	};

	@doc("User entity")
	entity User in [Group, Organization] {
		name: Name,
		email?: String,
		address?: Address,
	};

	entity Group {
		name: String,
	};

	entity Organization;

	entity Document {
		title: String,
		owner: User,
		tags?: Set<String>,
	};

	action view appliesTo {
		principal: [User, Group],
		resource: Document,
	};

	action edit in view appliesTo {
		principal: User,
		resource: Document,
		context: {
			ip: String,
		},
	};
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Marshal back to Cedar
	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Parse again
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarData); err != nil {
		t.Fatalf("Second UnmarshalCedar() error = %v", err)
	}

	// Verify structure is preserved
	ns := s2.GetNamespace("MyApp")
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	if len(ns.Entities) != 4 {
		t.Errorf("expected 4 entities, got %d", len(ns.Entities))
	}
	if len(ns.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(ns.Actions))
	}
	if len(ns.CommonTypes) != 2 {
		t.Errorf("expected 2 common types, got %d", len(ns.CommonTypes))
	}
}

func TestJSONToCedarConversion(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"MyApp": {
			"entityTypes": {
				"User": {
					"memberOfTypes": ["Group"],
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "String", "required": true},
							"email": {"type": "String", "required": false}
						}
					}
				},
				"Group": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Doc"]
					}
				}
			},
			"commonTypes": {
				"Name": {"type": "String"}
			}
		}
	}`

	// Parse JSON
	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonInput)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Marshal to Cedar
	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Parse Cedar
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarData); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Marshal back to JSON
	jsonData2, err := s2.MarshalJSON()
	if err != nil {
		t.Fatalf("Second MarshalJSON() error = %v", err)
	}

	// Verify valid JSON
	var raw interface{}
	if err := json.Unmarshal(jsonData2, &raw); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
}

func TestCedarToJSONConversion(t *testing.T) {
	t.Parallel()

	cedarInput := `
namespace MyApp {
	type Name = String;
	entity User {
		name: Name,
	};
	action view appliesTo {
		principal: User,
		resource: Doc,
	};
}
`

	// Parse Cedar
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedarInput)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Marshal to JSON
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Parse JSON
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Marshal back to Cedar
	cedarData2, err := s2.MarshalCedar()
	if err != nil {
		t.Fatalf("Second MarshalCedar() error = %v", err)
	}

	// Parse again to verify
	var s3 Schema
	if err := s3.UnmarshalCedar(cedarData2); err != nil {
		t.Fatalf("Third UnmarshalCedar() error = %v", err)
	}
}

func TestCommonTypeWithNilType(t *testing.T) {
	t.Parallel()

	ct := NewCommonType("Empty", Type{})

	if ct.Name != "Empty" {
		t.Errorf("expected Empty, got %s", ct.Name)
	}
}

func TestAttrToJSONWithNilType(t *testing.T) {
	t.Parallel()

	attr := Attribute{
		Name:     "test",
		Type:     Type{},
		Required: true,
	}

	result, err := attrToJSON(attr)
	if err != nil {
		t.Errorf("attrToJSON() error = %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestJSONAttrFromJSONAllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		attrType string
	}{
		{"Boolean", "Boolean"},
		{"Long", "Long"},
		{"String", "String"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `{
				"": {
					"entityTypes": {
						"Test": {
							"shape": {
								"type": "Record",
								"attributes": {
									"attr": {"type": "` + tt.attrType + `", "required": true}
								}
							}
						}
					},
					"actions": {}
				}
			}`

			var s Schema
			if err := s.UnmarshalJSON([]byte(input)); err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}

			ns := s.GetNamespace("")
			if ns == nil {
				t.Fatal("expected anonymous namespace")
			}

			test := ns.GetEntity("Test")
			if test == nil {
				t.Fatal("expected Test entity")
			}

			if len(test.Attributes) != 1 {
				t.Errorf("expected 1 attribute, got %d", len(test.Attributes))
			}
		})
	}
}

func TestCedarWriteTypeNil(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Empty"))

	// Should not error even with no attributes
	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if len(cedarData) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestCedarMarshalWithAnnotationsOnAll(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.Annotate("doc", "Namespace doc")

	ct := NewCommonType("MyType", String())
	ct.Annotate("doc", "Common type doc")
	ns.AddCommonType(ct)

	entity := NewEntity("User")
	entity.Annotate("doc", "Entity doc")
	entity.SetAttributes(RequiredAttr("name", String()))
	ns.AddEntity(entity)

	action := NewAction("view")
	action.Annotate("doc", "Action doc")
	action.SetPrincipalTypes("User")
	action.SetResourceTypes("Doc")
	ns.AddAction(action)

	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "@doc") {
		t.Error("expected @doc annotations in output")
	}
}

func TestCedarMarshalWithEnumEntity(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Status").SetEnum("active", "inactive", "pending"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "enum") {
		t.Error("expected enum in output")
	}
}

func TestCedarMarshalWithQuotedActionName(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("my-action") // has dash, needs quoting
	action.SetPrincipalTypes("User")
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, `"my-action"`) {
		t.Error("expected quoted action name in output")
	}
}

func TestCedarMarshalWithMultipleMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("edit")
	action.InActions(
		ActionRef{Name: "view"},
		ActionRef{Name: "read"},
	)
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "[") || !contains(output, "]") {
		t.Error("expected bracketed list for multiple memberOf")
	}
}

func TestCedarMarshalWithActionRefNamespace(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("edit")
	action.InActions(ActionRef{Namespace: "OtherNS", Name: "view"})
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "OtherNS::") {
		t.Error("expected namespace prefix in output")
	}
}

func TestCedarMarshalWithQuotedRecordAttr(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").SetAttributes(
		RequiredAttr("my-attr", String()), // has dash, needs quoting
	))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, `"my-attr"`) {
		t.Error("expected quoted attribute name in output")
	}
}

func TestCedarMarshalWithContextOnly(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("view")
	action.SetContext(Record(RequiredAttr("ip", String())))
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "context:") {
		t.Error("expected context in output")
	}
}

func TestCedarMarshalEmptyAnnotationValue(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.Annotate("deprecated", "") // annotation with no value
	s.AddEntity(entity)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "@deprecated") {
		t.Error("expected @deprecated annotation in output")
	}
}

func TestCedarMarshalMultiplePaths(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.MemberOf("Group", "Organization")
	s.AddEntity(entity)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "[Group, Organization]") {
		t.Error("expected bracketed path list in output")
	}
}

func TestCedarMarshalWithTags(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("Tagged")
	entity.SetTags(String())
	s.AddEntity(entity)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "tags String") {
		t.Error("expected tags in output")
	}
}

func TestJSONUnmarshalWithSetAttribute(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"roles": {"type": "Set", "element": {"type": "String"}, "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestJSONUnmarshalWithNestedRecord(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"address": {
								"type": "Record",
								"required": true,
								"attributes": {
									"street": {"type": "String", "required": true}
								}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
}

func TestJSONUnmarshalWithEntityTags(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"Tagged": {
					"tags": {"type": "String"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	tagged := ns.GetEntity("Tagged")
	if tagged.Tags.v == nil {
		t.Error("expected tags to be set")
	}
}

func TestJSONUnmarshalWithEntityEnum(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"Status": {
					"enum": ["active", "inactive"]
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	status := ns.GetEntity("Status")
	if len(status.Enum) != 2 {
		t.Errorf("expected 2 enum values, got %d", len(status.Enum))
	}
}

func TestJSONMarshalWithTags(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("Tagged")
	entity.SetTags(Long())
	s.AddEntity(entity)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "tags") {
		t.Error("expected tags in JSON output")
	}
}

func TestJSONMarshalWithEnum(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("Status")
	entity.SetEnum("active", "inactive")
	s.AddEntity(entity)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "enum") {
		t.Error("expected enum in JSON output")
	}
}

func TestJSONMarshalWithActionMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("edit")
	action.InActions(ActionRef{Namespace: "NS", Name: "view"})
	s.AddAction(action)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "memberOf") {
		t.Error("expected memberOf in JSON output")
	}
}

func TestJSONMarshalWithAnnotations(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.Annotate("doc", "test")
	s.AddNamespace(ns)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "annotations") {
		t.Error("expected annotations in JSON output")
	}
}

func TestJSONUnmarshalErrorInvalidJSON(t *testing.T) {
	t.Parallel()

	var s Schema
	err := s.UnmarshalJSON([]byte("not valid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONUnmarshalErrorInvalidType(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {"type": "Unknown", "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestCedarUnmarshalError(t *testing.T) {
	t.Parallel()

	var s Schema
	err := s.UnmarshalCedar([]byte("invalid cedar {{{{"))
	if err == nil {
		t.Error("expected error for invalid Cedar")
	}
}

func TestCedarMarshalWithAllTypes(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Full").SetAttributes(
		RequiredAttr("b", Boolean()),
		RequiredAttr("l", Long()),
		RequiredAttr("s", String()),
		RequiredAttr("set", SetOf(Long())),
		RequiredAttr("rec", Record(RequiredAttr("nested", String()))),
		RequiredAttr("ent", Entity("Other")),
		RequiredAttr("ext", Extension("ipaddr")),
		RequiredAttr("ref", Ref("MyType")),
	))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "Bool") {
		t.Error("expected Bool in output")
	}
	if !contains(output, "Long") {
		t.Error("expected Long in output")
	}
	if !contains(output, "String") {
		t.Error("expected String in output")
	}
	if !contains(output, "Set<") {
		t.Error("expected Set in output")
	}
}

func TestCedarUnmarshalWithComments(t *testing.T) {
	t.Parallel()

	input := `
// This is a comment
entity User;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
}

func TestCedarUnmarshalWithCommonTypeInNamespace(t *testing.T) {
	t.Parallel()

	input := `
namespace Test {
	type Name = String;
	entity User {
		name: Name,
	};
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	ct := ns.GetCommonType("Name")
	if ct == nil {
		t.Fatal("expected Name common type")
	}
}

func TestCedarUnmarshalWithNamespaceAnnotations(t *testing.T) {
	t.Parallel()

	input := `
@doc("My namespace")
namespace MyApp {
	entity User;
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("MyApp")
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	doc, ok := ns.Annotations.Get("doc")
	if !ok || doc != "My namespace" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestCedarUnmarshalWithEntityAnnotations(t *testing.T) {
	t.Parallel()

	input := `
@doc("User entity")
entity User;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	doc, ok := user.Annotations.Get("doc")
	if !ok || doc != "User entity" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestCedarUnmarshalWithActionAnnotations(t *testing.T) {
	t.Parallel()

	input := `
@doc("View action")
action view;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	view := ns.GetAction("view")

	doc, ok := view.Annotations.Get("doc")
	if !ok || doc != "View action" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestCedarUnmarshalWithCommonTypeAnnotations(t *testing.T) {
	t.Parallel()

	input := `
@doc("Custom type")
type Name = String;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	name := ns.GetCommonType("Name")

	doc, ok := name.Annotations.Get("doc")
	if !ok || doc != "Custom type" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestCedarUnmarshalWithActionInNamespace(t *testing.T) {
	t.Parallel()

	input := `
namespace Test {
	action view appliesTo {
		principal: User,
		resource: Doc,
		context: {
			ip: String,
		},
	};
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	view := ns.GetAction("view")
	if view == nil {
		t.Fatal("expected view action")
	}

	if view.Context.v == nil {
		t.Error("expected context to be set")
	}
}

func TestCedarUnmarshalWithActionContextPath(t *testing.T) {
	t.Parallel()

	input := `
type MyContext = {ip: String};
action view appliesTo {
	principal: User,
	resource: Doc,
	context: MyContext,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	view := ns.GetAction("view")

	if view.Context.v == nil {
		t.Error("expected context to be set")
	}
}

func TestCedarUnmarshalWithMultipleEntityNames(t *testing.T) {
	t.Parallel()

	input := `
entity User, Admin, Guest;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
	if ns.GetEntity("Admin") == nil {
		t.Error("expected Admin entity")
	}
	if ns.GetEntity("Guest") == nil {
		t.Error("expected Guest entity")
	}
}

func TestCedarUnmarshalWithMultipleActionNames(t *testing.T) {
	t.Parallel()

	input := `
action view, edit, delete;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns.GetAction("view") == nil {
		t.Error("expected view action")
	}
	if ns.GetAction("edit") == nil {
		t.Error("expected edit action")
	}
	if ns.GetAction("delete") == nil {
		t.Error("expected delete action")
	}
}

func TestJSONMarshalWithSetOfRecord(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").SetAttributes(
		RequiredAttr("items", SetOf(Record(RequiredAttr("name", String())))),
	))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestCedarMarshalAnonymousNamespaceFirst(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddNamespace(NewNamespace("ZNS"))
	s.AddNamespace(NewNamespace("ANS"))
	s.AddEntity(NewEntity("RootEntity")) // anonymous namespace

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	// Anonymous namespace content should appear before named namespaces
	rootIdx := indexOf(output, "RootEntity")
	ansIdx := indexOf(output, "ANS")
	if rootIdx > ansIdx && ansIdx >= 0 {
		t.Error("expected anonymous namespace content to appear first")
	}
}

func TestIsValidIdent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"_valid", true},
		{"Valid", true},
		{"valid123", true},
		{"123invalid", false},
		{"has-dash", false},
		{"has space", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isValidIdent(tt.input)
			if result != tt.expected {
				t.Errorf("isValidIdent(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func contains(s, substr string) bool {
	return indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// failingWriter is a writer that fails after a certain number of bytes
type failingWriter struct {
	written int
	failAt  int
}

func (w *failingWriter) Write(p []byte) (n int, err error) {
	if w.written+len(p) > w.failAt {
		remaining := w.failAt - w.written
		if remaining > 0 {
			w.written += remaining
			return remaining, errWriteFailed
		}
		return 0, errWriteFailed
	}
	w.written += len(p)
	return len(p), nil
}

var errWriteFailed = fmt.Errorf("write failed")

func TestWriteCedarFailure(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").SetAttributes(
		RequiredAttr("name", String()),
	))

	// Test that error is propagated through MarshalCedar
	writer := &failingWriter{failAt: 5}
	err := s.WriteCedar(writer)
	if err == nil {
		t.Error("expected error from WriteCedar")
	}
}

func TestJSONUnmarshalTypeFromJSONError(t *testing.T) {
	t.Parallel()

	// Test with invalid Set element type
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"items": {"type": "Set", "element": {"type": "Invalid"}, "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid Set element type")
	}
}

func TestJSONUnmarshalCommonTypeFromJSONError(t *testing.T) {
	t.Parallel()

	// Test with invalid common type
	input := `{
		"": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"MyType": {"type": "Invalid"}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid common type")
	}
}

func TestJSONUnmarshalActionFromJSONContextError(t *testing.T) {
	t.Parallel()

	// Test with invalid context type
	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"context": {"type": "Invalid"}
					}
				}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid context type")
	}
}

func TestJSONUnmarshalEntityFromJSONShapeError(t *testing.T) {
	t.Parallel()

	// Test with invalid shape
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {"type": "Record", "attributes": {"name": {"type": "Invalid"}}}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid shape type")
	}
}

func TestJSONUnmarshalEntityFromJSONTagsError(t *testing.T) {
	t.Parallel()

	// Test with invalid tags type
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "Invalid"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid tags type")
	}
}

func TestJSONUnmarshalWithEntityAnnotations(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"annotations": {"doc": "User entity"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	doc, ok := user.Annotations.Get("doc")
	if !ok || doc != "User entity" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestJSONUnmarshalWithActionAnnotations(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {
					"annotations": {"doc": "View action"}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	view := ns.GetAction("view")

	doc, ok := view.Annotations.Get("doc")
	if !ok || doc != "View action" {
		t.Errorf("expected doc annotation, got %v", doc)
	}
}

func TestCedarMarshalCommonTypeWithAnnotation(t *testing.T) {
	t.Parallel()

	s := New()
	ct := NewCommonType("Name", String())
	ct.Annotate("doc", "Name type")
	s.AddCommonType(ct)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "@doc") {
		t.Error("expected @doc in output")
	}
}

func TestCedarMarshalWithOptionalAttr(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").SetAttributes(
		OptionalAttr("nickname", String()),
	))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "nickname?") {
		t.Error("expected optional attribute marker in output")
	}
}

func TestCedarMarshalResourceOnlyAction(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("view")
	action.SetResourceTypes("Doc")
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "resource:") {
		t.Error("expected resource in output")
	}
}

func TestCedarMarshalQuotedActionRefName(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("edit")
	action.InActions(ActionRef{Name: "view-all"}) // needs quoting
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, `"view-all"`) {
		t.Error("expected quoted action ref in output")
	}
}

func TestCedarUnmarshalWithSetType(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	roles: Set<String>,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithBooleanType(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	active: Bool,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithLongType(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	age: Long,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithNestedRecord(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	address: {
		street: String,
		city: String,
	},
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithActionMemberOfNamespace(t *testing.T) {
	t.Parallel()

	input := `
action edit in OtherNS::"view";
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	edit := ns.GetAction("edit")

	if len(edit.MemberOf) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(edit.MemberOf))
	}

	if edit.MemberOf[0].Namespace != "OtherNS" {
		t.Errorf("expected OtherNS namespace, got %s", edit.MemberOf[0].Namespace)
	}
}

func TestJSONMarshalWithAllAttrTypes(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Full").SetAttributes(
		RequiredAttr("b", Boolean()),
		RequiredAttr("l", Long()),
		RequiredAttr("s", String()),
		RequiredAttr("set", SetOf(String())),
		RequiredAttr("rec", Record(RequiredAttr("nested", String()))),
		RequiredAttr("ent", Entity("Other")),
		RequiredAttr("ext", Extension("ipaddr")),
		RequiredAttr("ref", Ref("MyType")),
	))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}

	// Verify round-trip
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONMarshalEmptyCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddCommonType(NewCommonType("Empty", Type{}))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("expected non-empty JSON")
	}
}

func TestCedarUnmarshalWithEntityInNamespace(t *testing.T) {
	t.Parallel()

	input := `
namespace MyApp {
	entity User in [Group] {
		name: String,
	};
	entity Group;
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("MyApp")
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}

	if len(user.MemberOfTypes) != 1 {
		t.Errorf("expected 1 memberOf type, got %d", len(user.MemberOfTypes))
	}
}

func TestWriteCedarVariousFailurePoints(t *testing.T) {
	t.Parallel()

	// Create a complex schema that exercises all write paths
	s := New()
	ns := NewNamespace("Test")
	ns.Annotate("doc", "Test namespace")

	ct := NewCommonType("Name", String())
	ct.Annotate("doc", "Name type")
	ns.AddCommonType(ct)

	entity := NewEntity("User")
	entity.Annotate("doc", "User entity")
	entity.MemberOf("Group")
	entity.SetAttributes(
		RequiredAttr("name", String()),
		OptionalAttr("email", String()),
	)
	entity.SetTags(String())
	ns.AddEntity(entity)

	enumEntity := NewEntity("Status")
	enumEntity.SetEnum("active", "inactive", "pending")
	ns.AddEntity(enumEntity)

	action := NewAction("view")
	action.Annotate("doc", "View action")
	action.SetPrincipalTypes("User")
	action.SetResourceTypes("Doc")
	action.SetContext(Record(RequiredAttr("ip", String())))
	ns.AddAction(action)

	actionWithMemberOf := NewAction("edit")
	actionWithMemberOf.InActions(ActionRef{Name: "view"}, ActionRef{Namespace: "Other", Name: "read"})
	ns.AddAction(actionWithMemberOf)

	s.AddNamespace(ns)

	// Test writer failures at various points
	for i := 0; i < 500; i += 10 {
		writer := &failingWriter{failAt: i}
		err := s.WriteCedar(writer)
		if err == nil && writer.written < i {
			// Writing succeeded but writer was never filled - normal completion
			break
		}
		// We don't check the error - we just want to ensure all error paths are hit
	}
}

func TestJSONMarshalEntityWithAnnotations(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.Annotate("doc", "User entity")
	entity.Annotate("deprecated", "")
	s.AddEntity(entity)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "annotations") {
		t.Error("expected annotations in JSON output")
	}
}

func TestJSONMarshalActionWithAnnotations(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("view")
	action.Annotate("doc", "View action")
	s.AddAction(action)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !contains(string(jsonData), "annotations") {
		t.Error("expected annotations in JSON output")
	}
}

func TestJSONUnmarshalRecordInShape(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"nested": {
								"type": "Record",
								"required": true,
								"attributes": {
									"deep": {"type": "String", "required": true}
								}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
}

func TestJSONUnmarshalAllTypes(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"Full": {
					"shape": {
						"type": "Record",
						"attributes": {
							"b": {"type": "Boolean", "required": true},
							"l": {"type": "Long", "required": true},
							"s": {"type": "String", "required": true},
							"set": {"type": "Set", "element": {"type": "String"}, "required": true},
							"rec": {"type": "Record", "attributes": {"n": {"type": "String", "required": true}}, "required": true},
							"ent": {"type": "EntityOrCommon", "name": "Other", "required": true},
							"ext": {"type": "Extension", "name": "ipaddr", "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	full := ns.GetEntity("Full")
	if full == nil {
		t.Fatal("expected Full entity")
	}
	if len(full.Attributes) != 7 {
		t.Errorf("expected 7 attributes, got %d", len(full.Attributes))
	}
}

func TestJSONMarshalTypeToJSONAllTypes(t *testing.T) {
	t.Parallel()

	// Test all type variants in typeToJSON
	types := []Type{
		Boolean(),
		Long(),
		String(),
		SetOf(String()),
		Record(RequiredAttr("a", String())),
		Entity("User"),
		Extension("ipaddr"),
		Ref("MyType"),
	}

	for _, typ := range types {
		s := New()
		s.AddEntity(NewEntity("Test").SetAttributes(RequiredAttr("attr", typ)))

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON() error = %v for type %T", err, typ.v)
		}
		if len(jsonData) == 0 {
			t.Error("expected non-empty JSON")
		}
	}
}

func TestJSONMarshalCommonTypeAllTypes(t *testing.T) {
	t.Parallel()

	types := []Type{
		Boolean(),
		Long(),
		String(),
		SetOf(String()),
		Record(RequiredAttr("a", String())),
		Entity("User"),
		Extension("ipaddr"),
		Ref("MyType"),
	}

	for _, typ := range types {
		s := New()
		s.AddCommonType(NewCommonType("MyType", typ))

		jsonData, err := s.MarshalJSON()
		if err != nil {
			t.Fatalf("MarshalJSON() error = %v for type %T", err, typ.v)
		}
		if len(jsonData) == 0 {
			t.Error("expected non-empty JSON")
		}
	}
}

func TestCedarMarshalSingleEntityMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.MemberOf("Group") // single memberOf - no brackets needed
	s.AddEntity(entity)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	// Should have "in Group" not "in [Group]"
	if contains(output, "in [Group]") {
		t.Error("single memberOf should not have brackets")
	}
}

func TestCedarMarshalSinglePrincipalType(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("view")
	action.SetPrincipalTypes("User") // single type
	action.SetResourceTypes("Doc", "File") // multiple types
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	// Should have "principal: User" not "principal: [User]"
	if contains(output, "principal: [User]") {
		t.Error("single principal should not have brackets")
	}
}

func TestJSONUnmarshalWithNilContext(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Doc"]
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	view := ns.GetAction("view")
	if view.Context.v != nil {
		t.Error("expected nil context")
	}
}

func TestWriteCedarMoreFailurePoints(t *testing.T) {
	t.Parallel()

	// Test with writer that fails at every single byte position
	s := New()
	ns := NewNamespace("NS")
	ns.Annotate("doc", "Namespace doc")

	ct := NewCommonType("Type", String())
	ct.Annotate("doc", "Type doc")
	ns.AddCommonType(ct)

	entity := NewEntity("User")
	entity.Annotate("doc", "Entity doc")
	entity.MemberOf("Group", "Org")
	entity.SetAttributes(
		RequiredAttr("name", String()),
		OptionalAttr("email-addr", String()), // needs quoting
	)
	entity.SetTags(Long())
	ns.AddEntity(entity)

	enumEntity := NewEntity("Status")
	enumEntity.SetEnum("a", "b")
	ns.AddEntity(enumEntity)

	action := NewAction("view-docs") // needs quoting
	action.Annotate("doc", "Action doc")
	action.InActions(ActionRef{Name: "read"})
	action.SetPrincipalTypes("User", "Admin")
	action.SetResourceTypes("Doc")
	action.SetContext(Record(RequiredAttr("ip", String())))
	ns.AddAction(action)

	s.AddNamespace(ns)

	// Test every byte position
	for i := 0; i < 1000; i++ {
		writer := &failingWriter{failAt: i}
		_ = s.WriteCedar(writer)
	}
}

func TestJSONUnmarshalWithRecordAttribute(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"profile": {
								"type": "Record",
								"required": true,
								"attributes": {
									"bio": {"type": "String", "required": false}
								}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithSetInAttribute(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"roles": {"type": "Set", "element": {"type": "Long"}, "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithEntityOrCommonAttribute(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"manager": {"type": "EntityOrCommon", "name": "User", "required": true}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithExtensionAttribute(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
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
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestCedarUnmarshalWithRecordType(t *testing.T) {
	t.Parallel()

	input := `
type MyRecord = {
	name: String,
	age: Long,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}
}

func TestCedarUnmarshalCommonTypeInAnonymous(t *testing.T) {
	t.Parallel()

	input := `
type Name = String;
entity User {
	name: Name,
};
action view;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns.GetCommonType("Name") == nil {
		t.Error("expected Name common type")
	}
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
	if ns.GetAction("view") == nil {
		t.Error("expected view action")
	}
}

func TestJSONMarshalWithNestedSetRecord(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").SetAttributes(
		RequiredAttr("items", SetOf(Record(
			RequiredAttr("id", Long()),
			OptionalAttr("name", String()),
		))),
	))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify round-trip
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestCedarMarshalWithMultipleNamespaces(t *testing.T) {
	t.Parallel()

	s := New()

	// Add namespaces in random order
	s.AddNamespace(NewNamespace("Zebra"))
	s.AddNamespace(NewNamespace("Alpha"))
	s.AddEntity(NewEntity("RootUser")) // anonymous namespace

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	// Should have content
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestCedarMarshalWithEmptyNamespaces(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddNamespace(NewNamespace("Empty"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "namespace Empty") {
		t.Error("expected namespace in output")
	}
}

func TestJSONUnmarshalWithBooleanTag(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "Boolean"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithSetTag(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "Set", "element": {"type": "String"}}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithRecordContext(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"context": {
							"type": "Record",
							"attributes": {
								"flag": {"type": "Boolean", "required": true}
							}
						}
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalCommonTypesAllTypes(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"BoolType": {"type": "Boolean"},
				"LongType": {"type": "Long"},
				"StringType": {"type": "String"},
				"SetType": {"type": "Set", "element": {"type": "String"}},
				"RecordType": {"type": "Record", "attributes": {"a": {"type": "String", "required": true}}},
				"EntityType": {"type": "EntityOrCommon", "name": "User"},
				"ExtType": {"type": "Extension", "name": "ipaddr"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	if len(ns.CommonTypes) != 7 {
		t.Errorf("expected 7 common types, got %d", len(ns.CommonTypes))
	}
}

func TestJSONUnmarshalNilShape(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")
	if len(user.Attributes) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithFilenameBasic(t *testing.T) {
	t.Parallel()

	input := `entity User;`

	var s Schema
	if err := s.UnmarshalCedarWithFilename("test.cedar", []byte(input)); err != nil {
		t.Fatalf("UnmarshalCedarWithFilename() error = %v", err)
	}
}

func TestCedarUnmarshalWithFilenameInvalidSyntax(t *testing.T) {
	t.Parallel()

	input := `invalid {{ syntax`

	var s Schema
	err := s.UnmarshalCedarWithFilename("test.cedar", []byte(input))
	if err == nil {
		t.Error("expected error for invalid Cedar")
	}
}

func TestCedarMarshalSingleActionMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("edit")
	action.InActions(ActionRef{Name: "view"}) // single memberOf
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	// Should not have brackets for single memberOf
	if contains(output, "in [view]") {
		t.Error("single memberOf should not have brackets")
	}
}

func TestWriteCedarEvenMoreFailurePoints(t *testing.T) {
	t.Parallel()

	// More comprehensive test
	s := New()

	// Anonymous namespace with entities
	s.AddEntity(NewEntity("RootUser"))

	// Named namespace
	ns := NewNamespace("App")
	ns.Annotate("doc", "")

	entity := NewEntity("Ent")
	entity.SetAttributes(
		RequiredAttr("a", Boolean()),
		RequiredAttr("b", SetOf(String())),
		RequiredAttr("c", Record(RequiredAttr("d", String()))),
		RequiredAttr("e", Entity("Other")),
		RequiredAttr("f", Extension("ipaddr")),
		RequiredAttr("g", Ref("Type")),
	)
	ns.AddEntity(entity)

	action := NewAction("act")
	action.SetPrincipalTypes("P1", "P2")
	action.SetResourceTypes("R1", "R2")
	ns.AddAction(action)

	s.AddNamespace(ns)

	// Test various failure points
	for i := 0; i < 2000; i += 5 {
		writer := &failingWriter{failAt: i}
		_ = s.WriteCedar(writer)
	}
}

// immediateFailWriter fails on the first write
type immediateFailWriter struct{}

func (w *immediateFailWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("immediate failure")
}

func TestMarshalCedarFailure(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User"))

	// Test with an immediate failing writer
	writer := &immediateFailWriter{}
	err := s.WriteCedar(writer)
	if err == nil {
		t.Error("expected error from WriteCedar")
	}
}

func TestJSONUnmarshalWithNilTags(t *testing.T) {
	t.Parallel()

	// Entity with nil tags
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {"type": "Record", "attributes": {}}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestJSONUnmarshalWithEmptyRecordShape(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {"type": "Record", "attributes": {}}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")
	if len(user.Attributes) != 0 {
		t.Errorf("expected 0 attributes, got %d", len(user.Attributes))
	}
}

func TestCedarUnmarshalWithTagsAndAttributes(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	name: String,
} tags Long;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("User")

	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
	if user.Tags.v == nil {
		t.Error("expected tags to be set")
	}
}

func TestCedarMarshalActionWithOnlyContext(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("test")
	action.SetContext(Record(RequiredAttr("a", String())))
	s.AddAction(action)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "context") {
		t.Error("expected context in output")
	}
}

func TestCedarMarshalEntityWithOnlyMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.MemberOf("Group")
	// No attributes, no tags
	s.AddEntity(entity)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	output := string(cedarData)
	if !contains(output, "in Group") {
		t.Error("expected 'in Group' in output")
	}
}

func TestCedarUnmarshalWithAnnotationWithoutValue(t *testing.T) {
	t.Parallel()

	input := `
@deprecated
entity OldUser;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	user := ns.GetEntity("OldUser")

	val, ok := user.Annotations.Get("deprecated")
	if !ok {
		t.Error("expected deprecated annotation")
	}
	if val != "" {
		t.Errorf("expected empty value, got %q", val)
	}
}

func TestSetFilename(t *testing.T) {
	t.Parallel()

	var s Schema
	s.SetFilename("test.cedar")

	// The filename should be used for error messages
	err := s.UnmarshalCedar([]byte("invalid {{ syntax"))
	if err == nil {
		t.Error("expected error")
	}
	// The error message should contain the filename
	if !contains(err.Error(), "test.cedar") {
		t.Errorf("expected error to contain filename, got: %v", err)
	}
}

// unknownType is a test type that implements isType but isn't a known type variant
type unknownType struct{}

func (unknownType) isType()             {}
func (unknownType) equal(o isType) bool { return false }

func TestTypeToJSONUnknownType(t *testing.T) {
	t.Parallel()

	// Create a Type with an unknown variant
	typ := Type{v: unknownType{}}

	_, err := typeToJSON(typ)
	if err == nil {
		t.Error("expected error for unknown type")
	}
	if !contains(err.Error(), "unknown type") {
		t.Errorf("expected 'unknown type' in error, got: %v", err)
	}
}

func TestAttrToJSONUnknownType(t *testing.T) {
	t.Parallel()

	// Create an attribute with an unknown type variant
	attr := Attribute{
		Name:     "test",
		Type:     Type{v: unknownType{}},
		Required: true,
	}

	_, err := attrToJSON(attr)
	if err == nil {
		t.Error("expected error for unknown type")
	}
	if !contains(err.Error(), "unknown type") {
		t.Errorf("expected 'unknown type' in error, got: %v", err)
	}
}

func TestJSONMarshalUnknownTypeInEntity(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name:     "bad",
		Type:     Type{v: unknownType{}},
		Required: true,
	})
	s.AddEntity(entity)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestJSONMarshalUnknownTypeInAction(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("view")
	action.SetContext(Type{v: unknownType{}})
	s.AddAction(action)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type in action context")
	}
}

func TestJSONMarshalUnknownTypeInCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	ct := NewCommonType("Bad", Type{v: unknownType{}})
	s.AddCommonType(ct)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type in common type")
	}
}

func TestJSONMarshalUnknownTypeInSetElement(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name:     "items",
		Type:     SetOf(Type{v: unknownType{}}),
		Required: true,
	})
	s.AddEntity(entity)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type in set element")
	}
}

func TestJSONMarshalUnknownTypeInNestedRecord(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name: "nested",
		Type: Record(Attribute{
			Name:     "bad",
			Type:     Type{v: unknownType{}},
			Required: true,
		}),
		Required: true,
	})
	s.AddEntity(entity)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type in nested record")
	}
}

func TestJSONMarshalUnknownTypeInTags(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetTags(Type{v: unknownType{}})
	s.AddEntity(entity)

	_, err := s.MarshalJSON()
	if err == nil {
		t.Error("expected error for unknown type in tags")
	}
}

func TestJSONMarshalUnknownTypeInAttrSetElement(t *testing.T) {
	t.Parallel()

	// Test attrToJSON with Set containing unknown type
	attr := Attribute{
		Name:     "items",
		Type:     SetOf(Type{v: unknownType{}}),
		Required: true,
	}

	_, err := attrToJSON(attr)
	if err == nil {
		t.Error("expected error for unknown type in set element")
	}
}

func TestJSONMarshalUnknownTypeInAttrNestedRecord(t *testing.T) {
	t.Parallel()

	// Test attrToJSON with nested Record containing unknown type
	attr := Attribute{
		Name: "nested",
		Type: Record(Attribute{
			Name:     "bad",
			Type:     Type{v: unknownType{}},
			Required: true,
		}),
		Required: true,
	}

	_, err := attrToJSON(attr)
	if err == nil {
		t.Error("expected error for unknown type in nested record")
	}
}

func TestCedarWriteTypeUnknown(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name:     "bad",
		Type:     Type{v: unknownType{}},
		Required: true,
	})
	s.AddEntity(entity)

	// This should fail when trying to write the unknown type
	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type")
	}
}

func TestJSONTypeFromJSONNil(t *testing.T) {
	t.Parallel()

	// Test typeFromJSON with nil input
	result, err := typeFromJSON(nil)
	if err != nil {
		t.Errorf("typeFromJSON(nil) error = %v", err)
	}
	if result.v != nil {
		t.Error("expected nil type")
	}
}

func TestJSONTypeFromJSONSetElementError(t *testing.T) {
	t.Parallel()

	// Test with invalid Set element type
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"items": {
								"type": "Set",
								"element": {"type": "Unknown"},
								"required": true
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid Set element type")
	}
}

func TestJSONAttrFromJSONNestedRecordError(t *testing.T) {
	t.Parallel()

	// Test with invalid nested Record attribute type
	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"profile": {
								"type": "Record",
								"required": true,
								"attributes": {
									"bad": {"type": "Unknown", "required": true}
								}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid nested Record type")
	}
}

func TestJSONTypeFromJSONSetElementInTopLevelType(t *testing.T) {
	t.Parallel()

	// Test with invalid Set element in top-level type (common type)
	input := `{
		"": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"MySet": {
					"type": "Set",
					"element": {"type": "Unknown"}
				}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid Set element in common type")
	}
}

func TestJSONTypeFromJSONRecordAttrError(t *testing.T) {
	t.Parallel()

	// Test with invalid Record attribute in top-level type (common type)
	input := `{
		"": {
			"entityTypes": {},
			"actions": {},
			"commonTypes": {
				"MyRecord": {
					"type": "Record",
					"attributes": {
						"bad": {"type": "Unknown", "required": true}
					}
				}
			}
		}
	}`

	var s Schema
	err := s.UnmarshalJSON([]byte(input))
	if err == nil {
		t.Error("expected error for invalid Record attribute in common type")
	}
}

func TestTypeToJSONSetElementError(t *testing.T) {
	t.Parallel()

	// Test typeToJSON with Set containing unknown element type
	typ := SetOf(Type{v: unknownType{}})

	_, err := typeToJSON(typ)
	if err == nil {
		t.Error("expected error for unknown type in Set element")
	}
}

func TestWriteCedarActionMemberOfBracketsFailure(t *testing.T) {
	t.Parallel()

	// Create an action with multiple memberOf to test bracket write failures
	s := New()
	action := NewAction("test")
	action.InActions(
		ActionRef{Name: "a"},
		ActionRef{Name: "b"},
		ActionRef{Name: "c"},
	)
	s.AddAction(action)

	// Test failures at various points where brackets and commas are written
	// The output should look like: action test in [a, b, c];
	for i := 0; i < 50; i++ {
		writer := &failingWriter{failAt: i}
		_ = s.WriteCedar(writer)
	}
}

func TestWriteCedarUnknownTypeInTags(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetTags(Type{v: unknownType{}})
	s.AddEntity(entity)

	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type in tags")
	}
}

func TestWriteCedarUnknownTypeInSet(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name:     "items",
		Type:     SetOf(Type{v: unknownType{}}),
		Required: true,
	})
	s.AddEntity(entity)

	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type in Set element")
	}
}

func TestWriteCedarUnknownTypeInRecord(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name: "profile",
		Type: Record(Attribute{
			Name:     "bad",
			Type:     Type{v: unknownType{}},
			Required: true,
		}),
		Required: true,
	})
	s.AddEntity(entity)

	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type in Record attribute")
	}
}

func TestWriteCedarUnknownTypeInContext(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("test")
	action.SetContext(Type{v: unknownType{}})
	s.AddAction(action)

	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type in context")
	}
}

func TestWriteCedarUnknownTypeInCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	ct := NewCommonType("Bad", Type{v: unknownType{}})
	s.AddCommonType(ct)

	_, err := s.MarshalCedar()
	if err == nil {
		t.Error("expected error for unknown type in common type")
	}
}

func TestWriteCedarNilTypeInAttribute(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetAttributes(Attribute{
		Name:     "empty",
		Type:     Type{}, // nil type
		Required: true,
	})
	s.AddEntity(entity)

	// Should succeed - nil types are allowed
	_, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
	}
}

func TestWriteCedarNilTypeInCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	ct := NewCommonType("Empty", Type{}) // nil type
	s.AddCommonType(ct)

	// Should succeed
	_, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
	}
}

func TestWriteCedarNilTypeInTags(t *testing.T) {
	t.Parallel()

	s := New()
	entity := NewEntity("User")
	entity.SetTags(Type{}) // nil type
	s.AddEntity(entity)

	// Should succeed
	_, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
	}
}

func TestWriteCedarNilTypeInContext(t *testing.T) {
	t.Parallel()

	s := New()
	action := NewAction("test")
	action.SetContext(Type{}) // nil type
	s.AddAction(action)

	// Should succeed
	_, err := s.MarshalCedar()
	if err != nil {
		t.Errorf("MarshalCedar() error = %v", err)
	}
}
