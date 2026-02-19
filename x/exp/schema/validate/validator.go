package validate

import "github.com/cedar-policy/cedar-go/x/exp/schema/resolved"

// Validator validates Cedar policies, entities, and requests against a resolved schema.
type Validator struct {
	schema *resolved.Schema
	strict bool
}

// Option configures a Validator.
type Option func(*Validator)

// WithStrict returns an Option that enables strict validation mode (default).
func WithStrict() Option { return func(v *Validator) { v.strict = true } }

// WithPermissive returns an Option that enables permissive validation mode.
func WithPermissive() Option { return func(v *Validator) { v.strict = false } }

// New creates a Validator for the given schema. By default, strict mode is enabled.
func New(s *resolved.Schema, opts ...Option) *Validator {
	v := &Validator{schema: s, strict: true}
	for _, opt := range opts {
		opt(v)
	}
	return v
}
