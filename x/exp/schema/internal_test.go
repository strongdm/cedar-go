package schema

import (
	"testing"

	internalast "github.com/cedar-policy/cedar-go/internal/schema/ast"
)

// Tests for internal conversion functions to achieve 100% coverage.
// These tests use BadTypeForTesting from the internal package to trigger
// error paths that can't be triggered through the public API.

func TestConvertInternalTypeNil(t *testing.T) {
	t.Parallel()

	// Test with nil input
	result, err := convertInternalType(nil)
	if err != nil {
		t.Errorf("convertInternalType(nil) error = %v", err)
	}
	if result.v != nil {
		t.Error("expected nil type")
	}
}

func TestConvertInternalTypeUnknown(t *testing.T) {
	t.Parallel()

	// Test with unknown type using the test helper
	badType := internalast.NewBadTypeForTesting()
	_, err := convertInternalType(badType)
	if err == nil {
		t.Error("expected error for unknown internal type")
	}
}

func TestConvertInternalTypeSetWithBadElement(t *testing.T) {
	t.Parallel()

	// Test Set with bad element type
	setType := &internalast.SetType{
		Element: internalast.NewBadTypeForTesting(),
	}

	_, err := convertInternalType(setType)
	if err == nil {
		t.Error("expected error for Set with bad element type")
	}
}

func TestConvertInternalTypeRecordWithBadAttr(t *testing.T) {
	t.Parallel()

	// Test Record with bad attribute type
	recordType := &internalast.RecordType{
		Attributes: []*internalast.Attribute{
			{
				Key:        &internalast.Ident{Value: "bad"},
				Type:       internalast.NewBadTypeForTesting(),
				IsRequired: true,
			},
		},
	}

	_, err := convertInternalType(recordType)
	if err == nil {
		t.Error("expected error for Record with bad attribute type")
	}
}

func TestConvertInternalRecordAttrsWithBadType(t *testing.T) {
	t.Parallel()

	recordType := &internalast.RecordType{
		Attributes: []*internalast.Attribute{
			{
				Key:        &internalast.Ident{Value: "attr"},
				Type:       internalast.NewBadTypeForTesting(),
				IsRequired: true,
			},
		},
	}

	_, err := convertInternalRecordAttrs(recordType)
	if err == nil {
		t.Error("expected error for bad attribute type")
	}
}

func TestConvertInternalCommonTypeWithBadType(t *testing.T) {
	t.Parallel()

	ct := &internalast.CommonTypeDecl{
		Name:  &internalast.Ident{Value: "BadType"},
		Value: internalast.NewBadTypeForTesting(),
	}

	_, err := convertInternalCommonType(ct)
	if err == nil {
		t.Error("expected error for common type with bad type")
	}
}

func TestConvertInternalEntityWithBadShapeAttr(t *testing.T) {
	t.Parallel()

	entity := &internalast.Entity{
		Names: []*internalast.Ident{{Value: "BadEntity"}},
		Shape: &internalast.RecordType{
			Attributes: []*internalast.Attribute{
				{
					Key:        &internalast.Ident{Value: "bad"},
					Type:       internalast.NewBadTypeForTesting(),
					IsRequired: true,
				},
			},
		},
	}

	_, err := convertInternalEntity(entity)
	if err == nil {
		t.Error("expected error for entity with bad shape attribute type")
	}
}

func TestConvertInternalEntityWithBadTags(t *testing.T) {
	t.Parallel()

	entity := &internalast.Entity{
		Names: []*internalast.Ident{{Value: "BadEntity"}},
		Tags:  internalast.NewBadTypeForTesting(),
	}

	_, err := convertInternalEntity(entity)
	if err == nil {
		t.Error("expected error for entity with bad tags type")
	}
}

func TestConvertInternalActionWithBadContextRecord(t *testing.T) {
	t.Parallel()

	action := &internalast.Action{
		Names: []internalast.Name{&internalast.Ident{Value: "badAction"}},
		AppliesTo: &internalast.AppliesTo{
			ContextRecord: &internalast.RecordType{
				Attributes: []*internalast.Attribute{
					{
						Key:        &internalast.Ident{Value: "bad"},
						Type:       internalast.NewBadTypeForTesting(),
						IsRequired: true,
					},
				},
			},
		},
	}

	_, err := convertInternalAction(action)
	if err == nil {
		t.Error("expected error for action with bad context record attribute type")
	}
}

func TestConvertInternalNamespaceWithBadEntity(t *testing.T) {
	t.Parallel()

	ns := &internalast.Namespace{
		Name: &internalast.Path{Parts: []*internalast.Ident{{Value: "BadNS"}}},
		Decls: []internalast.Declaration{
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "BadEntity"}},
				Shape: &internalast.RecordType{
					Attributes: []*internalast.Attribute{
						{
							Key:        &internalast.Ident{Value: "bad"},
							Type:       internalast.NewBadTypeForTesting(),
							IsRequired: true,
						},
					},
				},
			},
		},
	}

	_, err := convertInternalNamespace(ns)
	if err == nil {
		t.Error("expected error for namespace with bad entity")
	}
}

func TestConvertInternalNamespaceWithBadAction(t *testing.T) {
	t.Parallel()

	ns := &internalast.Namespace{
		Name: &internalast.Path{Parts: []*internalast.Ident{{Value: "BadNS"}}},
		Decls: []internalast.Declaration{
			&internalast.Action{
				Names: []internalast.Name{&internalast.Ident{Value: "badAction"}},
				AppliesTo: &internalast.AppliesTo{
					ContextRecord: &internalast.RecordType{
						Attributes: []*internalast.Attribute{
							{
								Key:        &internalast.Ident{Value: "bad"},
								Type:       internalast.NewBadTypeForTesting(),
								IsRequired: true,
							},
						},
					},
				},
			},
		},
	}

	_, err := convertInternalNamespace(ns)
	if err == nil {
		t.Error("expected error for namespace with bad action")
	}
}

func TestConvertInternalNamespaceWithBadCommonType(t *testing.T) {
	t.Parallel()

	ns := &internalast.Namespace{
		Name: &internalast.Path{Parts: []*internalast.Ident{{Value: "BadNS"}}},
		Decls: []internalast.Declaration{
			&internalast.CommonTypeDecl{
				Name:  &internalast.Ident{Value: "BadType"},
				Value: internalast.NewBadTypeForTesting(),
			},
		},
	}

	_, err := convertInternalNamespace(ns)
	if err == nil {
		t.Error("expected error for namespace with bad common type")
	}
}

func TestConvertInternalSchemaWithBadNamespace(t *testing.T) {
	t.Parallel()

	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Namespace{
				Name: &internalast.Path{Parts: []*internalast.Ident{{Value: "BadNS"}}},
				Decls: []internalast.Declaration{
					&internalast.Entity{
						Names: []*internalast.Ident{{Value: "BadEntity"}},
						Shape: &internalast.RecordType{
							Attributes: []*internalast.Attribute{
								{
									Key:        &internalast.Ident{Value: "bad"},
									Type:       internalast.NewBadTypeForTesting(),
									IsRequired: true,
								},
							},
						},
					},
				},
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err == nil {
		t.Error("expected error for schema with bad namespace")
	}
}

func TestConvertInternalSchemaWithBadEntity(t *testing.T) {
	t.Parallel()

	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "BadEntity"}},
				Shape: &internalast.RecordType{
					Attributes: []*internalast.Attribute{
						{
							Key:        &internalast.Ident{Value: "bad"},
							Type:       internalast.NewBadTypeForTesting(),
							IsRequired: true,
						},
					},
				},
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err == nil {
		t.Error("expected error for schema with bad entity")
	}
}

func TestConvertInternalSchemaWithBadAction(t *testing.T) {
	t.Parallel()

	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Action{
				Names: []internalast.Name{&internalast.Ident{Value: "badAction"}},
				AppliesTo: &internalast.AppliesTo{
					ContextRecord: &internalast.RecordType{
						Attributes: []*internalast.Attribute{
							{
								Key:        &internalast.Ident{Value: "bad"},
								Type:       internalast.NewBadTypeForTesting(),
								IsRequired: true,
							},
						},
					},
				},
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err == nil {
		t.Error("expected error for schema with bad action")
	}
}

func TestConvertInternalSchemaWithBadCommonType(t *testing.T) {
	t.Parallel()

	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.CommonTypeDecl{
				Name:  &internalast.Ident{Value: "BadType"},
				Value: internalast.NewBadTypeForTesting(),
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err == nil {
		t.Error("expected error for schema with bad common type")
	}
}

func TestUnmarshalCedarWithFilenameConversionError(t *testing.T) {
	t.Parallel()

	// This test verifies the error path when convertInternalSchema fails
	// by testing the conversion function directly
	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "Test"}},
				Tags:  internalast.NewBadTypeForTesting(),
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err == nil {
		t.Error("expected error from convertInternalSchema")
	}
}

func TestConvertInternalSchemaWithComments(t *testing.T) {
	t.Parallel()

	// Test with CommentBlock in schema declarations
	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.CommentBlock{},
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "User"}},
			},
		},
	}

	s := New()
	err := convertInternalSchema(s, internal)
	if err != nil {
		t.Errorf("convertInternalSchema error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
}

func TestConvertInternalNamespaceWithComments(t *testing.T) {
	t.Parallel()

	// Test with CommentBlock in namespace declarations
	ns := &internalast.Namespace{
		Name: &internalast.Path{Parts: []*internalast.Ident{{Value: "Test"}}},
		Decls: []internalast.Declaration{
			&internalast.CommentBlock{},
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "User"}},
			},
		},
	}

	result, err := convertInternalNamespace(ns)
	if err != nil {
		t.Errorf("convertInternalNamespace error = %v", err)
	}

	if result.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
}

func TestUnmarshalFromInternalError(t *testing.T) {
	t.Parallel()

	// Test the error path in unmarshalFromInternal when conversion fails
	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "BadEntity"}},
				Tags:  internalast.NewBadTypeForTesting(),
			},
		},
	}

	s := New()
	err := s.unmarshalFromInternal("test.cedar", internal)
	if err == nil {
		t.Error("expected error from unmarshalFromInternal")
	}
}

func TestUnmarshalFromInternalSuccess(t *testing.T) {
	t.Parallel()

	// Test successful conversion
	internal := &internalast.Schema{
		Decls: []internalast.Declaration{
			&internalast.Entity{
				Names: []*internalast.Ident{{Value: "User"}},
			},
		},
	}

	s := New()
	s.SetFilename("original.cedar")
	err := s.unmarshalFromInternal("original.cedar", internal)
	if err != nil {
		t.Errorf("unmarshalFromInternal error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
}
