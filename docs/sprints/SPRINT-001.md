# Sprint 001: Programmatic Cedar Schema Interface

## Overview

This sprint introduces a programmatic interface for Cedar schema in cedar-go, following the established policy builder pattern. The implementation provides two distinct schema representations: a "parsed" schema that preserves the structure from JSON/Cedar text (including unqualified type references), and a "resolved" schema with fully-qualified EntityTypes, expanded common types, and computed entity hierarchy.

The design mirrors the Rust implementation's separation between `json_schema::Fragment` (parsed) and `ValidatorSchema` (resolved), while providing a Go-idiomatic fluent builder API that matches the existing policy AST builder patterns in cedar-go. This enables users to either parse existing schemas from JSON/Cedar text or construct schemas programmatically, then resolve them for use in validation.

**Key Design Decisions**:
- All code in `x/exp/schema2/` - completely separate from existing `x/exp/schema/`
- Minimal public API surface with internal implementation in `x/exp/schema2/internal/`
- Method chaining builder API (fluent style like policy builder)
- Common types inlined during resolution
- Validation performed during resolution (returns errors for invalid schemas)
- ResolvedSchema is immutable with accessor methods

## Use Cases

1. **Parse and use existing schemas**: Load schemas from JSON or Cedar text format, resolve them, and use for validation
2. **Programmatic schema construction**: Build schemas in code using fluent builder API for testing, code generation, or dynamic schema creation
3. **Schema transformation**: Parse a schema, modify it programmatically, and serialize back to JSON or Cedar text
4. **Schema resolution**: Convert parsed schemas with potentially unqualified names into fully-resolved schemas suitable for validation
5. **Enum entity support**: Define and work with enumerated entity types that map to specific EntityUIDs

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Public API (x/exp/schema2)                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────┐         ┌─────────────────────────────────┐   │
│  │     Schema      │         │       ResolvedSchema            │   │
│  │  (parsed/built) │ ──────► │  (fully-qualified types)        │   │
│  │                 │ Resolve │                                 │   │
│  │ - Namespaces    │         │  - EntityTypes map              │   │
│  │ - Entities      │         │  - Actions map                  │   │
│  │ - Actions       │         │  - Descendants computed         │   │
│  │ - CommonTypes   │         │  - CommonTypes inlined          │   │
│  └─────────────────┘         └─────────────────────────────────┘   │
│         ▲                                                           │
│         │ Build                                                     │
│  ┌──────┴──────────────────────────────────────────────────────┐   │
│  │          Minimal Builder Entry Points                        │   │
│  │  NewSchema(), ParseJSON(), ParseCedar()                      │   │
│  │  Type helpers: String(), Long(), Bool(), Set(), Record()    │   │
│  └──────────────────────────────────────────────────────────────┘   │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ uses
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│               Internal (x/exp/schema2/internal)                     │
├─────────────────────────────────────────────────────────────────────┤
│  internal/ast:    Schema AST types (unresolved)                     │
│  internal/resolve: Resolution logic, transitive closure             │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ reuses
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│              Existing Internal (internal/schema/ast)                │
├─────────────────────────────────────────────────────────────────────┤
│  JSONSchema, JSONNamespace, JSONEntity, JSONAction, JSONType        │
│  Schema (human), Namespace, Entity, Action, CommonTypeDecl          │
│  Parser, Formatter                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

### Key Type Mappings

| Parsed Schema (Schema)      | Resolved Schema (ResolvedSchema) |
|-----------------------------|----------------------------------|
| string entity type names    | types.EntityType (fully-qualified) |
| string action references    | types.EntityUID |
| CommonType references       | Types inlined |
| memberOfTypes (parents)     | Descendants computed (children) |
| Enum values (strings)       | types.EntityUID set |

### Example Usage

```go
// Build a schema programmatically
schema := schema2.NewSchema().
    Namespace("MyApp").
        Entity("User").In("Group").Attributes(
            schema2.Attr("name", schema2.String(), true),
            schema2.Attr("email", schema2.String(), true),
        ).
        Entity("Group").
        Entity("Document").In("Folder").Attributes(
            schema2.Attr("owner", schema2.Entity("User"), true),
        ).
        Action("read").
            Principals("User").
            Resources("Document").
        Action("write").
            Principals("User").
            Resources("Document")

// Resolve to get fully-qualified types
resolved, err := schema.Resolve()
if err != nil {
    // Handle validation errors (undefined types, cycles, etc.)
}

// Access resolved entity types
userType := resolved.EntityType(types.EntityType("MyApp::User"))
```

## Implementation Plan

### Phase 1: Internal AST and Type Definitions (~20%)

**Files:**
- `x/exp/schema2/internal/ast/types.go` - Internal schema AST types
- `x/exp/schema2/internal/ast/json.go` - JSON marshaling (reuses internal/schema/ast)

**Tasks:**
- [ ] Define internal `Schema` struct wrapping namespace definitions
- [ ] Define internal `Namespace` struct (entities, actions, common types)
- [ ] Define internal `EntityType` struct (name, memberOf, shape, tags, enum)
- [ ] Define internal `Action` struct (name, memberOf, appliesTo)
- [ ] Define internal `Type` interface (Primitive, Entity, Set, Record, Extension)
- [ ] Wire up JSON marshal/unmarshal using existing internal/schema/ast/json.go types
- [ ] Wire up Cedar text marshal/unmarshal using existing internal/schema/parser

### Phase 2: Resolved Schema Types (~20%)

**Files:**
- `x/exp/schema2/resolved.go` - Public resolved schema types

**Tasks:**
- [ ] Define `ResolvedSchema` struct with EntityTypes and Actions maps (immutable)
- [ ] Define `ResolvedEntityType` struct (name, descendants, attributes, tags, kind)
- [ ] Define `ResolvedEntityTypeKind` (Standard vs Enum)
- [ ] Define `ResolvedAction` struct (name, memberOf, appliesTo, context)
- [ ] Define `ResolvedAppliesTo` struct (principals, resources, context type)
- [ ] Define `ResolvedType` interface and implementations (Primitive, Entity, Set, Record)
- [ ] Define `ResolvedAttribute` struct (type, required)
- [ ] Add accessor methods for resolved schema introspection (return copies/iterators)

### Phase 3: Schema Builder API (~25%)

**Files:**
- `x/exp/schema2/schema.go` - Public Schema type and builder entry points
- `x/exp/schema2/builder.go` - Builder types and methods
- `x/exp/schema2/types.go` - Public type helpers

**Tasks:**
- [ ] Implement `NewSchema()` returning `*Schema` builder
- [ ] Implement `Schema.Namespace(name)` returning `*Schema` (method chaining)
- [ ] Implement `Schema.Entity(name)` returning `*Schema` (method chaining)
- [ ] Implement `Schema.In(types...)` for parent entity types (method chaining)
- [ ] Implement `Schema.Attributes(attrs...)` for entity shape (method chaining)
- [ ] Implement `Schema.Tags(Type)` for entity tags (method chaining)
- [ ] Implement `Schema.Enum(values...)` for enumerated entities (method chaining)
- [ ] Implement `Schema.Action(name)` returning `*Schema` (method chaining)
- [ ] Implement `Schema.Principals(types...)` for appliesTo principals (method chaining)
- [ ] Implement `Schema.Resources(types...)` for appliesTo resources (method chaining)
- [ ] Implement `Schema.Context(Type)` for appliesTo context (method chaining)
- [ ] Implement `Schema.CommonType(name, Type)` for type aliases (method chaining)
- [ ] Implement type helpers: `String()`, `Long()`, `Bool()`, `Set(Type)`, `Record(attrs...)`, `Entity(name)`
- [ ] Implement `Attr(name, type, required)` attribute helper

### Phase 4: Schema Resolution (~25%)

**Files:**
- `x/exp/schema2/internal/resolve/resolve.go` - Resolution logic
- `x/exp/schema2/internal/resolve/qualify.go` - Name qualification
- `x/exp/schema2/internal/resolve/tc.go` - Transitive closure computation
- `x/exp/schema2/resolve.go` - Public Resolve() method

**Tasks:**
- [ ] Implement name qualification (prepend namespace to unqualified names)
- [ ] Implement common type resolution (inline type definitions)
- [ ] Implement entity type resolution with fully-qualified types.EntityType
- [ ] Implement action resolution with types.EntityUID
- [ ] Implement enum resolution (convert string values to EntityUIDs)
- [ ] Compute transitive closure for entity hierarchy (descendants from memberOf)
- [ ] Implement validation during resolution (undefined types, cycles)
- [ ] Define resolution error types with descriptive messages
- [ ] Implement public `Schema.Resolve()` method returning `(*ResolvedSchema, error)`

### Phase 5: Parse/Serialize Integration (~10%)

**Files:**
- `x/exp/schema2/parse.go` - Parsing helpers
- `x/exp/schema2/marshal.go` - Serialization methods

**Tasks:**
- [ ] Implement `ParseJSON([]byte) (*Schema, error)` - parse JSON schema
- [ ] Implement `ParseCedar([]byte) (*Schema, error)` - parse Cedar text schema
- [ ] Implement `Schema.MarshalJSON() ([]byte, error)`
- [ ] Implement `Schema.MarshalCedar() ([]byte, error)`
- [ ] Ensure round-trip compatibility (parse -> serialize -> parse)

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| **Public API (x/exp/schema2/)** | | |
| `x/exp/schema2/schema.go` | Create | Public Schema type with builder entry points |
| `x/exp/schema2/builder.go` | Create | Builder types and method chaining implementation |
| `x/exp/schema2/types.go` | Create | Type helpers (String, Long, Bool, Set, Record, Entity) |
| `x/exp/schema2/resolved.go` | Create | ResolvedSchema and related immutable types |
| `x/exp/schema2/resolve.go` | Create | Public Resolve() method |
| `x/exp/schema2/parse.go` | Create | ParseJSON(), ParseCedar() functions |
| `x/exp/schema2/marshal.go` | Create | MarshalJSON(), MarshalCedar() methods |
| `x/exp/schema2/schema_test.go` | Create | Tests for builder and serialization |
| `x/exp/schema2/resolve_test.go` | Create | Tests for resolution |
| `x/exp/schema2/example_test.go` | Create | Example usage documentation |
| **Internal (x/exp/schema2/internal/)** | | |
| `x/exp/schema2/internal/ast/types.go` | Create | Internal AST types for unresolved schema |
| `x/exp/schema2/internal/ast/json.go` | Create | JSON marshaling adapter |
| `x/exp/schema2/internal/resolve/resolve.go` | Create | Core resolution logic |
| `x/exp/schema2/internal/resolve/qualify.go` | Create | Name qualification |
| `x/exp/schema2/internal/resolve/tc.go` | Create | Transitive closure computation |

## Definition of Done

- [ ] Schema builder API allows fluent method-chaining construction of schemas
- [ ] Schemas can be created via builder or parsed from JSON/Cedar
- [ ] `Schema.Resolve()` produces an immutable `ResolvedSchema` with fully-qualified types
- [ ] ResolvedSchema contains computed descendants for entity hierarchy
- [ ] Enum entity types resolve to EntityUID sets
- [ ] Common types are inlined during resolution
- [ ] Resolution validates and returns errors for invalid schemas (undefined types, cycles)
- [ ] Built schemas can be serialized to JSON and Cedar formats
- [ ] API matches patterns from policy builder (fluent, composable, method chaining)
- [ ] Example tests demonstrate all use cases
- [ ] All existing tests pass
- [ ] New tests cover builder, resolution, and round-trip scenarios

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Namespace resolution edge cases | Medium | High | Study Rust implementation thoroughly, add comprehensive tests |
| Transitive closure performance | Low | Medium | Use efficient algorithm, optimize only if benchmarks show need |
| API ergonomics unclear | Medium | Medium | Start with minimal API, gather feedback, iterate |
| Breaking changes to internal AST | Low | High | Keep internal types stable, adapt builder layer if needed |

## Security Considerations

- Schema resolution must handle malformed input gracefully (return errors, not panic)
- Transitive closure computation must detect cycles to prevent infinite loops
- Entity/action name validation should prevent injection of invalid characters

## Dependencies

- No external dependencies required
- Uses existing `internal/schema/ast` for JSON/human schema representation
- Uses existing `types` package for EntityType, EntityUID

## Design Decisions (from Interview)

1. **Builder API style**: Method chaining where each builder method returns the schema for continued chaining, similar to the existing policy builder pattern.

2. **Common types handling**: Inlined during resolution - common type references are expanded to their definitions wherever they appear in the resolved schema.

3. **Validation during resolution**: The `Resolve()` method performs validation and returns errors for invalid schemas (undefined type references, cycles in entity hierarchy).

4. **ResolvedSchema immutability**: The resolved schema is immutable after creation. Accessor methods return copies or iterators rather than exposing mutable internal state.
