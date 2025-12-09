package schema_test

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/x/exp/schema"
)

// This example shows how to build a schema programmatically using the builder pattern.
func Example() {
	s := schema.New().
		AddNamespace(schema.NewNamespace("PhotoApp").
			AddEntityType(schema.Entity("User").
				In("Group").
				WithShape(schema.Record().
					AddRequired("name", schema.String()).
					AddOptional("email", schema.String()))).
			AddEntityType(schema.Entity("Group")).
			AddEntityType(schema.Entity("Photo").
				In("Album").
				WithShape(schema.Record().
					AddRequired("title", schema.String()).
					AddOptional("public", schema.Boolean()))).
			AddEntityType(schema.Entity("Album")).
			AddAction(schema.NewAction("view").
				WithAppliesTo(schema.NewAppliesTo().
					WithPrincipals("User", "Group").
					WithResources("Photo"))).
			AddAction(schema.NewAction("upload").
				WithAppliesTo(schema.NewAppliesTo().
					WithPrincipals("User").
					WithResources("Album"))))

	cedarBytes, _ := s.MarshalCedar()
	fmt.Println(string(cedarBytes))

	// Output:
	// namespace PhotoApp {
	//   entity User in Group {
	//     name: String,
	//     email?: String,
	//   };
	//   entity Group;
	//   entity Photo in Album {
	//     title: String,
	//     public?: Bool,
	//   };
	//   entity Album;
	//   action view appliesTo {
	//     principal: [User, Group],
	//     resource: Photo,
	//   };
	//   action upload appliesTo {
	//     principal: User,
	//     resource: Album,
	//   };
	// }
}

// This example shows how to add annotations to schema elements.
func Example_annotations() {
	s := schema.New().
		AddNamespace(schema.NewNamespace("AnnotatedApp").
			Annotate("doc", "Application schema with documentation").
			AddCommonType(schema.NewCommonType("Context", schema.Record().
				AddOptional("authenticated", schema.Boolean())).
				Annotate("doc", "Common context type")).
			AddEntityType(schema.Entity("User").
				Annotate("doc", "A user in the system").
				WithShape(schema.Record().
					AddRequired("name", schema.String()))))

	cedarBytes, _ := s.MarshalCedar()
	fmt.Println(string(cedarBytes))

	// Output:
	// @doc("Application schema with documentation")
	// namespace AnnotatedApp {
	//   @doc("Common context type")
	//   type Context = {
	//     authenticated?: Bool,
	//   };
	//   @doc("A user in the system")
	//   entity User {
	//     name: String,
	//   };
	// }
}

// This example shows how to create common types and reference them.
func Example_commonTypes() {
	s := schema.New().
		AddNamespace(schema.NewNamespace("CommonTypesApp").
			AddCommonType(schema.NewCommonType("Address", schema.Record().
				AddRequired("street", schema.String()).
				AddRequired("city", schema.String()).
				AddOptional("zip", schema.String()))).
			AddEntityType(schema.Entity("User").
				WithShape(schema.Record().
					AddRequired("name", schema.String()).
					AddOptional("address", schema.EntityOrCommonType("Address")))))

	cedarBytes, _ := s.MarshalCedar()
	fmt.Println(string(cedarBytes))

	// Output:
	// namespace CommonTypesApp {
	//   type Address = {
	//     street: String,
	//     city: String,
	//     zip?: String,
	//   };
	//   entity User {
	//     name: String,
	//     address?: Address,
	//   };
	// }
}

// This example shows how to parse a schema and access its elements.
func Example_parsing() {
	cedarSchema := `
namespace PhotoApp {
    entity User;
    entity Photo;
    action view appliesTo {
        principal: User,
        resource: Photo,
    };
}
`

	var s schema.Schema
	_ = s.UnmarshalCedar([]byte(cedarSchema))

	ns := s.GetNamespace("PhotoApp")
	fmt.Printf("Namespace: %s\n", ns.Name)
	fmt.Printf("Entity types: %d\n", len(ns.EntityTypes))
	fmt.Printf("Actions: %d\n", len(ns.Actions))

	// Output:
	// Namespace: PhotoApp
	// Entity types: 2
	// Actions: 1
}
