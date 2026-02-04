package schema

func addAttribute(rec *RecordType, name string, typ Type, required bool) {
	rec.Attributes[name] = &Attribute{
		Type:        typ,
		Required:    required,
		Annotations: newAnnotations(),
	}
}

type SchemaBuilder struct {
	schema *Schema
}

// Example:
//
//	s := schema.NewBuilder().
//		Namespace("MyApp").
//			Entity("User").MemberOf("Group").Done().
//			Entity("Group").Done().
//			Action("view").Principal("User").Resource("Document").Done().
//		Done().
//		Build()
func NewBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{
			Namespaces: make(map[string]*Namespace),
		},
	}
}

// Use "" for the empty namespace.
func (b *SchemaBuilder) Namespace(name string) *NamespaceBuilder {
	ns := newNamespace()
	b.schema.Namespaces[name] = ns
	return &NamespaceBuilder{
		parent:    b,
		namespace: ns,
	}
}

func (b *SchemaBuilder) Build() *Schema {
	return b.schema
}

type NamespaceBuilder struct {
	parent    *SchemaBuilder
	namespace *Namespace
}

func (b *NamespaceBuilder) Entity(name string) *EntityBuilder {
	e := &EntityTypeDef{
		Annotations: newAnnotations(),
	}
	b.namespace.EntityTypes[name] = e
	return &EntityBuilder{
		parent: b,
		entity: e,
	}
}

func (b *NamespaceBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	b.namespace.EnumTypes[name] = &EnumTypeDef{
		Values:      values,
		Annotations: newAnnotations(),
	}
	return b
}

func (b *NamespaceBuilder) Action(name string) *ActionBuilder {
	a := &ActionDef{
		Annotations: newAnnotations(),
	}
	b.namespace.Actions[name] = a
	return &ActionBuilder{
		parent: b,
		action: a,
	}
}

func (b *NamespaceBuilder) CommonType(name string, typ Type) *NamespaceBuilder {
	b.namespace.CommonTypes[name] = &CommonTypeDef{
		Type:        typ,
		Annotations: newAnnotations(),
	}
	return b
}

func (b *NamespaceBuilder) Annotate(key, value string) *NamespaceBuilder {
	b.namespace.Annotations[key] = value
	return b
}

func (b *NamespaceBuilder) Done() *SchemaBuilder {
	return b.parent
}

type EntityBuilder struct {
	parent *NamespaceBuilder
	entity *EntityTypeDef
}

func (b *EntityBuilder) MemberOf(parentTypes ...string) *EntityBuilder {
	b.entity.MemberOfTypes = append(b.entity.MemberOfTypes, parentTypes...)
	return b
}

func (b *EntityBuilder) ensureShape() *RecordType {
	if b.entity.Shape == nil {
		b.entity.Shape = &RecordType{
			Attributes: make(map[string]*Attribute),
		}
	}
	return b.entity.Shape
}

func (b *EntityBuilder) Attr(name string, typ Type) *EntityBuilder {
	addAttribute(b.ensureShape(), name, typ, true)
	return b
}

func (b *EntityBuilder) OptionalAttr(name string, typ Type) *EntityBuilder {
	addAttribute(b.ensureShape(), name, typ, false)
	return b
}

func (b *EntityBuilder) Tags(typ Type) *EntityBuilder {
	b.entity.Tags = typ
	return b
}

func (b *EntityBuilder) Annotate(key, value string) *EntityBuilder {
	b.entity.Annotations[key] = value
	return b
}

func (b *EntityBuilder) Done() *NamespaceBuilder {
	return b.parent
}

type ActionBuilder struct {
	parent *NamespaceBuilder
	action *ActionDef
}

func (b *ActionBuilder) InGroup(refs ...*ActionRef) *ActionBuilder {
	b.action.MemberOf = append(b.action.MemberOf, refs...)
	return b
}

// InGroupByName assumes the action group is in the same namespace.
func (b *ActionBuilder) InGroupByName(names ...string) *ActionBuilder {
	for _, name := range names {
		b.action.MemberOf = append(b.action.MemberOf, &ActionRef{ID: name})
	}
	return b
}

func (b *ActionBuilder) ensureAppliesTo() *AppliesTo {
	if b.action.AppliesTo == nil {
		b.action.AppliesTo = &AppliesTo{}
	}
	return b.action.AppliesTo
}

func (b *ActionBuilder) Principal(types ...string) *ActionBuilder {
	b.ensureAppliesTo().PrincipalTypes = append(b.ensureAppliesTo().PrincipalTypes, types...)
	return b
}

func (b *ActionBuilder) Resource(types ...string) *ActionBuilder {
	b.ensureAppliesTo().ResourceTypes = append(b.ensureAppliesTo().ResourceTypes, types...)
	return b
}

func (b *ActionBuilder) Context(ctx *RecordType) *ActionBuilder {
	b.ensureAppliesTo().Context = ctx
	return b
}

func (b *ActionBuilder) ContextRef(typ Type) *ActionBuilder {
	b.ensureAppliesTo().ContextRef = typ
	return b
}

func (b *ActionBuilder) Annotate(key, value string) *ActionBuilder {
	b.action.Annotations[key] = value
	return b
}

func (b *ActionBuilder) Done() *NamespaceBuilder {
	return b.parent
}

func Long() Type {
	return PrimitiveType{Kind: PrimitiveLong}
}

func String() Type {
	return PrimitiveType{Kind: PrimitiveString}
}

func Bool() Type {
	return PrimitiveType{Kind: PrimitiveBool}
}

func Set(element Type) Type {
	return SetType{Element: element}
}

func NewRecordType(attrs map[string]*Attribute) *RecordType {
	if attrs == nil {
		attrs = make(map[string]*Attribute)
	}
	return &RecordType{Attributes: attrs}
}

func Entity(name string) Type {
	return EntityRef{Name: name}
}

func Extension(name string) Type {
	return ExtensionType{Name: name}
}

func IPAddr() Type {
	return ExtensionType{Name: "ipaddr"}
}

func Decimal() Type {
	return ExtensionType{Name: "decimal"}
}

func Datetime() Type {
	return ExtensionType{Name: "datetime"}
}

func Duration() Type {
	return ExtensionType{Name: "duration"}
}

func CommonType(name string) Type {
	return CommonTypeRef{Name: name}
}
