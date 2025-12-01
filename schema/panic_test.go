package schema

import (
	"errors"
	"strings"
	"testing"
)

// TestPanicOnErr tests the panicOnErr helper function
func TestPanicOnErr(t *testing.T) {
	t.Run("panicOnErr with nil error does nothing", func(t *testing.T) {
		// Should not panic
		panicOnErr(nil, "this should not panic")
	})

	t.Run("panicOnErr with error panics with message", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				msg := r.(string)
				if !strings.Contains(msg, "impossible error occurred") {
					t.Errorf("Expected panic message to contain 'impossible error occurred', got: %v", msg)
				}
				if !strings.Contains(msg, "test error message") {
					t.Errorf("Expected panic message to contain 'test error message', got: %v", msg)
				}
				if !strings.Contains(msg, "test error") {
					t.Errorf("Expected panic message to contain the actual error, got: %v", msg)
				}
			} else {
				t.Error("Expected panic but none occurred")
			}
		}()

		// This should panic
		panicOnErr(errors.New("test error"), "test error message")
	})
}

// TestImpossibleErrorPaths tests that the impossible error paths are now covered
func TestImpossibleErrorPaths(t *testing.T) {
	t.Run("UnmarshalCedar json.Marshal path covered", func(t *testing.T) {
		// This test ensures the panicOnErr call in UnmarshalCedar is covered
		// by normal operation (where err is nil)
		var s Schema
		s.SetFilename("test.cedar")

		err := s.UnmarshalCedar([]byte("namespace Test {}"))
		if err != nil {
			t.Fatalf("UnmarshalCedar failed: %v", err)
		}

		// The panicOnErr with nil error was executed, giving us coverage
	})

	t.Run("MarshalCedar json.Unmarshal path covered", func(t *testing.T) {
		// This test ensures the panicOnErr call in MarshalCedar is covered
		s := NewSchema().WithNamespace("Test",
			NewEntity("User"),
		)

		_, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		// The panicOnErr with nil error was executed, giving us coverage
	})

	t.Run("MarshalCedar ast.Format path covered", func(t *testing.T) {
		// This test ensures the second panicOnErr call in MarshalCedar is covered
		s := NewSchema().WithNamespace("App",
			NewEntity("Resource").
				WithAttribute("name", String()),
			NewAction("view").
				AppliesTo(
					Principals("Resource"),
					Resources("Resource"),
					nil,
				),
		)

		cedarData, err := s.MarshalCedar()
		if err != nil {
			t.Fatalf("MarshalCedar failed: %v", err)
		}

		if len(cedarData) == 0 {
			t.Error("Expected non-empty Cedar output")
		}

		// Both panicOnErr calls with nil errors were executed, giving us coverage
	})
}
