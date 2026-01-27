package resolver

import (
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func resolveType(rd *resolveData, in ast.IsType) (ast.IsType, error) {
	switch t := in.(type) {
	case ast.SetType:
		return resolveSet(rd, t)
	case ast.RecordType:
		return resolveRecord(rd, t)
	case ast.EntityTypeRef:
		return resolveEntityTypeRef(rd, t), nil
	case ast.TypeRef:
		return resolveTypeRef(rd, t)
	default:
		return in, nil
	}
}

// resolve returns a new SetType with the element type resolved.
func resolveSet(rd *resolveData, s ast.SetType) (ast.SetType, error) {
	resolved, err := resolveType(rd, s.Element)
	if err != nil {
		return ast.SetType{}, err
	}
	return ast.SetType{Element: resolved}, nil
}

// resolve returns a new RecordType with all attribute types resolved.
func resolveRecord(rd *resolveData, r ast.RecordType) (ast.RecordType, error) {
	resolved := make([]ast.Pair, len(r.Pairs))
	for i, p := range r.Pairs {
		resolvedType, err := resolveType(rd, p.Type)
		if err != nil {
			return ast.RecordType{}, err
		}
		resolved[i] = ast.Pair{
			Key:         p.Key,
			Type:        resolvedType,
			Optional:    p.Optional,
			Annotations: p.Annotations,
		}
	}
	return ast.RecordType{Pairs: resolved}, nil
}

// willResolve resolves the entity type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it checks if the entity exists
// in the empty namespace first before qualifying it with the current namespace.
// This method never returns an error.
func resolveEntityTypeRef(rd *resolveData, e ast.EntityTypeRef) ast.EntityTypeRef {
	if rd.namespace == nil {
		return e
	}

	name := string(e.Name)
	// If already qualified (contains "::"), return as-is
	if strings.Contains(name, "::") || (len(name) > 0 && name[0] == ':') {
		return e
	}

	// Check if this entity exists in the empty namespace (global)
	if rd.entityExistsInEmptyNamespace(e.Name) {
		// Keep it unqualified to reference the global entity
		return e
	}

	// Otherwise, qualify it with the current namespace
	return ast.EntityTypeRef{Name: types.EntityType(string(rd.namespace.Name) + "::" + name)}
}

// resolve resolves the type reference relative to the given namespace and schema.
// It searches for a matching CommonType in the namespace first, then in the entire schema.
// If found, it returns the resolved concrete type. Otherwise, it returns an error.
func resolveTypeRef(rd *resolveData, t ast.TypeRef) (ast.IsType, error) {
	name := string(t.Name)

	// Try to find the type in the current namespace first (for unqualified names)
	if rd.namespace != nil && len(name) > 0 && name[0] != ':' && !strings.Contains(name, "::") {
		// Check namespace-local cache first
		if entry, found := rd.namespaceCommonTypes[name]; found {
			// If already resolved, return cached type
			if entry.resolved {
				return entry.node.Type, nil
			}
			// Resolve lazily
			resolvedNode, err := resolveCommonTypeNode(rd, entry.node)
			if err != nil {
				return nil, err
			}
			// Cache the resolved node
			entry.node = resolvedNode
			entry.resolved = true
			return resolvedNode.Type, nil
		}

		// Not found in namespace, qualify the name for schema search
		// name = string(rd.namespace.Name) + "::" + name
	}

	// Check schema-wide cache
	if entry, found := rd.schemaCommonTypes[name]; found {
		// If already resolved, return cached type
		if entry.resolved {
			return entry.node.Type, nil
		}
		// Resolve lazily with the common type's namespace context
		// Find the namespace for this common type by checking where it's declared
		var ns *ast.NamespaceNode
		for nsNode, ct := range rd.schema.CommonTypes() {
			var fullName string
			if nsNode == nil {
				fullName = string(ct.Name)
			} else {
				fullName = string(nsNode.Name) + "::" + string(ct.Name)
			}
			if fullName == name {
				ns = nsNode
				break
			}
		}
		ctRd := rd.withNamespace(ns)

		resolvedNode, err := resolveCommonTypeNode(ctRd, entry.node)
		if err != nil {
			return nil, err
		}
		// Cache the resolved node
		entry.node = resolvedNode
		entry.resolved = true
		return resolvedNode.Type, nil
	}

	if _, ok := knownExtensions[name]; ok {
		return ast.ExtensionType{Name: types.Ident(name)}, nil
	}

	// Not found, return an error
	// return nil, fmt.Errorf("type %q not found", name)
	return ast.EntityTypeRef{Name: types.EntityType(name)}, nil
}

var knownExtensions = map[string]struct{}{
	"decimal":  {},
	"duration": {},
	"datetime": {},
	"ipaddr":   {},
}
