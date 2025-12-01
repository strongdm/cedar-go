package schema_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/schema"
)

// Example_comprehensive demonstrates the full capabilities of the schema package
func Example_comprehensive() {
	// Create a schema for a photo-sharing application
	s := schema.NewSchema().
		WithNamespace("PhotoApp",
			// Define entity types
			schema.NewEntity("User").
				WithAttribute("name", schema.String()).
				WithAttribute("email", schema.String()).
				WithAttribute("age", schema.Long()).
				WithOptionalAttribute("verified", schema.Bool()).
				WithAnnotation("doc", "Represents a user in the system"),

			schema.NewEntity("Photo").
				WithAttribute("owner", schema.EntityType("User")).
				WithAttribute("tags", schema.Set(schema.String())).
				WithOptionalAttribute("location", schema.String()).
				MemberOf("Resource").
				WithAnnotation("doc", "A photo uploaded by a user"),

			schema.NewEntity("Album").
				WithAttribute("name", schema.String()).
				WithAttribute("photos", schema.Set(schema.EntityType("Photo"))).
				WithAttribute("owner", schema.EntityType("User")).
				MemberOf("Resource"),

			// Define a resource parent type
			schema.NewEntity("Resource").
				WithAnnotation("doc", "Base resource type"),

			// Define actions
			schema.NewAction("viewPhoto").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Photo", "Album"),
					schema.Record(
						schema.Attr("ip", schema.String()),
						schema.OptionalAttr("reason", schema.String()),
					),
				).
				WithAnnotation("doc", "View a photo or album"),

			schema.NewAction("uploadPhoto").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Album"),
					nil,
				),

			schema.NewAction("deletePhoto").
				MemberOf(
					schema.ActionGroup("modifyPhoto"),
				).
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("Photo"),
					nil,
				),

			// Define common types
			schema.TypeDecl("Address", schema.Record(
				schema.Attr("street", schema.String()),
				schema.Attr("city", schema.String()),
				schema.Attr("zipCode", schema.String()),
			)),
		)

	// Marshal to JSON format
	jsonBytes, err := s.MarshalJSON()
	if err != nil {
		panic(err)
	}
	fmt.Println("Schema marshaled to JSON:", len(jsonBytes), "bytes")

	// Marshal to Cedar format
	cedarBytes, err := s.MarshalCedar()
	if err != nil {
		panic(err)
	}
	fmt.Println("Schema marshaled to Cedar:", len(cedarBytes), "bytes")

	// Parse from JSON
	var s2 schema.Schema
	if err := s2.UnmarshalJSON(jsonBytes); err != nil {
		panic(err)
	}
	fmt.Println("Schema parsed from JSON successfully")

	// Parse from Cedar
	var s3 schema.Schema
	s3.SetFilename("example.cedar")
	if err := s3.UnmarshalCedar(cedarBytes); err != nil {
		panic(err)
	}
	fmt.Println("Schema parsed from Cedar successfully")

	// Output:
	// Schema marshaled to JSON: 1600 bytes
	// Schema marshaled to Cedar: 915 bytes
	// Schema parsed from JSON successfully
	// Schema parsed from Cedar successfully
}

// Example_entityTypes demonstrates entity type definitions
func Example_entityTypes() {
	s := schema.NewSchema().
		WithNamespace("App",
			// Simple entity with attributes
			schema.NewEntity("User").
				WithAttribute("id", schema.String()).
				WithAttribute("name", schema.String()),

			// Entity with optional attributes
			schema.NewEntity("Document").
				WithAttribute("title", schema.String()).
				WithOptionalAttribute("description", schema.String()),

			// Entity with shape (record type)
			schema.NewEntity("Config").
				WithShape(schema.Record(
					schema.Attr("enabled", schema.Bool()),
					schema.Attr("timeout", schema.Long()),
				)),

			// Entity with tags
			schema.NewEntity("Resource").
				WithTags(schema.Set(schema.String())),

			// Entity hierarchy
			schema.NewEntity("Admin").
				MemberOf("User"),

			// Enum entity
			schema.NewEntity("Status").
				AsEnum("active", "inactive", "pending"),
		)

	jsonBytes, _ := s.MarshalJSON()
	fmt.Println("Entity types defined:", len(jsonBytes) > 0)

	// Output:
	// Entity types defined: true
}

// Example_actions demonstrates action definitions
func Example_actions() {
	s := schema.NewSchema().
		WithNamespace("FileSystem",
			schema.NewEntity("User"),
			schema.NewEntity("File"),
			schema.NewEntity("Directory"),

			// Simple action
			schema.NewAction("read").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("File", "Directory"),
					nil,
				),

			// Action with context
			schema.NewAction("write").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("File"),
					schema.Record(
						schema.Attr("timestamp", schema.Long()),
						schema.OptionalAttr("reason", schema.String()),
					),
				),

			// Action groups
			schema.NewAction("createFile").
				MemberOf(schema.ActionGroup("modify")),

			schema.NewAction("deleteFile").
				MemberOf(schema.ActionGroup("modify")),

			// Action with annotations
			schema.NewAction("admin").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("File"),
					nil,
				).
				WithAnnotation("dangerous", "true"),
		)

	cedarBytes, _ := s.MarshalCedar()
	fmt.Println("Actions defined:", len(cedarBytes) > 0)

	// Output:
	// Actions defined: true
}

// Example_types demonstrates the type system
func Example_types() {
	s := schema.NewSchema().
		WithNamespace("Types",
			schema.NewEntity("Example").
				// Primitive types
				WithAttribute("stringField", schema.String()).
				WithAttribute("longField", schema.Long()).
				WithAttribute("boolField", schema.Bool()).

				// Entity references
				WithAttribute("owner", schema.EntityType("User")).

				// Set types
				WithAttribute("tags", schema.Set(schema.String())).
				WithAttribute("members", schema.Set(schema.EntityType("User"))).

				// Nested record types
				WithAttribute("address", schema.Record(
					schema.Attr("street", schema.String()),
					schema.Attr("city", schema.String()),
					schema.Attr("coordinates", schema.Record(
						schema.Attr("lat", schema.Long()),
						schema.Attr("lon", schema.Long()),
					)),
				)),
		)

	jsonBytes, _ := s.MarshalJSON()
	fmt.Println("Type system example:", len(jsonBytes) > 0)

	// Output:
	// Type system example: true
}

// Example_annotations demonstrates annotation usage
func Example_annotations() {
	s := schema.NewSchema().
		WithNamespace("Annotated",
			schema.NewEntity("User").
				WithAttribute("email", schema.String()).
				WithAnnotation("doc", "User account").
				WithAnnotation("version", "2.0"),

			schema.NewAction("login").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("User"),
					nil,
				).
				WithAnnotation("rateLimit", "100").
				WithAnnotation("doc", "User login action"),

			// Annotations on types
			schema.TypeDecl("Email", schema.String().
				WithAnnotation("format", "email")),
		)

	cedarBytes, _ := s.MarshalCedar()
	fmt.Println("Annotations example:", len(cedarBytes) > 0)

	// Output:
	// Annotations example: true
}

// Example_multipleNamespaces demonstrates working with multiple namespaces
func Example_multipleNamespaces() {
	s := schema.NewSchema().
		WithNamespace("Users",
			schema.NewEntity("User").
				WithAttribute("name", schema.String()),
			schema.NewAction("login").
				AppliesTo(
					schema.Principals("User"),
					schema.Resources("User"),
					nil,
				),
		).
		WithNamespace("Documents",
			schema.NewEntity("Document").
				WithAttribute("title", schema.String()).
				WithAttribute("author", schema.EntityType("Users::User")),
			schema.NewAction("view").
				AppliesTo(
					schema.Principals("Users::User"),
					schema.Resources("Document"),
					nil,
				),
		).
		WithNamespace("", // Global namespace
			schema.NewEntity("System"),
		)

	jsonBytes, _ := s.MarshalJSON()
	fmt.Println("Multiple namespaces:", len(jsonBytes) > 0)

	// Output:
	// Multiple namespaces: true
}
