// Package schema provides functions for programmatically constructing and parsing Cedar schemas.
//
// Cedar schemas define entity types, actions, and their relationships. Schemas can be represented
// in either JSON or human-readable Cedar format, and this package supports parsing and marshaling
// between both formats.
//
// # Creating a Schema
//
// Schemas can be created programmatically using builder functions:
//
//	s := schema.NewSchema().
//		WithNamespace("PhotoApp",
//			schema.NewEntity("User").
//				WithAttribute("name", schema.String()).
//				WithAttribute("age", schema.Long()),
//			schema.NewAction("viewPhoto").
//				AppliesTo(
//					schema.Principals("User"),
//					schema.Resources("Photo"),
//					nil,
//				),
//		)
//
// # Parsing Schemas
//
// Schemas can be parsed from JSON or Cedar text format:
//
//	var s schema.Schema
//	err := s.UnmarshalJSON(jsonBytes)
//	// or
//	err := s.UnmarshalCedar(cedarBytes)
//
// # Marshaling Schemas
//
// Schemas can be marshaled to JSON or Cedar text format:
//
//	jsonBytes, err := s.MarshalJSON()
//	cedarBytes, err := s.MarshalCedar()
package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
)

// Schema represents a complete Cedar schema containing entity types, actions, and their relationships.
// A schema can contain multiple namespaces, and can be marshaled to or unmarshaled from JSON or Cedar text format.
type Schema struct {
	namespaces map[string]*Namespace
	filename   string
}

// NewSchema creates a new empty schema.
func NewSchema() *Schema {
	return &Schema{
		namespaces: make(map[string]*Namespace),
	}
}

// WithNamespace adds a namespace to the schema with the given declarations.
// The namespace parameter is the fully-qualified namespace name (e.g., "MyApp::Users").
// Use an empty string "" for the global namespace.
func (s *Schema) WithNamespace(namespace string, decls ...Declaration) *Schema {
	if s.namespaces == nil {
		s.namespaces = make(map[string]*Namespace)
	}

	ns := &Namespace{
		name:   namespace,
		schema: s,
	}

	for _, decl := range decls {
		decl.addToNamespace(ns)
	}

	s.namespaces[namespace] = ns
	return s
}

// SetFilename sets the filename used in error messages when unmarshaling.
func (s *Schema) SetFilename(filename string) *Schema {
	s.filename = filename
	return s
}

// UnmarshalJSON parses a schema from JSON format as specified by the Cedar documentation.
// See: https://docs.cedarpolicy.com/schema/json-schema.html
func (s *Schema) UnmarshalJSON(data []byte) error {
	var jsonSchema ast.JSONSchema
	if err := json.Unmarshal(data, &jsonSchema); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	s.namespaces = make(map[string]*Namespace)

	// Convert from internal JSON representation
	for nsName, jsonNS := range jsonSchema {
		ns := &Namespace{
			name:        nsName,
			schema:      s,
			entities:    make(map[string]*Entity),
			actions:     make(map[string]*Action),
			commonTypes: make(map[string]Type),
		}

		if jsonNS.Annotations != nil {
			ns.annotations = jsonNS.Annotations
		}

		// Convert entities
		for entityName, jsonEntity := range jsonNS.EntityTypes {
			entity := &Entity{
				name:        entityName,
				namespace:   ns,
				memberOf:    jsonEntity.MemberOfTypes,
				annotations: jsonEntity.Annotations,
			}

			if jsonEntity.Shape != nil {
				entity.shape = convertJSONTypeToType(jsonEntity.Shape)
			}

			if jsonEntity.Tags != nil {
				entity.tags = convertJSONTypeToType(jsonEntity.Tags)
			}

			if jsonEntity.Enum != nil {
				entity.enum = jsonEntity.Enum
			}

			ns.entities[entityName] = entity
		}

		// Convert actions
		for actionName, jsonAction := range jsonNS.Actions {
			action := &Action{
				name:        actionName,
				namespace:   ns,
				annotations: jsonAction.Annotations,
			}

			if jsonAction.MemberOf != nil {
				for _, member := range jsonAction.MemberOf {
					action.memberOf = append(action.memberOf, &ActionRef{
						id:       member.ID,
						typeName: member.Type,
					})
				}
			}

			if jsonAction.AppliesTo != nil {
				action.appliesTo = &AppliesTo{
					principals: jsonAction.AppliesTo.PrincipalTypes,
					resources:  jsonAction.AppliesTo.ResourceTypes,
				}
				if jsonAction.AppliesTo.Context != nil {
					action.appliesTo.context = convertJSONTypeToType(jsonAction.AppliesTo.Context)
				}
			}

			ns.actions[actionName] = action
		}

		// Convert common types
		for typeName, jsonCommonType := range jsonNS.CommonTypes {
			ns.commonTypes[typeName] = convertJSONTypeToType(jsonCommonType.JSONType)
		}

		s.namespaces[nsName] = ns
	}

	return nil
}

// MarshalJSON converts the schema to JSON format as specified by the Cedar documentation.
// See: https://docs.cedarpolicy.com/schema/json-schema.html
func (s *Schema) MarshalJSON() ([]byte, error) {
	if len(s.namespaces) == 0 {
		return []byte("{}"), nil
	}

	jsonSchema := make(ast.JSONSchema)

	for nsName, ns := range s.namespaces {
		jsonNS := &ast.JSONNamespace{
			EntityTypes: make(map[string]*ast.JSONEntity),
			Actions:     make(map[string]*ast.JSONAction),
			CommonTypes: make(map[string]*ast.JSONCommonType),
		}

		if len(ns.annotations) > 0 {
			jsonNS.Annotations = ns.annotations
		}

		// Convert entities
		for entityName, entity := range ns.entities {
			jsonEntity := &ast.JSONEntity{
				MemberOfTypes: entity.memberOf,
				Annotations:   entity.annotations,
			}

			if entity.shape != nil {
				jsonEntity.Shape = convertTypeToJSONType(entity.shape)
			}

			if entity.tags != nil {
				jsonEntity.Tags = convertTypeToJSONType(entity.tags)
			}

			if entity.enum != nil {
				jsonEntity.Enum = entity.enum
			}

			jsonNS.EntityTypes[entityName] = jsonEntity
		}

		// Convert actions
		for actionName, action := range ns.actions {
			jsonAction := &ast.JSONAction{
				Annotations: action.annotations,
			}

			if action.memberOf != nil {
				for _, member := range action.memberOf {
					jsonMember := &ast.JSONMember{
						ID:   member.id,
						Type: member.typeName,
					}
					jsonAction.MemberOf = append(jsonAction.MemberOf, jsonMember)
				}
			}

			if action.appliesTo != nil {
				jsonAction.AppliesTo = &ast.JSONAppliesTo{
					PrincipalTypes: action.appliesTo.principals,
					ResourceTypes:  action.appliesTo.resources,
				}
				if action.appliesTo.context != nil {
					jsonAction.AppliesTo.Context = convertTypeToJSONType(action.appliesTo.context)
				}
			}

			jsonNS.Actions[actionName] = jsonAction
		}

		// Convert common types
		for typeName, typeVal := range ns.commonTypes {
			jsonNS.CommonTypes[typeName] = &ast.JSONCommonType{
				JSONType: convertTypeToJSONType(typeVal),
			}
		}

		jsonSchema[nsName] = jsonNS
	}

	return json.Marshal(jsonSchema)
}

// UnmarshalCedar parses a schema from Cedar text format as specified by the Cedar documentation.
// See: https://docs.cedarpolicy.com/schema/human-readable-schema-grammar.html
func (s *Schema) UnmarshalCedar(data []byte) error {
	humanSchema, err := parser.ParseFile(s.filename, data)
	if err != nil {
		return err
	}

	// Convert to JSON format first (reusing existing conversion logic)
	jsonSchema := ast.ConvertHuman2JSON(humanSchema)

	// Then unmarshal from JSON
	jsonData, err := json.Marshal(jsonSchema)
	if err != nil {
		return fmt.Errorf("failed to convert schema: %w", err)
	}

	return s.UnmarshalJSON(jsonData)
}

// MarshalCedar converts the schema to Cedar text format as specified by the Cedar documentation.
// See: https://docs.cedarpolicy.com/schema/human-readable-schema-grammar.html
func (s *Schema) MarshalCedar() ([]byte, error) {
	// Convert to JSON format first
	jsonData, err := s.MarshalJSON()
	if err != nil {
		return nil, err
	}

	// Parse as JSON schema
	var jsonSchema ast.JSONSchema
	if err := json.Unmarshal(jsonData, &jsonSchema); err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	// Convert to human-readable format
	humanSchema := ast.ConvertJSON2Human(jsonSchema)

	// Format as Cedar text
	var buf bytes.Buffer
	if err := ast.Format(humanSchema, &buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
