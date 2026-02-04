package schema

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parse"
)

var (
	ErrCycle         = errors.New("cycle detected in common type definitions")
	ErrUndefinedType = errors.New("undefined type")
	ErrShadow        = errors.New("cannot shadow definition in empty namespace")
	ErrDuplicate     = errors.New("duplicate definition")
	ErrReservedName  = errors.New("reserved name")
	ErrParse         = errors.New("parse error")
)

type CycleError struct {
	Path []string // the types forming the cycle
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("%v: %s", ErrCycle, strings.Join(e.Path, " -> "))
}

func (e *CycleError) Unwrap() error {
	return ErrCycle
}

type UndefinedTypeError struct {
	Name      string // the undefined type name
	Namespace string // the namespace where the reference occurred
	Context   string // additional context (e.g., "in entity User attribute owner")
}

func (e *UndefinedTypeError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%v: %q %s", ErrUndefinedType, e.Name, e.Context)
	}
	return fmt.Sprintf("%v: %q", ErrUndefinedType, e.Name)
}

func (e *UndefinedTypeError) Unwrap() error {
	return ErrUndefinedType
}

type ShadowError struct {
	Name      string // the shadowed name
	Namespace string // the namespace attempting to shadow
}

func (e *ShadowError) Error() string {
	return fmt.Sprintf("%v: %q in namespace %q shadows definition in empty namespace",
		ErrShadow, e.Name, e.Namespace)
}

func (e *ShadowError) Unwrap() error {
	return ErrShadow
}

type DuplicateError struct {
	Kind      string // "entity type", "action", "common type", "attribute"
	Name      string // the duplicated name
	Namespace string // the namespace where the duplicate occurred
}

func (e *DuplicateError) Error() string {
	if e.Namespace == "" {
		return fmt.Sprintf("%v: %s %q in empty namespace", ErrDuplicate, e.Kind, e.Name)
	}
	return fmt.Sprintf("%v: %s %q in namespace %q", ErrDuplicate, e.Kind, e.Name, e.Namespace)
}

func (e *DuplicateError) Unwrap() error {
	return ErrDuplicate
}

type ReservedNameError struct {
	Name string // the reserved name that was used
	Kind string // what it was used as (e.g., "entity type", "common type")
}

func (e *ReservedNameError) Error() string {
	return fmt.Sprintf("%v: %q cannot be used as %s", ErrReservedName, e.Name, e.Kind)
}

func (e *ReservedNameError) Unwrap() error {
	return ErrReservedName
}

type ParseError struct {
	Filename string
	Line     int
	Column   int
	Message  string
}

func (e *ParseError) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("%v: %s:%d:%d: %s", ErrParse, e.Filename, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("%v: line %d, column %d: %s", ErrParse, e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("%v: %s", ErrParse, e.Message)
}

func (e *ParseError) Unwrap() error {
	return ErrParse
}

// IsPrimitiveTypeName checks if name is a built-in type name (Bool, Long, String, Entity, etc.)
// that cannot be used as a custom type name.
func IsPrimitiveTypeName(name string) bool {
	return parse.IsPrimitiveTypeName(name)
}
