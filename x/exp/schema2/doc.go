// Package schema2 provides a programmatic interface for building and
// working with Cedar schemas.
//
// # Overview
//
// This package provides two distinct schema representations:
//   - [Schema]: A mutable builder type for constructing schemas programmatically
//   - [ResolvedSchema]: An immutable type with fully-qualified names and computed hierarchies
//
// # Two-Phase Pattern
//
// Schema construction follows a build -> resolve pattern:
//
//  1. Build: Use [NewSchema] and the fluent builder API to construct your schema
//  2. Resolve: Call [Schema.Resolve] to validate and produce a [ResolvedSchema]
//
// This separation allows you to build schemas incrementally and catch
// validation errors at a well-defined point.
//
// # Building Schemas
//
// The [Schema] type provides a fluent API for building Cedar schemas:
//
//	schema := schema2.NewSchema().
//	    Namespace("MyApp").
//	    Entity("User").In("Group").Attributes(
//	        schema2.Attr("name", schema2.String()),
//	        schema2.Attr("email", schema2.String()),
//	    ).
//	    Entity("Group").
//	    Entity("Document").
//	    Action("read").Principals("User", "Group").Resources("Document").
//	    Action("write").Principals("User").Resources("Document")
//
// The builder uses separate [EntityBuilder] and [ActionBuilder] types that
// provide compile-time safety. Methods like [EntityBuilder.In] and
// [EntityBuilder.Attributes] are only available when building an entity.
//
// # Type Constructors
//
// The following functions create Cedar types for use in schema building:
//
// Primitive types:
//   - [String]: Cedar String type
//   - [Long]: Cedar Long (integer) type
//   - [Bool]: Cedar Bool type
//
// Complex types:
//   - [Set]: Set<T> type containing elements of a given type
//   - [Record]: Record type with named attributes
//   - [EntityRef]: Reference to an entity type
//
// Extension types:
//   - [Extension]: Generic extension type constructor
//   - [IPAddr]: ipaddr extension type
//   - [Decimal]: decimal extension type
//   - [Datetime]: datetime extension type
//   - [Duration]: duration extension type
//
// Attribute constructors:
//   - [Attr]: Required attribute with name and type
//   - [OptionalAttr]: Optional attribute with name and type
//
// # Resolving Schemas
//
// Call [Schema.Resolve] to validate and produce an immutable [ResolvedSchema]:
//
//	resolved, err := schema.Resolve()
//	if err != nil {
//	    // Handle validation errors (undefined types, cycles, etc.)
//	}
//
// For tests, [Schema.MustResolve] panics on error:
//
//	resolved := schema.MustResolve()
//
// # Working with Resolved Schemas
//
// [ResolvedSchema] provides methods to query entity types and actions:
//
//	// Get entity type (returns nil if not found)
//	userType := resolved.EntityType(types.EntityType("MyApp::User"))
//
//	// Lookup with found indicator
//	userType, found := resolved.LookupEntityType(types.EntityType("MyApp::User"))
//
//	// Check entity type kind
//	if userType.IsEnum() {
//	    enumKind, _ := userType.AsEnum()
//	    fmt.Println("Values:", enumKind.Values())
//	}
//
//	// Iterate over all entity types
//	for name, et := range resolved.EntityTypes() {
//	    fmt.Printf("Entity: %s\n", name)
//	}
//
// # Parsing and Serialization
//
// Parse schemas from JSON:
//
//	schema, err := schema2.ParseJSON(jsonData)
//
// Parse schemas from Cedar human-readable format:
//
//	schema, err := schema2.ParseCedar(cedarData)
//	schema, err := schema2.ParseCedar(cedarData, schema2.WithFilename("schema.cedarschema"))
//
// Serialize schemas:
//
//	jsonData, err := schema.MarshalJSON()
//	jsonData, err := schema.MarshalJSONIndent("", "  ")
//	cedarData, err := schema.MarshalCedar()
//
// # Common Patterns
//
// Enum entity types:
//
//	schema := schema2.NewSchema().
//	    Namespace("App").
//	    Entity("Status").Enum("Draft", "Published", "Archived")
//
// Reusable common types:
//
//	schema := schema2.NewSchema().
//	    Namespace("App").
//	    CommonType("Address", schema2.Record(
//	        schema2.Attr("street", schema2.String()),
//	        schema2.Attr("city", schema2.String()),
//	    )).
//	    Entity("User").Attributes(
//	        schema2.Attr("homeAddress", schema2.EntityRef("Address")),
//	    )
//
// Entity with tags:
//
//	schema := schema2.NewSchema().
//	    Namespace("App").
//	    Entity("Resource").Tags(schema2.String())
//
// Action with context:
//
//	schema := schema2.NewSchema().
//	    Namespace("App").
//	    Entity("User").
//	    Entity("Document").
//	    Action("edit").
//	        Principals("User").
//	        Resources("Document").
//	        Context(schema2.Record(
//	            schema2.Attr("reason", schema2.String()),
//	        ))
package schema2
