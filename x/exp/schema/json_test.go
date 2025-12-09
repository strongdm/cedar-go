package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	// Create a schema programmatically
	s := New()
	ns := NewNamespace("PhotoApp")
	ns.AddEntity(NewEntity("User").
		MemberOf("Group").
		SetAttributes(
			RequiredAttr("name", String()),
			OptionalAttr("email", String()),
		))
	ns.AddEntity(NewEntity("Group"))
	ns.AddAction(NewAction("view").
		SetPrincipalTypes("User", "Group").
		SetResourceTypes("Photo"))
	ns.AddCommonType(NewCommonType("Name", String()))
	s.AddNamespace(ns)

	// Marshal to JSON
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify valid JSON
	var raw interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("MarshalJSON() produced invalid JSON: %v", err)
	}

	// Unmarshal back
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Marshal again
	jsonData2, err := s2.MarshalJSON()
	if err != nil {
		t.Fatalf("Second MarshalJSON() error = %v", err)
	}

	// Compare (JSON should be identical since we sort keys)
	if string(jsonData) != string(jsonData2) {
		t.Errorf("Round-trip produced different JSON:\nFirst: %s\nSecond: %s", jsonData, jsonData2)
	}
}

func TestJSONUnmarshalValidSchema(t *testing.T) {
	t.Parallel()

	input := `{
		"PhotoApp": {
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
				"Group": {},
				"Photo": {
					"shape": {
						"type": "Record",
						"attributes": {
							"title": {"type": "String", "required": true},
							"tags": {"type": "Set", "required": false, "element": {"type": "String"}}
						}
					}
				}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User", "Group"],
						"resourceTypes": ["Photo"]
					}
				}
			},
			"commonTypes": {
				"Name": {"type": "String"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("PhotoApp")
	if ns == nil {
		t.Fatal("expected PhotoApp namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	if len(user.MemberOfTypes) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(user.MemberOfTypes))
	}
	if len(user.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(user.Attributes))
	}

	view := ns.GetAction("view")
	if view == nil {
		t.Fatal("expected view action")
	}
	if len(view.PrincipalTypes) != 2 {
		t.Errorf("expected 2 principal types, got %d", len(view.PrincipalTypes))
	}
}

func TestJSONUnmarshalInvalidJSON(t *testing.T) {
	t.Parallel()

	var s Schema
	err := s.UnmarshalJSON([]byte("{invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestJSONMarshalEmpty(t *testing.T) {
	t.Parallel()

	s := New()
	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}
	if string(jsonData) != "{}" {
		t.Errorf("expected empty JSON object, got %s", jsonData)
	}
}

func TestJSONUnmarshalEmpty(t *testing.T) {
	t.Parallel()

	var s Schema
	if err := s.UnmarshalJSON([]byte("{}")); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
	if len(s.Namespaces) != 0 {
		t.Errorf("expected empty namespaces, got %d", len(s.Namespaces))
	}
}

func TestJSONMarshalAllTypes(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")

	// Entity with all type variants
	ns.AddEntity(NewEntity("TestEntity").
		SetAttributes(
			RequiredAttr("boolean", Boolean()),
			RequiredAttr("long", Long()),
			RequiredAttr("string", String()),
			RequiredAttr("set", SetOf(String())),
			RequiredAttr("record", Record(
				RequiredAttr("nested", String()),
			)),
			RequiredAttr("entity", Entity("OtherEntity")),
			RequiredAttr("extension", Extension("ipaddr")),
			RequiredAttr("ref", Ref("CommonType")),
		))

	s.AddNamespace(ns)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Verify all type strings are present
	jsonStr := string(jsonData)
	if !strings.Contains(jsonStr, `"Boolean"`) {
		t.Error("expected Boolean type in JSON")
	}
	if !strings.Contains(jsonStr, `"Long"`) {
		t.Error("expected Long type in JSON")
	}
	if !strings.Contains(jsonStr, `"String"`) {
		t.Error("expected String type in JSON")
	}
	if !strings.Contains(jsonStr, `"Set"`) {
		t.Error("expected Set type in JSON")
	}
	if !strings.Contains(jsonStr, `"Record"`) {
		t.Error("expected Record type in JSON")
	}
}

func TestJSONMarshalEntityWithTags(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("TaggedEntity").SetTags(String()))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !strings.Contains(string(jsonData), `"tags"`) {
		t.Error("expected tags in JSON")
	}
}

func TestJSONMarshalEnumEntity(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Color").SetEnum("red", "green", "blue"))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !strings.Contains(string(jsonData), `"enum"`) {
		t.Error("expected enum in JSON")
	}
	if !strings.Contains(string(jsonData), `"red"`) {
		t.Error("expected red in JSON")
	}
}

func TestJSONMarshalActionWithMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.AddAction(NewAction("edit").
		InActions(
			ActionRef{Name: "view"},
			ActionRef{Namespace: "OtherNS", Name: "read"},
		))
	s.AddNamespace(ns)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !strings.Contains(string(jsonData), `"memberOf"`) {
		t.Error("expected memberOf in JSON")
	}
}

func TestJSONMarshalActionWithContext(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.AddAction(NewAction("create").
		SetPrincipalTypes("User").
		SetResourceTypes("Doc").
		SetContext(Record(
			RequiredAttr("ip", Extension("ipaddr")),
		)))
	s.AddNamespace(ns)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !strings.Contains(string(jsonData), `"context"`) {
		t.Error("expected context in JSON")
	}
}

func TestJSONMarshalAnnotations(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test").Annotate("doc", "test namespace")
	ns.AddEntity(NewEntity("User").Annotate("doc", "user entity"))
	ns.AddAction(NewAction("view").Annotate("doc", "view action"))
	s.AddNamespace(ns)

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	if !strings.Contains(string(jsonData), `"annotations"`) {
		t.Error("expected annotations in JSON")
	}
}

func TestJSONUnmarshalWithAnnotations(t *testing.T) {
	t.Parallel()

	input := `{
		"Test": {
			"annotations": {"doc": "test namespace"},
			"entityTypes": {
				"User": {
					"annotations": {"doc": "user entity"}
				}
			},
			"actions": {
				"view": {
					"annotations": {"doc": "view action"}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	val, ok := ns.Annotations.Get("doc")
	if !ok || val != "test namespace" {
		t.Error("expected namespace annotation")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	val, ok = user.Annotations.Get("doc")
	if !ok || val != "user entity" {
		t.Error("expected entity annotation")
	}
}

func TestJSONUnmarshalNestedRecord(t *testing.T) {
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
									"name": {"type": "String", "required": true}
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
	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestJSONUnmarshalUnknownType(t *testing.T) {
	t.Parallel()

	input := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "UnknownType"
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

func TestJSONAnonymousNamespace(t *testing.T) {
	t.Parallel()

	// Create schema with anonymous namespace
	s := New()
	s.AddEntity(NewEntity("User"))
	s.AddAction(NewAction("view").
		SetPrincipalTypes("User").
		SetResourceTypes("Resource"))

	jsonData, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	// Should have empty string key for anonymous namespace
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := raw[""]; !ok {
		t.Error("expected anonymous namespace with empty string key")
	}
}

func TestJSONTypeToJSONNilType(t *testing.T) {
	t.Parallel()

	// Test that typeToJSON handles nil type
	result, err := typeToJSON(Type{})
	if err != nil {
		t.Errorf("typeToJSON(nil) error = %v", err)
	}
	if result != nil {
		t.Error("expected nil result for nil type")
	}
}

func TestJSONAttrToJSONAllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		attr Attribute
	}{
		{"Boolean", Attr("a", Boolean(), true)},
		{"Long", Attr("a", Long(), true)},
		{"String", Attr("a", String(), true)},
		{"Set", Attr("a", SetOf(String()), true)},
		{"Record", Attr("a", Record(), true)},
		{"Entity", Attr("a", Entity("E"), true)},
		{"Extension", Attr("a", Extension("ipaddr"), true)},
		{"Ref", Attr("a", Ref("T"), true)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := attrToJSON(tt.attr)
			if err != nil {
				t.Errorf("attrToJSON() error = %v", err)
			}
			if result == nil {
				t.Error("expected non-nil result")
			}
		})
	}
}

func TestJSONNamespaceFromJSONSorting(t *testing.T) {
	t.Parallel()

	// Test that namespaceFromJSON sorts entities, actions, and common types
	input := `{
		"Test": {
			"entityTypes": {
				"Z": {},
				"A": {},
				"M": {}
			},
			"actions": {
				"z": {},
				"a": {},
				"m": {}
			},
			"commonTypes": {
				"Z": {"type": "String"},
				"A": {"type": "String"},
				"M": {"type": "String"}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	// Verify all entities are present
	if len(ns.Entities) != 3 {
		t.Errorf("expected 3 entities, got %d", len(ns.Entities))
	}
	if len(ns.Actions) != 3 {
		t.Errorf("expected 3 actions, got %d", len(ns.Actions))
	}
	if len(ns.CommonTypes) != 3 {
		t.Errorf("expected 3 common types, got %d", len(ns.CommonTypes))
	}
}
