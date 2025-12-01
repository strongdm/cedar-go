package schema

// Entity represents a Cedar entity type definition.
// Entities can have attributes (shape), tags, hierarchical relationships (memberOf),
// or be defined as enums.
type Entity struct {
	name        string
	namespace   *Namespace
	memberOf    []string
	shape       Type
	tags        Type
	enum        []string
	annotations map[string]string
}

// NewEntity creates a new entity type with the given name.
// The name should be a simple identifier (e.g., "User", "Photo").
func NewEntity(name string) *Entity {
	return &Entity{
		name: name,
	}
}

// WithAttribute adds an attribute to the entity's shape.
// If the entity doesn't have a shape yet, a new RecordType is created.
func (e *Entity) WithAttribute(name string, typ Type) *Entity {
	if e.shape == nil {
		e.shape = &RecordType{
			attributes: make(map[string]*Attribute),
		}
	}

	if record, ok := e.shape.(*RecordType); ok {
		record.attributes[name] = &Attribute{
			name:       name,
			attrType:   typ,
			isRequired: true, // default to required
		}
	}

	return e
}

// WithOptionalAttribute adds an optional attribute to the entity's shape.
func (e *Entity) WithOptionalAttribute(name string, typ Type) *Entity {
	if e.shape == nil {
		e.shape = &RecordType{
			attributes: make(map[string]*Attribute),
		}
	}

	if record, ok := e.shape.(*RecordType); ok {
		record.attributes[name] = &Attribute{
			name:       name,
			attrType:   typ,
			isRequired: false,
		}
	}

	return e
}

// WithShape sets the complete shape (record type) for the entity.
func (e *Entity) WithShape(shape Type) *Entity {
	e.shape = shape
	return e
}

// WithTags sets the tag type for the entity.
func (e *Entity) WithTags(tagType Type) *Entity {
	e.tags = tagType
	return e
}

// MemberOf specifies that this entity type can be a member of the given parent types.
// The parent types should be fully-qualified (e.g., "PhotoApp::UserGroup").
func (e *Entity) MemberOf(parentTypes ...string) *Entity {
	e.memberOf = append(e.memberOf, parentTypes...)
	return e
}

// AsEnum defines this entity as an enumeration with the given values.
func (e *Entity) AsEnum(values ...string) *Entity {
	e.enum = values
	return e
}

// WithAnnotation adds an annotation to the entity.
func (e *Entity) WithAnnotation(key, value string) *Entity {
	if e.annotations == nil {
		e.annotations = make(map[string]string)
	}
	e.annotations[key] = value
	return e
}

// addToNamespace implements the Declaration interface.
func (e *Entity) addToNamespace(ns *Namespace) {
	e.namespace = ns
	if ns.entities == nil {
		ns.entities = make(map[string]*Entity)
	}
	ns.entities[e.name] = e
}
