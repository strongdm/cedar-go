# Codex Critique of Claude's Sprint 001 Draft

## What's Strong
- Clear three-phase naming progression (Raw → Conditional → Internal) aligned to Rust.
- Solid file breakdown and layered architecture (API, Types, Resolve, Parser).
- Concrete task lists per phase with good coverage (duplicates, reserved names, tokenizer keywords).
- Practical Definition of Done with lint/vet gates and corpus parsing goals.
- Risk table identifies key failure modes and aligns mitigations to differential testing.

## Gaps and Suggestions

### Programmatic API specifics
- Add examples for builder usage and struct literal usage to mirror `ast/` patterns.
- Specify minimal constructors and helpers (e.g., `schema.New()`, `ns.AddEntity()`, `attr.Required()`).

### Cedar round-trip semantics
- Define equivalence: formatting-insensitive but AST-equal or normalized text? Call out idempotence expectations.

### Resolution details
- Clarify default namespace rules and how empty namespace resolves.
- State case sensitivity, path normalization, and escaping strategy for idents.
- List the extension types seeded and where they live.

### Error conformance
- Describe how to align messages with Rust (e.g., snapshot expected strings, shared wording).
- Add explicit typed error set and when callers should match them (Cycle, UndefinedType, Shadow).

### JSON schema mapping
- Call out exact field names and discriminators for types (e.g., "type": "Set", "element": …).
- Note `appliesTo` shape for actions (principal, resource, context) and required/optional attributes.
- Include behavior for empty namespace key and duplicate detection rules by section (entities, actions, common).

### Multi-fragment merging
- Intent mentions merging fragments; add design note or explicitly defer as future work with constraints.

### Tests and fixtures
- Enumerate minimal seed schemas for unit/integration tests before corpus availability.
- Define round-trip tests that assert both JSON and Cedar with mixed namespaces.

### Integration points
- Clarify reuse vs fork of `internal/parser` tokenizer; describe tradeoffs and chosen path.
- Note update to `corpus_test.go` (draft mentions it; add exact action and gating).

### Performance and memory
- Briefly set expectations (e.g., O(n) def indexing, recursion depth limits for cycle detection).

### Minor correctness issues
- The "Type Parameter Progression" code block appears truncated ("type Fragment[N Type …]").
- Consider recommending concrete non-generic internal types for simplicity in Go (Claude already suggests; reinforce choice and reflect in file APIs).

## Proposed Additions to Claude's Plan

### API Examples (short)
- Construct a schema with one namespace, one entity with record attributes, and one action applying to principal/resource.
- JSON and Cedar round-trip calls via `Schema.Marshal* / Unmarshal*`.

### Resolution Spec Checklist
- Disambiguation priority
- Shadowing rules
- Forward references
- Cycle detection reporting (which node pair, path)

### Acceptance Tests
- Table-driven tests mapping input → expected fully-qualified paths for edge cases (e.g., `__cedar::` prefix, ambiguity).

### Docs
- A short `x/exp/schema/README.md` for data model and examples.

## Verdict
Claude's draft is strong and implementable. With the additions above—particularly API examples, stricter JSON/Cedar mapping details, explicit resolution/namespace rules, and a clearer error parity strategy—the plan will better ensure correctness and developer ergonomics while staying aligned with the sprint intent. If helpful, I can fold these suggestions directly into Claude's draft structure.
