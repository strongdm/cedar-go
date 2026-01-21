package schema2

// Test helpers for ResolvedSchema - these are only used in tests.

// schema converts the resolved schema back to a valid Schema with proper namespace structure.
// All types remain fully resolved (common types are inlined).
// Entity, enum, and action names are unqualified within their namespaces.
func (r *ResolvedSchema) schema() *Schema {
	astSchema := r.ResolvedSchema.Schema()
	return &Schema{schema: *astSchema}
}

// marshalCedar converts the resolved schema to Cedar format.
// The schema is first converted to a valid Schema with proper namespace structure,
// then marshaled to Cedar format.
func (r *ResolvedSchema) marshalCedar() ([]byte, error) {
	schema := r.schema()
	return schema.MarshalCedar()
}

// marshalJSON converts the resolved schema to JSON format.
// The schema is first converted to a valid Schema with proper namespace structure,
// then marshaled to JSON format.
func (r *ResolvedSchema) marshalJSON() ([]byte, error) {
	schema := r.schema()
	return schema.MarshalJSON()
}
