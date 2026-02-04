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

// Example (explicit Done() calls):
//
//	s := schema.NewBuilder().
//		Namespace("MyApp").
//			Entity("User").MemberOf("Group").Done().
//			Entity("Group").Done().
//			Action("view").Principal("User").Resource("Document").Done().
//		Done().
//		Build()
//
// Example (chained without explicit Done() calls):
//
//	s := schema.NewBuilder().
//		Namespace("MyApp").
//			Entity("User").MemberOf("Group").
//			Entity("Group").
//			Action("view").Principal("User").Resource("Document").
//		Build()
//
// Example (multiple namespaces without Done() calls):
//
//	s := schema.NewBuilder().
//		Namespace("MyApp").
//			Entity("User").
//			Entity("Group").
//		Namespace("OtherApp").
//			Entity("Foo").
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

// Namespace completes the current namespace and starts a new namespace.
func (b *NamespaceBuilder) Namespace(name string) *NamespaceBuilder {
	return b.Done().Namespace(name)
}

// Build completes the current namespace and builds the schema.
func (b *NamespaceBuilder) Build() *Schema {
	return b.Done().Build()
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

// Entity completes the current entity and starts a new entity in the same namespace.
func (b *EntityBuilder) Entity(name string) *EntityBuilder {
	return b.Done().Entity(name)
}

// EnumType completes the current entity and adds an enum type to the namespace.
func (b *EntityBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	return b.Done().EnumType(name, values...)
}

// Action completes the current entity and starts a new action in the same namespace.
func (b *EntityBuilder) Action(name string) *ActionBuilder {
	return b.Done().Action(name)
}

// CommonType completes the current entity and adds a common type to the namespace.
func (b *EntityBuilder) CommonType(name string, typ Type) *NamespaceBuilder {
	return b.Done().CommonType(name, typ)
}

// Namespace completes the current entity and namespace, then starts a new namespace.
func (b *EntityBuilder) Namespace(name string) *NamespaceBuilder {
	return b.Done().Done().Namespace(name)
}

// Build completes the current entity and namespace, then builds the schema.
func (b *EntityBuilder) Build() *Schema {
	return b.Done().Done().Build()
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

// Entity completes the current action and starts a new entity in the same namespace.
func (b *ActionBuilder) Entity(name string) *EntityBuilder {
	return b.Done().Entity(name)
}

// EnumType completes the current action and adds an enum type to the namespace.
func (b *ActionBuilder) EnumType(name string, values ...string) *NamespaceBuilder {
	return b.Done().EnumType(name, values...)
}

// Action completes the current action and starts a new action in the same namespace.
func (b *ActionBuilder) Action(name string) *ActionBuilder {
	return b.Done().Action(name)
}

// CommonType completes the current action and adds a common type to the namespace.
func (b *ActionBuilder) CommonType(name string, typ Type) *NamespaceBuilder {
	return b.Done().CommonType(name, typ)
}

// Namespace completes the current action and namespace, then starts a new namespace.
func (b *ActionBuilder) Namespace(name string) *NamespaceBuilder {
	return b.Done().Done().Namespace(name)
}

// Build completes the current action and namespace, then builds the schema.
func (b *ActionBuilder) Build() *Schema {
	return b.Done().Done().Build()
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
