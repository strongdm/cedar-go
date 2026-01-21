package ast

import (
	"testing"
)

// TestEnumNodeEntityUIDsEarlyTermination tests early termination of EntityUIDs iterator
func TestEnumNodeEntityUIDsEarlyTermination(t *testing.T) {
	t.Parallel()

	t.Run("EnumNode.EntityUIDs early break", func(t *testing.T) {
		enum := Enum("Status", "active", "inactive", "pending")

		count := 0
		for range enum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}

		if count != 1 {
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})

	t.Run("ResolvedEnum.EntityUIDs early break", func(t *testing.T) {
		schema := NewSchema(Enum("Status", "active", "inactive", "pending"))
		resolved, err := schema.Resolve()
		if err != nil {
			t.Fatalf("Resolve() error = %v", err)
		}

		resolvedEnum := resolved.Enums["Status"]
		count := 0
		for range resolvedEnum.EntityUIDs() {
			count++
			if count == 1 {
				break
			}
		}

		if count != 1 {
			t.Errorf("expected to break after 1 iteration, got %d", count)
		}
	})
}

