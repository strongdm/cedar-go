package schema

import (
	"strings"
	"testing"
)

// mockType is a mock Type implementation used only for testing the panic path
type mockType struct{}

func (mockType) isType() { _ = 0 }

// TestConvertTypeToJSONTypePanic tests the panic path for unknown Type implementations
func TestConvertTypeToJSONTypePanic(t *testing.T) {
	t.Run("panic on unknown Type implementation", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				msg := r.(string)
				if !strings.Contains(msg, "unknown Type implementation") {
					t.Errorf("Expected panic message to contain 'unknown Type implementation', got: %v", msg)
				}
				if !strings.Contains(msg, "mockType") {
					t.Errorf("Expected panic message to contain type name, got: %v", msg)
				}
			} else {
				t.Error("Expected panic but none occurred")
			}
		}()

		// This should panic because mockType is not one of the known implementations
		var m mockType
		convertTypeToJSONType(&m)
	})
}
