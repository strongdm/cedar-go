# Schema Package Public API

## Overview

The `schema` package provides a fluent API for programmatically constructing, parsing, and marshaling Cedar schemas. It supports both JSON and human-readable Cedar formats.

## Core Types

### Schema
The main schema container that holds entity types, actions, and common types organized by namespace.

```go
type Schema struct { ... }
```

**Methods:**
- `NewSchema() *Schema` - Creates a new empty schema
- `WithNamespace(namespace string, decls ...Declaration) *Schema` - Adds a namespace with declarations
- `SetFilename(filename string) *Schema` - Sets filename for error messages
- `MarshalJSON() ([]byte, error)` - Marshals to Cedar JSON format
- `UnmarshalJSON(data []byte) error` - Unmarshals from Cedar JSON format
- `MarshalCedar() ([]byte, error)` - Marshals to Cedar human-readable format
- `UnmarshalCedar(data []byte) error` - Unmarshals from Cedar human-readable format

### Type
Interface representing Cedar schema types (primitives, sets, records, entity references).

```go
type Type interface { ... }
```

**Implementations:**
- `PathType` - Primitives (String, Long, Bool) and entity references
- `SetType` - Set types
- `RecordType` - Record types with named attributes

### Declaration
Interface for things that can be declared in a namespace (entities, actions, common types).

```go
type Declaration interface { ... }
```

**Implementations:**
- `Entity` - Entity type declarations
- `Action` - Action declarations
- `CommonTypeDecl` - Common type declarations

## Type Constructors

### Primitive Types
```go
String() *PathType                    // String type
Long() *PathType                      // Long (integer) type
Bool() *PathType                      // Boolean type
Boolean() *PathType                   // Alias for Bool()
```

### Reference Types
```go
EntityType(path string) *PathType     // Entity reference (e.g., "User", "App::User")
CommonType(path string) *PathType     // Common type reference
```

### Composite Types
```go
Set(element Type) *SetType            // Set of elements
Record(attrs ...*Attribute) *RecordType  // Record with attributes
```

### Attributes
```go
Attr(name string, typ Type) *Attribute              // Required attribute
OptionalAttr(name string, typ Type) *Attribute      // Optional attribute
```

## Entity Builder

```go
NewEntity(name string) *Entity
```

**Methods:**
- `WithAttribute(name string, typ Type) *Entity` - Adds required attribute
- `WithOptionalAttribute(name string, typ Type) *Entity` - Adds optional attribute
- `WithShape(shape Type) *Entity` - Sets entity shape
- `WithTags(tagType Type) *Entity` - Sets tag type for entity
- `MemberOf(parentTypes ...string) *Entity` - Sets parent entity types
- `AsEnum(values ...string) *Entity` - Makes entity an enum
- `WithAnnotation(key, value string) *Entity` - Adds annotation

## Action Builder

```go
NewAction(name string) *Action
```

**Methods:**
- `AppliesTo(principals, resources []string, context Type) *Action` - Sets action scope
- `MemberOf(parents ...*ActionRef) *Action` - Sets parent actions
- `WithAnnotation(key, value string) *Action` - Adds annotation

**Helper Functions:**
```go
Principals(types ...string) []string    // Creates principal list
Resources(types ...string) []string     // Creates resource list
ActionGroup(id string) *ActionRef       // Creates action reference
QualifiedActionGroup(id, typeName string) *ActionRef  // Creates qualified action reference
```

## Common Type Declarations

```go
TypeDecl(name string, typ Type) *CommonTypeDecl
```

Declares a reusable common type that can be referenced by name.

## Type Annotations

All type implementations support annotations:

```go
type RecordType struct { ... }
func (r *RecordType) WithAnnotation(key, value string) *RecordType

type SetType struct { ... }
func (s *SetType) WithAnnotation(key, value string) *SetType

type PathType struct { ... }
func (p *PathType) WithAnnotation(key, value string) *PathType

type Attribute struct { ... }
func (a *Attribute) WithAnnotation(key, value string) *Attribute
```

## Usage Patterns

### Creating a Schema Programmatically

```go
s := schema.NewSchema().
    WithNamespace("MyApp",
        schema.NewEntity("User").
            WithAttribute("name", schema.String()).
            WithAttribute("age", schema.Long()),
        schema.NewAction("view").
            AppliesTo(
                schema.Principals("User"),
                schema.Resources("User"),
                nil,
            ),
    )
```

### Parsing from JSON

```go
var s schema.Schema
err := s.UnmarshalJSON(jsonBytes)
```

### Parsing from Cedar Format

```go
var s schema.Schema
s.SetFilename("schema.cedar")  // Optional, for better error messages
err := s.UnmarshalCedar(cedarBytes)
```

### Marshaling to JSON

```go
jsonBytes, err := s.MarshalJSON()
```

### Marshaling to Cedar Format

```go
cedarBytes, err := s.MarshalCedar()
```

## Type System Hierarchy

```
Type (interface)
├── PathType         (String, Long, Bool, Entity references)
├── SetType          (Set<Element>)
└── RecordType       (Record with attributes)
    └── Attribute    (name, type, required/optional)
```

## Declaration Hierarchy

```
Declaration (interface)
├── Entity           (Entity types with attributes, shapes, tags, enums)
├── Action           (Actions with principals, resources, context)
└── CommonTypeDecl   (Reusable type declarations)
```

## Comparison with x/exp/schema

The new `schema` package is a complete redesign with a fluent builder API:

**Old (x/exp/schema):**
- Parse-only (Cedar text format)
- Limited API
- No programmatic construction
- Located in experimental package

**New (schema):**
- ✅ Full builder API with fluent interface
- ✅ Both JSON and Cedar format support
- ✅ Programmatic construction and parsing
- ✅ Round-trip marshaling/unmarshaling
- ✅ Public, stable API
- ✅ 100% test coverage

## Migration Guide

The old `x/exp/schema` package is now a thin wrapper around the new `schema` package for backward compatibility. New code should use `github.com/cedar-policy/cedar-go/schema`.

**Before:**
```go
import "github.com/cedar-policy/cedar-go/x/exp/schema"

var s schema.Schema
s.UnmarshalCedar(data)
```

**After:**
```go
import "github.com/cedar-policy/cedar-go/schema"

// Same API, but can also construct programmatically:
s := schema.NewSchema().
    WithNamespace("App", ...)
```
