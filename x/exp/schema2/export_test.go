package schema2

// This file exports internal test helpers for use by external test packages (schema2_test).

// TestMarshalCedar is a test helper that marshals a resolved schema to Cedar format.
func (r *ResolvedSchema) TestMarshalCedar() ([]byte, error) {
	return r.marshalCedar()
}

// TestMarshalJSON is a test helper that marshals a resolved schema to JSON format.
func (r *ResolvedSchema) TestMarshalJSON() ([]byte, error) {
	return r.marshalJSON()
}

// TestSchema is a test helper that converts a resolved schema back to a Schema.
func (r *ResolvedSchema) TestSchema() *Schema {
	return r.schema()
}
