package schema_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/schema"
)

// ExampleNewSchema demonstrates creating a schema programmatically
func ExampleNewSchema() {
	// Create a new schema with a PhotoApp namespace
	s := schema.NewSchema().
		WithNamespace("PhotoApp",
			// Define a User entity with attributes
			schema.NewEntity("User").
				WithAttribute("name", schema.String()).
				WithAttribute("age", schema.Long()).
				WithOptionalAttribute("email", schema.String()).
				MemberOf("UserGroup"),

			// Define a UserGroup entity
			schema.NewEntity("UserGroup"),

			// Define a Photo entity with a reference to User
			schema.NewEntity("Photo").
				WithAttribute("owner", schema.EntityType("User")).
				WithAttribute("private", schema.Bool()),

			// Define an action
			schema.NewAction("viewPhoto").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Photo"),
					schema.Record(
						schema.Attr("authenticated", schema.Bool()),
					),
				),
		)

	// Marshal to JSON
	jsonData, _ := s.MarshalJSON()
	fmt.Printf("Schema has %d bytes of JSON\n", len(jsonData))

	// Marshal to Cedar text format
	cedarData, _ := s.MarshalCedar()
	fmt.Printf("Schema has %d bytes of Cedar text\n", len(cedarData))

	// Output:
	// Schema has 593 bytes of JSON
	// Schema has 335 bytes of Cedar text
}

// ExampleSchema_UnmarshalJSON demonstrates parsing a JSON schema
func ExampleSchema_UnmarshalJSON() {
	jsonSchema := []byte(`{
		"MyApp": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"name": {
								"type": "String",
								"required": true
							}
						}
					}
				}
			},
			"actions": {
				"read": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["Document"]
					}
				}
			}
		}
	}`)

	var s schema.Schema
	if err := s.UnmarshalJSON(jsonSchema); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Convert to Cedar format
	cedarData, _ := s.MarshalCedar()
	fmt.Printf("Converted to Cedar:\n%s", cedarData)
}

// ExampleSchema_UnmarshalCedar demonstrates parsing a Cedar text schema
func ExampleSchema_UnmarshalCedar() {
	cedarSchema := []byte(`
namespace MyApp {
  entity User = {
    "name": String,
  };
  entity Document;
  action read appliesTo {
    principal: [User],
    resource: [Document]
  };
}
`)

	var s schema.Schema
	if err := s.UnmarshalCedar(cedarSchema); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Schema parsed successfully")
	// Output:
	// Schema parsed successfully
}

// ExampleRecord demonstrates creating complex nested record types
func ExampleRecord() {
	s := schema.NewSchema().
		WithNamespace("App",
			schema.NewEntity("User").
				WithAttribute("profile", schema.Record(
					schema.Attr("firstName", schema.String()),
					schema.Attr("lastName", schema.String()),
					schema.OptionalAttr("middleName", schema.String()),
				)).
				WithAttribute("addresses", schema.Set(
					schema.Record(
						schema.Attr("street", schema.String()),
						schema.Attr("city", schema.String()),
						schema.Attr("zipCode", schema.String()),
					),
				)),
		)

	jsonData, _ := s.MarshalJSON()
	fmt.Printf("Complex schema: %d bytes\n", len(jsonData))
	// Output:
	// Complex schema: 500 bytes
}

// ExampleEntity_AsEnum demonstrates creating an enumerated entity type
func ExampleEntity_AsEnum() {
	s := schema.NewSchema().
		WithNamespace("App",
			schema.NewEntity("FileType").AsEnum("pdf", "doc", "txt", "jpg"),
		)

	jsonData, _ := s.MarshalJSON()
	fmt.Printf("Enum schema: %d bytes\n", len(jsonData))
	// Output:
	// Enum schema: 84 bytes
}

// ExampleNewEntity demonstrates the fluent entity builder API
func ExampleNewEntity() {
	entity := schema.NewEntity("Document").
		WithAttribute("title", schema.String()).
		WithAttribute("content", schema.String()).
		WithOptionalAttribute("author", schema.EntityType("User")).
		WithAnnotation("doc", "Represents a document in the system")

	// Use the entity in a schema
	s := schema.NewSchema().
		WithNamespace("MyApp", entity)

	cedarData, _ := s.MarshalCedar()
	fmt.Printf("Schema length: %d\n", len(cedarData))
	// Output:
	// Schema length: 156
}

// ExampleSet demonstrates creating set types
func ExampleSet() {
	// A set of strings
	tags := schema.Set(schema.String())

	// A set of entity references
	users := schema.Set(schema.EntityType("User"))

	// Use in an entity
	s := schema.NewSchema().
		WithNamespace("App",
			schema.NewEntity("Document").
				WithAttribute("tags", tags).
				WithAttribute("editors", users),
		)

	jsonData, _ := s.MarshalJSON()
	fmt.Printf("Set schema: %d bytes\n", len(jsonData))
	// Output:
	// Set schema: 249 bytes
}
