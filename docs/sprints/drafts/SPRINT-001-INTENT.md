# Sprint 001 Intent: Programmatic Cedar Schema Interface

## Seed

We're going to create a programmatic interface for Cedar schema in the https://github.com/cedar-policy/cedar-go repository. It needs to have a way to parse Cedar schema from JSON and Cedar text and to serialize back to those formats, but it should also have an interface for programmatic construction that is similar to the Cedar policy builder from the same repo. One of the most important features is that you must be able to turn a parsed Schema into a "resolved" Schema, which is a separate type and deals in fully-qualified EntityTypes (and EntityUIDs for enums). This will be created completely out of place in `x/exp/schema2`. Only the minimum surface area should be exposed publicly; packages that shouldn't be exposed should go in an internal directory.

## Context

- **Current project state**: cedar-go has existing internal schema AST (`internal/schema/ast/`) supporting JSON and human-readable Cedar schema formats with parsing/formatting, but no public programmatic builder API similar to the policy AST builder
- **Pattern to follow**: The policy builder pattern in `ast/policy.go` wraps internal types from `x/exp/ast/` with a fluent builder interface and supports JSON/Cedar marshal/unmarshal
- **Key architectural insight**: The Rust implementation has two distinct schema representations: (1) `json_schema::Fragment<N>` (parsed, possibly with unresolved names where N=RawName) and (2) `ValidatorSchema` (resolved, with fully-qualified EntityTypes/EntityUIDs, computed descendants, expanded common types)
- **Existing infrastructure**: `x/exp/schema/schema.go` provides basic Schema type with marshal/unmarshal but no builder or resolved schema - we will NOT modify this, creating in `x/exp/schema2` instead
- **Key types to leverage**: `types.EntityType` (Path alias), `types.EntityUID`, existing `internal/schema/ast` types

## Recent Sprint Context

First sprint - no previous sprints exist.

## Relevant Codebase Areas

### Policy Builder Pattern (to emulate)
- `ast/policy.go` - Public fluent builder wrapping internal types
- `ast/node.go`, `ast/scope.go`, `ast/value.go` - Builder components
- `x/exp/ast/*.go` - Internal policy AST types

### Existing Schema Infrastructure
- `internal/schema/ast/ast.go` - Human-readable schema AST (Schema, Namespace, Entity, Action, CommonTypeDecl, Type, RecordType, SetType, Path, etc.)
- `internal/schema/ast/json.go` - JSON schema types (JSONSchema, JSONNamespace, JSONEntity, JSONAction, JSONType, etc.)
- `internal/schema/ast/convert_json.go` - JSON to human AST conversion
- `internal/schema/ast/convert_human.go` - Human AST to JSON conversion
- `internal/schema/ast/format.go` - Human-readable formatting
- `internal/schema/parser/` - Parser for Cedar schema text
- `x/exp/schema/schema.go` - Current public schema type (parse/serialize only)

### Types Package
- `types/entity_uid.go` - EntityType (Path alias), EntityUID types
- `types/ident.go` - Identifier handling

### Rust Reference (ValidatorSchema pattern)
- `cedar-policy-core/src/validator/schema.rs` - ValidatorSchema with HashMap<EntityType, ValidatorEntityType>, HashMap<EntityUID, ValidatorActionId>
- `cedar-policy-core/src/validator/schema/entity_type.rs` - ValidatorEntityType with name, descendants (computed TC), attributes, kind (Standard/Enum)
- `cedar-policy-core/src/validator/json_schema.rs` - Fragment<N> with RawName -> ConditionalName -> InternalName type progression

## Constraints

- Must create in `x/exp/schema2/` - completely out of place from existing `x/exp/schema`
- Only expose minimum necessary public API surface
- Internal implementation goes in `x/exp/schema2/internal/`
- Must follow cedar-go patterns (fluent builders, wrap internal types)
- Must support both JSON and Cedar text formats (parse and serialize)
- Must maintain separation between parsed schema (may have unqualified names) and resolved schema (fully-qualified)
- Resolved schema must use existing types.EntityType and types.EntityUID
- Must handle namespaces correctly during resolution
- Must compute transitive closure for entity hierarchy (descendants)
- Enums must be represented as EntityUIDs in resolved form

## Success Criteria

1. Users can programmatically construct a schema using a fluent builder API
2. Schemas can be parsed from JSON and Cedar text formats
3. Schemas can be serialized to JSON and Cedar text formats
4. Parsed schemas can be "resolved" into a separate resolved schema type
5. Resolved schema contains fully-qualified EntityTypes everywhere
6. Resolved schema represents enum values as EntityUIDs
7. Resolved schema computes and stores entity hierarchy (descendants)
8. API is consistent with existing policy builder patterns

## Open Questions

1. Should the builder produce parsed schema or resolved schema directly?
   - Recommendation: Builder produces parsed schema, separate Resolve() method creates resolved schema

2. How should we handle validation errors during resolution?
   - Need to detect: undefined type references, circular dependencies, invalid namespaces

3. Should common types be inlined in resolved schema (like Rust) or kept as references?
   - Recommendation: Inline for simplicity, matching Rust behavior

4. What should the public API surface look like?
   - Schema (parsed) with builder methods
   - ResolvedSchema (resolved) from Schema.Resolve()
   - Both support marshal/unmarshal

5. Should we support schema fragments (multiple namespaces from different sources) or just complete schemas?
   - Initial implementation: complete schemas only, fragments as future enhancement
