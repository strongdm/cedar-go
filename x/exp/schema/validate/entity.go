package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Entity validates a single entity against the schema.
func (v *Validator) Entity(entity types.Entity) error {
	et := entity.UID.Type

	// Check if it's an action entity
	if isActionEntity(et) {
		return v.validateActionEntity(entity)
	}

	// Look up in entities
	if schemaEntity, ok := v.schema.Entities[et]; ok {
		return v.validateEntity(entity, schemaEntity)
	}

	// Look up in enums
	if schemaEnum, ok := v.schema.Enums[et]; ok {
		return validateEnumEntity(entity, schemaEnum)
	}

	return fmt.Errorf("entity type %q not found in schema", et)
}

// Entities validates all entities in the map against the schema.
func (v *Validator) Entities(entities types.EntityMap) error {
	for _, entity := range entities {
		if err := v.Entity(entity); err != nil {
			return fmt.Errorf("entity %s: %w", entity.UID, err)
		}
	}
	return nil
}

func (v *Validator) validateActionEntity(entity types.Entity) error {
	action, ok := v.schema.Actions[entity.UID]
	if !ok {
		return fmt.Errorf("action %s not found in schema", entity.UID)
	}

	// Verify parents match
	schemaParents := make(map[types.EntityUID]bool)
	for parent := range action.Entity.Parents.All() {
		schemaParents[parent] = true
	}

	for parent := range entity.Parents.All() {
		if !schemaParents[parent] {
			return fmt.Errorf("action %s has unexpected parent %s", entity.UID, parent)
		}
	}
	for parent := range action.Entity.Parents.All() {
		if !entity.Parents.Contains(parent) {
			return fmt.Errorf("action %s missing expected parent %s", entity.UID, parent)
		}
	}

	return nil
}

func (v *Validator) validateEntity(entity types.Entity, schemaEntity resolved.Entity) error {
	// Validate parents
	for parent := range entity.Parents.All() {
		if !isValidParentType(parent.Type, schemaEntity.ParentTypes) {
			return fmt.Errorf("invalid parent type %q for entity type %q", parent.Type, entity.UID.Type)
		}
		// If parent is an enum type, validate the ID
		if schemaEnum, ok := v.schema.Enums[parent.Type]; ok {
			if !isValidEnumID(parent, schemaEnum) {
				return fmt.Errorf("invalid enum ID %q for enum type %q", parent.ID, parent.Type)
			}
		}
	}

	// Validate attributes
	if err := checkRecord(entity.Attributes, schemaEntity.Shape); err != nil {
		return fmt.Errorf("attributes: %w", err)
	}

	// Validate tags
	if schemaEntity.Tags == nil {
		if entity.Tags.Len() > 0 {
			return fmt.Errorf("entity type %q does not allow tags", entity.UID.Type)
		}
	} else {
		for _, tagVal := range entity.Tags.All() {
			if err := checkValue(tagVal, schemaEntity.Tags); err != nil {
				return fmt.Errorf("tag value: %w", err)
			}
		}
	}

	return nil
}

func validateEnumEntity(entity types.Entity, schemaEnum resolved.Enum) error {
	if !isValidEnumID(entity.UID, schemaEnum) {
		return fmt.Errorf("invalid enum value %q for enum type %q", entity.UID.ID, entity.UID.Type)
	}

	// Enum entities shouldn't have parents, attributes, or tags
	if entity.Parents.Len() > 0 {
		return fmt.Errorf("enum entity %s should not have parents", entity.UID)
	}
	if entity.Attributes.Len() > 0 {
		return fmt.Errorf("enum entity %s should not have attributes", entity.UID)
	}
	if entity.Tags.Len() > 0 {
		return fmt.Errorf("enum entity %s should not have tags", entity.UID)
	}

	return nil
}

func isValidParentType(parentType types.EntityType, allowedTypes []types.EntityType) bool {
	return slices.Contains(allowedTypes, parentType)
}

func isValidEnumID(uid types.EntityUID, schemaEnum resolved.Enum) bool {
	return slices.Contains(schemaEnum.Values, uid)
}
