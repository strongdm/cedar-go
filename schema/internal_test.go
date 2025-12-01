package schema

import (
	"testing"
)

// TestInternalMethods tests internal methods for 100% coverage.
// This file is in the schema package (not schema_test) to access private methods.
func TestInternalMethods(t *testing.T) {
	t.Run("isType marker methods", func(t *testing.T) {
		// Call the isType marker methods directly
		// These are empty marker methods but need to be called for coverage
		pathType := &PathType{path: "String"}
		setType := &SetType{element: pathType}
		recordType := &RecordType{attributes: make(map[string]*Attribute)}

		// Call the marker methods
		pathType.isType()
		setType.isType()
		recordType.isType()

		// Verify they're still valid Type implementations
		var _ Type = pathType
		var _ Type = setType
		var _ Type = recordType
	})

	t.Run("Namespace WithAnnotation", func(t *testing.T) {
		// Create a namespace and call WithAnnotation
		ns := &Namespace{
			name:        "Test",
			entities:    make(map[string]*Entity),
			actions:     make(map[string]*Action),
			commonTypes: make(map[string]Type),
		}

		// Call WithAnnotation multiple times
		ns = ns.WithAnnotation("doc", "Test namespace")
		ns = ns.WithAnnotation("version", "1.0")

		// Verify annotations were set
		if ns.annotations == nil {
			t.Error("Expected annotations to be initialized")
		}
		if ns.annotations["doc"] != "Test namespace" {
			t.Errorf("Expected 'doc' annotation, got %v", ns.annotations["doc"])
		}
		if ns.annotations["version"] != "1.0" {
			t.Errorf("Expected 'version' annotation, got %v", ns.annotations["version"])
		}
	})
}
