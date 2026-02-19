package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// typeOfExpr infers the type of an expression given a request environment, schema, and capabilities.
// Returns the inferred type, updated capabilities (from `has` guards), and any type error.
func (v *Validator) typeOfExpr(env *requestEnv, expr ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	switch n := expr.(type) {
	case ast.NodeValue:
		ty, err := v.typeOfValue(n.Value)
		return ty, caps, err

	case ast.NodeTypeVariable:
		return typeOfVariable(env, n.Name), caps, nil

	case ast.NodeTypeAnd:
		return v.typeOfAnd(env, n, caps)

	case ast.NodeTypeOr:
		return v.typeOfOr(env, n, caps)

	case ast.NodeTypeNot:
		return v.typeOfNot(env, n, caps)

	case ast.NodeTypeIfThenElse:
		return v.typeOfIfThenElse(env, n, caps)

	case ast.NodeTypeEquals:
		return v.typeOfEquality(env, n.Left, n.Right, caps)

	case ast.NodeTypeNotEquals:
		return v.typeOfEquality(env, n.Left, n.Right, caps)

	case ast.NodeTypeLessThan:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeLessThanOrEqual:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeGreaterThan:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeGreaterThanOrEqual:
		return v.typeOfComparison(env, n.Left, n.Right, caps, expectLong, expectLong)

	case ast.NodeTypeAdd:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeSub:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeMult:
		return v.typeOfArith(env, n.Left, n.Right, caps)

	case ast.NodeTypeNegate:
		return v.typeOfNegate(env, n, caps)

	case ast.NodeTypeIn:
		return v.typeOfIn(env, n, caps)

	case ast.NodeTypeContains:
		return v.typeOfContains(env, n, caps)

	case ast.NodeTypeContainsAll:
		return v.typeOfContainsAllAny(env, n.Left, n.Right, caps)

	case ast.NodeTypeContainsAny:
		return v.typeOfContainsAllAny(env, n.Left, n.Right, caps)

	case ast.NodeTypeIsEmpty:
		return v.typeOfIsEmpty(env, n, caps)

	case ast.NodeTypeLike:
		return v.typeOfLike(env, n, caps)

	case ast.NodeTypeIs:
		return v.typeOfIs(env, n, caps)

	case ast.NodeTypeIsIn:
		return v.typeOfIsIn(env, n, caps)

	case ast.NodeTypeHas:
		return v.typeOfHas(env, n, caps)

	case ast.NodeTypeAccess:
		return v.typeOfAccess(env, n, caps)

	case ast.NodeTypeHasTag:
		return v.typeOfHasTag(env, n, caps)

	case ast.NodeTypeGetTag:
		return v.typeOfGetTag(env, n, caps)

	case ast.NodeTypeRecord:
		return v.typeOfRecord(env, n, caps)

	case ast.NodeTypeSet:
		return v.typeOfSet(env, n, caps)

	case ast.NodeTypeExtensionCall:
		return v.typeOfExtensionCall(env, n, caps)

	default:
		return nil, caps, fmt.Errorf("unknown node type %T", expr)
	}
}

// typeOfValue validates and infers the type of a literal value.
// Entity UIDs are validated against the schema (type must exist).
func (v *Validator) typeOfValue(val types.Value) (cedarType, error) {
	switch val := val.(type) {
	case types.Boolean:
		if val {
			return typeTrue{}, nil
		}
		return typeFalse{}, nil
	case types.Long:
		return typeLong{}, nil
	case types.String:
		return typeString{}, nil
	case types.EntityUID:
		return v.typeOfEntityUID(val)
	case types.Set:
		var elemType cedarType = typeNever{}
		for elem := range val.All() {
			et, err := v.typeOfValue(elem)
			if err != nil {
				return nil, err
			}
			lub, err := v.leastUpperBound(elemType, et)
			if err != nil {
				return typeSet{element: typeNever{}}, nil
			}
			elemType = lub
		}
		return typeSet{element: elemType}, nil
	case types.Record:
		attrs := make(map[types.String]attributeType)
		for k, rv := range val.All() {
			vt, err := v.typeOfValue(rv)
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
func (v *Validator) typeOfEntityUID(uid types.EntityUID) (cedarType, error) {
	et := uid.Type
	if _, ok := v.schema.Entities[et]; ok {
		return typeEntity{lub: singleEntityLUB(et)}, nil
	}
	if _, ok := v.schema.Enums[et]; ok {
		return typeEntity{lub: singleEntityLUB(et)}, nil
	}
	// Check if it's an action type
	if isActionEntity(et) {
		if _, ok := v.schema.Actions[uid]; ok {
			return typeEntity{lub: singleEntityLUB(et)}, nil
		}
		// Action entity type exists if any action of this type exists
		for aUID := range v.schema.Actions {
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

func (v *Validator) typeOfAnd(env *requestEnv, n ast.NodeTypeAnd, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, lCaps, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(lt) {
		return nil, caps, fmt.Errorf("left operand of && must be boolean, got %T", lt)
	}

	// Short-circuit: false && _ → typeFalse (skip RHS type checking but validate entity refs)
	if _, ok := lt.(typeFalse); ok {
		if err := v.validateEntityRefs(n.Right); err != nil {
			return nil, caps, err
		}
		return typeFalse{}, caps, nil
	}

	// RHS gets LHS capabilities
	rt, rCaps, err := v.typeOfExpr(env, n.Right, caps.merge(lCaps))
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

func (v *Validator) typeOfOr(env *requestEnv, n ast.NodeTypeOr, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, lCaps, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(lt) {
		return nil, caps, fmt.Errorf("left operand of || must be boolean, got %T", lt)
	}

	// Short-circuit: true || _ → typeTrue (skip RHS type checking but validate entity refs)
	if _, ok := lt.(typeTrue); ok {
		if err := v.validateEntityRefs(n.Right); err != nil {
			return nil, caps, err
		}
		return typeTrue{}, lCaps, nil
	}

	// RHS does NOT get LHS capabilities
	rt, rCaps, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(rt) {
		return nil, caps, fmt.Errorf("right operand of || must be boolean, got %T", rt)
	}

	// LHS is false → result depends entirely on RHS
	if _, ok := lt.(typeFalse); ok {
		return rt, rCaps, nil
	}
	// RHS is true → always true, propagate RHS caps
	if _, ok := rt.(typeTrue); ok {
		return typeTrue{}, rCaps, nil
	}
	// RHS is false → result depends entirely on LHS
	if _, ok := rt.(typeFalse); ok {
		return lt, lCaps, nil
	}

	// Both unknown → intersect capabilities (only caps guaranteed by both branches)
	return typeBool{}, lCaps.intersect(rCaps), nil
}

func (v *Validator) typeOfNot(env *requestEnv, n ast.NodeTypeNot, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
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

func (v *Validator) typeOfIfThenElse(env *requestEnv, n ast.NodeTypeIfThenElse, caps capabilitySet) (cedarType, capabilitySet, error) {
	condType, condCaps, err := v.typeOfExpr(env, n.If, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isBoolType(condType) {
		return nil, caps, fmt.Errorf("condition of if-then-else must be boolean, got %T", condType)
	}

	thenCaps := caps.merge(condCaps) // then branch gets test capabilities
	// else branch gets only prior capabilities (test was false)

	// Constant condition: skip dead branch type checking but validate entity refs
	if _, ok := condType.(typeFalse); ok {
		if err := v.validateEntityRefs(n.Then); err != nil {
			return nil, caps, err
		}
		return v.typeOfExpr(env, n.Else, caps)
	}
	if _, ok := condType.(typeTrue); ok {
		if err := v.validateEntityRefs(n.Else); err != nil {
			return nil, caps, err
		}
		return v.typeOfExpr(env, n.Then, thenCaps)
	}

	thenType, thenResultCaps, err := v.typeOfExpr(env, n.Then, thenCaps)
	if err != nil {
		return nil, caps, err
	}
	elseType, elseResultCaps, err := v.typeOfExpr(env, n.Else, caps)
	if err != nil {
		return nil, caps, err
	}

	if err := v.checkStrictEntityLUB(thenType, elseType); err != nil {
		return nil, caps, fmt.Errorf("if-then-else branches have incompatible entity types")
	}
	result, err := v.leastUpperBound(thenType, elseType)
	if err != nil {
		return nil, caps, fmt.Errorf("if-then-else branches have incompatible types")
	}
	return result, thenResultCaps.intersect(elseResultCaps), nil
}

func (v *Validator) typeOfEquality(env *requestEnv, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, left, caps)
	if err != nil {
		return nil, caps, err
	}
	rt, _, err := v.typeOfExpr(env, right, caps)
	if err != nil {
		return nil, caps, err
	}
	// Types must be compatible (LUB must exist) for equality to make sense
	if _, err := v.leastUpperBound(lt, rt); err != nil {
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

func (v *Validator) typeOfComparison(env *requestEnv, left, right ast.IsNode, caps capabilitySet, expectLeft, expectRight typeExpectation) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, left, caps)
	if err != nil {
		return nil, caps, err
	}
	if expectLeft != nil {
		if err := expectLeft(lt); err != nil {
			return nil, caps, fmt.Errorf("left operand: %w", err)
		}
	}
	rt, _, err := v.typeOfExpr(env, right, caps)
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

func (v *Validator) typeOfArith(env *requestEnv, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, left, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := lt.(typeLong); !ok {
		return nil, caps, fmt.Errorf("left operand of arithmetic must be Long, got %T", lt)
	}
	rt, _, err := v.typeOfExpr(env, right, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := rt.(typeLong); !ok {
		return nil, caps, fmt.Errorf("right operand of arithmetic must be Long, got %T", rt)
	}
	return typeLong{}, caps, nil
}

func (v *Validator) typeOfNegate(env *requestEnv, n ast.NodeTypeNegate, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeLong); !ok {
		return nil, caps, fmt.Errorf("operand of negation must be Long, got %T", t)
	}
	return typeLong{}, caps, nil
}

func (v *Validator) typeOfIn(env *requestEnv, n ast.NodeTypeIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("left operand of 'in' must be entity, got %T", lt)
	}
	rt, _, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrSetOfEntity(rt) {
		return nil, caps, fmt.Errorf("right operand of 'in' must be entity or set of entities, got %T", rt)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfContains(env *requestEnv, n ast.NodeTypeContains, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	st, ok := lt.(typeSet)
	if !ok {
		return nil, caps, fmt.Errorf("operand of contains must be Set, got %T", lt)
	}
	rt, _, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	// Check element type compatibility
	if _, isNever := st.element.(typeNever); isNever {
		// Strict mode: empty set (Set<Never>) can never contain any element
		if v.strict {
			if _, argNever := rt.(typeNever); !argNever {
				return nil, caps, fmt.Errorf("contains: empty set can never contain element of type %T", rt)
			}
		}
	} else {
		if _, err := v.leastUpperBound(st.element, rt); err != nil {
			return nil, caps, fmt.Errorf("contains: element type incompatible with set element type")
		}
		// Strict mode: entity types must be related
		if err := v.checkStrictEntityLUB(st.element, rt); err != nil {
			return nil, caps, fmt.Errorf("contains: %w", err)
		}
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfContainsAllAny(env *requestEnv, left, right ast.IsNode, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, left, caps)
	if err != nil {
		return nil, caps, err
	}
	lSet, ok := lt.(typeSet)
	if !ok {
		return nil, caps, fmt.Errorf("left operand of containsAll/containsAny must be Set, got %T", lt)
	}
	rt, _, err := v.typeOfExpr(env, right, caps)
	if err != nil {
		return nil, caps, err
	}
	rSet, ok := rt.(typeSet)
	if !ok {
		return nil, caps, fmt.Errorf("right operand of containsAll/containsAny must be Set, got %T", rt)
	}
	// Strict mode: element types must be compatible
	if v.strict {
		if _, err := v.leastUpperBound(lSet.element, rSet.element); err != nil {
			return nil, caps, fmt.Errorf("containsAll/containsAny: element types are incompatible")
		}
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIsEmpty(env *requestEnv, n ast.NodeTypeIsEmpty, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeSet); !ok {
		return nil, caps, fmt.Errorf("operand of isEmpty must be Set, got %T", t)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfLike(env *requestEnv, n ast.NodeTypeLike, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := t.(typeString); !ok {
		return nil, caps, fmt.Errorf("operand of like must be String, got %T", t)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIs(env *requestEnv, n ast.NodeTypeIs, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(t) {
		return nil, caps, fmt.Errorf("operand of is must be entity, got %T", t)
	}

	// If the entity LUB is known, check if the `is` result is statically known
	if et, ok := t.(typeEntity); ok {
		if !slices.Contains(et.lub.elements, n.EntityType) {
			return typeFalse{}, caps, nil
		}
		if len(et.lub.elements) == 1 && et.lub.elements[0] == n.EntityType {
			return typeTrue{}, caps, nil
		}
	}

	return typeBool{}, caps, nil
}

func (v *Validator) typeOfIsIn(env *requestEnv, n ast.NodeTypeIsIn, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("left operand of is...in must be entity, got %T", lt)
	}
	rt, _, err := v.typeOfExpr(env, n.Entity, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(rt) {
		return nil, caps, fmt.Errorf("right operand of is...in must be entity, got %T", rt)
	}
	return typeBool{}, caps, nil
}

func (v *Validator) typeOfHas(env *requestEnv, n ast.NodeTypeHas, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrRecordType(t) {
		return nil, caps, fmt.Errorf("operand of has must be entity or record, got %T", t)
	}

	// Determine precise bool type based on attribute existence
	resultType := v.hasResultType(t, n.Value)

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
func (v *Validator) hasResultType(t cedarType, attr types.String) cedarType {
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
		return v.hasResultTypeEntity(tv.lub, attr)
	case typeAnyEntity:
		return typeBool{} // Can't know
	default:
		return typeBool{}
	}
}

func (v *Validator) hasResultTypeEntity(lub entityLUB, attr types.String) cedarType {
	if len(lub.elements) == 0 {
		return typeBool{}
	}
	anyHas := false
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
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
			if _, ok := v.schema.Entities[et]; ok {
				continue
			}
			if _, ok := v.schema.Enums[et]; ok {
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

func (v *Validator) typeOfAccess(env *requestEnv, n ast.NodeTypeAccess, caps capabilitySet) (cedarType, capabilitySet, error) {
	t, _, err := v.typeOfExpr(env, n.Arg, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityOrRecordType(t) {
		return nil, caps, fmt.Errorf("operand of attribute access must be entity or record, got %T", t)
	}

	attrType := v.lookupAttributeType(t, n.Value)
	if attrType == nil {
		if !v.mayHaveAttr(t, n.Value) {
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

func (v *Validator) typeOfHasTag(env *requestEnv, n ast.NodeTypeHasTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	if !isEntityType(lt) {
		return nil, caps, fmt.Errorf("operand of hasTag must be entity, got %T", lt)
	}

	// Type check the tag key expression
	rt, _, err := v.typeOfExpr(env, n.Right, caps)
	if err != nil {
		return nil, caps, err
	}
	if _, ok := rt.(typeString); !ok {
		return nil, caps, fmt.Errorf("hasTag key must be String, got %T", rt)
	}

	// Return typeFalse if entity doesn't support tags (not an error, just always false)
	if et, ok := lt.(typeEntity); ok {
		if !v.entityHasTags(et.lub) {
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

func (v *Validator) typeOfGetTag(env *requestEnv, n ast.NodeTypeGetTag, caps capabilitySet) (cedarType, capabilitySet, error) {
	lt, _, err := v.typeOfExpr(env, n.Left, caps)
	if err != nil {
		return nil, caps, err
	}
	et, ok := lt.(typeEntity)
	if !ok {
		return nil, caps, fmt.Errorf("operand of getTag must be entity, got %T", lt)
	}

	if !v.entityHasTags(et.lub) {
		return nil, caps, fmt.Errorf("entity type does not support tags")
	}

	varName := exprVarName(n.Left)
	tagKey := tagCapabilityKey(n.Right)
	if varName != "" && tagKey != "" {
		if !caps.has(capability{varName: varName, attr: types.String("__tag:" + tagKey)}) {
			return nil, caps, fmt.Errorf("tag access requires prior hasTag check")
		}
	}

	tagType := v.entityTagType(et.lub)
	return tagType, caps, nil
}

func (v *Validator) typeOfRecord(env *requestEnv, n ast.NodeTypeRecord, caps capabilitySet) (cedarType, capabilitySet, error) {
	attrs := make(map[types.String]attributeType, len(n.Elements))
	for _, elem := range n.Elements {
		elemType, _, err := v.typeOfExpr(env, elem.Value, caps)
		if err != nil {
			return nil, caps, err
		}
		attrs[elem.Key] = attributeType{typ: elemType, required: true}
	}
	return typeRecord{attrs: attrs}, caps, nil
}

func (v *Validator) typeOfSet(env *requestEnv, n ast.NodeTypeSet, caps capabilitySet) (cedarType, capabilitySet, error) {
	// Strict mode: forbid empty set literals
	if v.strict && len(n.Elements) == 0 {
		return nil, caps, fmt.Errorf("empty set literal is not allowed in strict mode")
	}
	var elemType cedarType = typeNever{}
	for _, elem := range n.Elements {
		et, _, err := v.typeOfExpr(env, elem, caps)
		if err != nil {
			return nil, caps, err
		}
		if err := v.checkStrictEntityLUB(elemType, et); err != nil {
			return nil, caps, fmt.Errorf("set elements have incompatible entity types")
		}
		lub, err := v.leastUpperBound(elemType, et)
		if err != nil {
			return nil, caps, fmt.Errorf("set elements have incompatible types")
		}
		elemType = lub
	}
	return typeSet{element: elemType}, caps, nil
}

func (v *Validator) typeOfExtensionCall(env *requestEnv, n ast.NodeTypeExtensionCall, caps capabilitySet) (cedarType, capabilitySet, error) {
	sig, ok := extFuncTypes[n.Name]
	if !ok {
		return nil, caps, fmt.Errorf("unknown extension function %q", n.Name)
	}

	if len(n.Args) != len(sig.argTypes) {
		return nil, caps, fmt.Errorf("extension function %q expects %d arguments, got %d", n.Name, len(sig.argTypes), len(n.Args))
	}

	// Strict mode: extension constructors require string literal arguments
	if v.strict {
		switch n.Name {
		case "ip", "decimal", "datetime", "duration":
			for _, arg := range n.Args {
				nv, ok := arg.(ast.NodeValue)
				if !ok {
					return nil, caps, fmt.Errorf("extension function %q requires a string literal argument in strict mode", n.Name)
				}
				if _, ok := nv.Value.(types.String); !ok {
					return nil, caps, fmt.Errorf("extension function %q requires a string literal argument in strict mode", n.Name)
				}
			}
		}
	}

	for i, arg := range n.Args {
		argType, _, err := v.typeOfExpr(env, arg, caps)
		if err != nil {
			return nil, caps, err
		}
		if !v.isSubtype(argType, sig.argTypes[i]) {
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
	switch nd := n.(type) {
	case ast.NodeTypeVariable:
		return nd.Name
	case ast.NodeTypeAccess:
		parent := exprVarName(nd.Arg)
		if parent != "" {
			return parent + "." + nd.Value
		}
	}
	return ""
}

// validateEntityRefs walks an AST subtree and validates that all entity UID
// references point to types that exist in the schema. This runs on dead code
// branches to catch issues even when full type checking is skipped.
func (v *Validator) validateEntityRefs(n ast.IsNode) error {
	switch nd := n.(type) {
	case ast.NodeValue:
		if uid, ok := nd.Value.(types.EntityUID); ok {
			if _, err := v.typeOfEntityUID(uid); err != nil {
				return err
			}
		}
		if set, ok := nd.Value.(types.Set); ok {
			for elem := range set.All() {
				if uid, ok := elem.(types.EntityUID); ok {
					if _, err := v.typeOfEntityUID(uid); err != nil {
						return err
					}
				}
			}
		}
	case ast.NodeTypeVariable:
		// no entity refs to validate
	case ast.NodeTypeIfThenElse:
		if err := v.validateEntityRefs(nd.If); err != nil {
			return err
		}
		if err := v.validateEntityRefs(nd.Then); err != nil {
			return err
		}
		return v.validateEntityRefs(nd.Else)
	case ast.NodeTypeExtensionCall:
		for _, arg := range nd.Args {
			if err := v.validateEntityRefs(arg); err != nil {
				return err
			}
		}
	case ast.NodeTypeRecord:
		for _, elem := range nd.Elements {
			if err := v.validateEntityRefs(elem.Value); err != nil {
				return err
			}
		}
	case ast.NodeTypeSet:
		for _, elem := range nd.Elements {
			if err := v.validateEntityRefs(elem); err != nil {
				return err
			}
		}
	case ast.NodeTypeAnd:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeOr:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeEquals:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeNotEquals:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeLessThan:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeLessThanOrEqual:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGreaterThan:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGreaterThanOrEqual:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeAdd:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeSub:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeMult:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeIn:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContains:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContainsAll:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeContainsAny:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeHasTag:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeGetTag:
		return v.validateEntityRefsPair(nd.Left, nd.Right)
	case ast.NodeTypeNegate:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeNot:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeIsEmpty:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeHas:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeAccess:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeLike:
		return v.validateEntityRefs(nd.Arg)
	case ast.NodeTypeIs:
		return v.validateEntityRefs(nd.Left)
	case ast.NodeTypeIsIn:
		return v.validateEntityRefsPair(nd.Left, nd.Entity)
	}
	return nil
}

func (v *Validator) validateEntityRefsPair(a, b ast.IsNode) error {
	if err := v.validateEntityRefs(a); err != nil {
		return err
	}
	return v.validateEntityRefs(b)
}

// tagCapabilityKey extracts a string key from a tag expression for capability tracking.
func tagCapabilityKey(n ast.IsNode) types.String {
	if nv, ok := n.(ast.NodeValue); ok {
		if s, ok := nv.Value.(types.String); ok {
			return s
		}
	}
	return ""
}
