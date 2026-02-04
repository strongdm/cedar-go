package schema_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// TestCycleDetection tests that cycles in common type definitions are detected.
// Ported from Rust: cedar-policy-core/src/validator/schema.rs test_cycles()
func TestCycleDetection(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "self_reference",
			schema: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "a"}
					}
				}
			}`,
		},
		{
			name: "two_node_loop",
			schema: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "b"},
						"b": {"type": "a"}
					}
				}
			}`,
		},
		{
			name: "three_node_loop",
			schema: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "b"},
						"b": {"type": "c"},
						"c": {"type": "a"}
					}
				}
			}`,
		},
		{
			name: "cross_namespace_two_node_loop",
			schema: `{
				"A": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "B::a"}
					}
				},
				"B": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "A::a"}
					}
				}
			}`,
		},
		{
			name: "cross_namespace_three_node_loop",
			schema: `{
				"A": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "B::a"}
					}
				},
				"B": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "C::a"}
					}
				},
				"C": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "A::a"}
					}
				}
			}`,
		},
		{
			name: "cross_namespace_indirect_loop",
			schema: `{
				"A": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "B::a"}
					}
				},
				"B": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"a": {"type": "c"},
						"c": {"type": "A::a"}
					}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			if err := json.Unmarshal([]byte(tt.schema), &s); err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			_, err := s.Resolve()
			if err == nil {
				t.Fatal("expected cycle error, got nil")
			}

			if !errors.Is(err, schema.ErrCycle) {
				t.Errorf("expected ErrCycle, got %T: %v", err, err)
			}
		})
	}
}

// TestShadowingValidation tests that shadowing of types across namespaces is detected.
// Ported from Rust: cedar-policy-core/src/validator/schema.rs common_common_conflict(), entity_entity_conflict(), common_entity_conflict()
func TestShadowingValidation(t *testing.T) {
	tests := []struct {
		name   string
		schema string
	}{
		{
			name: "common_common_conflict",
			schema: `{
				"": {
					"commonTypes": {"T": {"type": "String"}},
					"entityTypes": {},
					"actions": {}
				},
				"NS": {
					"commonTypes": {"T": {"type": "String"}},
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"t": {"type": "T"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
		},
		{
			name: "entity_entity_conflict",
			schema: `{
				"": {
					"entityTypes": {
						"T": {
							"memberOfTypes": ["T"],
							"shape": {
								"type": "Record",
								"attributes": {
									"foo": {"type": "String"}
								}
							}
						}
					},
					"actions": {}
				},
				"NS": {
					"entityTypes": {
						"T": {
							"shape": {
								"type": "Record",
								"attributes": {
									"bar": {"type": "String"}
								}
							}
						},
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"t": {"type": "Entity", "name": "T"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
		},
		{
			name: "common_entity_conflict",
			schema: `{
				"": {
					"entityTypes": {
						"T": {
							"memberOfTypes": ["T"],
							"shape": {
								"type": "Record",
								"attributes": {
									"foo": {"type": "String"}
								}
							}
						}
					},
					"actions": {}
				},
				"NS": {
					"commonTypes": {"T": {"type": "String"}},
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"t": {"type": "T"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			if err := json.Unmarshal([]byte(tt.schema), &s); err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			_, err := s.Resolve()
			if err == nil {
				t.Fatal("expected shadow error, got nil")
			}

			if !errors.Is(err, schema.ErrShadow) {
				t.Errorf("expected ErrShadow, got %T: %v", err, err)
			}
		})
	}
}

// TestUndefinedTypeErrors tests that references to undefined types produce proper errors.
// Ported from Rust: cedar-policy-core/src/validator/schema.rs test_from_schema_file_undefined_*
// Note: Go implementation reports first error encountered, unlike Rust which collects all errors.
func TestUndefinedTypeErrors(t *testing.T) {
	tests := []struct {
		name          string
		schema        string
		expectedInErr string // at least one of these should be in the error
	}{
		{
			name: "undefined_types_in_common",
			schema: `{
				"": {
					"commonTypes": {
						"My1": {"type": "What"}
					},
					"entityTypes": {"Test": {}},
					"actions": {}
				}
			}`,
			expectedInErr: "What",
		},
		{
			name: "undefined_entities_in_applies_to",
			schema: `{
				"": {
					"entityTypes": {"Test": {}},
					"actions": {
						"doTests": {
							"appliesTo": {
								"principalTypes": ["Usr"],
								"resourceTypes": ["Test"]
							}
						}
					}
				}
			}`,
			expectedInErr: "Usr",
		},
		{
			name: "undefined_entity_in_member_of",
			schema: `{
				"": {
					"entityTypes": {
						"User": {"memberOfTypes": ["Grop"]},
						"Group": {},
						"Photo": {}
					},
					"actions": {
						"view_photo": {
							"appliesTo": {
								"principalTypes": ["User"],
								"resourceTypes": ["Photo"]
							}
						}
					}
				}
			}`,
			expectedInErr: "Grop",
		},
		{
			name: "cross_namespace_undefined_member_of",
			schema: `{
				"Foo": {
					"entityTypes": {
						"User": {"memberOfTypes": ["Foo::Group", "Bar::Group"]},
						"Group": {}
					},
					"actions": {}
				}
			}`,
			expectedInErr: "Bar::Group",
		},
		{
			name: "cross_namespace_undefined_applies_to",
			schema: `{
				"Foo": {
					"entityTypes": {"User": {}, "Photo": {}},
					"actions": {
						"view_photo": {
							"appliesTo": {
								"principalTypes": ["Bar::User"],
								"resourceTypes": ["Photo"]
							}
						}
					}
				}
			}`,
			expectedInErr: "Bar::User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			if err := json.Unmarshal([]byte(tt.schema), &s); err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			_, err := s.Resolve()
			if err == nil {
				t.Fatal("expected undefined type error, got nil")
			}

			if !errors.Is(err, schema.ErrUndefinedType) {
				t.Errorf("expected ErrUndefinedType, got %T: %v", err, err)
				return
			}

			// Check that the error message mentions the expected undefined type
			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.expectedInErr) {
				t.Errorf("expected error to mention %q, got: %v", tt.expectedInErr, errMsg)
			}
		})
	}
}

// TestReservedNames tests that reserved names like __cedar are rejected.
// Ported from Rust: cedar-policy-core/src/validator/schema.rs reserved_namespace()
func TestReservedNames(t *testing.T) {
	tests := []struct {
		name      string
		schema    string
		wantError bool
	}{
		{
			name: "reserved_namespace_cedar",
			schema: `{
				"__cedar": {
					"commonTypes": {},
					"entityTypes": {},
					"actions": {}
				}
			}`,
			wantError: true,
		},
		{
			name: "reserved_namespace_cedar_prefix",
			schema: `{
				"__cedar::A": {
					"commonTypes": {},
					"entityTypes": {},
					"actions": {}
				}
			}`,
			wantError: true,
		},
		{
			name: "reserved_common_type_name",
			schema: `{
				"": {
					"commonTypes": {
						"__cedar": {"type": "String"}
					},
					"entityTypes": {},
					"actions": {}
				}
			}`,
			wantError: true,
		},
		{
			name: "reserved_type_reference",
			schema: `{
				"": {
					"commonTypes": {
						"A": {"type": "__cedar"}
					},
					"entityTypes": {},
					"actions": {}
				}
			}`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			err := json.Unmarshal([]byte(tt.schema), &s)

			// Either parse error or resolve error is acceptable
			if err == nil {
				_, err = s.Resolve()
			}

			if tt.wantError && err == nil {
				t.Error("expected error for reserved name, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestResolutionPriority tests that common types take precedence over entity types.
// Per RFC 24: when a name could be either a common type or entity type, common type wins.
func TestResolutionPriority(t *testing.T) {
	// This schema has both a common type "MyType" and an entity type "MyType" in NS1.
	// When referenced without qualification inside NS1, the common type should win.
	cedarSchema := `
namespace NS1 {
	type MyType = { inner: String };
	entity MyType { different: Long };
	entity User { data: MyType };
}
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces["NS1"]
	if ns == nil {
		t.Fatal("expected NS1 namespace")
	}

	// Get the User entity type
	userType := types.EntityType("NS1::User")
	user := ns.EntityTypes[userType]
	if user == nil {
		t.Fatalf("expected User entity type")
	}

	// The "data" attribute should be a Record type (from common type), not an Entity reference
	dataAttr := user.Shape.Attributes["data"]
	if dataAttr == nil {
		t.Fatal("expected data attribute on User")
	}

	// If common type took precedence, dataAttr.Type should be a *resolved.RecordType
	// If entity type won, it would be a resolved.EntityRef
	if _, ok := dataAttr.Type.(*resolved.RecordType); !ok {
		t.Errorf("expected common type (Record) to take precedence, got %T", dataAttr.Type)
	}
}

// TestNamespaceQualification tests that type names are properly qualified during resolution.
func TestNamespaceQualification(t *testing.T) {
	tests := []struct {
		name               string
		schema             string
		checkNS            string
		checkEntityType    string
		expectedMemberOf   []types.EntityType
		expectedPrincipals []types.EntityType
	}{
		{
			name: "unqualified_resolves_to_current_namespace",
			schema: `{
				"MyApp": {
					"entityTypes": {
						"User": {"memberOfTypes": ["Group"]},
						"Group": {}
					},
					"actions": {
						"view": {
							"appliesTo": {
								"principalTypes": ["User"],
								"resourceTypes": ["Group"]
							}
						}
					}
				}
			}`,
			checkNS:          "MyApp",
			checkEntityType:  "MyApp::User",
			expectedMemberOf: []types.EntityType{"MyApp::Group"},
		},
		{
			name: "qualified_cross_namespace_reference",
			schema: `{
				"NS1": {
					"entityTypes": {"Group": {}},
					"actions": {}
				},
				"NS2": {
					"entityTypes": {
						"User": {"memberOfTypes": ["NS1::Group"]}
					},
					"actions": {}
				}
			}`,
			checkNS:          "NS2",
			checkEntityType:  "NS2::User",
			expectedMemberOf: []types.EntityType{"NS1::Group"},
		},
		{
			name: "empty_namespace_no_prefix",
			schema: `{
				"": {
					"entityTypes": {
						"User": {"memberOfTypes": ["Group"]},
						"Group": {}
					},
					"actions": {}
				}
			}`,
			checkNS:          "",
			checkEntityType:  "User",
			expectedMemberOf: []types.EntityType{"Group"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s schema.Schema
			if err := json.Unmarshal([]byte(tt.schema), &s); err != nil {
				t.Fatalf("failed to parse schema: %v", err)
			}

			rs, err := s.Resolve()
			if err != nil {
				t.Fatalf("failed to resolve schema: %v", err)
			}

			ns := rs.Namespaces[types.Path(tt.checkNS)]
			if ns == nil {
				t.Fatalf("expected namespace %q", tt.checkNS)
			}

			entityType := types.EntityType(tt.checkEntityType)
			entity := ns.EntityTypes[entityType]
			if entity == nil {
				t.Fatalf("expected entity type %q", tt.checkEntityType)
			}

			if len(tt.expectedMemberOf) > 0 {
				if len(entity.MemberOfTypes) != len(tt.expectedMemberOf) {
					t.Errorf("expected %d memberOf types, got %d", len(tt.expectedMemberOf), len(entity.MemberOfTypes))
				}
				for i, expected := range tt.expectedMemberOf {
					if i < len(entity.MemberOfTypes) && entity.MemberOfTypes[i] != expected {
						t.Errorf("memberOf[%d]: expected %q, got %q", i, expected, entity.MemberOfTypes[i])
					}
				}
			}
		})
	}
}

// TestActionResolution tests that actions are properly resolved to EntityUIDs.
func TestActionResolution(t *testing.T) {
	schemaJSON := `{
		"MyApp": {
			"entityTypes": {"User": {}, "Document": {}},
			"actions": {
				"read": {},
				"write": {
					"memberOf": [{"id": "read"}]
				}
			}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	// Check that actions are keyed by EntityUID
	readUID := types.NewEntityUID("MyApp::Action", "read")
	writeUID := types.NewEntityUID("MyApp::Action", "write")

	if ns.Actions[readUID] == nil {
		t.Errorf("expected action with key %v", readUID)
	}

	writeAction := ns.Actions[writeUID]
	if writeAction == nil {
		t.Fatalf("expected action with key %v", writeUID)
	}

	// Check that memberOf is resolved to EntityUID
	if len(writeAction.MemberOf) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(writeAction.MemberOf))
	} else if writeAction.MemberOf[0] != readUID {
		t.Errorf("expected memberOf %v, got %v", readUID, writeAction.MemberOf[0])
	}
}

// TestExtensionTypes tests that extension types are properly resolved.
func TestExtensionTypes(t *testing.T) {
	schemaJSON := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ip": {"type": "Extension", "name": "ipaddr"},
							"balance": {"type": "Extension", "name": "decimal"},
							"created": {"type": "Extension", "name": "datetime"},
							"timeout": {"type": "Extension", "name": "duration"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces[""]
	user := ns.EntityTypes[types.EntityType("User")]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	expectedExtensions := map[string]string{
		"ip":      "ipaddr",
		"balance": "decimal",
		"created": "datetime",
		"timeout": "duration",
	}

	for attrName, expectedExt := range expectedExtensions {
		attr := user.Shape.Attributes[attrName]
		if attr == nil {
			t.Errorf("expected %s attribute", attrName)
			continue
		}

		ext, ok := attr.Type.(resolved.Extension)
		if !ok {
			t.Errorf("%s: expected resolved.Extension, got %T", attrName, attr.Type)
			continue
		}

		if ext.Name != expectedExt {
			t.Errorf("%s: expected extension %q, got %q", attrName, expectedExt, ext.Name)
		}
	}
}

// TestCommonTypeResolution tests that common type references are properly inlined.
func TestCommonTypeResolution(t *testing.T) {
	schemaJSON := `{
		"": {
			"commonTypes": {
				"Address": {
					"type": "Record",
					"attributes": {
						"street": {"type": "String"},
						"city": {"type": "String"}
					}
				}
			},
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"home": {"type": "Address"},
							"work": {"type": "Address"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces[""]
	user := ns.EntityTypes[types.EntityType("User")]
	if user == nil {
		t.Fatal("expected User entity type")
	}

	// Check that home and work are both resolved to Record types
	for _, attrName := range []string{"home", "work"} {
		attr := user.Shape.Attributes[attrName]
		if attr == nil {
			t.Errorf("expected %s attribute", attrName)
			continue
		}

		rec, ok := attr.Type.(*resolved.RecordType)
		if !ok {
			t.Errorf("%s: expected *resolved.RecordType, got %T", attrName, attr.Type)
			continue
		}

		if len(rec.Attributes) != 2 {
			t.Errorf("%s: expected 2 attributes, got %d", attrName, len(rec.Attributes))
		}

		if rec.Attributes["street"] == nil || rec.Attributes["city"] == nil {
			t.Errorf("%s: expected street and city attributes", attrName)
		}
	}
}

// TestNestedCommonTypes tests resolution of common types that reference other common types.
func TestNestedCommonTypes(t *testing.T) {
	schemaJSON := `{
		"": {
			"commonTypes": {
				"Inner": {
					"type": "Record",
					"attributes": {
						"value": {"type": "Long"}
					}
				},
				"Outer": {
					"type": "Record",
					"attributes": {
						"nested": {"type": "Inner"}
					}
				}
			},
			"entityTypes": {
				"Test": {
					"shape": {
						"type": "Record",
						"attributes": {
							"data": {"type": "Outer"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces[""]
	test := ns.EntityTypes[types.EntityType("Test")]
	if test == nil {
		t.Fatal("expected Test entity type")
	}

	dataAttr := test.Shape.Attributes["data"]
	if dataAttr == nil {
		t.Fatal("expected data attribute")
	}

	outer, ok := dataAttr.Type.(*resolved.RecordType)
	if !ok {
		t.Fatalf("expected outer to be *resolved.RecordType, got %T", dataAttr.Type)
	}

	nestedAttr := outer.Attributes["nested"]
	if nestedAttr == nil {
		t.Fatal("expected nested attribute")
	}

	inner, ok := nestedAttr.Type.(*resolved.RecordType)
	if !ok {
		t.Fatalf("expected inner to be *resolved.RecordType, got %T", nestedAttr.Type)
	}

	if inner.Attributes["value"] == nil {
		t.Error("expected value attribute in inner record")
	}
}

// TestEnumTypeResolution tests that enum types are properly resolved.
func TestEnumTypeResolution(t *testing.T) {
	cedarSchema := `
namespace MyApp {
	entity Status enum ["Active", "Inactive", "Pending"];
	entity User;
	action view appliesTo { principal: [User], resource: [Status] };
}
`

	var s schema.Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected MyApp namespace")
	}

	// Check that Status is in EnumTypes, not EntityTypes
	statusType := types.EntityType("MyApp::Status")
	status := ns.EnumTypes[statusType]
	if status == nil {
		t.Fatalf("expected Status enum type with key %q", statusType)
	}

	if len(status.Values) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(status.Values))
	}

	expected := []string{"Active", "Inactive", "Pending"}
	for i, v := range expected {
		if status.Values[i] != v {
			t.Errorf("enum value %d: expected %q, got %q", i, v, status.Values[i])
		}
	}

	// Check that User is in EntityTypes
	userType := types.EntityType("MyApp::User")
	if ns.EntityTypes[userType] == nil {
		t.Errorf("expected User entity type with key %q", userType)
	}

	// Check that the action's resource types can include enum types
	viewUID := types.NewEntityUID("MyApp::Action", "view")
	view := ns.Actions[viewUID]
	if view == nil {
		t.Fatalf("expected view action")
	}

	// Status should be resolvable as a resource type (enum types are entity types for resolution)
	foundStatus := false
	for _, rt := range view.ResourceTypes {
		if rt == statusType {
			foundStatus = true
			break
		}
	}
	if !foundStatus {
		t.Errorf("expected Status in resource types, got %v", view.ResourceTypes)
	}
}

// TestSetOfEntityType tests resolution of Set<EntityType>.
func TestSetOfEntityType(t *testing.T) {
	schemaJSON := `{
		"MyApp": {
			"entityTypes": {
				"User": {},
				"Team": {
					"shape": {
						"type": "Record",
						"attributes": {
							"members": {
								"type": "Set",
								"element": {"type": "Entity", "name": "User"}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s schema.Schema
	if err := json.Unmarshal([]byte(schemaJSON), &s); err != nil {
		t.Fatalf("failed to parse schema: %v", err)
	}

	rs, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve schema: %v", err)
	}

	ns := rs.Namespaces["MyApp"]
	team := ns.EntityTypes[types.EntityType("MyApp::Team")]
	if team == nil {
		t.Fatal("expected Team entity type")
	}

	membersAttr := team.Shape.Attributes["members"]
	if membersAttr == nil {
		t.Fatal("expected members attribute")
	}

	setType, ok := membersAttr.Type.(resolved.Set)
	if !ok {
		t.Fatalf("expected resolved.Set, got %T", membersAttr.Type)
	}

	entityRef, ok := setType.Element.(resolved.EntityRef)
	if !ok {
		t.Fatalf("expected resolved.EntityRef element, got %T", setType.Element)
	}

	if entityRef.EntityType != types.EntityType("MyApp::User") {
		t.Errorf("expected MyApp::User, got %v", entityRef.EntityType)
	}
}
