package ast

import (
	"encoding/json"

	"github.com/cedar-policy/cedar-go/types"
)

type policyJSON struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Effect      string            `json:"effect"`
	Principal   scopeJSON         `json:"principal"`
	Action      scopeJSON         `json:"action"`
	Resource    scopeJSON         `json:"resource"`
	Conditions  []conditionJSON   `json:"conditions,omitempty"`
}

type inJSON struct {
	Entity types.EntityUID `json:"entity"`
}

type scopeJSON struct {
	Op         string            `json:"op"`
	Entity     *types.EntityUID  `json:"entity,omitempty"`
	Entities   []types.EntityUID `json:"entities,omitempty"`
	EntityType string            `json:"entity_type,omitempty"`
	In         *inJSON           `json:"in,omitempty"`
}

type conditionJSON struct {
	Kind string   `json:"kind"`
	Body nodeJSON `json:"body"`
}

type binaryJSON struct {
	Left  nodeJSON `json:"left"`
	Right nodeJSON `json:"right"`
}

type unaryJSON struct {
	Arg nodeJSON `json:"arg"`
}

type strJSON struct {
	Left nodeJSON `json:"left"`
	Attr string   `json:"attr"`
}

type isJSON struct {
	Left       nodeJSON `json:"left"`
	EntityType string   `json:"entity_type"`
	In         *inJSON  `json:"in,omitempty"`
}

type ifThenElseJSON struct {
	If   nodeJSON `json:"if"`
	Then nodeJSON `json:"then"`
	Else nodeJSON `json:"else"`
}

type arrayJSON []nodeJSON

type recordJSON map[string]nodeJSON

type nodeJSON struct {
	// Value
	Value *json.RawMessage `json:"Value,omitempty"` // could be any

	// Var
	Var *string `json:"Var,omitempty"`

	// Slot
	// Unknown

	// ! or neg operators
	Not    *unaryJSON `json:"!,omitempty"`
	Negate *unaryJSON `json:"neg,omitempty"`

	// Binary operators: ==, !=, in, <, <=, >, >=, &&, ||, +, -, *, contains, containsAll, containsAny
	Equals             *binaryJSON `json:"==,omitempty"`
	NotEquals          *binaryJSON `json:"!=,omitempty"`
	In                 *binaryJSON `json:"in,omitempty"`
	LessThan           *binaryJSON `json:"<,omitempty"`
	LessThanOrEqual    *binaryJSON `json:"<=,omitempty"`
	GreaterThan        *binaryJSON `json:">,omitempty"`
	GreaterThanOrEqual *binaryJSON `json:">=,omitempty"`
	And                *binaryJSON `json:"&&,omitempty"`
	Or                 *binaryJSON `json:"||,omitempty"`
	Plus               *binaryJSON `json:"+,omitempty"`
	Minus              *binaryJSON `json:"-,omitempty"`
	Times              *binaryJSON `json:"*,omitempty"`
	Contains           *binaryJSON `json:"contains,omitempty"`
	ContainsAll        *binaryJSON `json:"containsAll,omitempty"`
	ContainsAny        *binaryJSON `json:"containsAny,omitempty"`

	// ., has
	Access *strJSON `json:".,omitempty"`
	Has    *strJSON `json:"has,omitempty"`

	// is
	Is *isJSON `json:"is,omitempty"`

	// like
	Like *strJSON `json:"like,omitempty"`

	// if-then-else
	IfThenElse *ifThenElseJSON `json:"if-then-else,omitempty"`

	// Set
	Set arrayJSON `json:"Set,omitempty"`

	// Record
	Record recordJSON `json:"Record,omitempty"`

	// Any other function: decimal, ip
	Decimal arrayJSON `json:"decimal,omitempty"`
	IP      arrayJSON `json:"ip,omitempty"`

	// Any other method: lessThan, lessThanOrEqual, greaterThan, greaterThanOrEqual, isIpv4, isIpv6, isLoopback, isMulticast, isInRange
	LessThanExt           arrayJSON `json:"lessThan,omitempty"`
	LessThanOrEqualExt    arrayJSON `json:"lessThanOrEqual,omitempty"`
	GreaterThanExt        arrayJSON `json:"greaterThan,omitempty"`
	GreaterThanOrEqualExt arrayJSON `json:"greaterThanOrEqual,omitempty"`
	IsIpv4Ext             arrayJSON `json:"isIpv4,omitempty"`
	IsIpv6Ext             arrayJSON `json:"isIpv6,omitempty"`
	IsLoopbackExt         arrayJSON `json:"isLoopback,omitempty"`
	IsMulticastExt        arrayJSON `json:"isMulticast,omitempty"`
	IsInRangeExt          arrayJSON `json:"isInRange,omitempty"`
}