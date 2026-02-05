# Sprint 002: Cedar Schema Rewrite

## Overview

Complete replacement of the Cedar schema package (`x/exp/schema/`). The existing implementation works but does not meet the quality bar. This rewrite delivers the same capabilities from scratch:

1. Parse and emit Cedar schemas in JSON and Cedar text formats
2. Programmatic schema construction via a builder API
3. Resolution of raw schemas into fully-qualified types (using `types.EntityType`, `types.EntityUID`)
4. Comprehensive testing against corpus, Rust test matrix, and round-trip verification

The rewrite takes no inspiration from the existing code. It is informed by the Cedar specification, the canonical Rust implementation, and Go idioms observed in the rest of this codebase.

## Non-Goals

- Policy validator (separate sprint)
- LSP/editor integration
- Source location tracking in AST nodes
- Multi-fragment schema merging (single schema at a time)

## Use Cases

1. Parse Cedar schemas from `.cedarschema` files (corpus tests)
2. Parse Cedar schemas from JSON format
3. Serialize schemas to Cedar text and JSON
4. Build schemas programmatically for testing and code generation
5. Resolve raw schemas to get fully-qualified `types.EntityType` and `types.EntityUID` values
6. Detect and report resolution errors: cycles, undefined types, shadowing, reserved names

## Architecture

```
x/exp/schema/
├── schema.go          # Schema type, public API surface
├── namespace.go       # Namespace, EntityTypeDef, ActionDef, CommonTypeDef, etc.
├── typexpr.go         # TypeExpr interface and variants (the unresolved type system)
├── builder.go         # Fluent builder API
├── json.go            # JSON marshal/unmarshal
├── cedar.go           # Cedar text marshal/unmarshal
├── resolve.go         # Resolution algorithm (raw → resolved)
├── errors.go          # Error types and sentinels
├── internal/
│   └── scan/
│       └── scanner.go # Low-level tokenizer for Cedar text
└── resolved/
    └── resolved.go    # Resolved schema types (output of resolution)
```

### Data Flow

```
                          ┌──────────┐
       JSON bytes ──────► │          │
                          │  Schema  │ ──► Resolve() ──► resolved.Schema
  Cedar text bytes ──────►│          │
                          │          │
     Builder calls ──────►└──────────┘ ──► MarshalJSON() ──► JSON bytes
                                       ──► MarshalCedar() ──► Cedar text bytes
```

### Key Design Decisions

**1. `TypeExpr` instead of `Type` for unresolved types.**
The unresolved type system represents type *expressions* — references that haven't been resolved yet. Calling it `TypeExpr` distinguishes it from the resolved `Type` in the `resolved` package and avoids colliding with Go's `reflect.Type`.

**2. Sealed interface via unexported marker method.**
`TypeExpr` is a sealed interface (`typeExpr()` marker) with concrete variants: `PrimitiveTypeExpr`, `SetTypeExpr`, `RecordTypeExpr`, `EntityRefExpr`, `ExtensionTypeExpr`, `NameExpr`. The last variant replaces the three-way `CommonTypeRef`/`EntityOrCommonRef`/`EntityRef` split — resolution determines what a bare name refers to.

**3. Single `NameExpr` for all unresolved name references.**
In the raw schema, a bare name like `"User"` could refer to a common type, entity type, or primitive/extension. Rather than pre-classifying into `CommonTypeRef` vs `EntityOrCommonRef`, use a single `NameExpr{Name: "User"}` and let resolution disambiguate. A `MustEntity bool` flag on `NameExpr` handles positions where only entity types are valid (e.g., `memberOfTypes`, `principalTypes`, `resourceTypes`).

**4. Resolution produces `resolved.Schema` in a separate package.**
Clean separation: the `schema` package deals in strings and `TypeExpr`, the `resolved` package deals in `types.EntityType`, `types.EntityUID`, and `resolved.Type`. Resolution is the bridge.

**5. Internal scanner, not a full parser sub-package.**
The Cedar text parser is a recursive descent parser in `cedar.go` that uses an internal `scan` package for tokenization. The scanner handles lexical concerns (identifiers, strings, comments, whitespace). The parser handles grammar.

## Type System (Unresolved)

### `TypeExpr` interface (in `typexpr.go`)

```go
type TypeExpr interface {
    typeExpr() // sealed marker
}

type PrimitiveKind int
const (
    Long PrimitiveKind = iota
    String
    Bool
)

type PrimitiveTypeExpr struct{ Kind PrimitiveKind }
type SetTypeExpr struct{ Element TypeExpr }
type RecordTypeExpr struct{ Attributes map[string]*Attribute }
type EntityRefExpr struct{ Name string }         // explicit entity ref: {"type":"Entity","name":"User"}
type ExtensionTypeExpr struct{ Name string }     // ipaddr, decimal, datetime, duration
type NameExpr struct{ Name string; MustEntity bool } // bare name, resolved later
```

`Attribute`:
```go
type Attribute struct {
    Type        TypeExpr
    Required    bool
    Annotations Annotations
}
```

### Builder type constructors (in `builder.go`)

```go
func LongType() TypeExpr
func StringType() TypeExpr
func BoolType() TypeExpr
func SetOf(element TypeExpr) TypeExpr
func RecordOf(attrs map[string]*Attribute) TypeExpr
func EntityType(name string) TypeExpr
func ExtType(name string) TypeExpr
func IPAddrType() TypeExpr
func DecimalType() TypeExpr
func DatetimeType() TypeExpr
func DurationType() TypeExpr
func NamedType(name string) TypeExpr
```

## Schema Types (in `namespace.go`)

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
    MemberOfTypes []string
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
    Type string // e.g., "MyNamespace::Action", empty for same namespace
    ID   string
}

type AppliesTo struct {
    PrincipalTypes []string
    ResourceTypes  []string
    Context        TypeExpr // either a RecordTypeExpr or NameExpr
}

type CommonTypeDef struct {
    Type        TypeExpr
    Annotations Annotations
}
```

## Schema Entry Point (in `schema.go`)

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

## Resolution Algorithm (in `resolve.go`)

Three-phase resolution following the Rust implementation:

### Phase 1: Build Definition Index

Collect all fully-qualified names across all namespaces:
- `entityDefs` — set of all `NS::EntityName` (or just `EntityName` for empty namespace)
- `commonDefs` — set of all `NS::CommonName`
- `actionDefs` — set of all `NS::Action::"ActionName"`

### Phase 2: Validate

- **Reserved names**: Entity types and common types cannot use primitive/extension names (`Bool`, `Long`, `String`, `Record`, `Set`, `Entity`, `Extension`, `Boolean`)
- **RFC 70 shadowing**: If both `NS::Foo` and `Foo` (empty namespace) are defined as the same kind (entity or common), error
- **Common type cycles**: Topological sort of common type dependency graph (Kahn's algorithm). If any node has in-degree > 0 after processing, there's a cycle.

### Phase 3: Resolve Names

For each `NameExpr{Name: n, MustEntity: false}` in namespace `ns`:
1. Generate candidates: if `n` is unqualified and `ns` is non-empty → `[ns::n, n]`, else `[n]`
2. For each candidate in order:
   a. If candidate is defined as a common type → resolve to that common type
   b. If candidate is defined as an entity type → resolve to `EntityRef{EntityType: types.EntityType(candidate)}`
3. If no candidate matched → `UndefinedTypeError`

For `NameExpr{Name: n, MustEntity: true}`:
- Same candidate generation
- Only check entity type definitions (skip common types)

For `EntityRefExpr{Name: n}`:
- Explicit entity reference — resolve directly to `resolved.EntityRef`

For primitives/extensions:
- Map directly to `resolved.Primitive` / `resolved.Extension`

Common type inlining:
- Process in topological order
- Replace `NameExpr` references to common types with the common type's resolved definition

### Resolved Output Types (in `resolved/resolved.go`)

Keep the existing structure but ensure clean types:

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

Note: `CommonTypes` removed from resolved `Namespace` — common types are inlined during resolution. If the user wants to know what common types were defined, that information is in the unresolved `Schema`.

## Builder API (in `builder.go`)

Fluent builder with chaining:

```go
s := schema.NewBuilder().
    Namespace("MyApp").
        Entity("User").MemberOf("Group").
            Attr("name", schema.StringType()).
            OptionalAttr("age", schema.LongType()).
        Entity("Group").
        Action("view").
            Principal("User").Resource("Document").
            Context(schema.RecordOf(map[string]*schema.Attribute{
                "ip": {Type: schema.IPAddrType(), Required: true},
            })).
    Build()
```

Same structure as existing but with renamed type constructors (`StringType()` instead of `String()` to avoid collision with Go's `fmt.Stringer`).

## JSON Format (in `json.go`)

Cedar JSON schema format:
```json
{
    "NamespaceName": {
        "entityTypes": {
            "TypeName": {
                "memberOfTypes": ["ParentType"],
                "shape": { "type": "Record", "attributes": { ... } },
                "tags": { "type": "String" },
                "annotations": { "key": "value" }
            },
            "EnumTypeName": {
                "enum": ["A", "B", "C"]
            }
        },
        "actions": {
            "ActionName": {
                "memberOf": [{ "id": "GroupName", "type": "NS::Action" }],
                "appliesTo": {
                    "principalTypes": ["User"],
                    "resourceTypes": ["Doc"],
                    "context": { "type": "Record", "attributes": { ... } }
                }
            }
        },
        "commonTypes": {
            "TypeName": { "type": "Record", "attributes": { ... } }
        },
        "annotations": { "key": "value" }
    }
}
```

Type discriminated union in JSON:
- `{"type": "Long"}` → PrimitiveTypeExpr{Long}
- `{"type": "String"}` → PrimitiveTypeExpr{String}
- `{"type": "Bool"}` or `{"type": "Boolean"}` → PrimitiveTypeExpr{Bool}
- `{"type": "Set", "element": {...}}` → SetTypeExpr
- `{"type": "Record", "attributes": {...}}` → RecordTypeExpr
- `{"type": "Entity", "name": "Foo"}` → EntityRefExpr{Name: "Foo"}
- `{"type": "Extension", "name": "ipaddr"}` → ExtensionTypeExpr
- `{"type": "EntityOrCommon", "name": "Foo"}` → NameExpr{Name: "Foo"}
- `{"type": "Foo"}` (not a keyword) → NameExpr{Name: "Foo"}

## Cedar Text Format (in `cedar.go`)

Parser: hand-written recursive descent using internal scanner.

Scanner tokens: identifiers, strings, `::`, `{`, `}`, `[`, `]`, `<`, `>`, `,`, `;`, `:`, `?`, `=`, `@`, `(`, `)`, EOF.

Grammar rules follow the Cedar specification grammar exactly. Key productions:
- `Schema → {Namespace}`
- `Namespace → Annotations 'namespace' Path '{' {Decl} '}' | Decl`
- `Decl → Entity | Action | TypeDecl`
- `Entity → Annotations 'entity' Idents ['in' EntOrTyps] [['='] RecType] ['tags' Type] ';'`
- `Entity → Annotations 'entity' Idents 'enum' '[' STR+ ']' ';'`
- `Action → Annotations 'action' Names ['in' RefOrRefs] [AppliesTo] ';'`
- `TypeDecl → Annotations 'type' TYPENAME '=' Type ';'`

Emitter: structured output with sorted keys for deterministic output.

## Implementation Plan

### Phase 1: Foundation Types

Files: `typexpr.go`, `namespace.go`, `schema.go`, `errors.go`

1. Define `TypeExpr` interface and all variants
2. Define `Namespace`, `EntityTypeDef`, `EnumTypeDef`, `ActionDef`, `CommonTypeDef`, `AppliesTo`, `ActionRef`, `Attribute`, `Annotations`
3. Define `Schema` struct with `Namespaces`, `filename`, `SetFilename`, `Filename`
4. Define error sentinels and typed errors: `ErrCycle`, `ErrUndefinedType`, `ErrShadow`, `ErrDuplicate`, `ErrReservedName`, `ErrParse`
5. Define `isPrimitiveTypeName()` check

### Phase 2: Builder

File: `builder.go`

1. Type constructor functions: `LongType()`, `StringType()`, `BoolType()`, `SetOf()`, `RecordOf()`, `EntityType()`, `ExtType()`, `IPAddrType()`, `DecimalType()`, `DatetimeType()`, `DurationType()`, `NamedType()`
2. Fluent builder: `NewBuilder()` → `SchemaBuilder` → `NamespaceBuilder` → `EntityBuilder`, `ActionBuilder`
3. Builder methods mirror namespace/entity/action creation with chaining

### Phase 3: JSON Serialization

File: `json.go`

1. `Schema.UnmarshalJSON` — parse top-level namespace map
2. Namespace unmarshaling — entityTypes (including enum detection), actions, commonTypes, annotations
3. Type expression unmarshaling — discriminated union dispatch
4. Attribute unmarshaling — type + required + annotations
5. `Schema.MarshalJSON` — reverse of unmarshal
6. Custom marshalers for all types (namespace, entity, enum, action, common type, record, attribute)

### Phase 4: Cedar Text Parser

Files: `internal/scan/scanner.go`, `cedar.go`

1. Scanner: tokenizer producing tokens (Ident, String, Symbol, EOF) with position tracking for error messages
2. Parser: recursive descent following the Cedar grammar
3. `Schema.UnmarshalCedar` — entry point
4. Namespace, entity, action, type declaration parsing
5. Type expression parsing (paths, Set<T>, records)
6. Annotation parsing
7. `Schema.MarshalCedar` — Cedar text emitter with sorted output

### Phase 5: Resolution

File: `resolve.go`, `resolved/resolved.go`

1. Resolved types (keep/refactor `resolved/resolved.go`)
2. Definition index builder — collect all qualified names
3. Reserved name validation
4. RFC 70 shadowing validation
5. Common type cycle detection (topological sort)
6. Name resolution — candidate generation, priority-ordered lookup
7. Common type inlining
8. Action resolution — member-of to `types.EntityUID`, principal/resource to `types.EntityType`
9. `Schema.Resolve()` → `*resolved.Schema`

### Phase 6: Testing

Files: `schema_test.go`, `resolve_test.go`

1. Builder tests — construct schemas, verify structure
2. JSON round-trip tests — unmarshal → marshal → unmarshal → compare
3. Cedar text round-trip tests — unmarshal → marshal → unmarshal → compare
4. Cross-format tests — parse Cedar, marshal JSON, parse JSON, compare
5. Resolution tests — builder → resolve → check resolved types
6. Error tests — cycles, undefined types, shadowing, reserved names, parse errors
7. Rust test matrix (test_579 equivalent) — 114 test cases across 6 positions × scenarios
8. Corpus test compatibility — ensure `corpus_test.go` passes

## Files Summary

| File | Action | Description |
|------|--------|-------------|
| `x/exp/schema/schema.go` | Create | Schema type, public API |
| `x/exp/schema/namespace.go` | Create | Namespace and definition types |
| `x/exp/schema/typexpr.go` | Create | TypeExpr interface and variants |
| `x/exp/schema/builder.go` | Create | Builder API and type constructors |
| `x/exp/schema/json.go` | Create | JSON marshal/unmarshal |
| `x/exp/schema/cedar.go` | Create | Cedar text marshal/unmarshal |
| `x/exp/schema/resolve.go` | Create | Resolution algorithm |
| `x/exp/schema/errors.go` | Create | Error types and sentinels |
| `x/exp/schema/internal/scan/scanner.go` | Create | Cedar text tokenizer |
| `x/exp/schema/resolved/resolved.go` | Create | Resolved schema types |
| `x/exp/schema/schema_test.go` | Create | Comprehensive tests |
| `x/exp/schema/resolve_test.go` | Create | Resolution + test_579 matrix |
| `x/exp/schema/types.go` | Delete | Replaced by typexpr.go + namespace.go + schema.go |
| `x/exp/schema/parse_cedar.go` | Delete | Replaced by cedar.go |
| `x/exp/schema/internal/parse/` | Delete | Replaced by internal/scan/ |
| `corpus_test.go` | Modify | Update import if API changes |

## Definition of Done

1. `go test ./x/exp/schema/...` passes
2. `go test ./...` passes (including corpus tests)
3. `go vet ./...` clean
4. JSON round-trip: parse → marshal → parse → compare = identical for all test schemas
5. Cedar text round-trip: parse → marshal → parse → compare = identical for all test schemas
6. Resolution matches Rust behavior for test_579 matrix (114 cases)
7. All corpus .cedarschema files parse without error
8. Cycle detection catches all cyclic common type definitions
9. Shadowing validation catches all RFC 70 violations
10. Reserved name validation catches all uses of primitive/extension names

## Risks & Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Cedar text parser doesn't handle all edge cases | Medium | High | Test against full corpus (6,084+ files) |
| Resolution algorithm misses edge cases | Medium | High | Port test_579 matrix from Rust |
| JSON format has undocumented variants | Low | Medium | Cross-reference Rust implementation |
| `corpus_test.go` API coupling | Low | High | Maintain `Schema`, `SetFilename`, `UnmarshalCedar` API surface |
| Scope creep into validation | Low | Medium | Strict non-goal: no policy validation this sprint |

## Open Questions

1. Should `CommonTypes` remain in the resolved namespace (as a "what was defined" reference) or be fully removed since common types are inlined?
2. The existing `resolved.Namespace.CommonTypes` uses `map[types.Path]*Type` — is this useful post-resolution or just noise?
3. Should the scanner be its own `internal/scan` package or just unexported types in the `schema` package?
4. Should `NameExpr.MustEntity` be a separate type (`EntityNameExpr`) instead of a bool flag?
