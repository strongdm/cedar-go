# Sprint 002 Merge Notes

## Sources
- Claude draft: `SPRINT-002-CLAUDE-DRAFT.md`
- Codex draft: `SPRINT-002-CODEX-DRAFT.md`
- Codex critique: `SPRINT-002-CLAUDE-DRAFT-CODEX-CRITIQUE.md`
- User interview decisions

## Critique Disposition

### 1. Resolved CommonTypes removal — REJECT
Codex recommends keeping `resolved.Namespace.CommonTypes`. The user explicitly chose "Remove CommonTypes" during the interview. Common types are inlined during resolution; callers who need the original common type definitions can access the unresolved `Schema`.

### 2. Builder API renaming — ACCEPT (partially)
Codex flags `StringType()` / `BoolType()` as unnecessary churn. Agreed. Keep `String()`, `Bool()`, `Long()` — in the `schema` package namespace (`schema.String()`), these read naturally and don't conflict. However, the user chose "Split types" for name references, so we use `EntityNameExpr` and `TypeNameExpr` instead of a single `NameExpr` with bool flag.

### 3. NameExpr guardrails — ACCEPT (already resolved)
User chose split types: `EntityNameExpr` (for memberOf, principalTypes, resourceTypes) and `TypeNameExpr` (for attribute types, common type bodies). Plus `EntityRefExpr` for explicit JSON entity references `{"type": "Entity", "name": "..."}`. Three distinct reference types eliminate ambiguity.

### 4. Code quality contract reflection — ACCEPT
Add explicit quality gates in the implementation section.

### 5. Compatibility checklist — ACCEPT
Add explicit compatibility check for `corpus_test.go` API surface: `Schema` type, `SetFilename()`, `UnmarshalCedar()`.

### 6. Action resolution underspecified — ACCEPT
Expand action resolution to cover: `ActionRef.Type` qualification, member-of validation (action groups must be declared), error handling for undeclared action parents.

## Key Decisions from Interview
- CommonTypes removed from resolved output
- Internal scanner in `internal/scan` sub-package
- Split reference types: `EntityNameExpr` + `TypeNameExpr` (not single `NameExpr`)

## Architecture Differences

Claude draft provides more implementation detail (concrete Go types, JSON format mapping). Codex draft is higher-level and more task-oriented. The merged plan uses Claude's detail with Codex's quality and compatibility recommendations incorporated.

## Merged Plan Changes vs Claude Draft
1. Keep `String()`, `Bool()`, `Long()` constructor names (not `StringType()` etc.)
2. Three reference types: `EntityRefExpr`, `EntityNameExpr`, `TypeNameExpr`
3. Add code quality gates section
4. Add compatibility checklist
5. Expand action resolution specification
6. Remove `CommonTypes` from resolved.Namespace (per user)
