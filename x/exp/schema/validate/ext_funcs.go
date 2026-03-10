package validate

import "github.com/cedar-policy/cedar-go/types"

type extFuncSig struct {
	argTypes   []cedarType
	returnType cedarType
}

var extFuncTypes = map[types.Path]extFuncSig{
	// Constructors
	"ip":       {argTypes: []cedarType{typeString{}}, returnType: typeExtension{name: "ipaddr"}},
	"decimal":  {argTypes: []cedarType{typeString{}}, returnType: typeExtension{name: "decimal"}},
	"datetime": {argTypes: []cedarType{typeString{}}, returnType: typeExtension{name: "datetime"}},
	"duration": {argTypes: []cedarType{typeString{}}, returnType: typeExtension{name: "duration"}},

	// Decimal methods
	"lessThan":           {argTypes: []cedarType{typeExtension{name: "decimal"}, typeExtension{name: "decimal"}}, returnType: typeBool{}},
	"lessThanOrEqual":    {argTypes: []cedarType{typeExtension{name: "decimal"}, typeExtension{name: "decimal"}}, returnType: typeBool{}},
	"greaterThan":        {argTypes: []cedarType{typeExtension{name: "decimal"}, typeExtension{name: "decimal"}}, returnType: typeBool{}},
	"greaterThanOrEqual": {argTypes: []cedarType{typeExtension{name: "decimal"}, typeExtension{name: "decimal"}}, returnType: typeBool{}},

	// IPAddr methods
	"isIpv4":      {argTypes: []cedarType{typeExtension{name: "ipaddr"}}, returnType: typeBool{}},
	"isIpv6":      {argTypes: []cedarType{typeExtension{name: "ipaddr"}}, returnType: typeBool{}},
	"isLoopback":  {argTypes: []cedarType{typeExtension{name: "ipaddr"}}, returnType: typeBool{}},
	"isMulticast": {argTypes: []cedarType{typeExtension{name: "ipaddr"}}, returnType: typeBool{}},
	"isInRange":   {argTypes: []cedarType{typeExtension{name: "ipaddr"}, typeExtension{name: "ipaddr"}}, returnType: typeBool{}},

	// Datetime methods
	"toDate":        {argTypes: []cedarType{typeExtension{name: "datetime"}}, returnType: typeExtension{name: "datetime"}},
	"toTime":        {argTypes: []cedarType{typeExtension{name: "datetime"}}, returnType: typeExtension{name: "duration"}},
	"offset":        {argTypes: []cedarType{typeExtension{name: "datetime"}, typeExtension{name: "duration"}}, returnType: typeExtension{name: "datetime"}},
	"durationSince": {argTypes: []cedarType{typeExtension{name: "datetime"}, typeExtension{name: "datetime"}}, returnType: typeExtension{name: "duration"}},

	// Duration methods
	"toDays":         {argTypes: []cedarType{typeExtension{name: "duration"}}, returnType: typeLong{}},
	"toHours":        {argTypes: []cedarType{typeExtension{name: "duration"}}, returnType: typeLong{}},
	"toMinutes":      {argTypes: []cedarType{typeExtension{name: "duration"}}, returnType: typeLong{}},
	"toSeconds":      {argTypes: []cedarType{typeExtension{name: "duration"}}, returnType: typeLong{}},
	"toMilliseconds": {argTypes: []cedarType{typeExtension{name: "duration"}}, returnType: typeLong{}},
}
