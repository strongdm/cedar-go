# Sprint 001: Cedar Schema Support

## Overview

This sprint implements comprehensive Cedar schema support for the cedar-go project. Cedar schemas define the structure of entity types, actions, and common types that your application uses for authorization. Schema support enables:

1. **Validation readiness**: While this sprint doesn't implement the validator itself, the resolved schema provides the type information needed for policy validation
2. **Type safety**: Programmatic schema construction with Go's type system catches errors at compile time
3. **Interoperability**: Round-trip parsing and serialization in both JSON and Cedar text formats

The implementation follows the Rust reference implementation architecture with a three-phase approach: parse to raw types, convert to conditional names, then resolve to fully-qualified internal names. This mirrors the progression `RawName` → `ConditionalName` → `InternalName` in Rust.

## Use Cases

1. **Parse existing schemas**: Load schemas from JSON or Cedar text files to understand application authorization structure
2. **Serialize schemas**: Export schemas to JSON or Cedar text for storage, transfer, or debugging
3. **Build schemas programmatically**: Construct schemas in Go code for applications that need dynamic schema generation
4. **Resolve type references**: Convert parsed or constructed schemas to a resolved form where all type references are fully qualified
5. **Corpus testing**: Unblock the corpus tests that currently fail due to missing `x/exp/schema` package

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Public API Layer                         │
│  x/exp/schema/schema.go                                         │
│  - Schema type (main entry point)                               │
│  - UnmarshalJSON, UnmarshalCedar, MarshalJSON, MarshalCedar     │
│  - Resolve() → ResolvedSchema                                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      AST/Type Layer                             │
│  x/exp/schema/types.go                                          │
│  - Fragment[N]      (map of namespace → NamespaceDefinition)    │
│  - NamespaceDefinition[N] (entity types, actions, common types) │
│  - EntityType[N], ActionType[N], CommonType[N], Type[N]         │
│  - Attributes, TypeVariant (Long, String, Bool, Set, Record...) │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                     Resolution Layer                            │
│  x/exp/schema/resolve.go                                        │
│  - RawName → ConditionalName → InternalName                     │
│  - AllDefs (tracks defined types across namespaces)             │
│  - Type disambiguation (common > entity > primitive/extension)  │
│  - Cycle detection in common types                              │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      Parser Layer                               │
│  x/exp/schema/parse_cedar.go                                    │
│  - Tokenizer extension for schema keywords                      │
│  - Grammar implementation (namespace, entity, action, type)     │
│  x/exp/schema/parse_json.go                                     │
│  - JSON unmarshaling with custom deserializers                  │
└─────────────────────────────────────────────────────────────────┘
```

### Type Parameter Progression

Following Rust's pattern, schema types are generic over name representation:

```go
// Phase 1: Parsed (unresolved names)
type Fragment[N TypeName] struct { ... }

// Phase 2: Conditional (multiple resolution candidates)
type ConditionalName struct {
    Possibilities []InternalName  // tried in priority order
    ReferenceType ReferenceType   // entity or common
    Raw           RawName
}

// Phase 3: Resolved (fully qualified)
// Uses types.EntityType (which is types.Path) directly
```

### Key Data Structures

```go
// Fragment represents a schema, potentially spanning multiple namespaces
type Fragment[N TypeName] struct {
    Namespaces map[string]*NamespaceDefinition[N] // "" = empty namespace
}

// NamespaceDefinition contains all definitions within a namespace
type NamespaceDefinition[N TypeName] struct {
    CommonTypes map[string]*CommonType[N]
    EntityTypes map[string]*EntityType[N]
    Actions     map[string]*ActionType[N]
    Annotations Annotations
}

// EntityType describes an entity type in the schema
type EntityType[N TypeName] struct {
    MemberOfTypes []N                 // parent entity types
    Shape         *RecordType[N]      // attribute definitions
    Tags          Type[N]             // tag type (optional)
    Enum          []string            // for enumerated entities
    Annotations   Annotations
}

// ActionType describes an action in the schema
type ActionType[N TypeName] struct {
    MemberOf  []ActionRef           // action group membership
    AppliesTo *AppliesTo[N]         // principal/resource/context
    Annotations Annotations
}

// Type represents any Cedar type in the schema
type Type[N TypeName] interface {
    typeNode()
}

// Concrete type variants
type PrimitiveType struct { Kind PrimitiveKind } // Long, String, Bool
type EntityRef[N TypeName] struct { Name N }
type ExtensionType struct { Name string }       // ipaddr, decimal, datetime, duration
type SetType[N TypeName] struct { Element Type[N] }
type RecordType[N TypeName] struct { Attributes map[string]*Attribute[N] }
type CommonTypeRef[N TypeName] struct { Name N }
type EntityOrCommon[N TypeName] struct { Name N } // ambiguous until resolution
```

## Implementation Plan

### Phase 1: Core Types (~25%)

**Files:**
- `x/exp/schema/types.go` - Core schema type definitions
- `x/exp/schema/name.go` - Name types (RawName, ConditionalName, InternalName)
- `x/exp/schema/annotations.go` - Annotation handling

**Tasks:**
- [ ] Define `TypeName` interface constraint
- [ ] Implement `RawName` type (wraps string path)
- [ ] Implement `ConditionalName` with resolution priority
- [ ] Define `Fragment[N]` generic struct
- [ ] Define `NamespaceDefinition[N]` struct
- [ ] Define `EntityType[N]` with MemberOfTypes, Shape, Tags, Enum
- [ ] Define `ActionType[N]` with MemberOf, AppliesTo
- [ ] Define `CommonType[N]` as type alias wrapper
- [ ] Define type variant interfaces and structs (Primitive, Set, Record, Entity, Extension)
- [ ] Define `Annotations` map type with get/set methods
- [ ] Add `Attribute` struct with Name, Type, Required, Annotations

### Phase 2: JSON Parsing (~20%)

**Files:**
- `x/exp/schema/json.go` - JSON marshaling/unmarshaling

**Tasks:**
- [ ] Implement custom JSON unmarshaler for `Fragment`
- [ ] Handle empty namespace key ("" → nil)
- [ ] Implement `Type` variant JSON unmarshaling (discriminated by "type" field)
- [ ] Handle `EntityOrCommon` type (resolves later)
- [ ] Implement custom JSON marshaler for round-trip support
- [ ] Add validation for duplicate keys (entity types, actions, common types)
- [ ] Add validation for reserved type names (Bool, Long, String, Set, Record, Entity, Extension)

### Phase 3: Cedar Text Parsing (~25%)

**Files:**
- `x/exp/schema/parse_cedar.go` - Cedar schema parser
- `x/exp/schema/tokenize.go` - Schema-specific tokenization

**Tasks:**
- [ ] Extend tokenizer for schema keywords: `namespace`, `entity`, `action`, `type`, `in`, `appliesTo`, `enum`, `tags`, `Set`
- [ ] Implement namespace parsing (with and without explicit namespace block)
- [ ] Implement entity type parsing with membership, shape, tags, enum variants
- [ ] Implement action parsing with memberOf and appliesTo
- [ ] Implement common type parsing
- [ ] Implement type expression parsing (primitives, Set<>, records, entity refs)
- [ ] Handle annotations (@doc, etc.)
- [ ] Support multi-entity declarations (`entity A, B, C in [D]`)
- [ ] Support multi-action declarations (`action read, write appliesTo {...}`)

### Phase 4: Resolution (~20%)

**Files:**
- `x/exp/schema/resolve.go` - Name resolution logic
- `x/exp/schema/alldefs.go` - Definition tracking

**Tasks:**
- [ ] Implement `AllDefs` struct to track all defined types across namespaces
- [ ] Add primitive types to AllDefs (`__cedar::Long`, `__cedar::String`, `__cedar::Bool`)
- [ ] Add extension types to AllDefs (`__cedar::ipaddr`, `__cedar::decimal`, etc.)
- [ ] Implement `RawName.ConditionallyQualifyWith(namespace)` → `ConditionalName`
- [ ] Implement `ConditionalName.Resolve(AllDefs)` → `types.Path` (InternalName)
- [ ] Apply resolution priority: common type > entity type > primitive/extension
- [ ] Implement cycle detection in common type references
- [ ] Implement shadowing validation (can't shadow empty namespace defs)
- [ ] Convert `EntityOrCommon` to specific variant based on AllDefs

### Phase 5: Public API & Integration (~10%)

**Files:**
- `x/exp/schema/schema.go` - Public Schema type and methods

**Tasks:**
- [ ] Define `Schema` as wrapper around `Fragment[RawName]`
- [ ] Implement `UnmarshalJSON([]byte) error`
- [ ] Implement `UnmarshalCedar([]byte) error`
- [ ] Implement `MarshalJSON() ([]byte, error)`
- [ ] Implement `MarshalCedar() ([]byte, error)`
- [ ] Implement `Resolve() (*ResolvedSchema, error)`
- [ ] Define `ResolvedSchema` with `Fragment[types.Path]`
- [ ] Add `SetFilename(string)` for error reporting
- [ ] Expose namespace iteration methods

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `x/exp/schema/schema.go` | Create | Public API entry point |
| `x/exp/schema/types.go` | Create | Core type definitions |
| `x/exp/schema/name.go` | Create | RawName, ConditionalName types |
| `x/exp/schema/annotations.go` | Create | Annotation handling |
| `x/exp/schema/json.go` | Create | JSON marshal/unmarshal |
| `x/exp/schema/parse_cedar.go` | Create | Cedar text parser |
| `x/exp/schema/tokenize.go` | Create | Schema tokenization |
| `x/exp/schema/resolve.go` | Create | Name resolution |
| `x/exp/schema/alldefs.go` | Create | Definition tracking |
| `x/exp/schema/errors.go` | Create | Schema-specific errors |
| `x/exp/schema/schema_test.go` | Create | Unit tests |
| `x/exp/schema/json_test.go` | Create | JSON round-trip tests |
| `x/exp/schema/parse_cedar_test.go` | Create | Cedar parsing tests |
| `x/exp/schema/resolve_test.go` | Create | Resolution tests |
| `x/exp/schema/corpus_test.go` | Create | Corpus schema tests |
| `corpus_test.go` | Modify | Remove/update schema import once implemented |

## Definition of Done

- [ ] All 6,084 corpus `.cedarschema` files parse without error
- [ ] JSON schema → Schema → JSON produces equivalent output
- [ ] Cedar schema → Schema → Cedar produces semantically equivalent output
- [ ] Schema can be constructed programmatically using builder pattern
- [ ] Resolution produces fully-qualified `types.Path` for all type references
- [ ] Cycle detection catches circular common type definitions
- [ ] Shadowing validation catches illegal shadowing of empty namespace
- [ ] All error conditions documented in Rust implementation are handled
- [ ] Unit tests cover edge cases from documentation
- [ ] `corpus_test.go` compiles and runs (currently broken by missing import)
- [ ] No compiler warnings in new code
- [ ] `go vet` passes
- [ ] `golangci-lint` passes

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Resolution edge cases not matching Rust | Medium | High | Differential testing against `cedar validate` CLI |
| Cedar grammar ambiguities | Low | Medium | Follow grammar spec exactly, test all productions |
| Performance with large schemas | Low | Low | Profile against corpus; optimize if needed |
| Missing error messages | Medium | Medium | Extract error formats from Rust implementation |
| Corpus format changes | Low | Medium | Version-pin corpus, update with Cedar releases |

## Security Considerations

- **Input validation**: Reject malformed JSON/Cedar input early to prevent DoS
- **Memory limits**: Consider limiting schema size to prevent resource exhaustion
- **Error information**: Don't expose internal paths in error messages

## Dependencies

- No external dependencies (uses only stdlib + existing cedar-go internal packages)
- Depends on `types.Path`, `types.EntityType`, `types.EntityUID` from `types/` package
- Reuses tokenizer infrastructure from `internal/parser/`

## Open Questions

1. **Generic vs concrete types**: Should we use Go generics for `Fragment[N]` or separate concrete types (`RawFragment`, `ConditionalFragment`, `ResolvedFragment`)?
   - **Recommendation**: Use concrete types for simplicity; Go generics add complexity without significant benefit here since each phase has different methods.

2. **Builder pattern style**: Should programmatic construction use method chaining or struct literals?
   - **Recommendation**: Follow `ast/` package patterns - provide both literal construction and builder methods.

3. **Error types**: Dedicated error types or `fmt.Errorf` with sentinels?
   - **Recommendation**: Dedicated types for errors that callers need to match (e.g., `CycleError`, `UndefinedTypeError`); `fmt.Errorf` for others.

4. **Source location tracking**: Include optional `Loc` fields for LSP support?
   - **Recommendation**: Add `Loc` fields but make population optional; helps future LSP integration.
