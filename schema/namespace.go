package schema

// Namespace represents a Cedar namespace containing entity types, actions, and common type definitions.
type Namespace struct {
	name        string
	schema      *Schema
	entities    map[string]*Entity
	actions     map[string]*Action
	commonTypes map[string]Type
	annotations map[string]string
}

// WithAnnotation adds an annotation to the namespace.
func (n *Namespace) WithAnnotation(key, value string) *Namespace {
	if n.annotations == nil {
		n.annotations = make(map[string]string)
	}
	n.annotations[key] = value
	return n
}

// Declaration is implemented by Entity, Action, and CommonType to allow them
// to be added to a namespace.
type Declaration interface {
	addToNamespace(ns *Namespace)
}
