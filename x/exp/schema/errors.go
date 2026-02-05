package schema

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrCycle         = errors.New("cycle detected in common type definitions")
	ErrUndefinedType = errors.New("undefined type")
	ErrShadow        = errors.New("cannot shadow definition in empty namespace")
	ErrReservedName  = errors.New("reserved name")
	ErrParse         = errors.New("parse error")
)

// CycleError indicates a cycle in common type definitions.
type CycleError struct {
	Path []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("%v: %s", ErrCycle, strings.Join(e.Path, " -> "))
}

func (e *CycleError) Unwrap() error { return ErrCycle }

// UndefinedTypeError indicates an undefined type reference.
type UndefinedTypeError struct {
	Name      string
	Namespace string
	Context   string
}

func (e *UndefinedTypeError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%v: %q %s", ErrUndefinedType, e.Name, e.Context)
	}
	return fmt.Sprintf("%v: %q", ErrUndefinedType, e.Name)
}

func (e *UndefinedTypeError) Unwrap() error { return ErrUndefinedType }

// ShadowError indicates a definition in a named namespace shadows
// a definition in the empty namespace (RFC 70).
type ShadowError struct {
	Name      string
	Namespace string
}

func (e *ShadowError) Error() string {
	return fmt.Sprintf("%v: %q in namespace %q shadows definition in empty namespace",
		ErrShadow, e.Name, e.Namespace)
}

func (e *ShadowError) Unwrap() error { return ErrShadow }

// ReservedNameError indicates a reserved name was used for a custom type.
type ReservedNameError struct {
	Name string
	Kind string
}

func (e *ReservedNameError) Error() string {
	return fmt.Sprintf("%v: %q cannot be used as %s", ErrReservedName, e.Name, e.Kind)
}

func (e *ReservedNameError) Unwrap() error { return ErrReservedName }

// ParseError indicates a syntax error during schema parsing.
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

func (e *ParseError) Unwrap() error { return ErrParse }

var reservedTypeNames = map[string]bool{
	"Bool": true, "Boolean": true, "Long": true, "String": true,
	"Record": true, "Set": true, "Entity": true, "Extension": true,
}

func isPrimitiveTypeName(name string) bool {
	return reservedTypeNames[name]
}
