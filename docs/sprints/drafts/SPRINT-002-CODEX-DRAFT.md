# Sprint 002: Cedar Schema Rewrite (Codex Draft)

## Overview

Sprint 002 replaces `x/exp/schema` end-to-end with a new implementation that is easier to reason about, more Go-idiomatic, and behaviorally aligned with Cedar spec and Rust reference behavior.

This sprint must deliver four outcomes:
1. Parse and emit Cedar schema in both Cedar text and JSON formats.
2. Provide ergonomic programmatic construction APIs.
3. Resolve raw schema names into fully-qualified, concrete types.
4. Validate correctness with high confidence (unit, round-trip, Rust parity, corpus).

The rewrite explicitly does **not** preserve internal design from Sprint 001. It preserves only external compatibility points required by the repository and tests.

## Scope

### In Scope
- Full replacement of implementation under `x/exp/schema/`.
- Public API compatibility for current call sites:
  - `schema.Schema`
  - `(*Schema).SetFilename(string)`
  - `(*Schema).UnmarshalCedar([]byte) error`
- Cedar text parser/emitter.
- JSON parser/emitter.
- Programmatic builder.
- Resolution engine with Cedar-consistent prioritization and validation behavior.
- Error model with sentinels + typed wrappers that support `errors.Is`/`errors.As`.
- Test matrix porting from Rust resolution coverage and corpus integration.

### Out of Scope
- Policy validation engine.
- CLI/LSP/editor integrations.
- Source location UX improvements beyond practical parser diagnostics.
- Multi-file merge semantics beyond current single schema load path.

## Project Constraints

- Package remains at `x/exp/schema`.
- Resolved output uses `types.EntityType`, `types.EntityUID`, and `types.Path`.
- Go 1.23 idioms expected (`iter.Seq2`, `maps`, `slices`, sentinel + typed errors).
- Support all schema features used in corpus and Cedar docs:
  - Namespaces (including empty namespace)
  - Entity types and enum entities
  - Actions and action groups
  - Common types
  - Attributes, tags, annotations
  - Extension types (`ipaddr`, `decimal`, `datetime`, `duration`)
  - `__cedar::` disambiguation behavior

## Architecture

```
x/exp/schema/
├── schema.go                 # Public entrypoint and API methods
├── types.go                  # Unresolved schema model (or split files if cleaner)
├── builder.go                # Programmatic construction API
├── json.go                   # JSON marshal/unmarshal
├── cedar.go                  # Cedar text marshal/unmarshal
├── resolve.go                # Resolution and validation pipeline
├── errors.go                 # Sentinels + typed error wrappers
├── internal/
│   └── parse/
│       └── parser.go         # Hand-written scanner/parser for Cedar schema text
└── resolved/
    └── resolved.go           # Resolved output model
```

### Data Model Direction

- Keep unresolved and resolved representations distinct.
- Use explicit sum-type style for schema types; avoid `any`.
- Separate name-reference categories to eliminate positional ambiguity:
  - `EntityNameExpr` for entity-only contexts
  - `TypeNameExpr` for type-position references
  - `EntityRefExpr` for explicit JSON entity refs
- Inline common types into resolved output during resolution.

## Resolution Specification

Resolution follows a strict staged pipeline.

### Stage 1: Index Definitions
- Build definition maps for entities, actions, common types across all namespaces.
- Include empty namespace as a first-class namespace.
- Seed built-in primitive/extension handling (`__cedar::` + unqualified behavior per Cedar conventions).

### Stage 2: Structural Validation
- Duplicate definitions by namespace/kind.
- Reserved-name validation (including `__cedar` reserved prefix rules).
- RFC 70 shadowing checks against empty-namespace definitions.
- Common type cycle detection (topological-sort or equivalent graph walk).

### Stage 3: Name Resolution
For unqualified `name` in namespace `ns`, candidate order is:
1. `ns::name`
2. `name`

Type-position lookup priority is:
1. Common type
2. Entity type
3. Primitive/extension built-ins

Entity-only positions resolve only against entity definitions.

### Stage 4: Output Construction
- Convert resolved entity names to `types.EntityType`.
- Convert resolved actions to `types.EntityUID`.
- Produce `resolved.Schema` with deterministic map population semantics where testability benefits.

## Implementation Plan

### Phase 1: Core Model + Errors (15%)
Files:
- `x/exp/schema/types.go`
- `x/exp/schema/errors.go`
- `x/exp/schema/resolved/resolved.go`

Tasks:
- Define unresolved and resolved structures.
- Define typed references and type expressions.
- Define sentinel errors and typed wrappers with `Unwrap()`.

### Phase 2: Builder API (15%)
File:
- `x/exp/schema/builder.go`

Tasks:
- Fluent builder for namespaces/entity types/actions/common types.
- Typed constructors for primitives, extension types, set/record/named references.
- Preserve ergonomics for expected call-site readability in Go.

### Phase 3: JSON I/O (15%)
File:
- `x/exp/schema/json.go`

Tasks:
- Implement discriminator-based parsing/marshal for type variants.
- Support canonical Cedar schema JSON layout.
- Enforce duplicate/reserved-name validations during load path where practical.

### Phase 4: Cedar Text I/O (20%)
Files:
- `x/exp/schema/cedar.go`
- `x/exp/schema/internal/parse/parser.go`

Tasks:
- Hand-written parser and emitter for schema Cedar format.
- Support annotations, tags, enum entity syntax, appliesTo forms, and namespace forms.
- Deterministic emission for stable round-trip tests.

### Phase 5: Resolver (20%)
File:
- `x/exp/schema/resolve.go`

Tasks:
- Index build, validation stages, name resolution, output projection.
- Full action/memberOf/appliesTo resolution semantics.
- Common type expansion/inlining into resolved forms.

### Phase 6: Test Matrix + Integration (15%)
Files:
- `x/exp/schema/schema_test.go`
- `x/exp/schema/resolve_test.go`
- `x/exp/schema/internal/parse/parser_test.go`
- `x/exp/schema/resolved/resolved_internal_test.go`
- `corpus_test.go` (compatibility confirmation only)

Tasks:
- Port relevant Rust resolution test scenarios.
- Add JSON and Cedar round-trip tests.
- Add builder parity tests.
- Run corpus parse coverage via existing test harness.

## Test Strategy

1. Unit tests
- Type variant parsing and construction.
- Resolution edge cases: ambiguity, undefined references, cycles, shadowing.
- Reserved names and namespace edge cases.

2. Round-trip tests
- JSON: parse -> marshal -> parse semantic identity.
- Cedar: parse -> marshal -> parse semantic identity.
- Cross-format: Cedar -> model -> JSON -> model and inverse.

3. Compatibility/integration tests
- Existing `corpus_test.go` remains green with public API expectations.
- `go test ./x/exp/schema/...` and `go test ./...`.

4. Differential/parity tests
- Recreate relevant Rust `test_579.rs` scenarios for resolution behavior.

## Files Summary

| File | Action | Purpose |
|------|--------|---------|
| `x/exp/schema/types.go` | Rewrite | Core unresolved schema model |
| `x/exp/schema/errors.go` | Rewrite | Error sentinels and typed wrappers |
| `x/exp/schema/builder.go` | Rewrite | Fluent API and type constructors |
| `x/exp/schema/json.go` | Rewrite | JSON marshal/unmarshal |
| `x/exp/schema/cedar.go` | Add/Rewrite | Cedar text marshal/unmarshal |
| `x/exp/schema/resolve.go` | Rewrite | Resolution pipeline |
| `x/exp/schema/internal/parse/parser.go` | Rewrite | Cedar schema parser implementation |
| `x/exp/schema/resolved/resolved.go` | Rewrite | Resolved output types |
| `x/exp/schema/*_test.go` | Rewrite/Expand | Coverage of parser/resolver/builder/I-O |

## Definition of Done

1. `go test ./x/exp/schema/...` passes.
2. `go test ./...` passes, including corpus.
3. JSON and Cedar round-trip tests pass with semantic identity assertions.
4. Resolution behavior matches intended Rust scenarios for priority/cycle/shadowing cases.
5. `corpus_test.go` schema parse path remains compatible.
6. Code quality contract is met:
   - No test-only production abstractions
   - Comments explain intent, not mechanics
   - No defensive nil checks against internal invariants
   - Strong concrete typing (no broad `any` use)

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Cedar grammar edge cases missed | Medium | High | Validate against corpus and targeted parser tests |
| Resolution mismatch vs Rust | Medium | High | Port matrix cases and diff expected outcomes |
| Shadowing/namespace semantics regressions | Medium | High | Focused tests for empty namespace and RFC 70 behavior |
| API break for corpus harness | Low | High | Keep `Schema`, `SetFilename`, `UnmarshalCedar` stable |
| Rewriter overfits tests | Low | Medium | Ensure tests use only public API and corpus fixtures |

## Dependencies

- Cedar documentation workspace: `/Volumes/src/ai/cedar-docs`
- Rust reference implementation: `/Volumes/src/ai/cedar`
- Existing repository corpus archive and harness: `corpus-tests.tar.gz`, `corpus_test.go`
- Existing core types package: `types/`

## Open Questions

1. Should parser internals remain in `internal/parse` or be split into scanner/parser files for maintainability?
2. Do we require strict text-format preservation for Cedar output, or only semantic normalization?
3. How strict should error-message wording parity be with Rust (exact text vs stable categories)?
