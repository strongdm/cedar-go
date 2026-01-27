package resolver

import (
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func resolveType(rd *resolveData, in ast.IsType) ast.IsType {
	switch t := in.(type) {
	case ast.SetType:
		return resolveSet(rd, t)
	case ast.RecordType:
		return resolveRecord(rd, t)
	case ast.EntityTypeRef:
		return resolveEntityTypeRef(rd, t)
	case ast.TypeRef:
		return resolveTypeRef(rd, t)
	default:
		return in
	}
}

// resolve returns a new SetType with the element type resolved.
func resolveSet(rd *resolveData, s ast.SetType) ast.SetType {
	resolved := resolveType(rd, s.Element)
	return ast.SetType{Element: resolved}
}

// resolve returns a new RecordType with all attribute types resolved.
func resolveRecord(rd *resolveData, r ast.RecordType) ast.RecordType {
	resolvedAttrs := make(ast.Attributes)
	for key, attr := range r.Attributes {
		resolvedType := resolveType(rd, attr.Type)
		resolvedAttrs[key] = ast.Attribute{
			Type:        resolvedType,
			Optional:    attr.Optional,
			Annotations: attr.Annotations,
		}
	}
	return ast.RecordType{Attributes: resolvedAttrs}
}

// willResolve resolves the entity type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it checks if the entity exists
// in the empty namespace first before qualifying it with the current namespace.
// This method never returns an error.
func resolveEntityTypeRef(rd *resolveData, e ast.EntityTypeRef) ast.EntityTypeRef {
	if rd.namespacePath == "" {
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
	return ast.EntityTypeRef{Name: types.EntityType(string(rd.namespacePath) + "::" + name)}
}

// resolve resolves the type reference relative to the given namespace and schema.
// It searches for a matching CommonType in the namespace first, then in the entire schema.
// If found, it returns the resolved concrete type. Otherwise, it treats it as an EntityTypeRef.
func resolveTypeRef(rd *resolveData, t ast.TypeRef) ast.IsType {
	name := string(t.Name)

	// Try to find the type in the current namespace first (for unqualified names)
	if rd.namespacePath != "" && len(name) > 0 && name[0] != ':' && !strings.Contains(name, "::") {
		// Check namespace-local cache first
		if entry, found := rd.namespaceCommonTypes[name]; found {
			// If already resolved, return cached type
			if entry.resolved {
				return entry.node.Type
			}
			// Resolve lazily
			resolvedNode := resolveCommonTypeNode(rd, entry.node)
			// Cache the resolved node
			entry.node = resolvedNode
			entry.resolved = true
			return resolvedNode.Type
		}
	}

	// Check schema-wide cache
	if entry, found := rd.schemaCommonTypes[name]; found {
		// If already resolved, return cached type
		if entry.resolved {
			return entry.node.Type
		}
		// Resolve lazily with the common type's namespace context
		// Find the namespace for this common type by checking where it's declared
		var nsPath types.Path
		// Check top-level common types first
		if _, exists := rd.schema.CommonTypes[types.Ident(name)]; exists {
			nsPath = ""
		} else {
			// Check namespace common types
			for path, ns := range rd.schema.Namespaces {
				// Extract unqualified name from fully qualified name
				prefix := string(path) + "::"
				if strings.HasPrefix(name, prefix) {
					unqualifiedName := strings.TrimPrefix(name, prefix)
					if _, exists := ns.CommonTypes[types.Ident(unqualifiedName)]; exists {
						nsPath = path
						break
					}
				}
			}
		}
		ctRd := rd.withNamespace(nsPath)

		// Resolve lazily
		resolvedNode := resolveCommonTypeNode(ctRd, entry.node)
		// Cache the resolved node
		entry.node = resolvedNode
		entry.resolved = true
		return resolvedNode.Type
	}

	// Check for known extension types (with or without __cedar:: prefix)
	extensionName := name
	if strings.HasPrefix(name, "__cedar::") {
		extensionName = strings.TrimPrefix(name, "__cedar::")
	}
	if _, ok := knownExtensions[extensionName]; ok {
		return ast.ExtensionType{Name: types.Ident(extensionName)}
	}

	// Not found, treat as EntityTypeRef
	return ast.EntityTypeRef{Name: types.EntityType(name)}
}

var knownExtensions = map[string]struct{}{
	"decimal":  {},
	"duration": {},
	"datetime": {},
	"ipaddr":   {},
}
