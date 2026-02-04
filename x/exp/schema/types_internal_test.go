package schema

import "testing"

// TestIsTypeMarkerMethods verifies that all Type implementations have the isType()
// marker method. These are private interface marker methods that exist only to
// satisfy the Type interface. The tests just call them to get coverage.
func TestIsTypeMarkerMethods(t *testing.T) {
	t.Run("PrimitiveType", func(t *testing.T) {
		p := PrimitiveType{Kind: PrimitiveLong}
		p.isType() // marker method - no return value
	})

	t.Run("SetType", func(t *testing.T) {
		s := SetType{Element: PrimitiveType{Kind: PrimitiveString}}
		s.isType() // marker method - no return value
	})

	t.Run("RecordType", func(t *testing.T) {
		r := &RecordType{Attributes: make(map[string]*Attribute)}
		r.isType() // marker method - no return value
	})

	t.Run("EntityRef", func(t *testing.T) {
		e := EntityRef{Name: "User"}
		e.isType() // marker method - no return value
	})

	t.Run("ExtensionType", func(t *testing.T) {
		ext := ExtensionType{Name: "ipaddr"}
		ext.isType() // marker method - no return value
	})

	t.Run("CommonTypeRef", func(t *testing.T) {
		c := CommonTypeRef{Name: "MyType"}
		c.isType() // marker method - no return value
	})

	t.Run("EntityOrCommonRef", func(t *testing.T) {
		e := EntityOrCommonRef{Name: "Ambiguous"}
		e.isType() // marker method - no return value
	})
}

// TestTypeInterfaceSatisfaction verifies that all types can be assigned to the
// Type interface variable, ensuring the marker methods satisfy the interface.
func TestTypeInterfaceSatisfaction(t *testing.T) {
	var _ Type = PrimitiveType{}
	var _ Type = SetType{}
	var _ Type = &RecordType{}
	var _ Type = EntityRef{}
	var _ Type = ExtensionType{}
	var _ Type = CommonTypeRef{}
	var _ Type = EntityOrCommonRef{}
}
