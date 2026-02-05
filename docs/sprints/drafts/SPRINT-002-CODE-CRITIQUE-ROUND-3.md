# Sprint 002 Code Critique (Round 3)

## Context
- Reviewed intent in `docs/sprints/SPRINT-002.md`.
- `CLAUDE.md` was requested but is not present in this workspace (no `CLAUDE.md` found under repo root).
- Scope: all Sprint 002 changed/created files currently present under `x/exp/schema` plus deleted-file notes from git status.

## Findings (Ordered by Severity)

### 1) `CONSEQUENTIAL` - Untyped `any` maps are used where concrete encoding structs are feasible
- Criteria: **4. STRONG TYPING**
- File: `x/exp/schema/json.go:291`, `x/exp/schema/json.go:315`, `x/exp/schema/json.go:340`, `x/exp/schema/json.go:395`, `x/exp/schema/json.go:414`
- Problem: JSON marshaling is built around `map[string]any` and helpers returning `any`.
- Why this matters: This weakens compile-time guarantees and creates runtime-shape coupling. Most output shapes here are known and can be represented with typed marshal DTOs.

### 2) `CONSEQUENTIAL` - Default-branch type assertions can panic instead of returning structured errors
- Criteria: **5. IDIOMATIC**
- File: `x/exp/schema/json.go:410`, `x/exp/schema/json.go:429`, `x/exp/schema/resolve.go:425`
- Problem: Default branches assume exact concrete types via assertions (`t.(EntityNameExpr)`, `expr.(TypeNameExpr)`).
- Why this matters: This is panic-prone API behavior for malformed/partial schema values (for example nil `TypeExpr` values introduced via public mutability), and is not idiomatic for library code that otherwise returns errors.

### 3) `CONSEQUENTIAL` - Tests use non-public helpers and white-box construction paths
- Criteria: **1. NO TEST-ONLY CODE**
- File: `x/exp/schema/schema_test.go:2735`, `x/exp/schema/schema_test.go:2745`, `x/exp/schema/schema_test.go:2774`, `x/exp/schema/schema_test.go:2806`, `x/exp/schema/schema_test.go:2834`
- Problem: Tests call unexported `newAnnotations()` and directly wire internals in ways not used by external callers.
- Why this matters: The sprint contract requires tests to exercise the same public API surface as production callers.

### 4) `INCONSEQUENTIAL` - Multiple comments restate the next line instead of explaining intent
- Criteria: **2. COMMENTS**
- File: `x/exp/schema/cedar.go:387`, `x/exp/schema/cedar.go:472`, `x/exp/schema/cedar.go:523`, `x/exp/schema/internal/scan/scanner.go:188`, `x/exp/schema/internal/scan/scanner.go:194`, `x/exp/schema/internal/scan/scanner.go:247`
- Problem: Comments like `// skip "entity"` and `// skip opening "` narrate mechanics rather than non-obvious reasoning.

### 5) `INCONSEQUENTIAL` - Coverage-comment block is tightly coupled to parser line numbers
- Criteria: **2. COMMENTS**, **5. IDIOMATIC**
- File: `x/exp/schema/schema_test.go:2870`
- Problem: The long comment block enumerates exact parser line numbers and micro-branches.
- Why this matters: High churn for low value; it will stale quickly as implementation moves.

## Per-File Evaluation

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
