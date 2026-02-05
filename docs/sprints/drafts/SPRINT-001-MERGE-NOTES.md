# Sprint 001 Merge Notes

## Claude Draft Strengths
- Comprehensive architecture diagram with clear layer separation
- Detailed task breakdown with specific checklist items
- Generic type parameter progression explanation mirrors Rust design
- Strong security considerations section
- Thorough risk table with mitigations

## Codex Draft Strengths
- Explicit Non-Goals section clarifies boundaries
- Concrete milestones (M1-M5) provide checkpoints
- Time estimates per phase add planning context
- Testing Strategy section is well-organized
- Notes section captures important reminders (corpus_test.go update, extension types)

## Valid Critiques Accepted

### From Codex Critique:
1. **API Examples needed** - Will add builder pattern examples
2. **Cedar round-trip semantics** - Define AST-equivalence, not text-equivalence
3. **Resolution details** - Add extension types list and namespace rules
4. **Error conformance** - Add explicit typed error list
5. **JSON schema mapping** - Document discriminator fields
6. **corpus_test.go action** - Add explicit action item

## Critiques Rejected (with reasoning)

1. **Multi-fragment merging** - Defer to future sprint; current scope is single-fragment parsing. Noted as out of scope.
2. **Performance expectations** - Premature optimization; will profile after implementation works

## Interview Refinements Applied

1. **Test vectors from Rust** - Extract golden files from Rust unit tests
2. **Builder pattern API** - Use fluent builder methods, not just struct literals
3. **Concrete types** - Use `RawSchema` and `ResolvedSchema` instead of generics
4. **No source locations** - Skip location tracking for simplicity

## Final Decisions

1. **Type design**: Concrete types (`Schema`, `ResolvedSchema`) for public API
2. **API style**: Builder pattern with fluent methods
3. **Round-trip equivalence**: AST-equal, not text-equal (formatting may differ)
4. **Extension types**: `ipaddr`, `decimal`, `datetime`, `duration` seeded in `__cedar` namespace
5. **Error types**: `CycleError`, `UndefinedTypeError`, `ShadowError`, `DuplicateError`, `ReservedNameError`
6. **Test approach**: Extract Rust test vectors + corpus parsing
7. **Resolved types**: Use `types.EntityType` for entity types, `types.EntityUID` for actions (not just string paths)
