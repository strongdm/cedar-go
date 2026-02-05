# Sprint 002 Code Critique (Round 2)

## Context
- Reviewed intent in `docs/sprints/SPRINT-002.md`.
- `CLAUDE.md` was requested but is not present in this workspace (`find .. -name CLAUDE.md` returned no matches).
- Scope is the Sprint 002 changed/created files reported by git under `x/exp/schema`.

## Findings (Ordered by Severity)

### 1) `CONSEQUENTIAL` - Tests bypass the public API contract
- File: `x/exp/schema/schema_test.go:2732`, `x/exp/schema/schema_test.go:2794`, `x/exp/schema/schema_test.go:2819`, `x/exp/schema/schema_test.go:2844`
- Problem: Many tests construct internal state directly (`&Schema{...}`, `&Namespace{...}`, direct map wiring, use of `newAnnotations()`), rather than using public constructors/parsers/builders.
- Why this violates the contract: The sprint quality gate says tests must use the same public API as production callers.

### 2) `CONSEQUENTIAL` - JSON marshaling path relies on untyped `any` maps where concrete shapes are feasible
- File: `x/exp/schema/json.go:291`, `x/exp/schema/json.go:315`, `x/exp/schema/json.go:340`, `x/exp/schema/json.go:395`, `x/exp/schema/json.go:414`
- Problem: Marshaling is built via `map[string]any` and `any` return values (`marshalTypeExprValue`), rather than typed encoding structs.
- Why this violates the contract: Explicitly conflicts with the strong-typing criterion (`any/interface{}` where concrete types can work).

### 3) `CONSEQUENTIAL` - Shallow cloning introduces aliasing across parsed multi-entity declarations
- File: `x/exp/schema/cedar.go:926`, `x/exp/schema/cedar.go:930`
- Problem: `cloneRecordType` copies `Attribute` structs but assigns `Type: v.Type` directly. Nested record/set type expressions are shared between supposedly independent cloned declarations.
- Impact: Mutating one parsed entity/type can mutate another unexpectedly.
- Why this violates the contract: Non-idiomatic and behaviorally risky for a library API that exposes mutable structs.

### 4) `CONSEQUENTIAL` - Type-switch defaults can panic on `nil` `TypeExpr`
- File: `x/exp/schema/json.go:409`, `x/exp/schema/json.go:428`, `x/exp/schema/resolve.go:424`
- Problem: Default branches do unchecked assertions (`t.(EntityNameExpr)`, `expr.(TypeNameExpr)`). Public builder methods accept `TypeExpr`, so callers can pass `nil` and trigger panic during marshal/resolve.
- Why this violates the contract: Non-idiomatic error handling and brittle API behavior.

### 5) `INCONSEQUENTIAL` - Redundant comments that restate adjacent code
- File: `x/exp/schema/cedar.go:387`, `x/exp/schema/cedar.go:472`, `x/exp/schema/cedar.go:523`, `x/exp/schema/internal/scan/scanner.go:188`, `x/exp/schema/internal/scan/scanner.go:194`, `x/exp/schema/internal/scan/scanner.go:247`
- Problem: Comments like `// skip "entity"`, `// skip opening "` restate exactly what the next line does.

### 6) `INCONSEQUENTIAL` - Coverage-targeted comment block is overly coupled to line numbers
- File: `x/exp/schema/schema_test.go:2868`
- Problem: Large block comments enumerate parser line numbers and micro-paths; this is high-churn and not durable documentation.

## Per-File Evaluation

- `x/exp/schema/builder.go`
  - No findings against criteria 1-5.
- `x/exp/schema/cedar.go`
  - Findings: #3 (`CONSEQUENTIAL`), #5 (`INCONSEQUENTIAL`).
- `x/exp/schema/errors.go`
  - No findings against criteria 1-5.
- `x/exp/schema/internal/scan/scanner.go`
  - Findings: #5 (`INCONSEQUENTIAL`).
- `x/exp/schema/internal/scan/scanner_test.go`
  - No findings against criteria 1-5.
- `x/exp/schema/json.go`
  - Findings: #2 (`CONSEQUENTIAL`), #4 (`CONSEQUENTIAL`).
- `x/exp/schema/namespace.go`
  - No findings against criteria 1-5.
- `x/exp/schema/resolve.go`
  - Findings: #4 (`CONSEQUENTIAL`).
- `x/exp/schema/resolved/resolved.go`
  - No findings against criteria 1-5.
- `x/exp/schema/resolved/resolved_test.go`
  - No findings against criteria 1-5.
- `x/exp/schema/schema.go`
  - No findings against criteria 1-5.
- `x/exp/schema/schema_test.go`
  - Findings: #1 (`CONSEQUENTIAL`), #6 (`INCONSEQUENTIAL`).
- `x/exp/schema/typexpr.go`
  - No findings against criteria 1-5.

### Deleted Sprint Files (Reviewed as removed)
- `x/exp/schema/coverage_internal_test.go` - deleted, no active code to evaluate.
- `x/exp/schema/internal/parse/parser.go` - deleted, no active code to evaluate.
- `x/exp/schema/internal/parse/parser_test.go` - deleted, no active code to evaluate.
- `x/exp/schema/parse_cedar.go` - deleted, no active code to evaluate.
- `x/exp/schema/resolve_test.go` - deleted, no active code to evaluate.
- `x/exp/schema/resolved/resolved_internal_test.go` - deleted, no active code to evaluate.
- `x/exp/schema/types.go` - deleted, no active code to evaluate.
- `x/exp/schema/types_internal_test.go` - deleted, no active code to evaluate.
