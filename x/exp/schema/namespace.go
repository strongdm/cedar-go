package schema

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

// Namespace defines a namespace in a Cedar schema, containing entity types,
// actions, and common types.
type Namespace struct {
	Name        types.Path // empty for the anonymous namespace
	Annotations Annotations
	EntityTypes []*EntityType
	EnumTypes   []*EnumEntityType
	Actions     []*Action
	CommonTypes []*CommonType
}

// NewNamespace creates a new namespace with the given name.
// Use an empty string for the anonymous namespace.
func NewNamespace(name types.Path) *Namespace {
	return &Namespace{Name: name}
}

// Annotate adds an annotation to the namespace.
func (n *Namespace) Annotate(key types.Ident, value types.String) *Namespace {
	n.Annotations = append(n.Annotations, Annotation{Key: key, Value: value})
	return n
}

// AddEntityType adds an entity type to the namespace.
func (n *Namespace) AddEntityType(e *EntityType) *Namespace {
	n.EntityTypes = append(n.EntityTypes, e)
	return n
}

// AddEnumType adds an enumerated entity type to the namespace.
func (n *Namespace) AddEnumType(e *EnumEntityType) *Namespace {
	n.EnumTypes = append(n.EnumTypes, e)
	return n
}

// AddAction adds an action to the namespace.
func (n *Namespace) AddAction(a *Action) *Namespace {
	n.Actions = append(n.Actions, a)
	return n
}

// AddCommonType adds a common type to the namespace.
func (n *Namespace) AddCommonType(c *CommonType) *Namespace {
	n.CommonTypes = append(n.CommonTypes, c)
	return n
}

// MarshalCedar returns the Cedar format representation of the namespace.
func (n *Namespace) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write annotations
	for _, ann := range n.Annotations {
		buf.Write(ann.MarshalCedar())
		buf.WriteByte('\n')
	}

	// If named namespace, wrap in namespace block
	if n.Name != "" {
		buf.WriteString("namespace ")
		buf.WriteString(string(n.Name))
		buf.WriteString(" {\n")
	}

	indent := ""
	if n.Name != "" {
		indent = "  "
	}

	// Write common types first
	for _, ct := range n.CommonTypes {
		lines := bytes.Split(ct.MarshalCedar(), []byte{'\n'})
		for _, line := range lines {
			buf.WriteString(indent)
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}

	// Write entity types
	for _, et := range n.EntityTypes {
		lines := bytes.Split(et.MarshalCedar(), []byte{'\n'})
		for _, line := range lines {
			buf.WriteString(indent)
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}

	// Write enum entity types
	for _, et := range n.EnumTypes {
		lines := bytes.Split(et.MarshalCedar(), []byte{'\n'})
		for _, line := range lines {
			buf.WriteString(indent)
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}

	// Write actions
	for _, act := range n.Actions {
		lines := bytes.Split(act.MarshalCedar(), []byte{'\n'})
		for _, line := range lines {
			buf.WriteString(indent)
			buf.Write(line)
			buf.WriteByte('\n')
		}
	}

	// Close namespace block
	if n.Name != "" {
		buf.WriteString("}\n")
	}

	return buf.Bytes()
}

// CommonType defines a reusable type alias in a Cedar schema.
type CommonType struct {
	Name        types.Ident
	Annotations Annotations
	Type        Type
}

// NewCommonType creates a new common type definition.
func NewCommonType(name types.Ident, t Type) *CommonType {
	return &CommonType{Name: name, Type: t}
}

// Annotate adds an annotation to the common type.
func (c *CommonType) Annotate(key types.Ident, value types.String) *CommonType {
	c.Annotations = append(c.Annotations, Annotation{Key: key, Value: value})
	return c
}

// MarshalCedar returns the Cedar format representation of the common type.
func (c *CommonType) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write annotations
	for _, ann := range c.Annotations {
		buf.Write(ann.MarshalCedar())
		buf.WriteByte('\n')
	}

	buf.WriteString("type ")
	buf.WriteString(string(c.Name))
	buf.WriteString(" = ")
	buf.Write(c.Type.MarshalCedar())
	buf.WriteByte(';')
	return buf.Bytes()
}
