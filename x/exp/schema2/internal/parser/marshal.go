package parser

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// MarshalSchema converts the schema to Cedar human-readable format.
func MarshalSchema(s *ast.Schema) []byte {
	var buf bytes.Buffer
	for i, node := range s.Nodes {
		if i > 0 {
			buf.WriteString("\n")
		}
		marshalNode(&buf, node, "")
	}
	return buf.Bytes()
}

func marshalNode(buf *bytes.Buffer, node ast.IsNode, indent string) {
	switch n := node.(type) {
	case ast.NamespaceNode:
		marshalNamespace(buf, n, indent)
	case ast.CommonTypeNode:
		marshalCommonType(buf, n, indent)
	case ast.EntityNode:
		marshalEntity(buf, n, indent)
	case ast.EnumNode:
		marshalEnum(buf, n, indent)
	case ast.ActionNode:
		marshalAction(buf, n, indent)
	}
}

func marshalAnnotations(buf *bytes.Buffer, annotations []ast.Annotation, indent string) {
	for _, ann := range annotations {
		buf.WriteString(indent)
		buf.WriteString("@")
		buf.WriteString(string(ann.Key))
		if ann.Value != "" {
			buf.WriteString("(")
			buf.WriteString(quoteString(string(ann.Value)))
			buf.WriteString(")")
		}
		buf.WriteString("\n")
	}
}

func marshalNamespace(buf *bytes.Buffer, ns ast.NamespaceNode, indent string) {
	marshalAnnotations(buf, ns.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("namespace ")
	buf.WriteString(string(ns.Name))
	buf.WriteString(" {\n")

	innerIndent := indent + "  "
	for i, decl := range ns.Declarations {
		if i > 0 {
			buf.WriteString("\n")
		}
		marshalNode(buf, decl, innerIndent)
	}

	buf.WriteString(indent)
	buf.WriteString("}\n")
}

func marshalCommonType(buf *bytes.Buffer, ct ast.CommonTypeNode, indent string) {
	marshalAnnotations(buf, ct.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("type ")
	buf.WriteString(string(ct.Name))
	buf.WriteString(" = ")
	marshalTypeIndented(buf, ct.Type, indent)
	buf.WriteString(";\n")
}

func marshalEntity(buf *bytes.Buffer, e ast.EntityNode, indent string) {
	marshalAnnotations(buf, e.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("entity ")
	buf.WriteString(string(e.Name))

	if len(e.MemberOfVal) > 0 {
		buf.WriteString(" in ")
		marshalEntityTypeRefs(buf, e.MemberOfVal)
	}

	if e.ShapeVal != nil && len(e.ShapeVal.Pairs) > 0 {
		buf.WriteString(" = ")
		marshalRecordType(buf, *e.ShapeVal, indent)
	}

	if e.TagsVal != nil {
		buf.WriteString(" tags ")
		marshalTypeIndented(buf, e.TagsVal, indent+"  ")
	}

	buf.WriteString(";\n")
}

func marshalEnum(buf *bytes.Buffer, e ast.EnumNode, indent string) {
	marshalAnnotations(buf, e.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("entity ")
	buf.WriteString(string(e.Name))
	buf.WriteString(" enum [")

	for i, v := range e.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(quoteString(string(v)))
	}

	buf.WriteString("];\n")
}

func marshalAction(buf *bytes.Buffer, a ast.ActionNode, indent string) {
	marshalAnnotations(buf, a.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("action ")
	marshalActionName(buf, string(a.Name))

	if len(a.MemberOfVal) > 0 {
		buf.WriteString(" in ")
		marshalEntityRefs(buf, a.MemberOfVal)
	}

	if a.AppliesToVal != nil {
		buf.WriteString(" appliesTo {\n")
		innerIndent := indent + "  "

		if len(a.AppliesToVal.PrincipalTypes) > 0 {
			buf.WriteString(innerIndent)
			buf.WriteString("principal: ")
			marshalEntityTypeRefs(buf, a.AppliesToVal.PrincipalTypes)
			buf.WriteString(",\n")
		}

		if len(a.AppliesToVal.ResourceTypes) > 0 {
			buf.WriteString(innerIndent)
			buf.WriteString("resource: ")
			marshalEntityTypeRefs(buf, a.AppliesToVal.ResourceTypes)
			buf.WriteString(",\n")
		}

		if a.AppliesToVal.Context != nil {
			buf.WriteString(innerIndent)
			buf.WriteString("context: ")
			marshalTypeIndented(buf, a.AppliesToVal.Context, innerIndent)
			buf.WriteString("\n")
		}

		buf.WriteString(indent)
		buf.WriteString("}")
	}

	buf.WriteString(";\n")
}

func marshalActionName(buf *bytes.Buffer, name string) {
	if needsQuoting(name) {
		buf.WriteString(quoteString(name))
	} else {
		buf.WriteString(name)
	}
}

func marshalEntityTypeRefs(buf *bytes.Buffer, refs []ast.EntityTypeRef) {
	buf.WriteString("[")
	for i, ref := range refs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(string(ref.Name))
	}
	buf.WriteString("]")
}

func marshalEntityRefs(buf *bytes.Buffer, refs []ast.EntityRef) {
	if len(refs) == 1 {
		marshalEntityRef(buf, refs[0])
		return
	}

	buf.WriteString("[")
	for i, ref := range refs {
		if i > 0 {
			buf.WriteString(", ")
		}
		marshalEntityRef(buf, ref)
	}
	buf.WriteString("]")
}

func marshalEntityRef(buf *bytes.Buffer, ref ast.EntityRef) {
	if ref.Type.Name != types.EntityType("Action") {
		buf.WriteString(string(ref.Type.Name))
		buf.WriteString("::")
	}
	buf.WriteString(quoteString(string(ref.ID)))
}

func marshalTypeIndented(buf *bytes.Buffer, t ast.IsType, indent string) {
	switch v := t.(type) {
	case ast.StringType:
		buf.WriteString("String")
	case ast.LongType:
		buf.WriteString("Long")
	case ast.BoolType:
		buf.WriteString("Bool")
	case ast.ExtensionType:
		buf.WriteString("__cedar::")
		buf.WriteString(string(v.Name))
	case ast.SetType:
		buf.WriteString("Set<")
		marshalTypeIndented(buf, v.Element, indent+"  ")
		buf.WriteString(">")
	case ast.RecordType:
		marshalRecordType(buf, v, indent)
	case ast.EntityTypeRef:
		buf.WriteString(string(v.Name))
	case ast.TypeRef:
		buf.WriteString(string(v.Name))
	}
}

func marshalRecordType(buf *bytes.Buffer, r ast.RecordType, indent string) {
	buf.WriteString("{")
	if len(r.Pairs) == 0 {
		buf.WriteString("}")
		return
	}

	buf.WriteString("\n")
	innerIndent := indent + "  "
	for i, pair := range r.Pairs {
		if i > 0 {
			buf.WriteString(",\n")
		}
		buf.WriteString(innerIndent)
		buf.WriteString(quoteString(string(pair.Key)))
		if pair.Optional {
			buf.WriteString("?")
		}
		buf.WriteString(": ")
		marshalTypeIndented(buf, pair.Type, innerIndent)
	}
	buf.WriteString(",\n")
	buf.WriteString(indent)
	buf.WriteString("}")
}

// marshalRecordTypeCompact marshals a record type in Rust format for entity shapes
// e.g., {"name": String, "age": Long}
func marshalRecordTypeCompact(buf *bytes.Buffer, r ast.RecordType, indent string) {
	buf.WriteString("{")
	for i, pair := range r.Pairs {
		if i > 0 {
			buf.WriteString(", ")
		}
		key := string(pair.Key)
		if pair.Optional {
			key += "?"
		}
		buf.WriteString(quoteString(key))
		buf.WriteString(": ")
		marshalTypeIndented(buf, pair.Type, indent)
	}
	buf.WriteString("}")
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for i, c := range s {
		if i == 0 {
			if !isIdentStart(c) {
				return true
			}
		} else {
			if !isIdentChar(c) {
				return true
			}
		}
	}
	return false
}

func isIdentStart(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func isIdentChar(c rune) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func quoteString(s string) string {
	return fmt.Sprintf("%q", s)
}
