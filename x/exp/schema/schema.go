package schema

// Schema represents a parsed but unresolved Cedar schema.
// Type references are raw strings that may or may not be qualified.
// Call Resolve() to get a resolved.Schema with fully-qualified type references.
type Schema struct {
	Namespaces map[string]*Namespace
	filename   string
}

// SetFilename sets the filename used in error messages.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// Filename returns the filename set for error messages.
func (s *Schema) Filename() string {
	return s.filename
}
