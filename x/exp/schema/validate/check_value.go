package validate

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// checkValue validates that a runtime value matches the expected schema type.
func checkValue(v types.Value, expected resolved.IsType) error {
	switch expected := expected.(type) {
	case resolved.StringType:
		if _, ok := v.(types.String); !ok {
			return fmt.Errorf("expected String, got %T", v)
		}
	case resolved.LongType:
		if _, ok := v.(types.Long); !ok {
			return fmt.Errorf("expected Long, got %T", v)
		}
	case resolved.BoolType:
		if _, ok := v.(types.Boolean); !ok {
			return fmt.Errorf("expected Boolean, got %T", v)
		}
	case resolved.EntityType:
		uid, ok := v.(types.EntityUID)
		if !ok {
			return fmt.Errorf("expected EntityUID, got %T", v)
		}
		if uid.Type != types.EntityType(expected) {
			return fmt.Errorf("expected entity type %q, got %q", expected, uid.Type)
		}
	case resolved.SetType:
		set, ok := v.(types.Set)
		if !ok {
			return fmt.Errorf("expected Set, got %T", v)
		}
		for elem := range set.All() {
			if err := checkValue(elem, expected.Element); err != nil {
				return fmt.Errorf("set element: %w", err)
			}
		}
	case resolved.RecordType:
		rec, ok := v.(types.Record)
		if !ok {
			return fmt.Errorf("expected Record, got %T", v)
		}
		return checkRecord(rec, expected)
	case resolved.ExtensionType:
		return checkExtensionValue(v, expected)
	default:
		return fmt.Errorf("unknown schema type %T", expected)
	}
	return nil
}

// checkRecord validates a record against a record schema type.
func checkRecord(rec types.Record, expected resolved.RecordType) error {
	// Check all required attributes are present
	for name, attr := range expected {
		v, ok := rec.Get(name)
		if !ok {
			if !attr.Optional {
				return fmt.Errorf("missing required attribute %q", name)
			}
			continue
		}
		if err := checkValue(v, attr.Type); err != nil {
			return fmt.Errorf("attribute %q: %w", name, err)
		}
	}

	// Check for unexpected attributes (closed record)
	for k := range rec.All() {
		if _, ok := expected[k]; !ok {
			return fmt.Errorf("unexpected attribute %q", k)
		}
	}
	return nil
}

// checkExtensionValue checks that a value matches an extension type.
func checkExtensionValue(v types.Value, expected resolved.ExtensionType) error {
	switch types.Ident(expected) {
	case "ipaddr":
		if _, ok := v.(types.IPAddr); !ok {
			return fmt.Errorf("expected IPAddr, got %T", v)
		}
	case "decimal":
		if _, ok := v.(types.Decimal); !ok {
			return fmt.Errorf("expected Decimal, got %T", v)
		}
	case "datetime":
		if _, ok := v.(types.Datetime); !ok {
			return fmt.Errorf("expected Datetime, got %T", v)
		}
	case "duration":
		if _, ok := v.(types.Duration); !ok {
			return fmt.Errorf("expected Duration, got %T", v)
		}
	default:
		return fmt.Errorf("unknown extension type %q", expected)
	}
	return nil
}
