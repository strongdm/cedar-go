# Sprint 002 Intent: Cedar Schema Rewrite

## Seed

Replace the existing implementation of schema parsing in this repo. Throw whatever is there away and take no inspiration from it. What we want is the ability to do the following things:

1. Parse and emit Cedar schema in the Cedar text and JSON formats into an in-memory representation
2. Build the in-memory representation in a Go-idiomatic and ergonomic way
3. Resolve the raw schema into a different type which deals in fully-qualified EntityTypes, taking into account all the intricacies of the resolution algorithm
4. Validate with high confidence including incorporating the schemas in the corpus tests and replicating relevant tests from the Rust implementation for schema resolution

The Cedar documentation is available in /Volumes/src/ai/cedar-docs and the canonical Rust implementation is available in /Volumes/src/ai/cedar.

## Context

- cedar-go is a Go Cedar implementation (`github.com/cedar-policy/cedar-go`, Go 1.23.0)
- Sprint 001 shipped a working schema implementation — this sprint replaces it entirely
- `corpus_test.go` imports `x/exp/schema` and tests against 6,084+ corpus .cedarschema files
- The existing code works and all tests pass, but the user considers the implementation unsatisfactory
- The `types` package provides `EntityType`, `EntityUID`, `Path` — the resolved schema must produce these
- The `resolved` sub-package provides the output types for resolution

## Code Quality Contract

All shipped code must satisfy:
1. No code exists solely for tests — tests use the public API
2. Comments explain intent only, never restate what code does
3. No defensive coding — trust types and internal invariants
4. Strong typing — concrete types, generics, enums over `any`/`interface{}`
5. Idiomatic to an expert in Go — follows conventions observed in the codebase:
   - `MarshalJSON`/`UnmarshalJSON`, `MarshalCedar`/`UnmarshalCedar` method pairs
   - `iter.Seq2` for iteration, `maps.Clone`, `slices.Sort`
   - Sentinel errors + typed error structs with `Unwrap()`
   - Constructor functions delegating to internal packages
   - Generics with constraints where appropriate

## Recent Sprint Context

- **Sprint 001**: Cedar Schema Support — implemented parsing (JSON + Cedar text), programmatic builder, resolution with cycle detection and shadowing validation. This is the implementation being replaced.

## Relevant Codebase Areas

### Types Package (`types/`)
- `EntityType` (alias for `Path`) — fully-qualified entity type name
- `EntityUID` — entity identifier combining type and ID
- `Path` — series of idents separated by `::`

### Current Schema Package (`x/exp/schema/`)
- `types.go` — Type interface with marker method, PrimitiveType, SetType, RecordType, EntityRef, ExtensionType, CommonTypeRef, EntityOrCommonRef
- `builder.go` — Fluent builder: SchemaBuilder → NamespaceBuilder → EntityBuilder/ActionBuilder
- `resolve.go` — `Schema.Resolve()` with resolver struct, cycle detection, shadowing, disambiguation
- `json.go` — JSON marshal/unmarshal with discriminated union types
- `parse_cedar.go` — Cedar text marshal/unmarshal via internal parser
- `errors.go` — Sentinel errors + typed structs (CycleError, UndefinedTypeError, etc.)
- `resolved/resolved.go` — Output types for resolution
- `internal/parse/parser.go` — Hand-written recursive descent parser

### Reference Implementation (Rust)
- Three-phase resolution: `RawName → ConditionalName → InternalName`
- `ConditionalName` holds ordered list of possibilities (e.g., `[NS::Name, Name]`)
- RFC 24: Within each possibility, common type checked before entity type
- RFC 70: Named namespaces cannot shadow empty namespace definitions
- `ReferenceType` enum: `Common`, `Entity`, `CommonOrEntity`
- Cycle detection via topological sort (Kahn's algorithm)
- `AllDefs` struct tracks entity_defs, common_defs, action_defs
- Comprehensive test matrix in `test_579.rs` covering all positions × scenarios

### Cedar Schema Format
- JSON: `{"namespace": {"entityTypes": {...}, "actions": {...}, "commonTypes": {...}, "annotations": {...}}}`
- Cedar text grammar with namespaces, entities (including enum), actions, common types, annotations, tags
- Extension types: ipaddr, decimal, datetime, duration
- `__cedar::` namespace for built-in type disambiguation
- Annotations on namespaces, entities, actions, common types, attributes

## Constraints

- Must live in `x/exp/schema/` — same import path as existing code
- Must support `corpus_test.go` which calls `schema.UnmarshalCedar`
- Must produce resolved output using `types.EntityType` and `types.EntityUID`
- Must handle all Cedar schema features: namespaces, entities, enum entities, actions, common types, tags, annotations, extension types
- Must implement correct resolution priority (common > entity > primitive/extension)
- Must detect cycles in common type definitions
- Must enforce RFC 70 shadowing rules
- Error handling must support `errors.Is()` via sentinel errors and `Unwrap()`
- Go 1.23 — can use `iter.Seq2`, generics, `maps`, `slices` packages

## Success Criteria

1. All corpus tests pass (6,084+ schemas parse correctly)
2. All existing `schema_test.go` and `resolve_test.go` test scenarios are covered (may be rewritten)
3. Resolution algorithm matches Rust behavior for all `test_579.rs` scenarios
4. JSON round-trip: parse → marshal → parse produces identical schemas
5. Cedar text round-trip: parse → marshal → parse produces identical schemas
6. Builder produces same schemas as parsing equivalent text/JSON
7. An expert Go developer reads the code and sees nothing to wince at

## Verification Strategy

- **Reference implementation**: Rust at `/Volumes/src/ai/cedar` — replicate resolution test matrix
- **Documentation**: Cedar docs at `/Volumes/src/ai/cedar-docs` — conformance to spec
- **Corpus**: 6,084+ .cedarschema files in corpus-tests.tar.gz — parser coverage
- **Round-trip**: JSON ↔ Cedar text ↔ in-memory round-trips preserve semantics
- **Differential**: Where possible, compare Go output against Rust output for same inputs
- **Edge cases**: cycles, shadowing, reserved names, cross-namespace refs, enum entities, `__cedar::` disambiguation, empty namespaces

## Open Questions

1. Should the `resolved` sub-package be kept as-is, redesigned, or folded into the main schema package?
2. Should the internal parser be a separate `internal/parse` sub-package or inline in the schema package?
3. What should the builder API look like — keep the fluent chain style or try a different approach?
4. Should we support multi-fragment schemas (multiple schema files merged) or just single schemas?
5. How closely should error messages match the Rust implementation's wording?
