package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// typeOfExpr infers the type of an expression given a request environment, schema, and capabilities.
// Returns the inferred type, updated capabilities (from `has` guards), and any type error.
func typeOfExpr(env *requestEnv, s *resolved.Schema, expr ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	switch n := expr.(type) {
	case ast.NodeValue:
		ty, err := typeOfValue(s, n.Value)
		return ty, caps, err

	case ast.NodeTypeVariable:
		return typeOfVariable(env, n.Name), caps, nil

	case ast.NodeTypeAnd:
		return typeOfAnd(env, s, n, caps)

	case ast.NodeTypeOr:
		return typeOfOr(env, s, n, caps)

	case ast.NodeTypeNot:
		return typeOfNot(env, s, n, caps)

	case ast.NodeTypeIfThenElse:
		return typeOfIfThenElse(env, s, n, caps)

	case ast.NodeTypeEquals:
		return typeOfEquality(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeNotEquals:
		return typeOfEquality(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeLessThan:
		return typeOfComparison(env, s, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeLessThanOrEqual:
		return typeOfComparison(env, s, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeGreaterThan:
		return typeOfComparison(env, s, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeGreaterThanOrEqual:
		return typeOfComparison(env, s, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeAdd:
		return typeOfArith(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeSub:
		return typeOfArith(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeMult:
		return typeOfArith(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeNegate:
		return typeOfNegate(env, s, n, caps)

	case ast.NodeTypeIn:
		return typeOfIn(env, s, n, caps)

	case ast.NodeTypeContains:
		return typeOfContains(env, s, n, caps)

	case ast.NodeTypeContainsAll:
		return typeOfContainsAllAny(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeContainsAny:
		return typeOfContainsAllAny(env, s, n.Left, n.Right, caps)

	case ast.NodeTypeIsEmpty:
		return typeOfIsEmpty(env, s, n, caps)

	case ast.NodeTypeLike:
		return typeOfLike(env, s, n, caps)

	case ast.NodeTypeIs:
		return typeOfIs(env, s, n, caps)

	case ast.NodeTypeIsIn:
		return typeOfIsIn(env, s, n, caps)

	case ast.NodeTypeHas:
		return typeOfHas(env, s, n, caps)

	case ast.NodeTypeAccess:
		return typeOfAccess(env, s, n, caps)

	case ast.NodeTypeHasTag:
		return typeOfHasTag(env, s, n, caps)

	case ast.NodeTypeGetTag:
		return typeOfGetTag(env, s, n, caps)

	case ast.NodeTypeRecord:
		return typeOfRecord(env, s, n, caps)

	case ast.NodeTypeSet:
		return typeOfSet(env, s, n, caps)

	case ast.NodeTypeExtensionCall:
		return typeOfExtensionCall(env, s, n, caps)

	default:
		return nil, caps, fmt.Errorf("unknown node type %T", expr)
	}
}

// typeOfValue validates and infers the type of a literal value.
// Entity UIDs are validated against the schema (type must exist).
func typeOfValue(s *resolved.Schema, v types.Value) (cedarType, error) {
	switch v := v.(type) {
	case types.Boolean:
		if v {
			return typeTrue{}, nil
		}
		return typeFalse{}, nil
	case types.Long:
		return typeLong{}, nil
	case types.String:
		return typeString{}, nil
	case types.EntityUID:
		return typeOfEntityUID(s, v)
	case types.Set:
		var elemType cedarType = typeNever{}
		for elem := range v.All() {
			et, err := typeOfValue(s, elem)
			if err != nil {
				return nil, err
			}
			lub, err := leastUpperBound(elemType, et)
			if err != nil {
				return typeSet{element: typeNever{}}, nil
			}
			elemType = lub
		}
		return typeSet{element: elemType}, nil
	case types.Record:
		attrs := make(map[types.String]attributeType)
		for k, val := range v.All() {
			vt, err := typeOfValue(s, val)
			if err != nil {
				return nil, err
			}
			attrs[k] = attributeType{typ: vt, required: true}
		}
		return typeRecord{attrs: attrs}, nil
	case types.IPAddr:
		return typeExtension{"ipaddr"}, nil
	case types.Decimal:
		return typeExtension{"decimal"}, nil
	case types.Datetime:
		return typeExtension{"datetime"}, nil
	case types.Duration:
		return typeExtension{"duration"}, nil
	default:
		return typeNever{}, nil
	}
}

// typeOfEntityUID validates an entity UID's type exists in the schema.
func typeOfEntityUID(s *resolved.Schema, uid types.EntityUID) (cedarType, error) {
	et := uid.Type
	if _, ok := s.Entities[et]; ok {
		return typeEntity{lub: singleEntityLUB(et)}, nil
	}
	if _, ok := s.Enums[et]; ok {
		return typeEntity{lub: singleEntityLUB(et)}, nil
	}
	// Check if it's an action type
	if isActionEntity(et) {
		if _, ok := s.Actions[uid]; ok {
			return typeEntity{lub: singleEntityLUB(et)}, nil
		}
		// Action entity type exists if any action of this type exists
		for aUID := range s.Actions {
			if aUID.Type == et {
				return typeEntity{lub: singleEntityLUB(et)}, nil
			}
		}
	}
	return nil, fmt.Errorf("entity type %q not found in schema", et)
}

func typeOfVariable(env *requestEnv, name types.String) cedarType {
	switch name {
	case "principal":
		return typeEntity{lub: singleEntityLUB(env.principalType)}
	case "action":
		return typeEntity{lub: singleEntityLUB(env.actionUID.Type)}
	case "resource":
		return typeEntity{lub: singleEntityLUB(env.resourceType)}
	case "context":
		return env.contextType
	default:
		return typeNever{}
	}
}

func typeOfAnd(env *requestEnv, s *resolved.Schema, n ast.NodeTypeAnd, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, lCaps, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(lt) {
		return nil, caps, fmt.Errorf("left operand of && must be boolean, got %T", lt)
	}

	// Short-circuit: false && _ → typeFalse (skip RHS type checking but validate entity refs)
	if _, ok := lt.(typeFalse); ok {
		if err := validateEntityRefs(s, n.Right); err != nil {
			return nil, caps, err
		}
		return typeFalse{}, caps, nil
	}

	// RHS gets LHS capabilities
	rt, rCaps, err := typeOfExpr(env, s, n.Right, caps.merge(lCaps))
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(rt) {
		return nil, caps, fmt.Errorf("right operand of && must be boolean, got %T", rt)
	}

	// Propagate precise type: true && false → false, true && true → true
	if _, ok := lt.(typeTrue); ok {
		return rt, rCaps, nil
	}
	// false && true → false
	if _, ok := rt.(typeFalse); ok {
		return typeFalse{}, rCaps, nil
	}

	return typeBool{}, rCaps, nil
}

func typeOfOr(env *requestEnv, s *resolved.Schema, n ast.NodeTypeOr, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(lt) {
		return nil, caps, fmt.Errorf("left operand of || must be boolean, got %T", lt)
	}

	// Short-circuit: true || _ → typeTrue (skip RHS type checking but validate entity refs)
	if _, ok := lt.(typeTrue); ok {
		if err := validateEntityRefs(s, n.Right); err != nil {
			return nil, caps, err
		}
		return typeTrue{}, caps, nil
	}

	// RHS does NOT get LHS capabilities
	rt, _, err := typeOfExpr(env, s, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(rt) {
		return nil, caps, fmt.Errorf("right operand of || must be boolean, got %T", rt)
	}

	return typeBool{}, caps, nil
}

func typeOfNot(env *requestEnv, s *resolved.Schema, n ast.NodeTypeNot, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(t) {
		return nil, caps, fmt.Errorf("operand of ! must be boolean, got %T", t)
	}
	switch t.(type) {
	case typeTrue:
		return typeFalse{}, caps, nil
	case typeFalse:
		return typeTrue{}, caps, nil
	default:
		return typeBool{}, caps, nil
	}
}

func typeOfIfThenElse(env *requestEnv, s *resolved.Schema, n ast.NodeTypeIfThenElse, caps capabilitySet) (cedarType, capabilitySet, error) {
	condType, condCaps, err := typeOfExpr(env, s, n.If, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(condType) {
		return nil, caps, fmt.Errorf("condition of if-then-else must be boolean, got %T", condType)
	}

	branchCaps := caps.merge(condCaps)

	// Constant condition: skip dead branch type checking but validate entity refs
	if _, ok := condType.(typeFalse); ok {
		if err := validateEntityRefs(s, n.Then); err != nil {
			return nil, caps, err
		}
		return typeOfExpr(env, s, n.Else, branchCaps)
	}
	if _, ok := condType.(typeTrue); ok {
		if err := validateEntityRefs(s, n.Else); err != nil {
			return nil, caps, err
		}
		return typeOfExpr(env, s, n.Then, branchCaps)
	}

	thenType, _, err := typeOfExpr(env, s, n.Then, branchCaps)
	if err != nil {
		return nil, caps, err
	}
	elseType, _, err := typeOfExpr(env, s, n.Else, branchCaps)
	if err != nil {
		return nil, caps, err
	}

	if err := checkStrictEntityLUB(s, thenType, elseType); err != nil {
		return nil, caps, fmt.Errorf("if-then-else branches have incompatible entity types")
	}
	result, err := leastUpperBound(thenType, elseType)
	if err != nil {
		return nil, caps, fmt.Errorf("if-then-else branches have incompatible types")
	}
	return result, caps, nil
}

func typeOfEquality(env *requestEnv, s *resolved.Schema, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, left, caps)
	if err != nil {
		return nil, caps, err
	}
	rt, _, err := typeOfExpr(env, s, right, caps)
	if err != nil {
		return nil, caps, err
	}
	// Types must be compatible (LUB must exist) for equality to make sense
	if _, err := leastUpperBound(lt, rt); err != nil {
		return nil, caps, fmt.Errorf("equality comparison between incompatible types %T and %T", lt, rt)
	}
	return typeBool{}, caps, nil
}

type typeExpectation func(cedarType) error

var expectLong typeExpectation = func(t cedarType) error {
	if _, ok := t.(typeLong); !ok {
		return fmt.Errorf("expected Long, got %T", t)
	}
	return nil
}

func typeOfComparison(env *requestEnv, s *resolved.Schema, left, right ast.IsNode, caps capabilitySet, expectLeft, expectRight typeExpectation) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, left, caps)
	if err != nil {
		return nil, caps, err
	}
	if expectLeft != nil {
		if err := expectLeft(lt); err != nil {
			return nil, caps, fmt.Errorf("left operand: %w", err)
		}
	}
	rt, _, err := typeOfExpr(env, s, right, caps)
	if err != nil {
		return nil, caps, err
	}
	if expectRight != nil {
		if err := expectRight(rt); err != nil {
			return nil, caps, fmt.Errorf("right operand: %w", err)
		}
	}
	return typeBool{}, caps, nil
}

func typeOfArith(env *requestEnv, s *resolved.Schema, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, left, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := lt.(typeLong); !ok {
		return nil, caps, fmt.Errorf("left operand of arithmetic must be Long, got %T", lt)
	}
	rt, _, err := typeOfExpr(env, s, right, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := rt.(typeLong); !ok {
		return nil, caps, fmt.Errorf("right operand of arithmetic must be Long, got %T", rt)
	}
	return typeLong{}, caps, nil
}

func typeOfNegate(env *requestEnv, s *resolved.Schema, n ast.NodeTypeNegate, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeLong); !ok {
		return nil, caps, fmt.Errorf("operand of negation must be Long, got %T", t)
	}
	return typeLong{}, caps, nil
}

func typeOfIn(env *requestEnv, s *resolved.Schema, n ast.NodeTypeIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("left operand of 'in' must be entity, got %T", lt)
	}
	rt, _, err := typeOfExpr(env, s, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrSetOfEntity(rt) {
		return nil, caps, fmt.Errorf("right operand of 'in' must be entity or set of entities, got %T", rt)
	}
	return typeBool{}, caps, nil
}

func typeOfContains(env *requestEnv, s *resolved.Schema, n ast.NodeTypeContains, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	st, ok := lt.(typeSet)
	if !ok {
		return nil, caps, fmt.Errorf("operand of contains must be Set, got %T", lt)
	}
	rt, _, err := typeOfExpr(env, s, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	// Strict mode: check element type compatibility
	if _, isNever := st.element.(typeNever); isNever {
		// Empty set (Set<Never>) can never contain any element — strict mode error
		if _, argNever := rt.(typeNever); !argNever {
			return nil, caps, fmt.Errorf("contains: empty set can never contain element of type %T", rt)
		}
	} else {
		if _, err := leastUpperBound(st.element, rt); err != nil {
			return nil, caps, fmt.Errorf("contains: element type incompatible with set element type")
		}
		// Strict mode: entity types must be related
		if err := checkStrictEntityLUB(s, st.element, rt); err != nil {
			return nil, caps, fmt.Errorf("contains: %w", err)
		}
	}
	return typeBool{}, caps, nil
}

func typeOfContainsAllAny(env *requestEnv, s *resolved.Schema, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, left, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := lt.(typeSet); !ok {
		return nil, caps, fmt.Errorf("left operand of containsAll/containsAny must be Set, got %T", lt)
	}
	rt, _, err := typeOfExpr(env, s, right, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := rt.(typeSet); !ok {
		return nil, caps, fmt.Errorf("right operand of containsAll/containsAny must be Set, got %T", rt)
	}
	return typeBool{}, caps, nil
}

func typeOfIsEmpty(env *requestEnv, s *resolved.Schema, n ast.NodeTypeIsEmpty, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeSet); !ok {
		return nil, caps, fmt.Errorf("operand of isEmpty must be Set, got %T", t)
	}
	return typeBool{}, caps, nil
}

func typeOfLike(env *requestEnv, s *resolved.Schema, n ast.NodeTypeLike, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeString); !ok {
		return nil, caps, fmt.Errorf("operand of like must be String, got %T", t)
	}
	return typeBool{}, caps, nil
}

func typeOfIs(env *requestEnv, s *resolved.Schema, n ast.NodeTypeIs, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(t) {
		return nil, caps, fmt.Errorf("operand of is must be entity, got %T", t)
	}

	// If the entity LUB is known and doesn't include the tested type, always false
	if et, ok := t.(typeEntity); ok {
		if !slices.Contains(et.lub.elements, n.EntityType) {
			return typeFalse{}, caps, nil
		}
	}

	return typeBool{}, caps, nil
}

func typeOfIsIn(env *requestEnv, s *resolved.Schema, n ast.NodeTypeIsIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("left operand of is...in must be entity, got %T", lt)
	}
	rt, _, err := typeOfExpr(env, s, n.Entity, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(rt) {
		return nil, caps, fmt.Errorf("right operand of is...in must be entity, got %T", rt)
	}
	return typeBool{}, caps, nil
}

func typeOfHas(env *requestEnv, s *resolved.Schema, n ast.NodeTypeHas, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrRecordType(t) {
		return nil, caps, fmt.Errorf("operand of has must be entity or record, got %T", t)
	}

	// Determine precise bool type based on attribute existence
	resultType := hasResultType(s, t, n.Value)

	// For entity types with required/optional attributes that returned typeBool,
	// check if the entity is already known to exist via a prior capability.
	// If so, we can upgrade to typeTrue for required attributes.
	if _, isBool := resultType.(typeBool); isBool {
		if varName := exprVarName(n.Arg); varName != "" {
			if caps.has(capability{varName: varName, attr: n.Value}) {
				resultType = typeTrue{}
			}
		}
	}

	newCaps := caps
	// Add capability based on the expression
	if varName := exprVarName(n.Arg); varName != "" {
		newCaps = caps.add(capability{varName: varName, attr: n.Value})
	}

	return resultType, newCaps, nil
}

// hasResultType returns the precise bool type for a `has` check.
func hasResultType(s *resolved.Schema, t cedarType, attr types.String) cedarType {
	switch tv := t.(type) {
	case typeRecord:
		if tv.openAttributes {
			return typeBool{}
		}
		a, ok := tv.attrs[attr]
		if !ok {
			return typeFalse{} // Closed record, attr definitely doesn't exist
		}
		if a.required {
			return typeTrue{} // Required attr always exists
		}
		return typeBool{} // Optional attr
	case typeEntity:
		return hasResultTypeEntity(s, tv.lub, attr)
	case typeAnyEntity:
		return typeBool{} // Can't know
	default:
		return typeBool{}
	}
}

func hasResultTypeEntity(s *resolved.Schema, lub entityLUB, attr types.String) cedarType {
	if len(lub.elements) == 0 {
		return typeBool{}
	}
	anyHas := false
	for _, et := range lub.elements {
		entity, ok := s.Entities[et]
		if !ok {
			continue
		}
		if _, ok := entity.Shape[attr]; ok {
			anyHas = true
		}
	}
	if !anyHas {
		// Check if all entity types are known and none have the attr
		allKnown := true
		for _, et := range lub.elements {
			if _, ok := s.Entities[et]; ok {
				continue
			}
			if _, ok := s.Enums[et]; ok {
				continue
			}
			if isActionEntity(et) {
				continue // Action entities are known but have no attributes
			}
			allKnown = false
			break
		}
		if allKnown {
			return typeFalse{} // Attribute definitely doesn't exist on any type
		}
		return typeBool{}
	}
	// For entity types, we can't conclude `has` is true even for required attributes,
	// because the entity might not exist in the entity store at runtime (`has` returns
	// false for non-existent entities). Only return typeBool.
	return typeBool{}
}

func typeOfAccess(env *requestEnv, s *resolved.Schema, n ast.NodeTypeAccess, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := typeOfExpr(env, s, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrRecordType(t) {
		return nil, caps, fmt.Errorf("operand of attribute access must be entity or record, got %T", t)
	}

	attrType := lookupAttributeType(s, t, n.Value)
	if attrType == nil {
		if !mayHaveAttr(s, t, n.Value) {
			return nil, caps, fmt.Errorf("attribute %q not found on type", n.Value)
		}
		return typeNever{}, caps, nil
	}

	// Check if the attribute is optional and requires a `has` guard
	if !attrType.required {
		varName := exprVarName(n.Arg)
		if varName == "" || !caps.has(capability{varName: varName, attr: n.Value}) {
			return nil, caps, fmt.Errorf("attribute %q is optional and may not be present; use `has` to check first", n.Value)
		}
	}

	return attrType.typ, caps, nil
}

func typeOfHasTag(env *requestEnv, s *resolved.Schema, n ast.NodeTypeHasTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("operand of hasTag must be entity, got %T", lt)
	}

	// Type check the tag key expression
	rt, _, err := typeOfExpr(env, s, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := rt.(typeString); !ok {
		return nil, caps, fmt.Errorf("hasTag key must be String, got %T", rt)
	}

	// Return typeFalse if entity doesn't support tags (not an error, just always false)
	if et, ok := lt.(typeEntity); ok {
		if !entityHasTags(s, et.lub) {
			return typeFalse{}, caps, nil
		}
	}

	newCaps := caps
	if varName := exprVarName(n.Left); varName != "" {
		tagKey := tagCapabilityKey(n.Right)
		if tagKey != "" {
			newCaps = caps.add(capability{varName: varName, attr: types.String("__tag:" + tagKey)})
		}
	}

	return typeBool{}, newCaps, nil
}

func typeOfGetTag(env *requestEnv, s *resolved.Schema, n ast.NodeTypeGetTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := typeOfExpr(env, s, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("operand of getTag must be entity, got %T", lt)
	}

	if et, ok := lt.(typeEntity); ok {
		if !entityHasTags(s, et.lub) {
			return nil, caps, fmt.Errorf("entity type does not support tags")
		}

		varName := exprVarName(n.Left)
		tagKey := tagCapabilityKey(n.Right)
		if varName != "" && tagKey != "" {
			if !caps.has(capability{varName: varName, attr: types.String("__tag:" + tagKey)}) {
				return nil, caps, fmt.Errorf("tag access requires prior hasTag check")
			}
		}

		tagType := entityTagType(s, et.lub)
		return tagType, caps, nil
	}

	return typeNever{}, caps, nil
}

func typeOfRecord(env *requestEnv, s *resolved.Schema, n ast.NodeTypeRecord, caps capabilitySet) (cedarType, capabilitySet, error) {
	attrs := make(map[types.String]attributeType, len(n.Elements))
	for _, elem := range n.Elements {
		elemType, _, err := typeOfExpr(env, s, elem.Value, caps)
		if err != nil {
			return nil, caps, err
		}
		attrs[elem.Key] = attributeType{typ: elemType, required: true}
	}
	return typeRecord{attrs: attrs}, caps, nil
}

func typeOfSet(env *requestEnv, s *resolved.Schema, n ast.NodeTypeSet, caps capabilitySet) (cedarType, capabilitySet, error) {
	var elemType cedarType = typeNever{}
	for _, elem := range n.Elements {
		et, _, err := typeOfExpr(env, s, elem, caps)
		if err != nil {
			return nil, caps, err
		}
		// Strict mode: entity types must be related
		if err := checkStrictEntityLUB(s, elemType, et); err != nil {
			return nil, caps, fmt.Errorf("set elements have incompatible entity types")
		}
		lub, err := leastUpperBound(elemType, et)
		if err != nil {
			return nil, caps, fmt.Errorf("set elements have incompatible types")
		}
		elemType = lub
	}
	return typeSet{element: elemType}, caps, nil
}

func typeOfExtensionCall(env *requestEnv, s *resolved.Schema, n ast.NodeTypeExtensionCall, caps capabilitySet) (cedarType, capabilitySet, error) {
	sig, ok := extFuncTypes[n.Name]
	if !ok {
		return nil, caps, fmt.Errorf("unknown extension function %q", n.Name)
	}

	if len(n.Args) != len(sig.argTypes) {
		return nil, caps, fmt.Errorf("extension function %q expects %d arguments, got %d", n.Name, len(sig.argTypes), len(n.Args))
	}

	for i, arg := range n.Args {
		argType, _, err := typeOfExpr(env, s, arg, caps)
		if err != nil {
			return nil, caps, err
		}
		if !isSubtype(argType, sig.argTypes[i]) {
			return nil, caps, fmt.Errorf("extension function %q argument %d: expected %T, got %T", n.Name, i, sig.argTypes[i], argType)
		}
	}

	return sig.returnType, caps, nil
}

func isBoolType(t cedarType) bool {
	switch t.(type) {
	case typeBool, typeTrue, typeFalse:
		return true
	}
	return false
}

func isEntityType(t cedarType) bool {
	switch t.(type) {
	case typeEntity, typeAnyEntity:
		return true
	}
	return false
}

func isEntityOrRecordType(t cedarType) bool {
	switch t.(type) {
	case typeEntity, typeAnyEntity, typeRecord:
		return true
	}
	return false
}

func isEntityOrSetOfEntity(t cedarType) bool {
	if isEntityType(t) {
		return true
	}
	if st, ok := t.(typeSet); ok {
		return isEntityType(st.element)
	}
	return false
}

// exprVarName extracts a variable name from an expression if it is a simple variable reference
// or a chain of accesses on a variable.
func exprVarName(n ast.IsNode) types.String {
	switch v := n.(type) {
	case ast.NodeTypeVariable:
		return v.Name
	case ast.NodeTypeAccess:
		parent := exprVarName(v.Arg)
		if parent != "" {
			return parent + "." + v.Value
		}
	}
	return ""
}

// validateEntityRefs walks an AST subtree and validates that all entity UID
// references point to types that exist in the schema. This runs on dead code
// branches to catch issues even when full type checking is skipped.
func validateEntityRefs(s *resolved.Schema, n ast.IsNode) error {
	switch v := n.(type) {
	case ast.NodeValue:
		if uid, ok := v.Value.(types.EntityUID); ok {
			if _, err := typeOfEntityUID(s, uid); err != nil {
				return err
			}
		}
		if set, ok := v.Value.(types.Set); ok {
			for elem := range set.All() {
				if uid, ok := elem.(types.EntityUID); ok {
					if _, err := typeOfEntityUID(s, uid); err != nil {
						return err
					}
				}
			}
		}
	case ast.NodeTypeVariable:
		// no entity refs to validate
	case ast.NodeTypeIfThenElse:
		if err := validateEntityRefs(s, v.If); err != nil {
			return err
		}
		if err := validateEntityRefs(s, v.Then); err != nil {
			return err
		}
		return validateEntityRefs(s, v.Else)
	case ast.NodeTypeExtensionCall:
		for _, arg := range v.Args {
			if err := validateEntityRefs(s, arg); err != nil {
				return err
			}
		}
	case ast.NodeTypeRecord:
		for _, elem := range v.Elements {
			if err := validateEntityRefs(s, elem.Value); err != nil {
				return err
			}
		}
	case ast.NodeTypeSet:
		for _, elem := range v.Elements {
			if err := validateEntityRefs(s, elem); err != nil {
				return err
			}
		}
	case ast.NodeTypeAnd:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeOr:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeEquals:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeNotEquals:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeLessThan:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeLessThanOrEqual:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeGreaterThan:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeGreaterThanOrEqual:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeAdd:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeSub:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeMult:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeIn:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeContains:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeContainsAll:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeContainsAny:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeHasTag:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeGetTag:
		return validateEntityRefsPair(s, v.Left, v.Right)
	case ast.NodeTypeNegate:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeNot:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeIsEmpty:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeHas:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeAccess:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeLike:
		return validateEntityRefs(s, v.Arg)
	case ast.NodeTypeIs:
		return validateEntityRefs(s, v.Left)
	case ast.NodeTypeIsIn:
		return validateEntityRefsPair(s, v.Left, v.Entity)
	}
	return nil
}

func validateEntityRefsPair(s *resolved.Schema, a, b ast.IsNode) error {
	if err := validateEntityRefs(s, a); err != nil {
		return err
	}
	return validateEntityRefs(s, b)
}

// tagCapabilityKey extracts a string key from a tag expression for capability tracking.
func tagCapabilityKey(n ast.IsNode) types.String {
	if v, ok := n.(ast.NodeValue); ok {
		if s, ok := v.Value.(types.String); ok {
			return s
		}
	}
	return ""
}
