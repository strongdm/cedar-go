// Package schema2 provides a new implementation of Cedar schema parsing and serialization.
package schema2

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/json"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/parser"
)

// Schema represents a Cedar schema with parsing and marshaling capabilities.
type Schema struct {
	filename string
	schema   *ast.Schema
}

// SetFilename sets the filename for this schema.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// MarshalJSON encodes the Schema in the JSON format specified by the Cedar documentation.
func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.schema == nil {
		return []byte{}, nil
	}
	jsonSchema := (*json.Schema)(s.schema)
	return jsonSchema.MarshalJSON()
}

// UnmarshalJSON parses a Schema in the JSON format specified by the Cedar documentation.
func (s *Schema) UnmarshalJSON(b []byte) error {
	var jsonSchema json.Schema
	if err := jsonSchema.UnmarshalJSON(b); err != nil {
		return err
	}
	s.schema = (*ast.Schema)(&jsonSchema)
	return nil
}

// MarshalCedar encodes the Schema in the human-readable format specified by the Cedar documentation.
func (s *Schema) MarshalCedar() ([]byte, error) {
	if s.schema == nil {
		return nil, &SchemaError{Err: "cannot marshal empty schema"}
	}

	var buf bytes.Buffer
	cedarBytes := s.schema.MarshalCedar()
	buf.Write(cedarBytes)
	return buf.Bytes(), nil
}

// UnmarshalCedar parses a Schema in the human-readable format specified by the Cedar documentation.
func (s *Schema) UnmarshalCedar(b []byte) error {
	schema, err := parser.ParseSchema(b)
	if err != nil {
		return err
	}
	s.schema = schema
	return nil
}

// SchemaError represents an error in schema processing.
type SchemaError struct {
	Err string
}

func (e *SchemaError) Error() string {
	return e.Err
}
