# Sprint 002: Schema API Ergonomics Improvements

## Overview

This sprint addresses high-priority API ergonomics issues identified in the design review of the `x/exp/schema2` package. The focus is on eliminating silent failures, improving type discoverability, and adopting idiomatic Go patterns.

## Issues to Address

### 1. Silent Failures on Out-of-Context Method Calls (High Priority)

**Problem**: Methods like `In()`, `Attributes()`, `Principals()` silently do nothing when called without a current entity/action context.

**Solution**: Use separate builder types (`EntityBuilder`, `ActionBuilder`) that provide compile-time safety and return to the parent `Schema` via a terminal method.

### 2. Type Interface Not Discoverable (High Priority)

**Problem**: The `Type` interface has only an unexported method, making it impossible for IDEs to show implementations.

**Solution**: Add helper methods and improve documentation to guide users to available type constructors.

### 3. Functional Options for Parsing (High Priority)

**Problem**: `ParseCedarWithFilename` exists as a separate function; doesn't scale for future options.

**Solution**: Adopt functional options pattern: `ParseCedar(data, WithFilename("x"))`

### 4. EntityTypeKind Ergonomics (Medium Priority)

**Problem**: Requires type assertions to check if entity is enum.

**Solution**: Add `IsEnum()` and `AsEnum()` helper methods.

### 5. Lookup Patterns (Medium Priority)

**Problem**: `EntityType()` returns nil for missing entries with no way to distinguish "not found".

**Solution**: Add `LookupEntityType() (*ResolvedEntityType, bool)` pattern.

### 6. Early Validation (Medium Priority)

**Problem**: No way to check for errors before `Resolve()`.

**Solution**: Add `MustResolve()` for tests and improve error collection during building.

## Implementation Plan

### Phase 1: Separate Builder Types
- Create `EntityBuilder` with `In()`, `Attributes()`, `Tags()`, `Enum()` methods
- Create `ActionBuilder` with `In()`, `Principals()`, `Resources()`, `Context()` methods
- Both return `*Schema` via terminal method or next builder call
- Maintain backward compatibility where possible

### Phase 2: Functional Options for Parsing
- Create `ParseOption` type
- Add `WithFilename(string) ParseOption`
- Update `ParseCedar()` to accept options
- Deprecate `ParseCedarWithFilename()`

### Phase 3: Improved Lookup and Kind Helpers
- Add `LookupEntityType()` and `LookupAction()` with bool returns
- Add `IsEnum()`, `AsEnum()` to `ResolvedEntityType`
- Add `MustResolve()` for test convenience

### Phase 4: Documentation and Discoverability
- Add doc.go with package overview
- Document all type constructors prominently
- Add examples for common patterns

## Definition of Done

- [ ] EntityBuilder and ActionBuilder provide compile-time safety
- [ ] Functional options pattern for ParseCedar
- [ ] Lookup methods with boolean returns
- [ ] IsEnum/AsEnum helpers
- [ ] MustResolve for tests
- [ ] All existing tests pass
- [ ] New tests for new patterns
- [ ] Documentation updated
