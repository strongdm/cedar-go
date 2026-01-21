package ast

import (
	"bytes"
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

// MarshalCedar converts the schema to Cedar human-readable format.
func (s *Schema) MarshalCedar() []byte {
	var buf bytes.Buffer
	for i, node := range s.Nodes {
		if i > 0 {
			buf.WriteString("\n")
		}
		marshalNode(&buf, node, "")
	}
	return buf.Bytes()
}

func marshalNode(buf *bytes.Buffer, node IsNode, indent string) {
	switch n := node.(type) {
	case NamespaceNode:
		marshalNamespace(buf, n, indent)
	case CommonTypeNode:
		marshalCommonType(buf, n, indent)
	case EntityNode:
		marshalEntity(buf, n, indent)
	case EnumNode:
		marshalEnum(buf, n, indent)
	case ActionNode:
		marshalAction(buf, n, indent)
	}
}

func marshalAnnotations(buf *bytes.Buffer, annotations []Annotation, indent string) {
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

func marshalNamespace(buf *bytes.Buffer, ns NamespaceNode, indent string) {
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

func marshalCommonType(buf *bytes.Buffer, ct CommonTypeNode, indent string) {
	marshalAnnotations(buf, ct.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("type ")
	buf.WriteString(string(ct.Name))
	buf.WriteString(" = ")
	marshalType(buf, ct.Type)
	buf.WriteString(";\n")
}

func marshalEntity(buf *bytes.Buffer, e EntityNode, indent string) {
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
		marshalRecordTypeCompact(buf, *e.ShapeVal, indent)
	}

	if e.TagsVal != nil {
		buf.WriteString(" tags ")
		marshalType(buf, e.TagsVal)
	}

	buf.WriteString(";\n")
}

func marshalEnum(buf *bytes.Buffer, e EnumNode, indent string) {
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

func marshalAction(buf *bytes.Buffer, a ActionNode, indent string) {
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

		buf.WriteString(innerIndent)
		buf.WriteString("context: ")
		if a.AppliesToVal.Context != nil {
			marshalTypeIndented(buf, a.AppliesToVal.Context, innerIndent)
		} else {
			buf.WriteString("{}")
		}
		buf.WriteString(",\n")

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

func marshalEntityTypeRefs(buf *bytes.Buffer, refs []EntityTypeRef) {
	if len(refs) == 1 {
		buf.WriteString(string(refs[0].Name))
		return
	}

	buf.WriteString("[")
	for i, ref := range refs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(string(ref.Name))
	}
	buf.WriteString("]")
}

func marshalEntityRefs(buf *bytes.Buffer, refs []EntityRef) {
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

func marshalEntityRef(buf *bytes.Buffer, ref EntityRef) {
	if ref.Type.Name != types.EntityType("Action") {
		buf.WriteString(string(ref.Type.Name))
		buf.WriteString("::")
	}
	buf.WriteString(quoteString(string(ref.ID)))
}

func marshalType(buf *bytes.Buffer, t IsType) {
	marshalTypeIndented(buf, t, "")
}

func marshalTypeIndented(buf *bytes.Buffer, t IsType, indent string) {
	switch v := t.(type) {
	case StringType:
		buf.WriteString("String")
	case LongType:
		buf.WriteString("Long")
	case BoolType:
		buf.WriteString("Bool")
	case ExtensionType:
		buf.WriteString("__cedar::")
		buf.WriteString(string(v.Name))
	case SetType:
		buf.WriteString("Set<")
		marshalType(buf, v.Element)
		buf.WriteString(">")
	case RecordType:
		marshalRecordType(buf, v, indent)
	case EntityTypeRef:
		buf.WriteString(string(v.Name))
	case TypeRef:
		buf.WriteString(string(v.Name))
	}
}

func marshalRecordType(buf *bytes.Buffer, r RecordType, indent string) {
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
func marshalRecordTypeCompact(buf *bytes.Buffer, r RecordType, indent string) {
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
