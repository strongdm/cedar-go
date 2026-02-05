# Sprint 001 Intent: Cedar Schema Support

## Seed

I've got a Go implementation of the Cedar policy language in this directory. I want to add support for Cedar schema, in particular: (1) serialization and deserialization from JSON and Cedar text (2) programmatic construction of schema (3) the ability to take a schema that's built either through parsing or programmatically and convert it into a resolved form with fully-qualified type information (e.g. `types.EntityType`s and `types.EntityUID`s). There is an existing reference implementation in Rust located in /Volumes/src/ai/cedar and the documentation for the language is located in /Volumes/src/ai/cedar-docs. The Go implementation should conform to the docs and reference implementation in terms of expected behavior both in success and error conditions. Completely ignore the deleted implementation of schema. I want you to slurp as many example schemas for verifying the parser from the Rust implementation. There are also examples in the Cedar corpus samples.

## Context

**Current Project State:**
- cedar-go is a Go implementation of the Cedar policy language at v1.4.0
- Core authorizer, JSON marshalling, all core/extended types exist
- `corpus_test.go:19` imports `x/exp/schema` which doesn't exist (causing build failure)
- Missing: schema validator, formatter, partial evaluation, policy templates

**Recent Sprint Context:**
- No prior sprints exist - this is SPRINT-001
- Project has been stable with incremental improvements (datetime/duration support, entity tags, batch authorization)

**Relevant Codebase Areas:**
- `types/` - Contains `EntityType`, `EntityUID`, `Path` types essential for schema resolution
- `internal/parser/` - Cedar language parser infrastructure (tokenizer, AST nodes)
- `internal/json/` - JSON marshalling utilities
- `ast/` - Programmatic AST construction patterns
- `x/exp/` - Experimental features directory (target location)

## Relevant Codebase Areas

### Types Package (`types/`)
- `EntityType` (alias for `Path`) - Type portion of EntityUID
- `EntityUID` - Principal/action/resource identifiers
- `Path` - Series of idents separated by `::`
- `Value` interface - Used for attribute values
- `Record`, `Set`, `String`, `Long`, `Boolean` - Cedar types

### Internal Parser (`internal/parser/`)
- `cedar_tokenize.go` - Tokenizer for Cedar syntax
- `cedar_unmarshal.go` - Policy parsing
- `node.go` - AST node types

### AST Package (`ast/`)
- Programmatic construction patterns (builder methods)
- Operator nodes, scope nodes, value nodes

### Reference Implementation Insights (Rust)
- `cedar-policy-core/src/validator/schema/` - Schema validation types
- `raw_name.rs` - `RawName` → `ConditionalName` resolution
- `namespace_def.rs` - Namespace definition processing
- `json_schema.rs` - JSON schema parsing (~157K lines)
- Resolution priority: common type > entity type > primitive/extension type

## Constraints

- Must follow Go idioms and existing project patterns
- Must place schema in `x/exp/schema` (experimental)
- Must support both JSON and Cedar text formats
- Must support programmatic construction
- Must produce resolved form with fully-qualified types
- Must handle namespace qualification correctly
- Must handle type disambiguation per Cedar spec
- Error messages should match Rust behavior where applicable

## Success Criteria

1. **Parsing**: Parse all 6,084 corpus schemas without error
2. **JSON Round-trip**: Schema → JSON → Schema produces equivalent result
3. **Cedar Round-trip**: Schema → Cedar → Schema produces equivalent result
4. **Programmatic API**: Builder pattern for constructing schemas in code
5. **Resolution**: Convert parsed/constructed schema to resolved form with:
   - All entity type references become `types.EntityType` values (e.g., `types.EntityType("PhotoFlash::User")`)
   - All action references become `types.EntityUID` values (e.g., `types.NewEntityUID("PhotoFlash::Action", "viewPhoto")`)
   - Parent types (memberOf) are `[]types.EntityType`
   - AppliesTo principal/resource types are `[]types.EntityType`
6. **Error Conformance**: Error conditions match Rust implementation

## Verification Strategy

### Primary: Differential Testing
- Parse all 6,084 corpus `.cedarschema` files
- Compare resolved output against expected fully-qualified names
- Run parallel parsing with Rust CLI and compare error/success conditions

### Secondary: Unit Tests
- Test each component (tokenizer, parser, resolver) in isolation
- Test edge cases from documentation:
  - Type disambiguation (`__cedar::` prefix)
  - Shadowing rules (can't shadow empty namespace definitions)
  - Forward references in common types
  - Cycle detection in common type definitions

### Tertiary: Integration Tests
- Parse and resolve schemas from `cedar-policy-cli/sample-data/`
- Verify programmatic construction produces identical resolved form to parsing

## Uncertainty Assessment

- **Correctness uncertainty**: HIGH - Complex resolution rules, namespace qualification, type disambiguation require exact conformance to spec
- **Scope uncertainty**: MEDIUM - Core requirements are clear, but edge cases may emerge during implementation
- **Architecture uncertainty**: LOW - Existing patterns in `ast/` and `internal/parser/` provide clear guidance

## Open Questions

1. **Resolution Representation**: What Go types should represent the resolved schema? Options:
   - Separate `ResolvedSchema` type with `types.EntityType` fields
   - Same type with resolved flag/variant
   - Generic over name type (like Rust's `ValidatorSchemaFragment<N, A>`)

2. **Error Handling**: Should we use dedicated error types (like Rust) or Go's idiomatic `error` interface with sentinel errors?

3. **Incremental Parsing**: Should we support parsing multiple schema fragments and merging them (like Rust)?

4. **Validation vs Parsing**: Should validation (cycle detection, undefined reference detection) happen during parsing or as a separate pass?

5. **Common Type Inlining**: Should we inline common types during resolution (like Rust) or preserve references?

6. **Location Tracking**: Should we track source locations for LSP support (optional feature)?
