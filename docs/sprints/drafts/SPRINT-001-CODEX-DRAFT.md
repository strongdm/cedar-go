# Sprint 001: Cedar Schema Support (Codex Draft)

## Goals
- Implement Cedar schema support in `x/exp/schema`:
  - JSON and Cedar text parsing/serialization
  - Programmatic construction APIs
  - Resolution to fully-qualified names (`types.Path`, `types.EntityType`, `types.EntityUID`)
- Achieve round-trip equivalence for JSON and Cedar formats
- Unblock corpus tests and align behavior with Rust reference

## Scope
- New experimental package `x/exp/schema` with:
  - Data model for schema fragments, namespaces, entities, actions, common types, attributes, types
  - JSON and Cedar text serializers/deserializers
  - Resolver implementing namespace qualification and disambiguation rules
  - Programmatic builder API
- Test coverage: unit, integration, corpus (when available)
- Adjust existing `corpus_test.go` to target implemented package

## Non-Goals
- Full policy validator (out of scope)
- LSP/editor integration (optional hooks only)
- Runtime authorization changes beyond schema support

## Success Criteria
- Parse all 6,084 corpus `.cedarschema` files without error
- JSON round-trip: Schema → JSON → Schema equivalent
- Cedar round-trip: Schema → Cedar → Schema equivalent
- Builder API constructs schemas equivalent to parsed forms
- Resolution:
  - Fully-qualifies names per namespace rules
  - Resolves entity/common type refs to `types.Path`/`types.EntityType`
  - Resolves actions to `types.EntityUID` where applicable
- Error behavior conforms to Rust implementation (messages and conditions)

## Architecture
- Public API: `x/exp/schema`
  - `Schema` (entry) wraps a fragment of namespaces
  - `UnmarshalJSON`, `MarshalJSON`
  - `UnmarshalCedar`, `MarshalCedar`
  - `Resolve() (*ResolvedSchema, error)`
  - Optional `WithFilename(string)` for diagnostics
- Internal layers:
  - Types: Fragment, NamespaceDefinition, EntityTypeDef, ActionDef, CommonTypeDef, Attribute, Type variants (Primitive, Record, Set, EntityRef, Extension, CommonRef)
  - Names: RawName, ConditionalName, InternalName (`types.Path`)
  - JSON parser: discriminated unions by "type", strict duplicate checks
  - Cedar parser: extends tokenizer with schema grammar
  - Resolver: definition index, disambiguation (common > entity > primitive/extension), cycle detection, shadowing validation

## Files
- x/exp/schema/schema.go — Public `Schema`, `ResolvedSchema`, loaders/savers
- x/exp/schema/types.go — Core schema data model
- x/exp/schema/name.go — Raw/Conditional/Internal name types
- x/exp/schema/json.go — JSON marshal/unmarshal
- x/exp/schema/tokenize.go — Schema tokens (if not reusing internal tokenizer)
- x/exp/schema/parse_cedar.go — Cedar text parser
- x/exp/schema/resolve.go — Resolution engine
- x/exp/schema/alldefs.go — Index of known defs (primitives, extensions, entities, common)
- x/exp/schema/errors.go — Typed errors (CycleError, UndefinedTypeError, ShadowError)
- x/exp/schema/*.test.go — Unit + integration tests
- corpus_test.go — Update/remove placeholder import to use new package

## Implementation Plan
- Phase 1: Core Types (15%)
  - Define schema structs and type variants
  - Name types and utilities (join, qualify, normalize)
  - Annotations map with light helpers
- Phase 2: JSON Parsing (20%)
  - Fragment unmarshaler with namespace handling (empty namespace allowed)
  - Discriminated union parsing for types
  - Validation for duplicates and reserved names
  - Marshaler to preserve semantics for round-trip
- Phase 3: Cedar Text Parsing (25%)
  - Tokenizer: namespace, entity, action, type, in, appliesTo, enum, tags, Set
  - Grammar productions for namespaces, entity defs (memberships, shape, tags, enum), action defs (memberOf, appliesTo), common types
  - Type expression parsing (primitives, sets, records, entity/common refs, extensions)
  - Optional annotations and multiple declarations
- Phase 4: Resolution (25%)
  - Build AllDefs from fragment, seed with primitives and extensions (`__cedar::Long`, `String`, `Bool`, extension set)
  - `RawName` → `ConditionalName` (consider namespace) → resolve to `types.Path`
  - Disambiguation: common > entity > primitive/extension
  - Cycle detection in common refs; shadowing checks
  - Resolve EntityOrCommon references concretely
- Phase 5: Public API & Integration (10%)
  - `Schema` wrapper methods (JSON/Cedar load/save, Resolve)
  - `ResolvedSchema` view typed on `types.Path`/`types.EntityType`
  - Adjust `corpus_test.go` import
- Phase 6: Testing & Docs (5%)
  - Unit tests for each layer
  - Integration tests for mixed features
  - Round-trip tests
  - Short developer doc in `x/exp/schema/README.md` (optional)

## Testing Strategy
- Unit:
  - JSON: variant parsing, duplicates, reserved names
  - Cedar: grammar productions and edge cases
  - Resolver: disambiguation, cycles, shadowing, forward refs
- Integration:
  - Programmatic builder parity with parsed schema
  - Round-trip JSON and Cedar
- Corpus:
  - Parse and resolve all corpuses; sync results with Rust CLI outcomes when available

## Risks and Mitigations
- Resolution edge cases (Med/High): build exhaustive tests from docs and mirrored Rust outcomes
- Grammar drift (Low/Med): adhere to current spec; isolate tokenizer extensions
- Error parity (Med/Med): codify messages in tests based on Rust behavior
- Performance (Low/Low): profile against corpus; optimize hot paths
- API creep (Low/Med): keep experimental scope in `x/exp`

## Estimates
- Phase 1: 0.5–1 day
- Phase 2: 1–1.5 days
- Phase 3: 2–3 days
- Phase 4: 2–3 days
- Phase 5: 0.5–1 day
- Phase 6: 0.5 day
- Total: ~6–9 days, depending on parser complexity and edge cases

## Dependencies
- Existing `types` package (`Path`, `EntityType`, `EntityUID`)
- `internal/parser` tokenizer patterns (reuse where practical)
- No new third-party libraries

## Out of Scope
- Policy validator; partial evaluation; policy templates
- Full LSP feature set; advanced diagnostics UX

## Open Questions
- Resolved representation: separate `ResolvedSchema` vs single type with flag; leaning separate for clarity
- Error taxonomy: typed errors for matchable cases vs fmt errors
- Multi-fragment merging strategy and conflict policy
- How strict should round-trip formatting be for Cedar (semantics vs exact text)

## Milestones
- M1 (Core Types + JSON load/save): JSON round-trip passes basic cases
- M2 (Cedar parser): Cedar round-trip parity for controlled samples
- M3 (Resolver): Fully-qualified names and disambiguation pass unit tests
- M4 (Integration): Programmatic builder parity + integration tests
- M5 (Corpus): All corpus schemas parse; error matches spec

## Notes
- Ensure `corpus_test.go` stops importing non-existent `x/exp/schema` until package lands, then re-enable.
- Document extension types supported; keep list alongside resolver seeding.
