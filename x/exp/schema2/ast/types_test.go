package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestPrimitiveTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeFn   func() ast.IsType
		expected ast.IsType
	}{
		{"StringType", func() ast.IsType { return ast.String() }, ast.StringType{}},
		{"LongType", func() ast.IsType { return ast.Long() }, ast.LongType{}},
		{"BoolType", func() ast.IsType { return ast.Bool() }, ast.BoolType{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.typeFn()
			testutil.Equals(t, result, tt.expected)
		})
	}
}

func TestExtensionTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		typeFn       func() ast.ExtensionType
		expectedName types.Ident
	}{
		{"IPAddr", ast.IPAddr, types.Ident("ipaddr")},
		{"Decimal", ast.Decimal, types.Ident("decimal")},
		{"Datetime", ast.Datetime, types.Ident("datetime")},
		{"Duration", ast.Duration, types.Ident("duration")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ext := tt.typeFn()
			testutil.Equals(t, ext.Name, tt.expectedName)
			var _ ast.IsType = ext
		})
	}
}

func TestSetType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setType ast.SetType
		check   func(t *testing.T, s ast.SetType)
	}{
		{
			name:    "set of string",
			setType: ast.Set(ast.String()),
			check: func(t *testing.T, s ast.SetType) {
				_, ok := s.Element.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			name:    "set of entity ref",
			setType: ast.Set(ast.EntityType("User")),
			check: func(t *testing.T, s ast.SetType) {
				ref, ok := s.Element.(ast.EntityTypeRef)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, ref.Name, types.EntityType("User"))
			},
		},
		{
			name:    "nested set",
			setType: ast.Set(ast.Set(ast.Long())),
			check: func(t *testing.T, s ast.SetType) {
				inner, ok := s.Element.(ast.SetType)
				testutil.Equals(t, ok, true)
				_, ok = inner.Element.(ast.LongType)
				testutil.Equals(t, ok, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var _ ast.IsType = tt.setType
			tt.check(t, tt.setType)
		})
	}
}

func TestRecordType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		record ast.RecordType
		check  func(t *testing.T, r ast.RecordType)
	}{
		{
			name:   "empty record",
			record: ast.Record(),
			check: func(t *testing.T, r ast.RecordType) {
				testutil.Equals(t, len(r.Pairs), 0)
			},
		},
		{
			name: "record with attributes",
			record: ast.Record(
				ast.Attribute("name", ast.String()),
				ast.Attribute("age", ast.Long()),
			),
			check: func(t *testing.T, r ast.RecordType) {
				testutil.Equals(t, len(r.Pairs), 2)
				testutil.Equals(t, r.Pairs[0].Key, types.String("name"))
				testutil.Equals(t, r.Pairs[0].Optional, false)
				testutil.Equals(t, r.Pairs[1].Key, types.String("age"))
			},
		},
		{
			name: "record with optional attributes",
			record: ast.Record(
				ast.Attribute("required", ast.String()),
				ast.Optional("optional", ast.String()),
			),
			check: func(t *testing.T, r ast.RecordType) {
				testutil.Equals(t, r.Pairs[0].Optional, false)
				testutil.Equals(t, r.Pairs[1].Optional, true)
			},
		},
		{
			name: "nested record",
			record: ast.Record(
				ast.Attribute("address", ast.Record(
					ast.Attribute("street", ast.String()),
					ast.Attribute("city", ast.String()),
				)),
			),
			check: func(t *testing.T, r ast.RecordType) {
				testutil.Equals(t, len(r.Pairs), 1)
				inner, ok := r.Pairs[0].Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(inner.Pairs), 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var _ ast.IsType = tt.record
			tt.check(t, tt.record)
		})
	}
}

func TestEntityTypeRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		ref          ast.EntityTypeRef
		expectedName types.EntityType
	}{
		{"EntityType", ast.EntityType("User"), types.EntityType("User")},
		{"Ref alias", ast.Ref("MyApp::User"), types.EntityType("MyApp::User")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.ref.Name, tt.expectedName)
			var _ ast.IsType = tt.ref
		})
	}
}

func TestTypeRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		typeRef      ast.TypeRef
		expectedName types.Path
	}{
		{"Type", ast.Type("Address"), types.Path("Address")},
		{"qualified type", ast.Type("MyApp::Types::Address"), types.Path("MyApp::Types::Address")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.typeRef.Name, tt.expectedName)
			var _ ast.IsType = tt.typeRef
		})
	}
}

func TestEntityRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		refFn        func() ast.EntityRef
		expectedType types.EntityType
		expectedID   types.String
	}{
		{
			"UID with default Action type",
			func() ast.EntityRef { return ast.UID("view") },
			types.EntityType("Action"),
			types.String("view"),
		},
		{
			"EntityUID with explicit type",
			func() ast.EntityRef { return ast.EntityUID("MyApp::Action", "view") },
			types.EntityType("MyApp::Action"),
			types.String("view"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ref := tt.refFn()
			testutil.Equals(t, ref.Type.Name, tt.expectedType)
			testutil.Equals(t, ref.ID, tt.expectedID)
		})
	}
}
