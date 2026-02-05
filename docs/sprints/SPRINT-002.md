# Sprint 002: Cedar Schema Rewrite

## Overview

Complete replacement of the Cedar schema package (`x/exp/schema/`). The existing implementation is discarded. This rewrite delivers:

1. Parsing and emitting Cedar schemas in JSON and Cedar text formats
2. Programmatic schema construction via a builder API
3. Resolution of raw schemas into fully-qualified types (`types.EntityType`, `types.EntityUID`)
4. Comprehensive testing: corpus (6,084+ schemas), Rust test matrix (114 cases), round-trips

Informed by the Cedar specification, the canonical Rust implementation, and Go idioms observed in this codebase.

## Non-Goals

- Policy validator
- LSP/editor integration
- Source location tracking
- Multi-fragment schema merging

## Use Cases

1. Parse Cedar schemas from `.cedarschema` files and JSON
2. Serialize schemas to Cedar text and JSON
3. Build schemas programmatically
4. Resolve to fully-qualified `types.EntityType` and `types.EntityUID`
5. Detect resolution errors: cycles, undefined types, shadowing, reserved names

## Architecture

```
x/exp/schema/
├── schema.go          # Schema type, public API surface
├── namespace.go       # Namespace, EntityTypeDef, ActionDef, CommonTypeDef
├── typexpr.go         # TypeExpr interface and variants
├── builder.go         # Fluent builder API and type constructors
├── json.go            # JSON marshal/unmarshal
├── cedar.go           # Cedar text marshal/unmarshal
├── resolve.go         # Resolution algorithm
├── errors.go          # Error types and sentinels
├── internal/
│   └── scan/
│       └── scanner.go # Cedar text tokenizer
└── resolved/
    └── resolved.go    # Resolved schema types
```

### Data Flow

```
                          ┌──────────┐
       JSON bytes ──────► │          │
                          │  Schema  │ ──► Resolve() ──► resolved.Schema
  Cedar text bytes ──────►│          │
                          │          │
     Builder calls ──────►└──────────┘ ──► MarshalJSON() ──► JSON bytes
                                       ──► MarshalCedar() ──► Cedar text
```

## Type System (Unresolved)

### TypeExpr Interface (`typexpr.go`)

Sealed interface with unexported marker method:

```go
type TypeExpr interface{ typeExpr() }
```

Variants:
- `PrimitiveTypeExpr{Kind PrimitiveKind}` — Long, String, Bool
- `SetTypeExpr{Element TypeExpr}` — Set<T>
- `RecordTypeExpr{Attributes map[string]*Attribute}` — Record with named attributes
- `EntityRefExpr{Name string}` — Explicit entity reference (JSON `{"type":"Entity","name":"..."}`)
- `ExtensionTypeExpr{Name string}` — ipaddr, decimal, datetime, duration
- `EntityNameExpr{Name string}` — Unresolved name in entity-only position (memberOf, principal, resource)
- `TypeNameExpr{Name string}` — Unresolved name in type position (attributes, common type bodies) — resolves via common > entity > primitive/extension priority

### Type Constructors (`builder.go`)

```go
func Long() TypeExpr
func String() TypeExpr
func Bool() TypeExpr
func Set(element TypeExpr) TypeExpr
func Record(attrs map[string]*Attribute) TypeExpr
func Entity(name string) TypeExpr
func Extension(name string) TypeExpr
func IPAddr() TypeExpr
func Decimal() TypeExpr
func Datetime() TypeExpr
func Duration() TypeExpr
func NamedType(name string) TypeExpr
```

## Schema Types (`namespace.go`)

```go
type Annotations map[string]string

type Namespace struct {
    EntityTypes map[string]*EntityTypeDef
    EnumTypes   map[string]*EnumTypeDef
    Actions     map[string]*ActionDef
    CommonTypes map[string]*CommonTypeDef
    Annotations Annotations
}

type EntityTypeDef struct {
    MemberOfTypes []string    // parent entity type names (unresolved)
    Shape         *RecordTypeExpr
    Tags          TypeExpr
    Annotations   Annotations
}

type EnumTypeDef struct {
    Values      []string
    Annotations Annotations
}

type ActionDef struct {
    MemberOf    []*ActionRef
    AppliesTo   *AppliesTo
    Annotations Annotations
}

type ActionRef struct {
    Type string // namespace-qualified action entity type (e.g., "MyNS::Action"), empty = same namespace
    ID   string // action name
}

type AppliesTo struct {
    PrincipalTypes []string    // entity type names
    ResourceTypes  []string    // entity type names
    Context        TypeExpr    // RecordTypeExpr or TypeNameExpr (common type ref)
}

type CommonTypeDef struct {
    Type        TypeExpr
    Annotations Annotations
}
```

## Schema Entry Point (`schema.go`)

```go
type Schema struct {
    Namespaces map[string]*Namespace
    filename   string
}

func (s *Schema) SetFilename(name string)
func (s *Schema) Filename() string
func (s *Schema) UnmarshalJSON(data []byte) error
func (s *Schema) MarshalJSON() ([]byte, error)
func (s *Schema) UnmarshalCedar(data []byte) error
func (s *Schema) MarshalCedar() ([]byte, error)
func (s *Schema) Resolve() (*resolved.Schema, error)
```

## Resolution Algorithm (`resolve.go`)

### Phase 1: Build Definition Index

Collect all fully-qualified names:
- `entityDefs` set: `NS::EntityName` or `EntityName` (empty namespace)
- `commonDefs` set: `NS::CommonName` or `CommonName`
- `actionDefs` set: `NS::Action::"ActionName"` or `Action::"ActionName"`

### Phase 2: Validate

- **Reserved names**: Entity and common types cannot use primitive/extension names (Bool, Boolean, Long, String, Record, Set, Entity, Extension)
- **RFC 70 shadowing**: If `NS::Foo` and `Foo` are both defined as the same kind → error
- **Common type cycles**: Topological sort (Kahn's algorithm). Remaining nodes with in-degree > 0 → cycle error

### Phase 3: Resolve Names

For `TypeNameExpr{Name: n}` in namespace `ns`:
1. Generate candidates: unqualified `n` in non-empty `ns` → `[ns::n, n]`; else `[n]`
2. For each candidate:
   a. If defined as common type → resolve (inline the common type's resolved definition)
   b. If defined as entity type → `resolved.EntityRef{EntityType: types.EntityType(candidate)}`
3. No match → `UndefinedTypeError`

For `EntityNameExpr{Name: n}` in namespace `ns`:
1. Same candidate generation
2. Only check entity type definitions (skip common types)
3. No match → `UndefinedTypeError`

For `EntityRefExpr{Name: n}`:
- Direct entity reference → `resolved.EntityRef{EntityType: types.EntityType(qualifiedName)}`

For primitives/extensions:
- Direct mapping to `resolved.Primitive` / `resolved.Extension`

### Action Resolution

For `ActionRef{Type: t, ID: id}`:
- If `t` is empty: qualify as `ns::Action::"id"` where `ns` is the containing namespace
- If `t` is present: use `t::"id"` directly
- Action must be declared in `actionDefs` → error if not found

For `AppliesTo.PrincipalTypes` / `ResourceTypes`:
- Each name resolves via `EntityNameExpr` rules (entity types only)

For `AppliesTo.Context`:
- If `RecordTypeExpr` → resolve recursively
- If `TypeNameExpr` → resolve via type-position rules (common > entity)

## Resolved Types (`resolved/resolved.go`)

```go
type Schema struct {
    Namespaces map[types.Path]*Namespace
}

type Namespace struct {
    EntityTypes map[types.EntityType]*EntityType
    EnumTypes   map[types.EntityType]*EnumType
    Actions     map[types.EntityUID]*Action
    Annotations Annotations
}

type EntityType struct {
    MemberOfTypes []types.EntityType
    Shape         *RecordType
    Tags          Type
    Annotations   Annotations
}

type EnumType struct {
    Values      []string
    Annotations Annotations
}

type Action struct {
    MemberOf       []types.EntityUID
    PrincipalTypes []types.EntityType
    ResourceTypes  []types.EntityType
    Context        *RecordType
    Annotations    Annotations
}

type Type interface{ resolvedType() }
type Primitive struct{ Kind PrimitiveKind }
type Set struct{ Element Type }
type RecordType struct{ Attributes map[string]*Attribute }
type EntityRef struct{ EntityType types.EntityType }
type Extension struct{ Name string }
```

CommonTypes are inlined during resolution and not present in the resolved output.

## Builder API (`builder.go`)

```go
s := schema.NewBuilder().
    Namespace("MyApp").
        Entity("User").MemberOf("Group").
            Attr("name", schema.String()).
            OptionalAttr("age", schema.Long()).
        Entity("Group").
        Action("view").
            Principal("User").Resource("Document").
    Build()
```

## JSON Format (`json.go`)

Type discriminated union:
- `{"type": "Long"}` → `PrimitiveTypeExpr{Long}`
- `{"type": "String"}` → `PrimitiveTypeExpr{String}`
- `{"type": "Bool"}` or `{"type": "Boolean"}` → `PrimitiveTypeExpr{Bool}`
- `{"type": "Set", "element": {...}}` → `SetTypeExpr`
- `{"type": "Record", "attributes": {...}}` → `RecordTypeExpr`
- `{"type": "Entity", "name": "Foo"}` → `EntityRefExpr{Name: "Foo"}`
- `{"type": "Extension", "name": "ipaddr"}` → `ExtensionTypeExpr`
- `{"type": "EntityOrCommon", "name": "Foo"}` → `TypeNameExpr{Name: "Foo"}`
- `{"type": "Foo"}` (not a keyword) → `TypeNameExpr{Name: "Foo"}`

## Cedar Text Format (`cedar.go`)

Hand-written recursive descent parser using `internal/scan` tokenizer. Emitter with sorted keys for deterministic output. Grammar follows the Cedar specification exactly.

## Implementation Plan

### Phase 1: Foundation Types
Files: `typexpr.go`, `namespace.go`, `schema.go`, `errors.go`
- TypeExpr interface and all variants (PrimitiveTypeExpr, SetTypeExpr, RecordTypeExpr, EntityRefExpr, ExtensionTypeExpr, EntityNameExpr, TypeNameExpr)
- Namespace, EntityTypeDef, EnumTypeDef, ActionDef, AppliesTo, ActionRef, CommonTypeDef, Attribute, Annotations
- Schema struct with SetFilename/Filename
- Error sentinels and typed errors

### Phase 2: Builder
File: `builder.go`
- Type constructors: Long(), String(), Bool(), Set(), Record(), Entity(), Extension(), IPAddr(), Decimal(), Datetime(), Duration(), NamedType()
- Fluent builder: NewBuilder() → SchemaBuilder → NamespaceBuilder → EntityBuilder, ActionBuilder

### Phase 3: JSON Serialization
File: `json.go`
- UnmarshalJSON/MarshalJSON for Schema, Namespace, all definition types
- Type expression discriminated union parsing
- Attribute parsing with required/annotations

### Phase 4: Cedar Text Parser
Files: `internal/scan/scanner.go`, `cedar.go`
- Scanner: tokenizer with position tracking
- Parser: recursive descent for full Cedar schema grammar
- Emitter: MarshalCedar with sorted, deterministic output

### Phase 5: Resolution
Files: `resolve.go`, `resolved/resolved.go`
- Resolved type definitions
- Definition index, reserved name check, shadowing check, cycle detection
- Name resolution with candidate generation and priority lookup
- Action resolution (member-of, principal/resource types, context)
- Common type inlining

### Phase 6: Testing
Files: `schema_test.go`, `resolve_test.go`
- Builder tests
- JSON round-trip tests
- Cedar text round-trip tests
- Cross-format tests (Cedar → JSON → Cedar)
- Resolution tests
- Error tests (cycles, undefined types, shadowing, reserved names, parse errors)
- Rust test_579 matrix (114 cases)
- Corpus test compatibility verification

## Compatibility Checklist

The following public API must be preserved for `corpus_test.go`:
- [x] `schema.Schema` type
- [x] `(*Schema).SetFilename(string)`
- [x] `(*Schema).UnmarshalCedar([]byte) error`

## Code Quality Gates

During implementation, enforce:
- No functions/types that exist solely for tests
- Comments explain why, never what
- No nil checks for values the type system guarantees non-nil
- No `any`/`interface{}` where concrete types work
- Match naming, error handling, and module patterns from existing codebase

## Definition of Done

1. `go test ./x/exp/schema/...` passes
2. `go test ./...` passes (including corpus tests)
3. `go vet ./...` clean
4. 100% code coverage across all packages: `go test -cover ./x/exp/schema/...` and `go test -cover ./x/exp/schema/internal/...` and `go test -cover ./x/exp/schema/resolved/...`
5. JSON round-trip: parse → marshal → parse = identical
6. Cedar text round-trip: parse → marshal → parse = identical
7. Resolution matches Rust for test_579 matrix (114 cases)
8. All corpus .cedarschema files parse without error
9. Cycle detection, shadowing validation, reserved name checks all work
10. Code quality gates satisfied

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cedar text parser edge cases | Medium | High | Test against 6,084+ corpus files |
| Resolution algorithm mismatches | Medium | High | Port test_579 matrix from Rust |
| JSON format undocumented variants | Low | Medium | Cross-reference Rust implementation |
| corpus_test.go API coupling | Low | High | Preserve Schema, SetFilename, UnmarshalCedar |

## Dependencies

- `types/` package: `EntityType`, `EntityUID`, `Path`
- Cedar docs: `/Volumes/src/ai/cedar-docs`
- Rust reference: `/Volumes/src/ai/cedar`
- Corpus: `corpus-tests.tar.gz`
- Go 1.23 standard library
