# Codex Critique of Claude's Sprint 002 Draft

## Strengths

- Strong overall structure: clear scope, non-goals, architecture, phased plan, and completion criteria.
- Resolution section captures the key behavior requirements: candidate generation, priority ordering, shadowing checks, and cycle detection.
- Testing emphasis is good: corpus coverage, Rust parity, and round-trip testing are all explicitly called out.
- Compatibility concern for `corpus_test.go` is acknowledged, which is critical for this repo.

## Key Gaps and Risks

### 1. `NameExpr{MustEntity bool}` weakens type safety

Using one reference type with a mode flag creates invalid state potential and adds branching complexity in parser/builder/resolver paths. The sprint intent explicitly pushes for strong typing and Go-idiomatic design; split types (`EntityNameExpr`, `TypeNameExpr`, plus `EntityRefExpr`) are clearer and safer.

### 2. Constructor renames create avoidable churn

The proposal switches to `LongType()/StringType()/BoolType()` and `SetOf()/RecordOf()`. This is not aligned with established package ergonomics (`schema.Long()`, `schema.String()`, `schema.Set(...)`) and adds migration cost without clear value.

### 3. File-level replacement plan is too rigid

The draft assumes wholesale split into `namespace.go` + `typexpr.go` and deletion of `types.go`, `parse_cedar.go`, and `internal/parse/`. The sprint intent is about replacing implementation quality, not enforcing file names. Mandating a file topology can create unnecessary refactor churn and review noise.

### 4. JSON mapping leaves ambiguity for context typing

`AppliesTo.Context` is typed as `TypeExpr` in prose, but the builder and examples imply record-centric behavior. The plan should explicitly define allowed forms and error behavior when context resolves to non-record types.

### 5. Action resolution is under-specified in failure cases

The draft covers positive path but not enough negative-path semantics:
- undeclared action group in `memberOf`
- cross-namespace `ActionRef.Type` qualification and normalization
- duplicate action UID collisions
- exact error categories for these failures

### 6. Code-quality contract from intent is not reflected strongly enough

Intent specifies strict standards (no test-only production code, no defensive coding against internal invariants, comments explain intent only). The draft references quality indirectly but should include these as explicit gates to avoid drift during implementation.

## Recommended Changes

1. Replace `NameExpr{MustEntity bool}` with three explicit reference variants:
- `EntityNameExpr`
- `TypeNameExpr`
- `EntityRefExpr`

2. Keep constructor names concise and idiomatic:
- `Long()`, `String()`, `Bool()`, `Set()`, `Record()`, `Entity()`, `Extension()`.

3. Make file structure guidance outcome-based, not prescriptive:
- define responsibilities per component
- allow implementation to stay in existing files where clean.

4. Expand action resolution spec with explicit error paths and sentinel/type mapping.

5. Add explicit quality gates section mirroring sprint intent contract.

6. Add a compatibility checklist section that locks the `corpus_test.go` touchpoints:
- `schema.Schema`
- `(*Schema).SetFilename(string)`
- `(*Schema).UnmarshalCedar([]byte) error`

## Suggested Additions to Claude Draft

### Resolution Checklist (must-pass)
- Priority ordering: common > entity > primitive/extension in type positions
- Entity-only resolution for memberOf/principal/resource positions
- RFC 70 shadowing against empty namespace
- Common type cycle detection with deterministic error reporting
- `__cedar::` handling and reserved prefix rules

### Testing Additions
- Table-driven tests for action `memberOf` error cases
- Context type validation tests (`appliesTo.context` must resolve to record)
- Namespace qualification tests for empty vs named namespace precedence

## Verdict

Claude's draft is solid and implementable, but it should tighten type-safety decisions, reduce avoidable API/file churn, and explicitly codify action failure semantics plus code-quality gates. With those changes, it will align more closely with Sprint 002 intent and repository conventions.
