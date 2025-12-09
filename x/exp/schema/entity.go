package schema

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

// EntityType defines an entity type in a Cedar schema.
type EntityType struct {
	Name        types.Ident
	Annotations Annotations
	MemberOf    []types.Path // entity types this entity can be a member of
	Shape       *TypeRecord  // optional shape (attributes)
	Tags        Type         // optional tags type
}

// Entity creates a new entity type definition with the given name.
func Entity(name types.Ident) *EntityType {
	return &EntityType{Name: name}
}

// Annotate adds an annotation to the entity type.
func (e *EntityType) Annotate(key types.Ident, value types.String) *EntityType {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// In specifies that this entity type can be a member of the given entity types.
func (e *EntityType) In(entityTypes ...types.Path) *EntityType {
	e.MemberOf = append(e.MemberOf, entityTypes...)
	return e
}

// WithShape sets the shape (record type) for this entity type.
func (e *EntityType) WithShape(shape TypeRecord) *EntityType {
	e.Shape = &shape
	return e
}

// WithTags sets the tags type for this entity type.
func (e *EntityType) WithTags(tagsType Type) *EntityType {
	e.Tags = tagsType
	return e
}

// MarshalCedar returns the Cedar format representation of the entity type.
func (e *EntityType) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write annotations
	for _, ann := range e.Annotations {
		buf.Write(ann.MarshalCedar())
		buf.WriteByte('\n')
	}

	// Write entity declaration
	buf.WriteString("entity ")
	buf.WriteString(string(e.Name))

	// Write memberOf clause
	if len(e.MemberOf) > 0 {
		buf.WriteString(" in ")
		if len(e.MemberOf) == 1 {
			buf.WriteString(string(e.MemberOf[0]))
		} else {
			buf.WriteByte('[')
			for i, m := range e.MemberOf {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(string(m))
			}
			buf.WriteByte(']')
		}
	}

	// Write shape
	if e.Shape != nil {
		buf.WriteByte(' ')
		buf.Write(e.Shape.MarshalCedar())
	}

	// Write tags
	if e.Tags != nil {
		buf.WriteString(" tags ")
		buf.Write(e.Tags.MarshalCedar())
	}

	buf.WriteByte(';')
	return buf.Bytes()
}

// EnumEntityType defines an enumerated entity type in a Cedar schema.
type EnumEntityType struct {
	Name        types.Ident
	Annotations Annotations
	Values      []types.String
}

// EnumEntity creates a new enumerated entity type with the given name and values.
func EnumEntity(name types.Ident, values ...types.String) *EnumEntityType {
	return &EnumEntityType{Name: name, Values: values}
}

// Annotate adds an annotation to the enumerated entity type.
func (e *EnumEntityType) Annotate(key types.Ident, value types.String) *EnumEntityType {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// AddValue adds a value to the enumeration.
func (e *EnumEntityType) AddValue(value types.String) *EnumEntityType {
	e.Values = append(e.Values, value)
	return e
}

// MarshalCedar returns the Cedar format representation of the enum entity type.
func (e *EnumEntityType) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write annotations
	for _, ann := range e.Annotations {
		buf.Write(ann.MarshalCedar())
		buf.WriteByte('\n')
	}

	// Write entity declaration
	buf.WriteString("entity ")
	buf.WriteString(string(e.Name))

	// Write enum values
	buf.WriteString(" enum [")
	for i, v := range e.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.Write(marshalString(v))
	}
	buf.WriteByte(']')

	buf.WriteByte(';')
	return buf.Bytes()
}
