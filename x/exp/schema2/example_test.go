package schema2_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

func Example_basicUsage() {
	// Build a schema programmatically
	s := schema2.NewSchema().
		Namespace("PhotoApp").
		Entity("User").In("Group").Attributes(
			schema2.Attr("name", schema2.String()),
			schema2.Attr("email", schema2.String()),
		).
		Entity("Group").
		Entity("Photo").In("Album").Attributes(
			schema2.Attr("owner", schema2.EntityRef("User")),
		).
		Entity("Album").
		Action("view").Principals("User", "Group").Resources("Photo", "Album").
		Action("edit").Principals("User").Resources("Photo")

	// Resolve to get fully-qualified types
	resolved, err := s.Resolve()
	if err != nil {
		fmt.Printf("Resolution error: %v\n", err)
		return
	}

	// Access resolved entity types
	userType := resolved.EntityType(types.EntityType("PhotoApp::User"))
	fmt.Printf("User type: %s\n", userType.Name())

	// Check entity hierarchy
	groupType := resolved.EntityType(types.EntityType("PhotoApp::Group"))
	if groupType.HasDescendant(types.EntityType("PhotoApp::User")) {
		fmt.Println("Users can be members of Groups")
	}

	// Output:
	// User type: PhotoApp::User
	// Users can be members of Groups
}

func Example_enumEntity() {
	// Build a schema with an enumerated entity type
	s := schema2.NewSchema().
		Namespace("App").
		Entity("User").
		Entity("Status").Enum("Draft", "Published", "Archived").
		Entity("Document").Attributes(
			schema2.Attr("status", schema2.EntityRef("Status")),
		).
		Action("publish").Principals("User").Resources("Document")

	resolved, err := s.Resolve()
	if err != nil {
		fmt.Printf("Resolution error: %v\n", err)
		return
	}

	statusType := resolved.EntityType(types.EntityType("App::Status"))
	if enumKind, ok := statusType.Kind().(schema2.EnumEntityType); ok {
		fmt.Printf("Status values: %v\n", len(enumKind.Values()))
	}

	// Output:
	// Status values: 3
}

func Example_parseJSON() {
	jsonSchema := `{
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
						"resourceTypes": ["User", "Group"]
					}
				}
			}
		}
	}`

	s, err := schema2.ParseJSON([]byte(jsonSchema))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	resolved, err := s.Resolve()
	if err != nil {
		fmt.Printf("Resolution error: %v\n", err)
		return
	}

	userType := resolved.EntityType(types.EntityType("App::User"))
	fmt.Printf("User type found: %v\n", userType != nil)

	// Output:
	// User type found: true
}

func Example_marshalJSON() {
	s := schema2.NewSchema().
		Namespace("App").
		Entity("User").
		Action("read").Principals("User").Resources("User")

	jsonData, err := s.MarshalJSONIndent("", "  ")
	if err != nil {
		fmt.Printf("Marshal error: %v\n", err)
		return
	}

	fmt.Println("Schema marshaled successfully")
	_ = jsonData // Use the JSON data

	// Output:
	// Schema marshaled successfully
}

func Example_complexTypes() {
	// Build a schema with complex types
	s := schema2.NewSchema().
		Namespace("App").
		CommonType("Address", schema2.Record(
			schema2.Attr("street", schema2.String()),
			schema2.Attr("city", schema2.String()),
			schema2.Attr("zip", schema2.String()),
		)).
		Entity("User").Attributes(
			schema2.Attr("name", schema2.String()),
			schema2.Attr("addresses", schema2.Set(schema2.EntityRef("Address"))),
			schema2.OptionalAttr("tags", schema2.Set(schema2.String())),
		)

	resolved, err := s.Resolve()
	if err != nil {
		fmt.Printf("Resolution error: %v\n", err)
		return
	}

	userType := resolved.EntityType(types.EntityType("App::User"))
	fmt.Printf("User has attributes: %v\n", userType.Attributes() != nil)

	// Output:
	// User has attributes: true
}
