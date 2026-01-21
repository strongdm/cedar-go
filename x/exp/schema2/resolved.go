package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// ResolvedSchema wraps ast.ResolvedSchema to provide access to resolved schema data.
type ResolvedSchema struct {
	ast.ResolvedSchema
}

// Schema converts the resolved schema back to a valid Schema with proper namespace structure.
// All types remain fully resolved (common types are inlined).
// Entity, enum, and action names are unqualified within their namespaces.
func (r *ResolvedSchema) Schema() *Schema {
	astSchema := r.ResolvedSchema.Schema()
	return &Schema{schema: *astSchema}
}

// MarshalCedar converts the resolved schema to Cedar format.
// The schema is first converted to a valid Schema with proper namespace structure,
// then marshaled to Cedar format.
func (r *ResolvedSchema) MarshalCedar() ([]byte, error) {
	schema := r.Schema()
	return schema.MarshalCedar()
}

// MarshalJSON converts the resolved schema to JSON format.
// The schema is first converted to a valid Schema with proper namespace structure,
// then marshaled to JSON format.
func (r *ResolvedSchema) MarshalJSON() ([]byte, error) {
	schema := r.Schema()
	return schema.MarshalJSON()
}
