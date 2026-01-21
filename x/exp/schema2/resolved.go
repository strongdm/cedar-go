package schema2

import (
	"cmp"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// ResolvedSchema wraps ast.ResolvedSchema to provide marshaling methods.
type ResolvedSchema struct {
	ast.ResolvedSchema
}

// MarshalCedar converts the resolved schema to Cedar format.
// All types are fully resolved with qualified names.
// Namespaces are not included in the output.
// Common types are inlined into entities/actions where they are used.
func (r *ResolvedSchema) MarshalCedar() ([]byte, error) {
	// Convert ResolvedSchema to Schema by creating top-level declarations
	schema := r.toSchema()
	return schema.MarshalCedar()
}

// MarshalJSON converts the resolved schema to JSON format.
// All types are fully resolved with qualified names.
// Namespaces are not included in the output.
// Common types are inlined into entities/actions where they are used.
func (r *ResolvedSchema) MarshalJSON() ([]byte, error) {
	// Convert ResolvedSchema to Schema by creating top-level declarations
	schema := r.toSchema()
	return schema.MarshalJSON()
}

// toSchema converts a ResolvedSchema to a Schema with all declarations at the top level.
// All names are fully qualified, and all types are fully resolved.
// Namespaces and common types are not included.
func (r *ResolvedSchema) toSchema() Schema {
	var schema ast.Schema
	var nodes []ast.IsNode

	// Add entities sorted by name for deterministic output
	entityNames := make([]types.EntityType, 0, len(r.Entities))
	for name := range r.Entities {
		entityNames = append(entityNames, name)
	}
	slices.SortFunc(entityNames, func(a, b types.EntityType) int {
		return cmp.Compare(string(a), string(b))
	})
	for _, name := range entityNames {
		nodes = append(nodes, r.Entities[name])
	}

	// Add enums sorted by name for deterministic output
	enumNames := make([]types.EntityType, 0, len(r.Enums))
	for name := range r.Enums {
		enumNames = append(enumNames, name)
	}
	slices.SortFunc(enumNames, func(a, b types.EntityType) int {
		return cmp.Compare(string(a), string(b))
	})
	for _, name := range enumNames {
		nodes = append(nodes, r.Enums[name])
	}

	// Add actions sorted by type then ID for deterministic output
	actionUIDs := make([]types.EntityUID, 0, len(r.Actions))
	for uid := range r.Actions {
		actionUIDs = append(actionUIDs, uid)
	}
	slices.SortFunc(actionUIDs, func(a, b types.EntityUID) int {
		if c := cmp.Compare(string(a.Type), string(b.Type)); c != 0 {
			return c
		}
		return cmp.Compare(string(a.ID), string(b.ID))
	})
	for _, uid := range actionUIDs {
		nodes = append(nodes, r.Actions[uid])
	}

	schema.Nodes = nodes
	return Schema{schema: schema}
}
