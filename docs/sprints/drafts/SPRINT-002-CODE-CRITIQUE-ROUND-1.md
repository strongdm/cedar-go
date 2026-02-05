# SPRINT-002 Code Critique - Round 1

## Scope
- Sprint intent source: `docs/sprints/SPRINT-002.md`
- Conventions source requested: `CLAUDE.md` (not found in repository; review performed against repo idioms + sprint quality gates)
- Files reviewed (current Sprint 002 surface):
  - `x/exp/schema/schema.go`
  - `x/exp/schema/namespace.go`
  - `x/exp/schema/typexpr.go`
  - `x/exp/schema/builder.go`
  - `x/exp/schema/json.go`
  - `x/exp/schema/cedar.go`
  - `x/exp/schema/resolve.go`
  - `x/exp/schema/errors.go`
  - `x/exp/schema/schema_test.go`
  - `x/exp/schema/internal/scan/scanner.go`
  - `x/exp/schema/internal/scan/scanner_test.go`
  - `x/exp/schema/resolved/resolved.go`
  - `x/exp/schema/resolved/resolved_test.go`

## Findings (Ordered by Severity)

1. **CONSEQUENTIAL** - Tests rely on internal/non-public API and coverage-only internals, violating the "tests use public API" contract.
- Evidence:
  - `x/exp/schema/schema_test.go:1802` (`TestTypeExprMarkers`) directly calls unexported marker methods.
  - `x/exp/schema/schema_test.go:2367` (`TestWalkTypeExprNil`) calls unexported `walkTypeExpr`.
  - `x/exp/schema/schema_test.go:2372` through `x/exp/schema/schema_test.go:2394` call unexported resolution helpers (`resolutionCandidates`, `qualifyName`).
  - `x/exp/schema/resolved/resolved_test.go:24` (`TestResolvedTypeMarkers`) directly calls unexported `resolvedType()` methods.
- Why this matters: these tests are coupled to implementation details rather than caller-observable behavior, which weakens refactor safety and violates Sprint 002’s stated quality gate.

2. **CONSEQUENTIAL** - `ErrDuplicate` / `DuplicateError` appear test-only in current implementation.
- Evidence:
  - Defined in `x/exp/schema/errors.go:13` and `x/exp/schema/errors.go:59`.
  - No non-test production call sites construct or return `DuplicateError`.
  - Only referenced by tests in `x/exp/schema/schema_test.go`.
- Why this matters: this is production surface that currently exists only to satisfy tests/coverage, violating "NO TEST-ONLY CODE".

3. **CONSEQUENTIAL** - JSON layer uses pervasive `map[string]any` and untyped JSON assembly where concrete typed structs are viable.
- Evidence:
  - `x/exp/schema/json.go:291`, `x/exp/schema/json.go:293`, `x/exp/schema/json.go:315`, `x/exp/schema/json.go:340`, `x/exp/schema/json.go:380`, `x/exp/schema/json.go:395`, `x/exp/schema/json.go:414`.
- Why this matters: this violates the sprint’s strong-typing gate (`any` where concrete structures would work). It also makes schema evolution less safe (field drift becomes runtime behavior instead of compile-time breakage).

4. **CONSEQUENTIAL** - Multi-name Cedar parse cloning is shallow for some TypeExpr paths, allowing aliasing between declarations.
- Evidence:
  - In multi-name entity declarations, `tags` is reused by pointer/reference: `x/exp/schema/cedar.go:468` to `x/exp/schema/cedar.go:470`.
  - `cloneAppliesTo` does not deep-clone `Context`: `x/exp/schema/cedar.go:954` to `x/exp/schema/cedar.go:959`.
- Why this matters: mutating one declaration’s parsed AST can implicitly mutate siblings parsed from the same multi-name declaration. That is a hidden side effect and non-idiomatic for AST construction.

5. **INCONSEQUENTIAL** - Multiple comments restate immediate code behavior instead of rationale.
- Evidence examples:
  - `x/exp/schema/builder.go:3`, `x/exp/schema/builder.go:25`, `x/exp/schema/builder.go:45`, `x/exp/schema/builder.go:91`, `x/exp/schema/builder.go:152`.
  - `x/exp/schema/cedar.go:343`, `x/exp/schema/cedar.go:392`, `x/exp/schema/cedar.go:398`, `x/exp/schema/cedar.go:419`, `x/exp/schema/cedar.go:432`, `x/exp/schema/cedar.go:481`, `x/exp/schema/cedar.go:923`.
  - `x/exp/schema/schema_test.go` has many `Coverage:` comments that describe what the next lines do rather than why.
- Why this matters: readability noise; does not currently introduce behavioral risk.

6. **INCONSEQUENTIAL** - Non-idiomatic reimplementation of substring search in tests with unnecessary defensive conditions.
- Evidence:
  - `x/exp/schema/schema_test.go:1814` to `x/exp/schema/schema_test.go:1825` (`contains` / `searchContains`).
- Why this matters: `strings.Contains` is the idiomatic standard library approach. Current helper adds avoidable complexity and redundant checks.

## Per-File Status
- `x/exp/schema/schema.go`: No consequential findings.
- `x/exp/schema/namespace.go`: No consequential findings.
- `x/exp/schema/typexpr.go`: No consequential findings.
- `x/exp/schema/builder.go`: Inconsequential comment-style findings only.
- `x/exp/schema/json.go`: Consequential strong-typing finding.
- `x/exp/schema/cedar.go`: Consequential shallow-clone/aliasing finding; inconsequential comment-style findings.
- `x/exp/schema/resolve.go`: No consequential findings.
- `x/exp/schema/errors.go`: Consequential test-only API surface finding.
- `x/exp/schema/schema_test.go`: Consequential test-internal-API coupling; inconsequential helper/comment style findings.
- `x/exp/schema/internal/scan/scanner.go`: No consequential findings.
- `x/exp/schema/internal/scan/scanner_test.go`: No consequential findings.
- `x/exp/schema/resolved/resolved.go`: No consequential findings.
- `x/exp/schema/resolved/resolved_test.go`: Consequential test-internal-API coupling.
