package schema

import "github.com/cedar-policy/cedar-go/internal/schema/ast"

// convertJSONTypeToType converts an internal JSON type to a public Type.
func convertJSONTypeToType(jsonType *ast.JSONType) Type {
	if jsonType == nil {
		return nil
	}

	switch jsonType.Type {
	case "String", "Long", "Boolean", "Bool":
		// Normalize Boolean to Bool
		typeName := jsonType.Type
		if typeName == "Boolean" {
			typeName = "Bool"
		}
		p := &PathType{path: typeName}
		if jsonType.Annotations != nil {
			p.annotations = jsonType.Annotations
		}
		return p

	case "Set":
		s := &SetType{
			element: convertJSONTypeToType(jsonType.Element),
		}
		if jsonType.Annotations != nil {
			s.annotations = jsonType.Annotations
		}
		return s

	case "Record":
		r := &RecordType{
			attributes: make(map[string]*Attribute),
		}
		if jsonType.Annotations != nil {
			r.annotations = jsonType.Annotations
		}

		for attrName, jsonAttr := range jsonType.Attributes {
			attr := convertJSONAttributeToAttribute(attrName, jsonAttr)
			r.attributes[attrName] = attr
		}

		return r

	case "Entity", "EntityOrCommon":
		p := &PathType{path: jsonType.Name}
		if jsonType.Annotations != nil {
			p.annotations = jsonType.Annotations
		}
		return p

	case "Extension":
		// Extensions are represented as path types with special names
		p := &PathType{path: jsonType.Name}
		if jsonType.Annotations != nil {
			p.annotations = jsonType.Annotations
		}
		return p

	default:
		// Unknown type, treat as path
		p := &PathType{path: jsonType.Type}
		if jsonType.Annotations != nil {
			p.annotations = jsonType.Annotations
		}
		return p
	}
}

// convertJSONAttributeToAttribute converts an internal JSON attribute to a public Attribute.
func convertJSONAttributeToAttribute(name string, jsonAttr *ast.JSONAttribute) *Attribute {
	attr := &Attribute{
		name:       name,
		isRequired: jsonAttr.Required,
	}

	if jsonAttr.Annotations != nil {
		attr.annotations = jsonAttr.Annotations
	}

	// Convert the attribute type based on its structure
	switch jsonAttr.Type {
	case "String", "Long", "Boolean", "Bool":
		// Normalize Boolean to Bool
		typeName := jsonAttr.Type
		if typeName == "Boolean" {
			typeName = "Bool"
		}
		p := &PathType{path: typeName}
		attr.attrType = p

	case "Set":
		s := &SetType{
			element: convertJSONAttributeElementToType(jsonAttr),
		}
		attr.attrType = s

	case "Record":
		r := &RecordType{
			attributes: make(map[string]*Attribute),
		}

		for nestedName, nestedAttr := range jsonAttr.Attributes {
			nestedAttribute := convertJSONAttributeToAttribute(nestedName, nestedAttr)
			r.attributes[nestedName] = nestedAttribute
		}

		attr.attrType = r

	case "Entity", "EntityOrCommon":
		p := &PathType{path: jsonAttr.Name}
		attr.attrType = p

	case "Extension":
		p := &PathType{path: jsonAttr.Name}
		attr.attrType = p

	default:
		// Unknown type, treat as path
		p := &PathType{path: jsonAttr.Type}
		attr.attrType = p
	}

	return attr
}

// convertJSONAttributeElementToType converts a JSON attribute's element field to a Type.
func convertJSONAttributeElementToType(jsonAttr *ast.JSONAttribute) Type {
	if jsonAttr.Element == nil {
		return nil
	}

	return convertJSONTypeToType(jsonAttr.Element)
}

// convertTypeToJSONType converts a public Type to an internal JSON type.
func convertTypeToJSONType(typ Type) *ast.JSONType {
	if typ == nil {
		return nil
	}

	switch t := typ.(type) {
	case *PathType:
		// Determine if this is a primitive, entity, or extension type
		jsonType := &ast.JSONType{}

		switch t.path {
		case "String", "Long", "Bool", "Boolean":
			jsonType.Type = t.path
			// Normalize Bool to Boolean for JSON output
			if jsonType.Type == "Bool" {
				jsonType.Type = "Boolean"
			}
		default:
			// Assume it's an entity or common type reference
			jsonType.Type = "EntityOrCommon"
			jsonType.Name = t.path
		}

		if t.annotations != nil {
			jsonType.Annotations = t.annotations
		}

		return jsonType

	case *SetType:
		jsonType := &ast.JSONType{
			Type:    "Set",
			Element: convertTypeToJSONType(t.element),
		}

		if t.annotations != nil {
			jsonType.Annotations = t.annotations
		}

		return jsonType

	case *RecordType:
		jsonType := &ast.JSONType{
			Type:       "Record",
			Attributes: make(map[string]*ast.JSONAttribute),
		}

		if t.annotations != nil {
			jsonType.Annotations = t.annotations
		}

		for attrName, attr := range t.attributes {
			jsonAttr := convertAttributeToJSONAttribute(attr)
			jsonType.Attributes[attrName] = jsonAttr
		}

		return jsonType

	default:
		return nil
	}
}

// convertAttributeToJSONAttribute converts a public Attribute to an internal JSON attribute.
func convertAttributeToJSONAttribute(attr *Attribute) *ast.JSONAttribute {
	jsonAttr := &ast.JSONAttribute{
		Required: attr.isRequired,
	}

	if attr.annotations != nil {
		jsonAttr.Annotations = attr.annotations
	}

	// Convert the attribute type
	switch t := attr.attrType.(type) {
	case *PathType:
		switch t.path {
		case "String", "Long", "Bool", "Boolean":
			jsonAttr.Type = t.path
			// Normalize Bool to Boolean for JSON output
			if jsonAttr.Type == "Bool" {
				jsonAttr.Type = "Boolean"
			}
		default:
			// Entity or common type reference
			jsonAttr.Type = "EntityOrCommon"
			jsonAttr.Name = t.path
		}

	case *SetType:
		jsonAttr.Type = "Set"
		jsonAttr.Element = convertTypeToJSONType(t.element)

	case *RecordType:
		jsonAttr.Type = "Record"
		jsonAttr.Attributes = make(map[string]*ast.JSONAttribute)

		for nestedName, nestedAttr := range t.attributes {
			nestedJSONAttr := convertAttributeToJSONAttribute(nestedAttr)
			jsonAttr.Attributes[nestedName] = nestedJSONAttr
		}
	}

	return jsonAttr
}
