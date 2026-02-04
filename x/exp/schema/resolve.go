package schema

import (
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Resolve converts a Schema with unresolved type names into a resolved.Schema
// with fully-qualified types.EntityType and types.EntityUID values.
// Returns an error if the schema contains undefined types, cycles, or shadowing violations.
func (s *Schema) Resolve() (*resolved.Schema, error) {
	if s == nil {
		return nil, fmt.Errorf("schema: cannot resolve nil Schema")
	}
	r := &resolver{
		schema:      s,
		entityTypes: make(map[string]bool),
		commonTypes: make(map[string]bool),
		visiting:    make(map[string]bool),
	}

	// Build the definition index
	if err := r.buildIndex(); err != nil {
		return nil, err
	}

	// Check for shadowing violations
	if err := r.checkShadowing(); err != nil {
		return nil, err
	}

	// Resolve all type references
	return r.resolve()
}

type resolver struct {
	schema      *Schema
	entityTypes map[string]bool // fully-qualified entity type names
	commonTypes map[string]bool // fully-qualified common type names
	visiting    map[string]bool // for cycle detection
}

// buildIndex populates the entityTypes and commonTypes maps with all defined types.
// Note: Built-in types (Long, String, Bool, ipaddr, decimal, datetime, duration)
// are resolved dynamically by isBuiltinTypeName/resolveBuiltinType rather than
// being pre-populated in the commonTypes map. This avoids conflicts with
// user-defined types of the same name.
func (r *resolver) buildIndex() error {
	// Add all defined types from the schema
	for nsName, ns := range r.schema.Namespaces {
		// Check for reserved namespace names
		if isReservedNamespace(nsName) {
			return &ReservedNameError{Name: nsName, Kind: "namespace"}
		}

		for etName := range ns.EntityTypes {
			// Check for __cedar reserved prefix
			if hasCedarReservedPrefix(etName) {
				return &ReservedNameError{Name: etName, Kind: "entity type"}
			}
			fqn := qualifyName(nsName, etName)
			if r.entityTypes[fqn] {
				return &DuplicateError{Kind: "entity type", Name: etName, Namespace: nsName}
			}
			r.entityTypes[fqn] = true
		}

		// Enum types are also entity types for resolution purposes
		for enumName := range ns.EnumTypes {
			// Check for __cedar reserved prefix
			if hasCedarReservedPrefix(enumName) {
				return &ReservedNameError{Name: enumName, Kind: "enum type"}
			}
			fqn := qualifyName(nsName, enumName)
			if r.entityTypes[fqn] {
				return &DuplicateError{Kind: "enum type", Name: enumName, Namespace: nsName}
			}
			r.entityTypes[fqn] = true
		}

		for ctName := range ns.CommonTypes {
			// Check for __cedar reserved prefix
			if hasCedarReservedPrefix(ctName) {
				return &ReservedNameError{Name: ctName, Kind: "common type"}
			}
			fqn := qualifyName(nsName, ctName)
			if r.commonTypes[fqn] {
				return &DuplicateError{Kind: "common type", Name: ctName, Namespace: nsName}
			}
			r.commonTypes[fqn] = true
		}
	}

	return nil
}

// isReservedNamespace checks if a namespace name is reserved.
func isReservedNamespace(name string) bool {
	return name == "__cedar" || strings.HasPrefix(name, "__cedar::")
}

// hasCedarReservedPrefix checks if a type name uses the reserved "__cedar" prefix.
func hasCedarReservedPrefix(name string) bool {
	return name == "__cedar" || strings.HasPrefix(name, "__cedar::")
}

// builtinExtensionTypes maps extension type names to themselves for quick lookup.
// These are the built-in extension types available in Cedar.
var builtinExtensionTypes = map[string]bool{
	"ipaddr":   true,
	"decimal":  true,
	"datetime": true,
	"duration": true,
}

// isBuiltinTypeName checks if a name (qualified or unqualified) is a built-in type.
// Returns true if the name is a built-in primitive (Long, String, Bool) or
// extension type (ipaddr, decimal, datetime, duration).
// This function supports both qualified (__cedar::Long) and unqualified (Long) names.
// The resolveName function falls back to this helper when a name is not found
// in the user-defined type maps.
func isBuiltinTypeName(name string) bool {
	// Strip __cedar:: prefix if present
	baseName := name
	if strings.HasPrefix(name, "__cedar::") {
		baseName = strings.TrimPrefix(name, "__cedar::")
	}

	// Check primitives
	switch baseName {
	case "Long", "String", "Bool":
		return true
	}

	// Check extension types
	return builtinExtensionTypes[baseName]
}

// resolveBuiltinType resolves a built-in type name to its resolved.Type.
// Returns nil if the name is not a built-in type.
func resolveBuiltinType(name string) resolved.Type {
	// Strip __cedar:: prefix if present
	baseName := name
	if strings.HasPrefix(name, "__cedar::") {
		baseName = strings.TrimPrefix(name, "__cedar::")
	}

	// Check primitives
	switch baseName {
	case "Long":
		return resolved.Primitive{Kind: resolved.PrimitiveLong}
	case "String":
		return resolved.Primitive{Kind: resolved.PrimitiveString}
	case "Bool":
		return resolved.Primitive{Kind: resolved.PrimitiveBool}
	}

	// Check extension types
	if builtinExtensionTypes[baseName] {
		return resolved.Extension{Name: baseName}
	}

	return nil
}

// checkShadowing verifies that namespaced definitions don't illegally shadow empty namespace definitions.
// Per Cedar spec: any type in a named namespace that has the same name as any type in the empty
// namespace is a shadowing violation, regardless of whether they are entity, enum, or common types.
func (r *resolver) checkShadowing() error {
	emptyNs := r.schema.Namespaces[""]
	if emptyNs == nil {
		return nil
	}

	// Build a set of all names in empty namespace
	emptyNames := make(map[string]bool)
	for name := range emptyNs.EntityTypes {
		emptyNames[name] = true
	}
	for name := range emptyNs.EnumTypes {
		emptyNames[name] = true
	}
	for name := range emptyNs.CommonTypes {
		emptyNames[name] = true
	}

	for nsName, ns := range r.schema.Namespaces {
		if nsName == "" {
			continue
		}

		// Check entity types in NS against ALL names in empty namespace
		for name := range ns.EntityTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}

		// Check enum types in NS against ALL names in empty namespace
		for name := range ns.EnumTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}

		// Check common types in NS against ALL names in empty namespace
		for name := range ns.CommonTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}
	}

	return nil
}

// resolve converts all type references to fully-qualified names.
func (r *resolver) resolve() (*resolved.Schema, error) {
	rs := &resolved.Schema{
		Namespaces: make(map[types.Path]*resolved.Namespace),
	}

	for nsName, ns := range r.schema.Namespaces {
		rns, err := r.resolveNamespace(nsName, ns)
		if err != nil {
			return nil, err
		}
		rs.Namespaces[types.Path(nsName)] = rns
	}

	return rs, nil
}

func (r *resolver) resolveNamespace(nsName string, ns *Namespace) (*resolved.Namespace, error) {
	rns := &resolved.Namespace{
		EntityTypes: make(map[types.EntityType]*resolved.EntityType),
		EnumTypes:   make(map[types.EntityType]*resolved.EnumType),
		Actions:     make(map[types.EntityUID]*resolved.Action),
		CommonTypes: make(map[types.Path]*resolved.Type),
		Annotations: resolved.Annotations(ns.Annotations),
	}

	// Resolve entity types
	for etName, et := range ns.EntityTypes {
		fqn := types.EntityType(qualifyName(nsName, etName))
		ret, err := r.resolveEntityType(nsName, et)
		if err != nil {
			return nil, fmt.Errorf("entity type %s: %w", fqn, err)
		}
		rns.EntityTypes[fqn] = ret
	}

	// Resolve enum types
	for enumName, enum := range ns.EnumTypes {
		fqn := types.EntityType(qualifyName(nsName, enumName))
		rns.EnumTypes[fqn] = &resolved.EnumType{
			Values:      enum.Values,
			Annotations: resolved.Annotations(enum.Annotations),
		}
	}

	// Resolve actions
	for actName, act := range ns.Actions {
		// Actions are EntityUIDs with type "Action" (qualified by namespace)
		actionType := types.EntityType(qualifyName(nsName, "Action"))
		uid := types.NewEntityUID(actionType, types.String(actName))
		ract, err := r.resolveAction(nsName, act)
		if err != nil {
			return nil, fmt.Errorf("action %s: %w", actName, err)
		}
		rns.Actions[uid] = ract
	}

	// Resolve common types (inline them)
	for ctName, ct := range ns.CommonTypes {
		fqn := types.Path(qualifyName(nsName, ctName))
		rct, err := r.resolveType(nsName, ct.Type)
		if err != nil {
			return nil, fmt.Errorf("common type %s: %w", fqn, err)
		}
		rns.CommonTypes[fqn] = &rct
	}

	return rns, nil
}

func (r *resolver) resolveEntityType(nsName string, et *EntityTypeDef) (*resolved.EntityType, error) {
	ret := &resolved.EntityType{
		Annotations: resolved.Annotations(et.Annotations),
	}

	// Resolve parent types
	for _, parent := range et.MemberOfTypes {
		resolved, err := r.resolveEntityTypeName(nsName, parent)
		if err != nil {
			return nil, fmt.Errorf("memberOf %q: %w", parent, err)
		}
		ret.MemberOfTypes = append(ret.MemberOfTypes, resolved)
	}

	// Resolve shape
	if et.Shape != nil {
		shape, err := r.resolveRecordType(nsName, et.Shape)
		if err != nil {
			return nil, fmt.Errorf("shape: %w", err)
		}
		ret.Shape = shape
	}

	// Resolve tags
	if et.Tags != nil {
		tags, err := r.resolveType(nsName, et.Tags)
		if err != nil {
			return nil, fmt.Errorf("tags: %w", err)
		}
		ret.Tags = tags
	}

	return ret, nil
}

func (r *resolver) resolveAction(nsName string, act *ActionDef) (*resolved.Action, error) {
	ract := &resolved.Action{
		Annotations: resolved.Annotations(act.Annotations),
	}

	// Resolve memberOf action refs
	for _, ref := range act.MemberOf {
		resolved, err := r.resolveActionRef(nsName, ref)
		if err != nil {
			return nil, fmt.Errorf("memberOf: %w", err)
		}
		ract.MemberOf = append(ract.MemberOf, resolved)
	}

	// Resolve appliesTo
	if act.AppliesTo != nil {
		// Resolve principal types
		for _, pt := range act.AppliesTo.PrincipalTypes {
			resolved, err := r.resolveEntityTypeName(nsName, pt)
			if err != nil {
				return nil, fmt.Errorf("principal type %q: %w", pt, err)
			}
			ract.PrincipalTypes = append(ract.PrincipalTypes, resolved)
		}

		// Resolve resource types
		for _, rt := range act.AppliesTo.ResourceTypes {
			resolved, err := r.resolveEntityTypeName(nsName, rt)
			if err != nil {
				return nil, fmt.Errorf("resource type %q: %w", rt, err)
			}
			ract.ResourceTypes = append(ract.ResourceTypes, resolved)
		}

		// Resolve context - either inline Context or ContextRef (common type reference)
		if act.AppliesTo.ContextRef != nil {
			// This is a reference to a common type, resolve and inline it
			ctx, err := r.resolveType(nsName, act.AppliesTo.ContextRef)
			if err != nil {
				return nil, fmt.Errorf("context: %w", err)
			}
			if rrt, ok := ctx.(*resolved.RecordType); ok {
				ract.Context = rrt
			} else {
				return nil, fmt.Errorf("context type must resolve to a record")
			}
		} else if act.AppliesTo.Context != nil {
			ctx, err := r.resolveRecordType(nsName, act.AppliesTo.Context)
			if err != nil {
				return nil, fmt.Errorf("context: %w", err)
			}
			ract.Context = ctx
		}
	}

	return ract, nil
}

func (r *resolver) resolveActionRef(nsName string, ref *ActionRef) (types.EntityUID, error) {
	var actionTypeName string

	if ref.Type != "" {
		// Fully qualified action type
		actionTypeName = ref.Type
	} else {
		// Action in current namespace
		actionTypeName = qualifyName(nsName, "Action")
	}

	return types.NewEntityUID(types.EntityType(actionTypeName), types.String(ref.ID)), nil
}

func (r *resolver) resolveType(nsName string, t Type) (resolved.Type, error) {
	switch v := t.(type) {
	case PrimitiveType:
		return resolved.Primitive{Kind: resolved.PrimitiveKind(v.Kind)}, nil

	case SetType:
		elem, err := r.resolveType(nsName, v.Element)
		if err != nil {
			return nil, err
		}
		return resolved.Set{Element: elem}, nil

	case *RecordType:
		return r.resolveRecordType(nsName, v)

	case EntityRef:
		et, err := r.resolveEntityTypeName(nsName, v.Name)
		if err != nil {
			return nil, err
		}
		return resolved.EntityRef{EntityType: et}, nil

	case ExtensionType:
		return resolved.Extension{Name: v.Name}, nil

	case CommonTypeRef:
		return r.resolveCommonTypeRef(nsName, v.Name)

	case EntityOrCommonRef:
		return r.resolveEntityOrCommon(nsName, v.Name)

	default:
		return nil, fmt.Errorf("unknown type: %T", t)
	}
}

func (r *resolver) resolveRecordType(nsName string, rt *RecordType) (*resolved.RecordType, error) {
	rrt := &resolved.RecordType{
		Attributes: make(map[string]*resolved.Attribute),
	}

	for name, attr := range rt.Attributes {
		rtype, err := r.resolveType(nsName, attr.Type)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		rrt.Attributes[name] = &resolved.Attribute{
			Type:        rtype,
			Required:    attr.Required,
			Annotations: resolved.Annotations(attr.Annotations),
		}
	}

	return rrt, nil
}

// resolveEntityTypeName resolves a possibly-unqualified entity type name to a fully-qualified types.EntityType.
func (r *resolver) resolveEntityTypeName(nsName string, name string) (types.EntityType, error) {
	resolved, err := r.resolveName(nsName, name, true)
	if err != nil {
		return "", err
	}
	return types.EntityType(resolved), nil
}

// resolveCommonTypeRef resolves a common type reference and inlines it.
func (r *resolver) resolveCommonTypeRef(nsName string, name string) (resolved.Type, error) {
	fqn, err := r.resolveName(nsName, name, false)
	if err != nil {
		return nil, err
	}

	// Check for cycle
	if r.visiting[fqn] {
		return nil, &CycleError{Path: []string{fqn}}
	}
	r.visiting[fqn] = true
	defer delete(r.visiting, fqn)

	// Find and inline the common type definition
	parts := strings.SplitN(fqn, "::", 2)
	var ctNsName, ctName string
	if len(parts) == 2 {
		ctNsName = parts[0]
		ctName = parts[1]
		// Handle multi-part namespaces
		for {
			if ns, ok := r.schema.Namespaces[ctNsName]; ok {
				if ct, ok := ns.CommonTypes[ctName]; ok {
					return r.resolveType(ctNsName, ct.Type)
				}
			}
			// Try adding more parts to namespace
			idx := strings.Index(ctName, "::")
			if idx == -1 {
				break
			}
			ctNsName = ctNsName + "::" + ctName[:idx]
			ctName = ctName[idx+2:]
		}
	} else {
		ctNsName = ""
		ctName = fqn
	}

	// Try direct lookup
	if ns, ok := r.schema.Namespaces[ctNsName]; ok {
		if ct, ok := ns.CommonTypes[ctName]; ok {
			return r.resolveType(ctNsName, ct.Type)
		}
	}

	// Check for built-in types
	if t := resolveBuiltinType(fqn); t != nil {
		return t, nil
	}

	return nil, &UndefinedTypeError{Name: name, Namespace: nsName}
}

// resolveEntityOrCommon resolves an ambiguous reference that could be entity or common type.
// Priority: common type > entity type > primitive/extension
func (r *resolver) resolveEntityOrCommon(nsName string, name string) (resolved.Type, error) {
	// Generate candidates in priority order
	candidates := r.generateCandidates(nsName, name)

	for _, candidate := range candidates {
		// Check common types first (higher priority)
		if r.commonTypes[candidate] {
			return r.resolveCommonTypeRef(nsName, candidate)
		}

		// Then check entity types
		if r.entityTypes[candidate] {
			return resolved.EntityRef{EntityType: types.EntityType(candidate)}, nil
		}
	}

	// Check for built-in types
	if t := resolveBuiltinType(name); t != nil {
		return t, nil
	}

	return nil, &UndefinedTypeError{Name: name, Namespace: nsName}
}

// resolveName resolves a possibly-unqualified name to a fully-qualified name.
// If entityOnly is true, only entity types are considered; otherwise common types take priority.
func (r *resolver) resolveName(nsName string, name string, entityOnly bool) (string, error) {
	// If already qualified (contains ::), use as-is but verify it exists
	if strings.Contains(name, "::") {
		if entityOnly {
			if r.entityTypes[name] {
				return name, nil
			}
		} else {
			if r.commonTypes[name] || r.entityTypes[name] {
				return name, nil
			}
		}
		// Check built-ins
		if strings.HasPrefix(name, "__cedar::") {
			return name, nil
		}
		return "", &UndefinedTypeError{Name: name, Namespace: nsName}
	}

	// Generate candidates: [nsName::name, name]
	candidates := r.generateCandidates(nsName, name)

	for _, candidate := range candidates {
		if entityOnly {
			if r.entityTypes[candidate] {
				return candidate, nil
			}
		} else {
			// Priority: common > entity
			if r.commonTypes[candidate] {
				return candidate, nil
			}
			if r.entityTypes[candidate] {
				return candidate, nil
			}
		}
	}

	// Check built-ins for unqualified names
	if isBuiltinTypeName(name) {
		return "__cedar::" + name, nil
	}

	return "", &UndefinedTypeError{Name: name, Namespace: nsName}
}

// generateCandidates returns possible fully-qualified names for an unqualified name.
func (r *resolver) generateCandidates(nsName string, name string) []string {
	if strings.Contains(name, "::") {
		// Already qualified
		return []string{name}
	}

	if nsName == "" {
		// In empty namespace, only one candidate
		return []string{name}
	}

	// In a named namespace: try ns::name first, then name in empty namespace
	return []string{
		qualifyName(nsName, name),
		name,
	}
}

func qualifyName(nsName, name string) string {
	if nsName == "" {
		return name
	}
	return nsName + "::" + name
}
