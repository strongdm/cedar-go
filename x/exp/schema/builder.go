package schema

func Long() TypeExpr      { return PrimitiveTypeExpr{Kind: PrimitiveLong} }
func String() TypeExpr    { return PrimitiveTypeExpr{Kind: PrimitiveString} }
func Bool() TypeExpr      { return PrimitiveTypeExpr{Kind: PrimitiveBool} }
func Set(element TypeExpr) TypeExpr { return SetTypeExpr{Element: element} }

func Record(attrs map[string]*Attribute) TypeExpr {
	if attrs == nil {
		attrs = make(map[string]*Attribute)
	}
	return &RecordTypeExpr{Attributes: attrs}
}

func Entity(name string) TypeExpr    { return EntityRefExpr{Name: name} }
func Extension(name string) TypeExpr { return ExtensionTypeExpr{Name: name} }
func IPAddr() TypeExpr               { return ExtensionTypeExpr{Name: "ipaddr"} }
func Decimal() TypeExpr              { return ExtensionTypeExpr{Name: "decimal"} }
func Datetime() TypeExpr             { return ExtensionTypeExpr{Name: "datetime"} }
func Duration() TypeExpr             { return ExtensionTypeExpr{Name: "duration"} }
func NamedType(name string) TypeExpr { return TypeNameExpr{Name: name} }

type SchemaBuilder struct {
	schema *Schema
}

func NewBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		schema: &Schema{Namespaces: make(map[string]*Namespace)},
	}
}

func (b *SchemaBuilder) Namespace(name string) *NamespaceBuilder {
	ns := newNamespace()
	b.schema.Namespaces[name] = ns
	return &NamespaceBuilder{parent: b, namespace: ns}
}

func (b *SchemaBuilder) Build() *Schema { return b.schema }

type NamespaceBuilder struct {
	parent    *SchemaBuilder
	namespace *Namespace
}

func (b *NamespaceBuilder) Entity(name string) *EntityBuilder {
	e := &EntityTypeDef{Annotations: newAnnotations()}
	b.namespace.EntityTypes[name] = e
	return &EntityBuilder{parent: b, entity: e}
}

func (b *NamespaceBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	b.namespace.EnumTypes[name] = &EnumTypeDef{
		Values:      values,
		Annotations: newAnnotations(),
	}
	return b
}

func (b *NamespaceBuilder) Action(name string) *ActionBuilder {
	a := &ActionDef{Annotations: newAnnotations()}
	b.namespace.Actions[name] = a
	return &ActionBuilder{parent: b, action: a}
}

func (b *NamespaceBuilder) CommonType(name string, typ TypeExpr) *NamespaceBuilder {
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

func (b *NamespaceBuilder) Namespace(name string) *NamespaceBuilder {
	return b.parent.Namespace(name)
}

func (b *NamespaceBuilder) Build() *Schema { return b.parent.Build() }

type EntityBuilder struct {
	parent *NamespaceBuilder
	entity *EntityTypeDef
}

func (b *EntityBuilder) MemberOf(parentTypes ...string) *EntityBuilder {
	b.entity.MemberOfTypes = append(b.entity.MemberOfTypes, parentTypes...)
	return b
}

func (b *EntityBuilder) ensureShape() *RecordTypeExpr {
	if b.entity.Shape == nil {
		b.entity.Shape = &RecordTypeExpr{Attributes: make(map[string]*Attribute)}
	}
	return b.entity.Shape
}

func (b *EntityBuilder) Attr(name string, typ TypeExpr) *EntityBuilder {
	b.ensureShape().Attributes[name] = &Attribute{Type: typ, Required: true, Annotations: newAnnotations()}
	return b
}

func (b *EntityBuilder) OptionalAttr(name string, typ TypeExpr) *EntityBuilder {
	b.ensureShape().Attributes[name] = &Attribute{Type: typ, Required: false, Annotations: newAnnotations()}
	return b
}

func (b *EntityBuilder) Tags(typ TypeExpr) *EntityBuilder {
	b.entity.Tags = typ
	return b
}

func (b *EntityBuilder) Annotate(key, value string) *EntityBuilder {
	b.entity.Annotations[key] = value
	return b
}

func (b *EntityBuilder) Entity(name string) *EntityBuilder {
	return b.parent.Entity(name)
}

func (b *EntityBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	return b.parent.EnumType(name, values...)
}

func (b *EntityBuilder) Action(name string) *ActionBuilder {
	return b.parent.Action(name)
}

func (b *EntityBuilder) CommonType(name string, typ TypeExpr) *NamespaceBuilder {
	return b.parent.CommonType(name, typ)
}

func (b *EntityBuilder) Namespace(name string) *NamespaceBuilder {
	return b.parent.Namespace(name)
}

func (b *EntityBuilder) Build() *Schema { return b.parent.Build() }

type ActionBuilder struct {
	parent *NamespaceBuilder
	action *ActionDef
}

func (b *ActionBuilder) InGroup(refs ...*ActionRef) *ActionBuilder {
	b.action.MemberOf = append(b.action.MemberOf, refs...)
	return b
}

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

func (b *ActionBuilder) Context(ctx TypeExpr) *ActionBuilder {
	b.ensureAppliesTo().Context = ctx
	return b
}

func (b *ActionBuilder) Annotate(key, value string) *ActionBuilder {
	b.action.Annotations[key] = value
	return b
}

func (b *ActionBuilder) Entity(name string) *EntityBuilder {
	return b.parent.Entity(name)
}

func (b *ActionBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	return b.parent.EnumType(name, values...)
}

func (b *ActionBuilder) Action(name string) *ActionBuilder {
	return b.parent.Action(name)
}

func (b *ActionBuilder) CommonType(name string, typ TypeExpr) *NamespaceBuilder {
	return b.parent.CommonType(name, typ)
}

func (b *ActionBuilder) Namespace(name string) *NamespaceBuilder {
	return b.parent.Namespace(name)
}

func (b *ActionBuilder) Build() *Schema { return b.parent.Build() }
