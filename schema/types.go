package schema

// Type represents a Cedar type in a schema.
// Types can be primitive (String, Long, Bool), complex (Record, Set),
// or references to entity/common types (Path).
type Type interface {
	isType()
}

// Attribute represents a field in a record type.
type Attribute struct {
	name        string
	attrType    Type
	isRequired  bool
	annotations map[string]string
}

// RecordType represents a structured record with named attributes.
type RecordType struct {
	attributes  map[string]*Attribute
	annotations map[string]string
}

// SetType represents a set of elements of a given type.
type SetType struct {
	element     Type
	annotations map[string]string
}

// PathType represents a reference to an entity type or common type.
// The path is namespace-qualified (e.g., "PhotoApp::User" or just "String").
type PathType struct {
	path        string
	annotations map[string]string
}

// Ensure types implement the Type interface
func (*RecordType) isType() { _ = 0 }
func (*SetType) isType()    { _ = 0 }
func (*PathType) isType()   { _ = 0 }

// Record creates a new record type with the given attributes.
func Record(attrs ...*Attribute) *RecordType {
	record := &RecordType{
		attributes: make(map[string]*Attribute),
	}
	for _, attr := range attrs {
		record.attributes[attr.name] = attr
	}
	return record
}

// WithAttribute adds an attribute to the record type.
func (r *RecordType) WithAttribute(name string, typ Type) *RecordType {
	r.attributes[name] = &Attribute{
		name:       name,
		attrType:   typ,
		isRequired: true,
	}
	return r
}

// WithOptionalAttribute adds an optional attribute to the record type.
func (r *RecordType) WithOptionalAttribute(name string, typ Type) *RecordType {
	r.attributes[name] = &Attribute{
		name:       name,
		attrType:   typ,
		isRequired: false,
	}
	return r
}

// WithAnnotation adds an annotation to the record type.
func (r *RecordType) WithAnnotation(key, value string) *RecordType {
	if r.annotations == nil {
		r.annotations = make(map[string]string)
	}
	r.annotations[key] = value
	return r
}

// Attr creates a required attribute with the given name and type.
func Attr(name string, typ Type) *Attribute {
	return &Attribute{
		name:       name,
		attrType:   typ,
		isRequired: true,
	}
}

// OptionalAttr creates an optional attribute with the given name and type.
func OptionalAttr(name string, typ Type) *Attribute {
	return &Attribute{
		name:       name,
		attrType:   typ,
		isRequired: false,
	}
}

// WithAnnotation adds an annotation to the attribute.
func (a *Attribute) WithAnnotation(key, value string) *Attribute {
	if a.annotations == nil {
		a.annotations = make(map[string]string)
	}
	a.annotations[key] = value
	return a
}

// Set creates a set type with the given element type.
func Set(element Type) *SetType {
	return &SetType{
		element: element,
	}
}

// WithAnnotation adds an annotation to the set type.
func (s *SetType) WithAnnotation(key, value string) *SetType {
	if s.annotations == nil {
		s.annotations = make(map[string]string)
	}
	s.annotations[key] = value
	return s
}

// EntityType creates a reference to an entity type.
// The path should be fully-qualified (e.g., "PhotoApp::User").
func EntityType(path string) *PathType {
	return &PathType{
		path: path,
	}
}

// CommonType creates a reference to a common type.
// The path should be the type name (e.g., "MyCommonType").
func CommonType(path string) *PathType {
	return &PathType{
		path: path,
	}
}

// String returns the Cedar String primitive type.
func String() *PathType {
	return &PathType{
		path: "String",
	}
}

// Long returns the Cedar Long primitive type.
func Long() *PathType {
	return &PathType{
		path: "Long",
	}
}

// Bool returns the Cedar Bool primitive type.
func Bool() *PathType {
	return &PathType{
		path: "Bool",
	}
}

// Boolean returns the Cedar Bool primitive type (alias for Bool).
func Boolean() *PathType {
	return Bool()
}

// WithAnnotation adds an annotation to the path type.
func (p *PathType) WithAnnotation(key, value string) *PathType {
	if p.annotations == nil {
		p.annotations = make(map[string]string)
	}
	p.annotations[key] = value
	return p
}

// CommonTypeDecl creates a common type declaration that can be added to a namespace.
type CommonTypeDecl struct {
	name        string
	typ         Type
	annotations map[string]string
}

// TypeDecl creates a new common type declaration.
func TypeDecl(name string, typ Type) *CommonTypeDecl {
	return &CommonTypeDecl{
		name: name,
		typ:  typ,
	}
}

// WithAnnotation adds an annotation to the common type declaration.
func (c *CommonTypeDecl) WithAnnotation(key, value string) *CommonTypeDecl {
	if c.annotations == nil {
		c.annotations = make(map[string]string)
	}
	c.annotations[key] = value
	return c
}

// addToNamespace implements the Declaration interface.
func (c *CommonTypeDecl) addToNamespace(ns *Namespace) {
	if ns.commonTypes == nil {
		ns.commonTypes = make(map[string]Type)
	}
	ns.commonTypes[c.name] = c.typ
}
