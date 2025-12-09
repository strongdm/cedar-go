package schema

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
)

func TestSchemaBuilderBasic(t *testing.T) {
	// Build a schema programmatically
	s := New().
		AddNamespace(NewNamespace("MyApp").
			AddEntityType(Entity("User").
				WithShape(Record().
					AddRequired("name", String()).
					AddOptional("email", String()))).
			AddEntityType(Entity("Document").
				In("Folder").
				WithShape(Record().
					AddRequired("title", String()).
					AddOptional("content", String()))).
			AddAction(NewAction("read").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("Document"))))

	// Test that we can marshal to Cedar format
	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Cedar output:\n%s", string(cedarBytes))

	// Test that we can marshal to JSON format
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	t.Logf("JSON output:\n%s", string(jsonBytes))

	// Verify JSON is valid
	var raw interface{}
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		t.Fatalf("MarshalJSON() produced invalid JSON: %v", err)
	}

	// Verify we can round-trip through Cedar
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarBytes); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Verify we can round-trip through JSON
	var s3 Schema
	if err := s3.UnmarshalJSON(jsonBytes); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}
}

func TestSchemaBuilderWithAnnotations(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("AnnotatedApp").
			Annotate("doc", "A test namespace").
			AddEntityType(Entity("User").
				Annotate("doc", "A user entity").
				WithShape(Record().
					AddAttribute(Attribute{
						Name:        "name",
						Type:        String(),
						Required:    true,
						Annotations: Annotations{{Key: "doc", Value: "The user's name"}},
					}))).
			AddCommonType(NewCommonType("Context", Record().
				AddOptional("authenticated", Boolean())).
				Annotate("doc", "Common context type")))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Cedar with annotations:\n%s", string(cedarBytes))

	// Verify round-trip
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarBytes); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Check that namespace was parsed
	ns := s2.GetNamespace("AnnotatedApp")
	if ns == nil {
		t.Fatal("Expected to find AnnotatedApp namespace")
	}
}

func TestSchemaBuilderComplexTypes(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("ComplexApp").
			AddEntityType(Entity("User").
				In("Group", "Organization").
				WithShape(Record().
					AddRequired("id", Long()).
					AddOptional("tags", Set(String())).
					AddOptional("manager", EntityOrCommonType("User")).
					AddOptional("metadata", Record().
						AddOptional("created", String()).
						AddOptional("updated", String())))).
			AddEntityType(Entity("Group").
				In("Organization")).
			AddEntityType(Entity("Organization")).
			AddAction(NewAction("manage").
				InAction("admin").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User", "Group").
					WithResources("User", "Group", "Organization").
					WithContext(Record().
						AddOptional("reason", String()).
						AddOptional("approved", Boolean())))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Complex schema:\n%s", string(cedarBytes))

	// Verify JSON round-trip
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var s2 Schema
	if err := s2.UnmarshalJSON(jsonBytes); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Verify Cedar round-trip
	var s3 Schema
	if err := s3.UnmarshalCedar(cedarBytes); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}
}

func TestSchemaBuilderCommonTypes(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("CommonTypesApp").
			AddCommonType(NewCommonType("Address", Record().
				AddRequired("street", String()).
				AddRequired("city", String()).
				AddOptional("zip", String()))).
			AddCommonType(NewCommonType("ContactInfo", Record().
				AddOptional("email", String()).
				AddOptional("phone", String()).
				AddOptional("address", EntityOrCommonType("Address")))).
			AddEntityType(Entity("User").
				WithShape(Record().
					AddRequired("name", String()).
					AddOptional("contact", EntityOrCommonType("ContactInfo")))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Schema with common types:\n%s", string(cedarBytes))

	// Round-trip
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarBytes); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s2.GetNamespace("CommonTypesApp")
	if ns == nil {
		t.Fatal("Expected to find CommonTypesApp namespace")
	}

	if len(ns.CommonTypes) != 2 {
		t.Errorf("Expected 2 common types, got %d", len(ns.CommonTypes))
	}
}

func TestSchemaBuilderMultipleNamespaces(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("").
			AddEntityType(Entity("GlobalEntity"))).
		AddNamespace(NewNamespace("App::Users").
			AddEntityType(Entity("User").
				WithShape(Record().AddRequired("name", String())))).
		AddNamespace(NewNamespace("App::Documents").
			AddEntityType(Entity("Document").
				WithShape(Record().AddRequired("title", String()))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Multi-namespace schema:\n%s", string(cedarBytes))

	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	t.Logf("Multi-namespace JSON:\n%s", string(jsonBytes))

	// Verify we can look up namespaces
	if s.GetNamespace("") == nil {
		t.Error("Expected to find anonymous namespace")
	}
	if s.GetNamespace("App::Users") == nil {
		t.Error("Expected to find App::Users namespace")
	}
	if s.GetNamespace("App::Documents") == nil {
		t.Error("Expected to find App::Documents namespace")
	}
}

func TestSchemaBuilderActionGroups(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("ActionGroupApp").
			AddEntityType(Entity("User")).
			AddEntityType(Entity("Resource")).
			AddAction(NewAction("read").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("Resource"))).
			AddAction(NewAction("write").
				InAction("mutate").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("Resource"))).
			AddAction(NewAction("delete").
				InAction("mutate").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("Resource"))).
			AddAction(NewAction("mutate").
				WithAppliesTo(NewAppliesTo().
					WithPrincipals("User").
					WithResources("Resource"))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Schema with action groups:\n%s", string(cedarBytes))
}

func TestSchemaBuilderEntityTags(t *testing.T) {
	s := New().
		AddNamespace(NewNamespace("TagsApp").
			AddEntityType(Entity("User").
				WithShape(Record().AddRequired("name", String())).
				WithTags(String())).
			AddEntityType(Entity("Document").
				WithTags(Record().
					AddOptional("classification", String()).
					AddOptional("sensitivity", Long()))))

	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Schema with entity tags:\n%s", string(cedarBytes))

	// JSON round-trip
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON() error = %v", err)
	}

	var s2 Schema
	if err := s2.UnmarshalJSON(jsonBytes); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	ns := s2.GetNamespace("TagsApp")
	if ns == nil {
		t.Fatal("Expected to find TagsApp namespace")
	}

	// Verify tags were preserved
	var userEntity *EntityType
	for _, e := range ns.EntityTypes {
		if e.Name == "User" {
			userEntity = e
			break
		}
	}
	if userEntity == nil {
		t.Fatal("Expected to find User entity")
	}
	if userEntity.Tags == nil {
		t.Error("Expected User entity to have tags")
	}
}

func TestParseAndAccessSchema(t *testing.T) {
	cedarSchema := `
namespace PhotoApp {
    type Context = {
        authenticated?: Bool,
    };

    entity User in [Group] {
        name: String,
        email?: String,
    };

    entity Group;

    entity Photo in [Album] {
        owner: User,
        public?: Bool,
    };

    entity Album;

    action view appliesTo {
        principal: [User, Group],
        resource: Photo,
        context: Context,
    };

    action upload appliesTo {
        principal: User,
        resource: Album,
    };
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(cedarSchema)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Access the parsed schema programmatically
	ns := s.GetNamespace("PhotoApp")
	if ns == nil {
		t.Fatal("Expected to find PhotoApp namespace")
	}

	t.Logf("Found %d entity types", len(ns.EntityTypes))
	t.Logf("Found %d actions", len(ns.Actions))
	t.Logf("Found %d common types", len(ns.CommonTypes))

	// Find User entity
	var userEntity *EntityType
	for _, e := range ns.EntityTypes {
		if e.Name == "User" {
			userEntity = e
			break
		}
	}

	if userEntity == nil {
		t.Fatal("Expected to find User entity")
	}

	if len(userEntity.MemberOf) != 1 || userEntity.MemberOf[0] != types.Path("Group") {
		t.Errorf("Expected User to be member of [Group], got %v", userEntity.MemberOf)
	}

	if userEntity.Shape == nil {
		t.Fatal("Expected User to have shape")
	}

	if len(userEntity.Shape.Attributes) != 2 {
		t.Errorf("Expected User to have 2 attributes, got %d", len(userEntity.Shape.Attributes))
	}

	// Verify we can marshal back to Cedar
	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("Re-marshaled Cedar:\n%s", string(cedarBytes))
}

func TestJSONSchemaRoundTrip(t *testing.T) {
	jsonSchema := `{
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
					"memberOfTypes": ["Album"],
					"shape": {
						"type": "Record",
						"attributes": {
							"owner": {"type": "EntityOrCommon", "name": "User", "required": true},
							"public": {"type": "Boolean", "required": false}
						}
					}
				},
				"Album": {}
			},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User", "Group"],
						"resourceTypes": ["Photo"],
						"context": {
							"type": "Record",
							"attributes": {
								"authenticated": {"type": "Boolean", "required": false}
							}
						}
					}
				}
			}
		}
	}`

	var s Schema
	if err := s.UnmarshalJSON([]byte(jsonSchema)); err != nil {
		t.Fatalf("UnmarshalJSON() error = %v", err)
	}

	// Access parsed schema
	ns := s.GetNamespace("PhotoApp")
	if ns == nil {
		t.Fatal("Expected to find PhotoApp namespace")
	}

	t.Logf("Parsed %d entity types from JSON", len(ns.EntityTypes))
	t.Logf("Parsed %d actions from JSON", len(ns.Actions))

	// Marshal to Cedar
	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	t.Logf("JSON to Cedar:\n%s", string(cedarBytes))

	// Marshal back to JSON
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

func TestTypeMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		typ      Type
		expected string
	}{
		{"String", String(), "String"},
		{"Long", Long(), "Long"},
		{"Boolean", Boolean(), "Bool"},
		{"Set of String", Set(String()), "Set<String>"},
		{"Set of Long", Set(Long()), "Set<Long>"},
		{"Nested Set", Set(Set(String())), "Set<Set<String>>"},
		{"EntityOrCommon", EntityOrCommonType("User"), "User"},
		{"Qualified EntityOrCommon", EntityOrCommonType("App::User"), "App::User"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.typ.MarshalCedar())
			if got != tt.expected {
				t.Errorf("MarshalCedar() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEntityMarshalCedar(t *testing.T) {
	tests := []struct {
		name     string
		entity   *EntityType
		contains []string
	}{
		{
			name:     "simple entity",
			entity:   Entity("User"),
			contains: []string{"entity User;"},
		},
		{
			name: "entity with shape",
			entity: Entity("User").
				WithShape(Record().AddRequired("name", String())),
			contains: []string{"entity User", "name: String"},
		},
		{
			name:     "entity with memberOf",
			entity:   Entity("User").In("Group"),
			contains: []string{"entity User in Group;"},
		},
		{
			name:     "entity with multiple memberOf",
			entity:   Entity("User").In("Group", "Organization"),
			contains: []string{"entity User in [Group, Organization];"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(tt.entity.MarshalCedar())
			for _, want := range tt.contains {
				if !contains(got, want) {
					t.Errorf("MarshalCedar() = %q, want to contain %q", got, want)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
