package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Resolve fully qualifies all type references and returns a resolved schema.
// It detects cycles in common type definitions, validates RFC 70 shadowing
// rules, and resolves names with priority: common > entity > primitive/extension.
func (s *Schema) Resolve() (*resolved.Schema, error) {
	r := newResolver(s)
	r.buildIndex()
	if err := r.checkShadowing(); err != nil {
		return nil, err
	}
	order, err := r.topoSortCommonTypes()
	if err != nil {
		return nil, err
	}
	if err := r.resolveCommonTypes(order); err != nil {
		return nil, err
	}
	return r.resolveAll()
}

type resolver struct {
	schema *Schema

	entityDefs map[string]bool // qualified entity + enum names
	commonDefs map[string]bool // qualified common type names (includes built-ins)
	actionDefs map[string]bool // action keys: "EntityType\x00ID"

	userCommon     map[string]*commonEntry
	resolvedCommon map[string]resolved.Type
}

type commonEntry struct {
	ns   string
	body TypeExpr
}

var builtinTypes = map[string]resolved.Type{
	"Long":     resolved.Primitive{Kind: resolved.PrimitiveLong},
	"String":   resolved.Primitive{Kind: resolved.PrimitiveString},
	"Bool":     resolved.Primitive{Kind: resolved.PrimitiveBool},
	"ipaddr":   resolved.Extension{Name: "ipaddr"},
	"decimal":  resolved.Extension{Name: "decimal"},
	"datetime": resolved.Extension{Name: "datetime"},
	"duration": resolved.Extension{Name: "duration"},
}

func newResolver(s *Schema) *resolver {
	r := &resolver{
		schema:         s,
		entityDefs:     make(map[string]bool),
		commonDefs:     make(map[string]bool),
		actionDefs:     make(map[string]bool),
		userCommon:     make(map[string]*commonEntry),
		resolvedCommon: make(map[string]resolved.Type),
	}
	for name, rt := range builtinTypes {
		r.commonDefs[name] = true
		r.commonDefs["__cedar::"+name] = true
		r.resolvedCommon[name] = rt
		r.resolvedCommon["__cedar::"+name] = rt
	}
	return r
}

func (r *resolver) buildIndex() {
	for nsName, ns := range r.schema.Namespaces {
		for name := range ns.EntityTypes {
			r.entityDefs[qualifyName(nsName, name)] = true
		}
		for name := range ns.EnumTypes {
			r.entityDefs[qualifyName(nsName, name)] = true
		}
		for name, ct := range ns.CommonTypes {
			qname := qualifyName(nsName, name)
			r.commonDefs[qname] = true
			r.userCommon[qname] = &commonEntry{ns: nsName, body: ct.Type}
		}
		aet := qualifyName(nsName, "Action")
		for name := range ns.Actions {
			r.actionDefs[aet+"\x00"+name] = true
		}
	}
}

// checkShadowing enforces RFC 70: definitions in named namespaces
// cannot shadow same-named definitions in the empty namespace.
func (r *resolver) checkShadowing() error {
	emptyNS, hasEmpty := r.schema.Namespaces[""]
	if !hasEmpty {
		return nil
	}

	emptyNames := make(map[string]bool)
	for name := range emptyNS.EntityTypes {
		emptyNames[name] = true
	}
	for name := range emptyNS.EnumTypes {
		emptyNames[name] = true
	}
	for name := range emptyNS.CommonTypes {
		emptyNames[name] = true
	}

	emptyActions := make(map[string]bool)
	for name := range emptyNS.Actions {
		emptyActions[name] = true
	}

	for nsName, ns := range r.schema.Namespaces {
		if nsName == "" {
			continue
		}
		for name := range ns.EntityTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}
		for name := range ns.EnumTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}
		for name := range ns.CommonTypes {
			if emptyNames[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}
		for name := range ns.Actions {
			if emptyActions[name] {
				return &ShadowError{Name: name, Namespace: nsName}
			}
		}
	}
	return nil
}

// topoSortCommonTypes performs Kahn's algorithm on user-defined common types
// and returns them in dependency order. Returns CycleError if a cycle exists.
func (r *resolver) topoSortCommonTypes() ([]string, error) {
	inDegree := make(map[string]int, len(r.userCommon))
	dependents := make(map[string][]string)

	for qname := range r.userCommon {
		inDegree[qname] = 0
	}

	for qname, entry := range r.userCommon {
		for _, dep := range r.findCommonDeps(entry.ns, entry.body) {
			dependents[dep] = append(dependents[dep], qname)
			inDegree[qname]++
		}
	}

	var queue []string
	for qname, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, qname)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)
		for _, dep := range dependents[curr] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				queue = append(queue, dep)
				sort.Strings(queue)
			}
		}
	}

	if len(order) < len(r.userCommon) {
		var cycle []string
		for qname, deg := range inDegree {
			if deg > 0 {
				cycle = append(cycle, qname)
			}
		}
		sort.Strings(cycle)
		return nil, &CycleError{Path: cycle}
	}
	return order, nil
}

// findCommonDeps walks a type expression and returns qualified names of
// user-defined common types it depends on.
func (r *resolver) findCommonDeps(ns string, expr TypeExpr) []string {
	var deps []string
	walkTypeExpr(expr, func(te TypeExpr) {
		tn, ok := te.(TypeNameExpr)
		if !ok {
			return
		}
		for _, c := range resolutionCandidates(ns, tn.Name) {
			if _, isUser := r.userCommon[c]; isUser {
				deps = append(deps, c)
				return
			}
			if r.commonDefs[c] || r.entityDefs[c] {
				return
			}
		}
	})
	return deps
}

func walkTypeExpr(expr TypeExpr, fn func(TypeExpr)) {
	fn(expr)
	switch v := expr.(type) {
	case SetTypeExpr:
		walkTypeExpr(v.Element, fn)
	case *RecordTypeExpr:
		for _, attr := range v.Attributes {
			walkTypeExpr(attr.Type, fn)
		}
	}
}

func (r *resolver) resolveCommonTypes(order []string) error {
	for _, qname := range order {
		entry := r.userCommon[qname]
		rt, err := r.resolveTypeExpr(entry.ns, entry.body)
		if err != nil {
			return fmt.Errorf("common type %q: %w", qname, err)
		}
		r.resolvedCommon[qname] = rt
	}
	return nil
}

func (r *resolver) resolveAll() (*resolved.Schema, error) {
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
		Annotations: resolved.Annotations(ns.Annotations),
	}

	for name, et := range ns.EntityTypes {
		ret, err := r.resolveEntityTypeDef(nsName, et)
		if err != nil {
			return nil, fmt.Errorf("entity type %q: %w", qualifyName(nsName, name), err)
		}
		rns.EntityTypes[types.EntityType(qualifyName(nsName, name))] = ret
	}

	for name, enum := range ns.EnumTypes {
		rns.EnumTypes[types.EntityType(qualifyName(nsName, name))] = &resolved.EnumType{
			Values:      enum.Values,
			Annotations: resolved.Annotations(enum.Annotations),
		}
	}

	aet := qualifyName(nsName, "Action")
	for name, act := range ns.Actions {
		ract, err := r.resolveActionDef(nsName, act)
		if err != nil {
			return nil, fmt.Errorf("action %q: %w", name, err)
		}
		uid := types.NewEntityUID(types.EntityType(aet), types.String(name))
		rns.Actions[uid] = ract
	}

	return rns, nil
}

func (r *resolver) resolveEntityTypeDef(ns string, et *EntityTypeDef) (*resolved.EntityType, error) {
	ret := &resolved.EntityType{
		Annotations: resolved.Annotations(et.Annotations),
	}

	for _, parentName := range et.MemberOfTypes {
		parent, err := r.resolveAsEntityType(ns, parentName)
		if err != nil {
			return nil, fmt.Errorf("memberOf %q: %w", parentName, err)
		}
		ret.MemberOfTypes = append(ret.MemberOfTypes, parent)
	}

	if et.Shape != nil {
		shape, err := r.resolveRecordType(ns, et.Shape)
		if err != nil {
			return nil, fmt.Errorf("shape: %w", err)
		}
		ret.Shape = shape
	}

	if et.Tags != nil {
		tags, err := r.resolveTypeExpr(ns, et.Tags)
		if err != nil {
			return nil, fmt.Errorf("tags: %w", err)
		}
		ret.Tags = tags
	}

	return ret, nil
}

func (r *resolver) resolveActionDef(ns string, act *ActionDef) (*resolved.Action, error) {
	ract := &resolved.Action{
		Annotations: resolved.Annotations(act.Annotations),
	}

	for _, ref := range act.MemberOf {
		uid, err := r.resolveActionRef(ns, ref)
		if err != nil {
			return nil, fmt.Errorf("memberOf: %w", err)
		}
		ract.MemberOf = append(ract.MemberOf, uid)
	}

	if act.AppliesTo != nil {
		for _, pt := range act.AppliesTo.PrincipalTypes {
			et, err := r.resolveAsEntityType(ns, pt)
			if err != nil {
				return nil, fmt.Errorf("principal type %q: %w", pt, err)
			}
			ract.PrincipalTypes = append(ract.PrincipalTypes, et)
		}
		for _, rt := range act.AppliesTo.ResourceTypes {
			et, err := r.resolveAsEntityType(ns, rt)
			if err != nil {
				return nil, fmt.Errorf("resource type %q: %w", rt, err)
			}
			ract.ResourceTypes = append(ract.ResourceTypes, et)
		}
		if act.AppliesTo.Context != nil {
			ctx, err := r.resolveTypeExpr(ns, act.AppliesTo.Context)
			if err != nil {
				return nil, fmt.Errorf("context: %w", err)
			}
			rt, ok := ctx.(*resolved.RecordType)
			if !ok {
				return nil, fmt.Errorf("context must resolve to a Record type")
			}
			ract.Context = rt
		}
	}

	return ract, nil
}

func (r *resolver) resolveActionRef(ns string, ref *ActionRef) (types.EntityUID, error) {
	aet := ref.Type
	if aet == "" {
		aet = qualifyName(ns, "Action")
	}
	key := aet + "\x00" + ref.ID
	if !r.actionDefs[key] {
		return types.EntityUID{}, &UndefinedTypeError{
			Name:    ref.ID,
			Context: fmt.Sprintf("in action group %q", aet),
		}
	}
	return types.NewEntityUID(types.EntityType(aet), types.String(ref.ID)), nil
}

// resolveAsEntityType resolves a name that must be an entity type
// (used for memberOfTypes, principalTypes, resourceTypes).
func (r *resolver) resolveAsEntityType(ns, name string) (types.EntityType, error) {
	for _, c := range resolutionCandidates(ns, name) {
		if r.entityDefs[c] {
			return types.EntityType(c), nil
		}
	}
	return "", &UndefinedTypeError{Name: name, Namespace: ns}
}

// resolveTypeExpr resolves any TypeExpr to a resolved.Type.
func (r *resolver) resolveTypeExpr(ns string, expr TypeExpr) (resolved.Type, error) {
	switch v := expr.(type) {
	case PrimitiveTypeExpr:
		return resolved.Primitive{Kind: toResolvedPrimitive(v.Kind)}, nil
	case SetTypeExpr:
		elem, err := r.resolveTypeExpr(ns, v.Element)
		if err != nil {
			return nil, err
		}
		return resolved.Set{Element: elem}, nil
	case *RecordTypeExpr:
		return r.resolveRecordType(ns, v)
	case EntityRefExpr:
		et, err := r.resolveAsEntityType(ns, v.Name)
		if err != nil {
			return nil, err
		}
		return resolved.EntityRef{EntityType: et}, nil
	case ExtensionTypeExpr:
		return resolved.Extension{Name: v.Name}, nil
	case EntityNameExpr:
		et, err := r.resolveAsEntityType(ns, v.Name)
		if err != nil {
			return nil, err
		}
		return resolved.EntityRef{EntityType: et}, nil
	default: // TypeNameExpr
		return r.resolveTypeName(ns, expr.(TypeNameExpr).Name)
	}
}

// resolveTypeName resolves a name in type position with priority:
// common types > entity types (RFC 24).
func (r *resolver) resolveTypeName(ns, name string) (resolved.Type, error) {
	for _, c := range resolutionCandidates(ns, name) {
		if r.commonDefs[c] {
			if rt, ok := r.resolvedCommon[c]; ok {
				return rt, nil
			}
		}
		if r.entityDefs[c] {
			return resolved.EntityRef{EntityType: types.EntityType(c)}, nil
		}
	}
	return nil, &UndefinedTypeError{Name: name, Namespace: ns}
}

func (r *resolver) resolveRecordType(ns string, rt *RecordTypeExpr) (*resolved.RecordType, error) {
	attrs := make(map[string]*resolved.Attribute, len(rt.Attributes))
	for name, attr := range rt.Attributes {
		rtype, err := r.resolveTypeExpr(ns, attr.Type)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", name, err)
		}
		attrs[name] = &resolved.Attribute{
			Type:        rtype,
			Required:    attr.Required,
			Annotations: resolved.Annotations(attr.Annotations),
		}
	}
	return &resolved.RecordType{Attributes: attrs}, nil
}

func toResolvedPrimitive(k PrimitiveKind) resolved.PrimitiveKind {
	switch k {
	case PrimitiveLong:
		return resolved.PrimitiveLong
	case PrimitiveString:
		return resolved.PrimitiveString
	default: // PrimitiveBool
		return resolved.PrimitiveBool
	}
}

// qualifyName returns the fully-qualified name ns::name, or just name if ns is empty.
func qualifyName(ns, name string) string {
	if ns == "" {
		return name
	}
	return ns + "::" + name
}

// resolutionCandidates returns the names to try when resolving name in namespace ns.
// If the name is already explicitly qualified (contains "::"), only that name is returned.
// Otherwise, for non-empty namespaces: [ns::name, name]; for empty namespace: [name].
func resolutionCandidates(ns, name string) []string {
	if strings.Contains(name, "::") {
		return []string{name}
	}
	if ns == "" {
		return []string{name}
	}
	return []string{qualifyName(ns, name), name}
}
