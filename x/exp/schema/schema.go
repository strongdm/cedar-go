// Package schema provides types and functions for working with Cedar schemas.
//
// Deprecated: This package is deprecated. Use github.com/cedar-policy/cedar-go/schema instead.
// This package is maintained for backwards compatibility and wraps the new schema package.
package schema

import (
	"fmt"

	publicschema "github.com/cedar-policy/cedar-go/schema"
)

// Schema is a description of entities and actions that are allowed for a PolicySet.
// They can be used to validate policies and entity definitions and also provide documentation.
//
// Schemas can be represented in either JSON (*JSON functions) or Human-readable formats
// (*Cedar functions) just like policies. Marshalling and unmarshalling between the formats is allowed.
//
// Deprecated: Use github.com/cedar-policy/cedar-go/schema.Schema instead.
type Schema struct {
	inner    *publicschema.Schema
	filename string
}

// UnmarshalCedar parses and stores the human-readable schema from src and returns an error if the schema is invalid.
//
// Any errors returned will have file positions matching filename.
func (s *Schema) UnmarshalCedar(src []byte) error {
	if s.inner == nil {
		s.inner = publicschema.NewSchema()
		if s.filename != "" {
			s.inner.SetFilename(s.filename)
		}
	}
	return s.inner.UnmarshalCedar(src)
}

// MarshalCedar serializes the schema into the human readable format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	if s.inner == nil {
		return nil, fmt.Errorf("schema is empty")
	}
	return s.inner.MarshalCedar()
}

// UnmarshalJSON deserializes the JSON schema from src or returns an error if the JSON is not valid schema JSON.
func (s *Schema) UnmarshalJSON(src []byte) error {
	if s.inner == nil {
		s.inner = publicschema.NewSchema()
		if s.filename != "" {
			s.inner.SetFilename(s.filename)
		}
	}
	return s.inner.UnmarshalJSON(src)
}

// MarshalJSON serializes the schema into the JSON format.
//
// If the schema was loaded from UnmarshalCedar, it will convert the human-readable format into the JSON format.
// An error is returned if the schema is invalid.
func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.inner == nil {
		return nil, nil
	}
	return s.inner.MarshalJSON()
}

// SetFilename sets the filename for the schema in the returned error messages from Unmarshal*.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
	if s.inner != nil {
		s.inner.SetFilename(filename)
	}
}
