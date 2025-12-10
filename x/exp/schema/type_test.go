package schema

import (
	"testing"
)

func TestTypeConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  Type
	}{
		{"Boolean", Boolean()},
		{"Long", Long()},
		{"String", String()},
		{"SetOf", SetOf(String())},
		{"Record", Record()},
		{"Entity", Entity("User")},
		{"Extension", Extension("ipaddr")},
		{"Ref", Ref("MyType")},
		{"IPAddr", IPAddr()},
		{"Decimal", Decimal()},
		{"Datetime", Datetime()},
		{"Duration", Duration()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typ.v == nil {
				t.Error("expected non-nil type variant")
			}
		})
	}
}

func TestTypeEquality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		a, b  Type
		equal bool
	}{
		{"Boolean/Boolean", Boolean(), Boolean(), true},
		{"Boolean/Long", Boolean(), Long(), false},
		{"Long/Long", Long(), Long(), true},
		{"String/String", String(), String(), true},
		{"Set/Set same", SetOf(String()), SetOf(String()), true},
		{"Set/Set diff", SetOf(String()), SetOf(Long()), false},
		{"Record/Record empty", Record(), Record(), true},
		{"Record/Record same", Record(Attr("a", String(), true)), Record(Attr("a", String(), true)), true},
		{"Record/Record diff name", Record(Attr("a", String(), true)), Record(Attr("b", String(), true)), false},
		{"Record/Record diff type", Record(Attr("a", String(), true)), Record(Attr("a", Long(), true)), false},
		{"Record/Record diff required", Record(Attr("a", String(), true)), Record(Attr("a", String(), false)), false},
		{"Entity/Entity same", Entity("User"), Entity("User"), true},
		{"Entity/Entity diff", Entity("User"), Entity("Group"), false},
		{"Extension/Extension same", Extension("ipaddr"), Extension("ipaddr"), true},
		{"Extension/Extension diff", Extension("ipaddr"), Extension("decimal"), false},
		{"Ref/Ref same", Ref("MyType"), Ref("MyType"), true},
		{"Ref/Ref diff", Ref("MyType"), Ref("OtherType"), false},
		{"nil/nil", Type{}, Type{}, true},
		{"nil/Boolean", Type{}, Boolean(), false},
		{"Boolean/nil", Boolean(), Type{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := typeEqual(tt.a, tt.b)
			if got != tt.equal {
				t.Errorf("typeEqual() = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestTypeAsIsType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		typ  Type
	}{
		{"Boolean", Boolean()},
		{"Long", Long()},
		{"String", String()},
		{"Set", SetOf(String())},
		{"Record", Record()},
		{"Entity", Entity("User")},
		{"Extension", Extension("ipaddr")},
		{"Ref", Ref("MyType")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.typ.AsIsType()
			if v == nil {
				t.Error("expected non-nil isType")
			}
		})
	}
}

func TestAttributeConstructors(t *testing.T) {
	t.Parallel()

	attr1 := Attr("name", String(), true)
	if attr1.Name != "name" || attr1.Required != true {
		t.Error("Attr() failed")
	}

	attr2 := RequiredAttr("email", String())
	if attr2.Name != "email" || attr2.Required != true {
		t.Error("RequiredAttr() failed")
	}

	attr3 := OptionalAttr("phone", String())
	if attr3.Name != "phone" || attr3.Required != false {
		t.Error("OptionalAttr() failed")
	}
}

func TestTypeIsTypeMethods(t *testing.T) {
	t.Parallel()

	// These calls exercise the isType() marker methods for coverage
	TypeBoolean{}.isType()
	TypeLong{}.isType()
	TypeString{}.isType()
	TypeSet{}.isType()
	TypeRecord{}.isType()
	TypeEntity{}.isType()
	TypeExtension{}.isType()
	TypeRef{}.isType()
}

func TestComplexTypes(t *testing.T) {
	t.Parallel()

	// Nested set
	setOfSets := SetOf(SetOf(String()))
	if setOfSets.v == nil {
		t.Error("expected non-nil set of sets")
	}

	// Record with multiple attributes
	rec := Record(
		RequiredAttr("name", String()),
		OptionalAttr("age", Long()),
		RequiredAttr("roles", SetOf(String())),
	)
	if rt, ok := rec.v.(TypeRecord); ok {
		if len(rt.Attributes) != 3 {
			t.Errorf("expected 3 attributes, got %d", len(rt.Attributes))
		}
	} else {
		t.Error("expected TypeRecord")
	}

	// Nested record
	nestedRec := Record(
		RequiredAttr("user", Record(
			RequiredAttr("name", String()),
		)),
	)
	if nestedRec.v == nil {
		t.Error("expected non-nil nested record")
	}
}

func TestExtensionTypes(t *testing.T) {
	t.Parallel()

	ip := IPAddr()
	if ext, ok := ip.v.(TypeExtension); ok {
		if ext.Name != "ipaddr" {
			t.Errorf("expected ipaddr, got %s", ext.Name)
		}
	} else {
		t.Error("expected TypeExtension")
	}

	dec := Decimal()
	if ext, ok := dec.v.(TypeExtension); ok {
		if ext.Name != "decimal" {
			t.Errorf("expected decimal, got %s", ext.Name)
		}
	} else {
		t.Error("expected TypeExtension")
	}

	dt := Datetime()
	if ext, ok := dt.v.(TypeExtension); ok {
		if ext.Name != "datetime" {
			t.Errorf("expected datetime, got %s", ext.Name)
		}
	} else {
		t.Error("expected TypeExtension")
	}

	dur := Duration()
	if ext, ok := dur.v.(TypeExtension); ok {
		if ext.Name != "duration" {
			t.Errorf("expected duration, got %s", ext.Name)
		}
	} else {
		t.Error("expected TypeExtension")
	}
}

func TestAttributeEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		a, b  Attribute
		equal bool
	}{
		{
			"same",
			Attribute{Name: "x", Type: String(), Required: true},
			Attribute{Name: "x", Type: String(), Required: true},
			true,
		},
		{
			"diff name",
			Attribute{Name: "x", Type: String(), Required: true},
			Attribute{Name: "y", Type: String(), Required: true},
			false,
		},
		{
			"diff type",
			Attribute{Name: "x", Type: String(), Required: true},
			Attribute{Name: "x", Type: Long(), Required: true},
			false,
		},
		{
			"diff required",
			Attribute{Name: "x", Type: String(), Required: true},
			Attribute{Name: "x", Type: String(), Required: false},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.a.equal(tt.b)
			if got != tt.equal {
				t.Errorf("Attribute.equal() = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestEntityType(t *testing.T) {
	t.Parallel()

	// Simple entity type
	user := Entity("User")
	if et, ok := user.v.(TypeEntity); ok {
		if et.Name != "User" {
			t.Errorf("expected User, got %s", et.Name)
		}
	} else {
		t.Error("expected TypeEntity")
	}

	// Qualified entity type
	appUser := Entity("MyApp::User")
	if et, ok := appUser.v.(TypeEntity); ok {
		if et.Name != "MyApp::User" {
			t.Errorf("expected MyApp::User, got %s", et.Name)
		}
	} else {
		t.Error("expected TypeEntity")
	}
}

func TestTypeRefConstructor(t *testing.T) {
	t.Parallel()

	ref := Ref("MyCommonType")
	if tr, ok := ref.v.(TypeRef); ok {
		if tr.Name != "MyCommonType" {
			t.Errorf("expected MyCommonType, got %s", tr.Name)
		}
	} else {
		t.Error("expected TypeRef")
	}
}

func TestRecordWithAnnotatedAttributes(t *testing.T) {
	t.Parallel()

	// Record with various attribute types
	rec := Record(
		Attr("id", Long(), true),
		Attr("name", String(), true),
		Attr("active", Boolean(), true),
		Attr("tags", SetOf(String()), false),
		Attr("metadata", Record(
			Attr("created", Datetime(), true),
		), false),
		Attr("owner", Entity("User"), false),
	)

	if rt, ok := rec.v.(TypeRecord); ok {
		if len(rt.Attributes) != 6 {
			t.Errorf("expected 6 attributes, got %d", len(rt.Attributes))
		}
	} else {
		t.Error("expected TypeRecord")
	}
}
