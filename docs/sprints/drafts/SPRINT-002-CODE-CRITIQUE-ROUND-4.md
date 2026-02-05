# Sprint 002 Code Critique (Round 4)

## Context
- Reviewed intent in `docs/sprints/SPRINT-002.md`.
- Requested `CLAUDE.md` is not present in this workspace (`find .. -name CLAUDE.md` returned no matches).
- Scope reviewed: all Sprint 002 code files currently changed/created under `x/exp/schema` in this workspace, plus deleted-file notes from git status.

## Findings (Ordered by Severity)

### 1) `CONSEQUENTIAL` - Marshal layer uses untyped `any` maps where concrete DTOs are feasible
- Criteria: **4. STRONG TYPING**
- File: `x/exp/schema/json.go:291`, `x/exp/schema/json.go:315`, `x/exp/schema/json.go:332`, `x/exp/schema/json.go:340`, `x/exp/schema/json.go:380`, `x/exp/schema/json.go:395`, `x/exp/schema/json.go:414`
- Problem: JSON emission relies on `map[string]any` and `any` return values rather than typed marshal structs.
- Why this matters: It weakens compile-time guarantees for output shape and pushes schema-shape mistakes to runtime.

### 2) `CONSEQUENTIAL` - Default-branch type assertions can panic
- Criteria: **3. NO DEFENSIVE CODING**, **5. IDIOMATIC**
- File: `x/exp/schema/json.go:410`, `x/exp/schema/json.go:429`, `x/exp/schema/resolve.go:425`
- Problem: `default` branches in type switches assert concrete types (`t.(EntityNameExpr)`, `expr.(TypeNameExpr)`).
- Why this matters: Invalid or partially-constructed `TypeExpr` values (possible via public mutable structs) can panic instead of returning structured errors; panic paths are non-idiomatic for library APIs that otherwise return errors.

### 3) `CONSEQUENTIAL` - Tests bypass public API and use package-internal construction paths
- Criteria: **1. NO TEST-ONLY CODE**
- File: `x/exp/schema/schema_test.go:2732`, `x/exp/schema/schema_test.go:2763`, `x/exp/schema/schema_test.go:2794`, `x/exp/schema/schema_test.go:2819`, `x/exp/schema/schema_test.go:2844`
- Problem: Tests directly instantiate internal structs/maps and call package-internal helper `newAnnotations()` instead of exercising only public entry points (`Unmarshal*`, `Marshal*`, builder API, `Resolve`).
- Why this matters: This violates the sprint quality contract that tests mirror production caller behavior.

### 4) `INCONSEQUENTIAL` - Comments restate mechanics instead of intent
- Criteria: **2. COMMENTS**
- File: `x/exp/schema/cedar.go:387`, `x/exp/schema/cedar.go:472`, `x/exp/schema/cedar.go:523`, `x/exp/schema/internal/scan/scanner.go:188`, `x/exp/schema/internal/scan/scanner.go:194`, `x/exp/schema/internal/scan/scanner.go:247`, `x/exp/schema/internal/scan/scanner.go:248`
- Problem: Inline comments like `// skip "entity"` and `// /` narrate exactly what the next line does.
- Why this matters: Noise without extra rationale; low signal comments reduce readability over time.

### 5) `INCONSEQUENTIAL` - Parser-error coverage block is tightly coupled to implementation line numbers
- Criteria: **2. COMMENTS**, **5. IDIOMATIC**
- File: `x/exp/schema/schema_test.go:2870`
- Problem: The long comment list references parser internals and exact line numbers.
- Why this matters: High maintenance overhead; comments become stale with routine refactors.

## Per-File Evaluation

- `docs/sprints/SPRINT-002.md`
  - No findings against criteria 1-5 (intent/spec document).
- `x/exp/schema/builder.go`
  - No findings against criteria 1-5.
- `x/exp/schema/cedar.go`
  - Finding: #4 (`INCONSEQUENTIAL`).
- `x/exp/schema/errors.go`
  - No findings against criteria 1-5.
- `x/exp/schema/internal/scan/scanner.go`
  - Finding: #4 (`INCONSEQUENTIAL`).
- `x/exp/schema/internal/scan/scanner_test.go`
  - No findings against criteria 1-5.
- `x/exp/schema/json.go`
  - Findings: #1 (`CONSEQUENTIAL`), #2 (`CONSEQUENTIAL`).
- `x/exp/schema/namespace.go`
  - No findings against criteria 1-5.
- `x/exp/schema/resolve.go`
  - Finding: #2 (`CONSEQUENTIAL`).
- `x/exp/schema/resolved/resolved.go`
  - No findings against criteria 1-5.
- `x/exp/schema/resolved/resolved_test.go`
  - No findings against criteria 1-5.
- `x/exp/schema/schema.go`
  - No findings against criteria 1-5.
- `x/exp/schema/schema_test.go`
  - Findings: #3 (`CONSEQUENTIAL`), #5 (`INCONSEQUENTIAL`).
- `x/exp/schema/typexpr.go`
  - No findings against criteria 1-5.

## Deleted Sprint Files (No Active Code to Review)
- `x/exp/schema/coverage_internal_test.go`
- `x/exp/schema/internal/parse/parser.go`
- `x/exp/schema/internal/parse/parser_test.go`
- `x/exp/schema/parse_cedar.go`
- `x/exp/schema/resolve_test.go`
- `x/exp/schema/resolved/resolved_internal_test.go`
- `x/exp/schema/types.go`
- `x/exp/schema/types_internal_test.go`
