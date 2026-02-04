package resolved

import "testing"

// TestIsResolvedTypeMarkerMethods verifies that all resolved.Type implementations
// have the isResolvedType() marker method. These are private interface marker
// methods that exist only to satisfy the Type interface. The tests just call
// them to get coverage.
func TestIsResolvedTypeMarkerMethods(t *testing.T) {
	t.Run("Primitive", func(t *testing.T) {
		p := Primitive{Kind: PrimitiveLong}
		p.isResolvedType() // marker method - no return value
	})

	t.Run("Set", func(t *testing.T) {
		s := Set{Element: Primitive{Kind: PrimitiveString}}
		s.isResolvedType() // marker method - no return value
	})

	t.Run("RecordType", func(t *testing.T) {
		r := &RecordType{Attributes: make(map[string]*Attribute)}
		r.isResolvedType() // marker method - no return value
	})

	t.Run("EntityRef", func(t *testing.T) {
		e := EntityRef{EntityType: "User"}
		e.isResolvedType() // marker method - no return value
	})

	t.Run("Extension", func(t *testing.T) {
		ext := Extension{Name: "ipaddr"}
		ext.isResolvedType() // marker method - no return value
	})
}

// TestResolvedTypeInterfaceSatisfaction verifies that all types can be assigned
// to the Type interface variable, ensuring the marker methods satisfy the interface.
func TestResolvedTypeInterfaceSatisfaction(t *testing.T) {
	var _ Type = Primitive{}
	var _ Type = Set{}
	var _ Type = &RecordType{}
	var _ Type = EntityRef{}
	var _ Type = Extension{}
}

// TestPrimitiveKindString tests the String() method on PrimitiveKind.
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
		got := tt.kind.String()
		if got != tt.want {
			t.Errorf("PrimitiveKind(%d).String() = %q, want %q", tt.kind, got, tt.want)
		}
	}
}
