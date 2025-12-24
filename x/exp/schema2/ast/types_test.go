package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestPrimitiveTypes(t *testing.T) {
	t.Parallel()

	t.Run("StringType", func(t *testing.T) {
		t.Parallel()
		s := ast.String()
		// Verify it implements IsType
		var _ ast.IsType = s
		testutil.Equals(t, s, ast.StringType{})
	})

	t.Run("LongType", func(t *testing.T) {
		t.Parallel()
		l := ast.Long()
		var _ ast.IsType = l
		testutil.Equals(t, l, ast.LongType{})
	})

	t.Run("BoolType", func(t *testing.T) {
		t.Parallel()
		b := ast.Bool()
		var _ ast.IsType = b
		testutil.Equals(t, b, ast.BoolType{})
	})
}

func TestExtensionTypes(t *testing.T) {
	t.Parallel()

	t.Run("IPAddr", func(t *testing.T) {
		t.Parallel()
		ip := ast.IPAddr()
		testutil.Equals(t, ip.Name, types.Ident("ipaddr"))
		var _ ast.IsType = ip
	})

	t.Run("Decimal", func(t *testing.T) {
		t.Parallel()
		d := ast.Decimal()
		testutil.Equals(t, d.Name, types.Ident("decimal"))
		var _ ast.IsType = d
	})

	t.Run("Datetime", func(t *testing.T) {
		t.Parallel()
		dt := ast.Datetime()
		testutil.Equals(t, dt.Name, types.Ident("datetime"))
		var _ ast.IsType = dt
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()
		dur := ast.Duration()
		testutil.Equals(t, dur.Name, types.Ident("duration"))
		var _ ast.IsType = dur
	})
}

func TestSetType(t *testing.T) {
	t.Parallel()

	t.Run("set of string", func(t *testing.T) {
		t.Parallel()
		s := ast.Set(ast.String())
		_, ok := s.Element.(ast.StringType)
		testutil.Equals(t, ok, true)
		var _ ast.IsType = s
	})

	t.Run("set of entity ref", func(t *testing.T) {
		t.Parallel()
		s := ast.Set(ast.EntityType("User"))
		ref, ok := s.Element.(ast.EntityTypeRef)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ref.Name, types.EntityType("User"))
	})

	t.Run("nested set", func(t *testing.T) {
		t.Parallel()
		s := ast.Set(ast.Set(ast.Long()))
		inner, ok := s.Element.(ast.SetType)
		testutil.Equals(t, ok, true)
		_, ok = inner.Element.(ast.LongType)
		testutil.Equals(t, ok, true)
	})
}

func TestRecordType(t *testing.T) {
	t.Parallel()

	t.Run("empty record", func(t *testing.T) {
		t.Parallel()
		r := ast.Record()
		testutil.Equals(t, len(r.Pairs), 0)
		var _ ast.IsType = r
	})

	t.Run("record with attributes", func(t *testing.T) {
		t.Parallel()
		r := ast.Record(
			ast.Attribute("name", ast.String()),
			ast.Attribute("age", ast.Long()),
		)
		testutil.Equals(t, len(r.Pairs), 2)
		testutil.Equals(t, r.Pairs[0].Key, types.String("name"))
		testutil.Equals(t, r.Pairs[0].Optional, false)
		testutil.Equals(t, r.Pairs[1].Key, types.String("age"))
	})

	t.Run("record with optional attributes", func(t *testing.T) {
		t.Parallel()
		r := ast.Record(
			ast.Attribute("required", ast.String()),
			ast.Optional("optional", ast.String()),
		)
		testutil.Equals(t, r.Pairs[0].Optional, false)
		testutil.Equals(t, r.Pairs[1].Optional, true)
	})

	t.Run("nested record", func(t *testing.T) {
		t.Parallel()
		r := ast.Record(
			ast.Attribute("address", ast.Record(
				ast.Attribute("street", ast.String()),
				ast.Attribute("city", ast.String()),
			)),
		)
		testutil.Equals(t, len(r.Pairs), 1)
		inner, ok := r.Pairs[0].Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(inner.Pairs), 2)
	})
}

func TestEntityTypeRef(t *testing.T) {
	t.Parallel()

	t.Run("EntityType", func(t *testing.T) {
		t.Parallel()
		ref := ast.EntityType("User")
		testutil.Equals(t, ref.Name, types.EntityType("User"))
		var _ ast.IsType = ref
	})

	t.Run("Ref alias", func(t *testing.T) {
		t.Parallel()
		ref := ast.Ref("MyApp::User")
		testutil.Equals(t, ref.Name, types.EntityType("MyApp::User"))
	})
}

func TestTypeRef(t *testing.T) {
	t.Parallel()

	t.Run("Type", func(t *testing.T) {
		t.Parallel()
		ref := ast.Type("Address")
		testutil.Equals(t, ref.Name, types.Path("Address"))
		var _ ast.IsType = ref
	})

	t.Run("qualified type", func(t *testing.T) {
		t.Parallel()
		ref := ast.Type("MyApp::Types::Address")
		testutil.Equals(t, ref.Name, types.Path("MyApp::Types::Address"))
	})
}

func TestEntityRef(t *testing.T) {
	t.Parallel()

	t.Run("UID with default Action type", func(t *testing.T) {
		t.Parallel()
		ref := ast.UID("view")
		testutil.Equals(t, ref.Type.Name, types.EntityType("Action"))
		testutil.Equals(t, ref.ID, types.String("view"))
	})

	t.Run("EntityUID with explicit type", func(t *testing.T) {
		t.Parallel()
		ref := ast.EntityUID("MyApp::Action", "view")
		testutil.Equals(t, ref.Type.Name, types.EntityType("MyApp::Action"))
		testutil.Equals(t, ref.ID, types.String("view"))
	})
}
