package resolved

import (
	"testing"
)

func TestPrimitiveKindString(t *testing.T) {
	tests := []struct {
		kind PrimitiveKind
		want string
	}{
		{PrimitiveLong, "Long"},
		{PrimitiveString, "String"},
		{PrimitiveBool, "Bool"},
		{PrimitiveKind(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("PrimitiveKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}
