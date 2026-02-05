# Sprint 001: Cedar Schema Support

## Overview

This sprint implements comprehensive Cedar schema support for the cedar-go project. Cedar schemas define the structure of entity types, actions, and common types that applications use for authorization. This implementation enables:

1. **Validation readiness**: Provides the type information needed for future policy validation
2. **Type safety**: Programmatic schema construction with Go's type system catches errors at compile time
3. **Interoperability**: Round-trip parsing and serialization in both JSON and Cedar text formats

The implementation follows a three-phase resolution approach:
- **Parse**: Load schema to raw (unresolved) form
- **Qualify**: Apply namespace context to create resolution candidates
- **Resolve**: Select final fully-qualified names based on what's defined

## Non-Goals

- Full policy validator (separate sprint)
- LSP/editor integration
- Source location tracking
- Multi-fragment merging (single schema at a time)
- Partial evaluation or policy templates

## Use Cases

1. **Parse existing schemas**: Load schemas from JSON or Cedar text files
2. **Serialize schemas**: Export schemas to JSON or Cedar text for storage, transfer, or debugging
3. **Build schemas programmatically**: Construct schemas in Go code using builder pattern
4. **Resolve type references**: Convert to resolved form with `types.EntityType` and `types.EntityUID` values
5. **Unblock corpus tests**: Fix the broken `corpus_test.go` import

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Public API Layer                         │
│  x/exp/schema/schema.go                                         │
│  - Schema (raw/unresolved entry point)                          │
│  - ResolvedSchema (fully-qualified output)                      │
│  - UnmarshalJSON, UnmarshalCedar, MarshalJSON, MarshalCedar     │
│  - Resolve() → (*ResolvedSchema, error)                         │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      Type Layer                                 │
│  x/exp/schema/types.go                                          │
│  - Namespace (entity types, actions, common types)              │
│  - EntityTypeDef, ActionDef, CommonTypeDef                      │
│  - Type variants: Primitive, Set, Record, EntityRef, Extension  │
│  - Attributes, Annotations                                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                     Resolution Layer                            │
│  x/exp/schema/resolve.go                                        │
│  - AllDefs (tracks all defined types)                           │
│  - Disambiguation: common > entity > primitive/extension        │
│  - Cycle detection, shadowing validation                        │
└───────────────────────────┬─────────────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│                      Parser Layer                               │
│  x/exp/schema/json.go - JSON marshal/unmarshal                  │
│  x/exp/schema/parse_cedar.go - Cedar text parser                │
└─────────────────────────────────────────────────────────────────┘
```

### Key Data Structures

```go
// Schema is the raw (unresolved) schema entry point
type Schema struct {
    namespaces map[string]*Namespace // "" = empty namespace
    filename   string                // for error messages
}

// ResolvedSchema has all type references as concrete types.* values
type ResolvedSchema struct {
    Namespaces map[types.Path]*ResolvedNamespace
}

// ResolvedNamespace contains resolved definitions
type ResolvedNamespace struct {
    EntityTypes map[types.EntityType]*ResolvedEntityType
    Actions     map[types.EntityUID]*ResolvedAction  // e.g., Action::"viewPhoto"
    CommonTypes map[types.Path]*ResolvedCommonType
}

// ResolvedEntityType has parent types as types.EntityType
type ResolvedEntityType struct {
    MemberOfTypes []types.EntityType
    Shape         *ResolvedRecordType
    Tags          ResolvedType
    Enum          []string
}

// ResolvedAction has principal/resource types as types.EntityType
type ResolvedAction struct {
    MemberOf      []types.EntityUID    // action group membership
    PrincipalTypes []types.EntityType
    ResourceTypes  []types.EntityType
    Context       *ResolvedRecordType
}

// Namespace contains definitions within a namespace
type Namespace struct {
    EntityTypes map[string]*EntityTypeDef
    Actions     map[string]*ActionDef
    CommonTypes map[string]*CommonTypeDef
    Annotations Annotations
}

// EntityTypeDef describes an entity type
type EntityTypeDef struct {
    MemberOfTypes []string            // parent entity type names
    Shape         *RecordType         // attribute definitions
    Tags          Type                // optional tag type
    Enum          []string            // for enumerated entities
    Annotations   Annotations
}

// ActionDef describes an action
type ActionDef struct {
    MemberOf    []ActionRef          // action group membership
    AppliesTo   *AppliesTo           // principal/resource/context
    Annotations Annotations
}
```

### Builder Pattern API

```go
// Example: Build a schema programmatically
s := schema.New().
    Namespace("PhotoFlash").
        Entity("User").
            MemberOf("UserGroup").
            Attr("name", schema.String()).
            Attr("age", schema.Long()).
            Done().
        Entity("Photo").
            MemberOf("Album").
            Attr("private", schema.Bool()).
            Done().
        Action("viewPhoto").
            Principal("User").
            Resource("Photo").
            Done().
        Done().
    Build()

// Resolve to types.EntityType and types.EntityUID
resolved, err := s.Resolve()

// Access resolved types
userType := resolved.Namespaces["PhotoFlash"].EntityTypes[types.EntityType("PhotoFlash::User")]
// userType.MemberOfTypes[0] == types.EntityType("PhotoFlash::UserGroup")

viewAction := resolved.Namespaces["PhotoFlash"].Actions[types.NewEntityUID("PhotoFlash::Action", "viewPhoto")]
// viewAction.PrincipalTypes[0] == types.EntityType("PhotoFlash::User")
// viewAction.ResourceTypes[0] == types.EntityType("PhotoFlash::Photo")
```

## Implementation Plan

### Phase 1: Core Types (~15%)

**Files:**
- `x/exp/schema/types.go` - Core schema type definitions
- `x/exp/schema/builder.go` - Builder pattern implementation
- `x/exp/schema/annotations.go` - Annotation handling

**Tasks:**
- [ ] Define `Schema` struct with namespace map
- [ ] Define `Namespace` with EntityTypes, Actions, CommonTypes
- [ ] Define `EntityTypeDef` with MemberOfTypes, Shape, Tags, Enum
- [ ] Define `ActionDef` with MemberOf, AppliesTo
- [ ] Define `CommonTypeDef` as type alias
- [ ] Define type interfaces and variants (Primitive, Set, Record, EntityRef, Extension, CommonRef)
- [ ] Define `Attribute` struct with Name, Type, Required
- [ ] Define `Annotations` map type
- [ ] Implement builder pattern methods for all types
- [ ] Add `SetFilename(string)` for error context

### Phase 2: JSON Parsing (~20%)

**Files:**
- `x/exp/schema/json.go` - JSON marshaling/unmarshaling

**Tasks:**
- [ ] Implement custom JSON unmarshaler for `Schema`
- [ ] Handle empty namespace key ("" → nil internally)
- [ ] Implement discriminated union parsing for Type variants:
  - `"type": "Long"` / `"String"` / `"Boolean"` → Primitive
  - `"type": "Set", "element": {...}` → Set
  - `"type": "Record", "attributes": {...}` → Record
  - `"type": "Entity", "name": "..."` → EntityRef
  - `"type": "Extension", "name": "..."` → Extension
  - `"type": "EntityOrCommon", "name": "..."` → ambiguous (resolved later)
- [ ] Validate duplicate keys in entityTypes, actions, commonTypes
- [ ] Validate reserved type names: Bool, Boolean, Long, String, Set, Record, Entity, Extension
- [ ] Implement JSON marshaler for round-trip support

### Phase 3: Cedar Text Parsing (~25%)

**Files:**
- `x/exp/schema/parse_cedar.go` - Cedar schema parser
- `x/exp/schema/tokenize.go` - Schema-specific tokenization (if needed)

**Tasks:**
- [ ] Add schema keywords: `namespace`, `entity`, `action`, `type`, `in`, `appliesTo`, `enum`, `tags`, `Set`
- [ ] Implement namespace parsing (explicit block and bare declarations)
- [ ] Implement entity type parsing:
  - Membership: `entity User in [Group]`
  - Shape: `entity User { name: String, age: Long }`
  - Tags: `entity User {...} tags String`
  - Enum: `entity Status enum ["Active", "Inactive"]`
  - Multi-declaration: `entity A, B, C in [D]`
- [ ] Implement action parsing:
  - MemberOf: `action view in [read]`
  - AppliesTo: `action view appliesTo { principal: User, resource: Photo, context: {...} }`
  - Multi-declaration: `action read, write appliesTo {...}`
- [ ] Implement common type parsing: `type Context = { ip: ipaddr, ... }`
- [ ] Implement type expression parsing
- [ ] Handle annotations: `@doc("description")`
- [ ] Implement Cedar text output (MarshalCedar)

### Phase 4: Resolution (~25%)

**Files:**
- `x/exp/schema/resolve.go` - Name resolution logic
- `x/exp/schema/errors.go` - Resolution errors

**Tasks:**
- [ ] Implement `AllDefs` to track all defined types:
  - Primitives: `__cedar::Long`, `__cedar::String`, `__cedar::Bool` (+ aliases in empty namespace)
  - Extensions: `__cedar::ipaddr`, `__cedar::decimal`, `__cedar::datetime`, `__cedar::duration`
- [ ] Implement resolution algorithm:
  1. For each unqualified name, generate candidates: `[ns::name, name]`
  2. Check candidates in priority order: common > entity > primitive/extension
  3. First match wins; error if none found
  4. Convert resolved entity type names to `types.EntityType`
  5. Convert resolved action names to `types.EntityUID` (e.g., `NS::Action::"name"`)
- [ ] Implement `EntityOrCommon` resolution based on AllDefs
- [ ] Implement cycle detection in common type references
- [ ] Implement shadowing validation (can't shadow empty namespace defs from other namespaces)
- [ ] Define error types:
  - `CycleError` - circular common type definition
  - `UndefinedTypeError` - referenced type not found
  - `ShadowError` - illegal shadowing of empty namespace
  - `DuplicateError` - duplicate entity/action/common type name
  - `ReservedNameError` - use of reserved keyword as type name

### Phase 5: Integration & Testing (~15%)

**Files:**
- `x/exp/schema/schema.go` - Public API finalization
- `x/exp/schema/schema_test.go` - Unit tests
- `x/exp/schema/corpus_test.go` - Corpus integration tests
- `corpus_test.go` - Update import

**Tasks:**
- [ ] Finalize `Schema.Resolve() (*ResolvedSchema, error)`
- [ ] Extract test vectors from Rust unit tests as golden files
- [ ] Implement unit tests for each layer
- [ ] Implement JSON round-trip tests
- [ ] Implement Cedar round-trip tests (AST-equivalence, not text-equivalence)
- [ ] Implement corpus parsing tests (all 6,084 `.cedarschema` files)
- [ ] Update `corpus_test.go` to use new `x/exp/schema` package

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `x/exp/schema/schema.go` | Create | Public Schema and ResolvedSchema types |
| `x/exp/schema/types.go` | Create | Core type definitions |
| `x/exp/schema/builder.go` | Create | Builder pattern API |
| `x/exp/schema/annotations.go` | Create | Annotation handling |
| `x/exp/schema/json.go` | Create | JSON marshal/unmarshal |
| `x/exp/schema/parse_cedar.go` | Create | Cedar text parser |
| `x/exp/schema/resolve.go` | Create | Name resolution |
| `x/exp/schema/errors.go` | Create | Typed error definitions |
| `x/exp/schema/schema_test.go` | Create | Unit tests |
| `x/exp/schema/corpus_test.go` | Create | Corpus integration tests |
| `corpus_test.go` | Modify | Update schema import |

## Definition of Done

- [ ] All 6,084 corpus `.cedarschema` files parse without error
- [ ] JSON round-trip produces equivalent schema (AST comparison)
- [ ] Cedar round-trip produces equivalent schema (AST comparison)
- [ ] Builder API produces schemas equivalent to parsed forms
- [ ] Resolution produces `types.EntityType` for entity type references
- [ ] Resolution produces `types.EntityUID` for action references (e.g., `Action::"viewPhoto"`)
- [ ] Cycle detection catches circular common type definitions
- [ ] Shadowing validation catches illegal empty namespace shadows
- [ ] All documented error conditions return appropriate error types
- [ ] `corpus_test.go` compiles and runs successfully
- [ ] `go vet` passes
- [ ] `golangci-lint` passes
- [ ] Unit test coverage > 80%

## Milestones

- **M1**: Core types + JSON parsing - JSON round-trip works for basic schemas
- **M2**: Cedar parser - Cedar round-trip works for controlled samples
- **M3**: Resolver - Disambiguation and qualification pass unit tests
- **M4**: Builder API - Programmatic construction matches parsed forms
- **M5**: Corpus - All corpus schemas parse; error conditions tested

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Resolution edge cases not matching Rust | Medium | High | Extract test vectors from Rust tests; differential testing |
| Cedar grammar ambiguities | Low | Medium | Follow grammar spec exactly; test all productions |
| Error message parity | Medium | Medium | Extract expected error strings from Rust as test fixtures |
| Performance with large schemas | Low | Low | Profile after correctness; optimize if needed |
| Corpus format drift | Low | Medium | Version-pin corpus; update with Cedar releases |

## Security Considerations

- Reject malformed JSON/Cedar input early to prevent resource exhaustion
- Limit recursion depth during parsing and cycle detection
- Don't expose internal file paths in error messages to end users

## Dependencies

- `types/` package:
  - `types.Path` - namespace names, common type names
  - `types.EntityType` - resolved entity type references (alias for `Path`)
  - `types.EntityUID` - resolved action references (Type + ID)
  - `types.String` - for EntityUID IDs
- `internal/parser/` - tokenizer patterns (reuse where practical)
- No new external dependencies

## Test Vector Extraction

Extract test cases from Rust implementation:

1. **Positive cases**: Schemas that should parse successfully
   - Location: `cedar-policy-core/src/validator/` test files
   - Extract: schema text + expected resolved type names

2. **Negative cases**: Schemas that should fail with specific errors
   - Cycle detection failures
   - Undefined type references
   - Shadowing violations
   - Duplicate definitions

3. **Resolution edge cases**:
   - `__cedar::` prefix disambiguation
   - Same name in common vs entity types
   - Forward references in common types
   - Cross-namespace references

## Extension Types

The following extension types are seeded in the `__cedar` namespace (and aliased in empty namespace if not shadowed):

- `ipaddr` - IP address type
- `decimal` - Decimal number type
- `datetime` - Date/time type (RFC 80)
- `duration` - Duration type (RFC 80)

## Round-Trip Semantics

- **JSON**: Byte-equivalent after re-parsing (modulo whitespace)
- **Cedar**: AST-equivalent, not text-equivalent
  - Formatting, whitespace, and comment preservation not required
  - Order of declarations may vary
  - Semantically identical schemas compare equal
