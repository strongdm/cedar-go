// Package resolve implements schema resolution and validation.
package resolve

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// ResolveError represents an error during schema resolution.
type ResolveError struct {
	Errors []error
}

func (e *ResolveError) Error() string {
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}
	var b strings.Builder
	b.WriteString("schema resolution failed with multiple errors:\n")
	for _, err := range e.Errors {
		b.WriteString("  - ")
		b.WriteString(err.Error())
		b.WriteString("\n")
	}
	return b.String()
}

func (e *ResolveError) Unwrap() []error {
	return e.Errors
}

// Result holds the resolved schema data.
type Result struct {
	EntityTypes map[types.EntityType]*EntityType
	Actions     map[types.EntityUID]*Action
}

// EntityType is the resolved entity type.
type EntityType struct {
	Name        types.EntityType
	Descendants []types.EntityType
	Attributes  *RecordType
	Tags        Type
	Kind        EntityTypeKind
}

// EntityTypeKind indicates standard or enum entity type.
type EntityTypeKind interface {
	isEntityTypeKind()
}

// StandardKind is a normal entity type.
type StandardKind struct{}

func (StandardKind) isEntityTypeKind() {}

// EnumKind is an enumerated entity type.
type EnumKind struct {
	Values []types.EntityUID
}

func (EnumKind) isEntityTypeKind() {}

// Action is the resolved action.
type Action struct {
	Name      types.EntityUID
	MemberOf  []types.EntityUID
	AppliesTo *AppliesTo
	Context   *RecordType
}

// AppliesTo defines principals and resources for an action.
type AppliesTo struct {
	Principals []types.EntityType
	Resources  []types.EntityType
}

// Type is a resolved type.
type Type interface {
	isResolvedType()
}

// PrimitiveType is a built-in type.
type PrimitiveType struct {
	Name string
}

func (PrimitiveType) isResolvedType() {}

// EntityRefType references an entity type.
type EntityRefType struct {
	Name types.EntityType
}

func (EntityRefType) isResolvedType() {}

// SetType is Set<T>.
type SetType struct {
	Element Type
}

func (SetType) isResolvedType() {}

// RecordType is a record with attributes.
type RecordType struct {
	Attributes map[string]*Attribute
}

func (RecordType) isResolvedType() {}

// ExtensionType is an extension type.
type ExtensionType struct {
	Name string
}

func (ExtensionType) isResolvedType() {}

// Attribute is a field in a record.
type Attribute struct {
	Name     string
	Type     Type
	Required bool
}

// resolver holds state during resolution.
type resolver struct {
	schema      *ast.Schema
	errors      []error
	entityTypes map[types.EntityType]*EntityType
	actions     map[types.EntityUID]*Action
	commonTypes map[string]ast.Type // fully qualified name -> type
}

// Resolve performs schema resolution and validation.
func Resolve(schema *ast.Schema) (*Result, error) {
	r := &resolver{
		schema:      schema,
		entityTypes: make(map[types.EntityType]*EntityType),
		actions:     make(map[types.EntityUID]*Action),
		commonTypes: make(map[string]ast.Type),
	}

	// Phase 1: Collect all definitions and build qualified name maps
	r.collectDefinitions()

	// Phase 2: Resolve entity types
	r.resolveEntityTypes()

	// Phase 3: Resolve actions
	r.resolveActions()

	// Phase 4: Compute transitive closure for entity hierarchy
	r.computeDescendants()

	if len(r.errors) > 0 {
		return nil, &ResolveError{Errors: r.errors}
	}

	return &Result{
		EntityTypes: r.entityTypes,
		Actions:     r.actions,
	}, nil
}

func (r *resolver) collectDefinitions() {
	for nsName, ns := range r.schema.Namespaces {
		// Collect common types with qualified names
		for name, t := range ns.CommonTypes {
			qname := qualifyName(nsName, name)
			r.commonTypes[qname] = t
		}
	}
}

func (r *resolver) resolveEntityTypes() {
	for nsName, ns := range r.schema.Namespaces {
		for name, et := range ns.Entities {
			qname := types.EntityType(qualifyName(nsName, name))
			resolved := &EntityType{
				Name: qname,
			}

			// Resolve shape
			if et.Shape != nil {
				resolved.Attributes = r.resolveRecordType(nsName, et.Shape)
			}

			// Resolve tags
			if et.Tags != nil {
				resolved.Tags = r.resolveType(nsName, et.Tags)
			}

			// Resolve kind (standard or enum)
			if len(et.Enum) > 0 {
				values := make([]types.EntityUID, len(et.Enum))
				for i, v := range et.Enum {
					values[i] = types.NewEntityUID(qname, types.String(v))
				}
				resolved.Kind = EnumKind{Values: values}
			} else {
				resolved.Kind = StandardKind{}
			}

			r.entityTypes[qname] = resolved
		}
	}
}

func (r *resolver) resolveActions() {
	for nsName, ns := range r.schema.Namespaces {
		for name, a := range ns.Actions {
			// Actions use "Action" as the entity type
			actionType := types.EntityType(qualifyName(nsName, "Action"))
			uid := types.NewEntityUID(actionType, types.String(name))

			resolved := &Action{
				Name: uid,
			}

			// Resolve memberOf
			if len(a.MemberOf) > 0 {
				resolved.MemberOf = make([]types.EntityUID, len(a.MemberOf))
				for i, ref := range a.MemberOf {
					refNs := nsName
					if ref.Namespace != "" {
						refNs = ref.Namespace
					}
					refType := types.EntityType(qualifyName(refNs, "Action"))
					resolved.MemberOf[i] = types.NewEntityUID(refType, types.String(ref.Name))
				}
			}

			// Resolve appliesTo
			if a.AppliesTo != nil {
				resolved.AppliesTo = &AppliesTo{}

				// Resolve principals
				for _, p := range a.AppliesTo.Principals {
					et := r.qualifyEntityType(nsName, p)
					if _, ok := r.entityTypes[et]; !ok {
						r.errors = append(r.errors, fmt.Errorf("action %q: unknown principal type %q", name, p))
					}
					resolved.AppliesTo.Principals = append(resolved.AppliesTo.Principals, et)
				}

				// Resolve resources
				for _, res := range a.AppliesTo.Resources {
					et := r.qualifyEntityType(nsName, res)
					if _, ok := r.entityTypes[et]; !ok {
						r.errors = append(r.errors, fmt.Errorf("action %q: unknown resource type %q", name, res))
					}
					resolved.AppliesTo.Resources = append(resolved.AppliesTo.Resources, et)
				}

				// Resolve context
				if a.AppliesTo.Context != nil {
					if rt, ok := r.resolveType(nsName, a.AppliesTo.Context).(*RecordType); ok {
						resolved.Context = rt
					}
				}
			}

			r.actions[uid] = resolved
		}
	}
}

func (r *resolver) computeDescendants() {
	// Build parent map from memberOf relationships
	// memberOf specifies what types THIS entity can be a member OF
	// descendants are the inverse - what entities can be members of THIS entity

	// First pass: collect direct children (inverse of memberOf)
	children := make(map[types.EntityType][]types.EntityType)
	for nsName, ns := range r.schema.Namespaces {
		for name, et := range ns.Entities {
			childType := types.EntityType(qualifyName(nsName, name))
			for _, parent := range et.MemberOf {
				parentType := r.qualifyEntityType(nsName, parent)
				if _, ok := r.entityTypes[parentType]; !ok {
					r.errors = append(r.errors, fmt.Errorf("entity %q: unknown parent type %q", name, parent))
					continue
				}
				children[parentType] = append(children[parentType], childType)
			}
		}
	}

	// Compute transitive closure using iterative approach
	changed := true
	for changed {
		changed = false
		for parentType, directChildren := range children {
			existing := make(map[types.EntityType]bool)
			for _, c := range children[parentType] {
				existing[c] = true
			}

			for _, child := range directChildren {
				// Add grandchildren
				for _, grandchild := range children[child] {
					if !existing[grandchild] {
						children[parentType] = append(children[parentType], grandchild)
						existing[grandchild] = true
						changed = true
					}
				}
			}
		}
	}

	// Detect cycles
	for et := range r.entityTypes {
		for _, desc := range children[et] {
			if desc == et {
				r.errors = append(r.errors, fmt.Errorf("entity %q: cycle detected in entity hierarchy", et))
			}
		}
	}

	// Assign descendants to entity types
	for et, resolved := range r.entityTypes {
		resolved.Descendants = children[et]
	}
}

func (r *resolver) resolveType(namespace string, t ast.Type) Type {
	switch t := t.(type) {
	case ast.PrimitiveType:
		return PrimitiveType{Name: t.Name}
	case ast.EntityRefType:
		return EntityRefType{Name: r.qualifyEntityType(namespace, t.Name)}
	case *ast.SetType:
		return SetType{Element: r.resolveType(namespace, t.Element)}
	case ast.SetType:
		return SetType{Element: r.resolveType(namespace, t.Element)}
	case *ast.RecordType:
		return r.resolveRecordType(namespace, t)
	case ast.RecordType:
		return r.resolveRecordType(namespace, &t)
	case ast.CommonTypeRef:
		// Inline common type
		qname := r.qualifyTypeName(namespace, t.Name)
		if ct, ok := r.commonTypes[qname]; ok {
			return r.resolveType(namespace, ct)
		}
		// Might be an entity type reference
		return EntityRefType{Name: types.EntityType(qname)}
	case ast.ExtensionType:
		return ExtensionType{Name: t.Name}
	default:
		return PrimitiveType{Name: "String"} // fallback
	}
}

func (r *resolver) resolveRecordType(namespace string, rt *ast.RecordType) *RecordType {
	result := &RecordType{
		Attributes: make(map[string]*Attribute),
	}
	if rt.Attributes != nil {
		for name, attr := range rt.Attributes {
			result.Attributes[name] = &Attribute{
				Name:     name,
				Type:     r.resolveType(namespace, attr.Type),
				Required: attr.Required,
			}
		}
	}
	return result
}

func (r *resolver) qualifyEntityType(namespace, name string) types.EntityType {
	return types.EntityType(r.qualifyTypeName(namespace, name))
}

func (r *resolver) qualifyTypeName(namespace, name string) string {
	// If already qualified (contains ::), return as-is
	if strings.Contains(name, "::") {
		return name
	}
	// Check if it's a primitive type
	if name == "String" || name == "Long" || name == "Bool" || name == "Boolean" {
		return name
	}
	return qualifyName(namespace, name)
}

func qualifyName(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "::" + name
}

var ErrCycleDetected = errors.New("cycle detected in entity hierarchy")
