package types

import "net/netip"

type Expr interface {
	isExpr()
}

type Boolean bool

func (Boolean) isExpr() {}

type String string

func (String) isExpr()           {}
func (String) isPatternElement() {}

type Long int

func (Long) isExpr() {}

type Set []Expr

func (Set) isExpr() {}

type Record map[string]Expr

func (Record) isExpr() {}

type Variable uint8

const (
	PrincipalVariable Variable = iota
	ActionVariable
	ResourceVariable
	ContextVariable
)

func (Variable) isExpr() {}

// String Operators

type Wildcard struct{}

func (Wildcard) isPatternElement() {}

type PatternElement interface {
	isPatternElement()
}

type Like struct {
	LHS     Expr
	Pattern []PatternElement
}

func (Like) isExpr() {}

// Comparison Operators
type Equals struct {
	LHS Expr
	RHS Expr
}

func (Equals) isExpr() {}

type NotEquals struct {
	LHS Expr
	RHS Expr
}

func (NotEquals) isExpr() {}

// Can be used for both Long and Decimal types
type LessThan struct {
	OrEqualTo bool
	LHS       Expr
	RHS       Expr
}

func (LessThan) isExpr() {}

// Can be used for both Long and Decimal types
type GreaterThan struct {
	OrEqualTo bool
	LHS       Expr
	RHS       Expr
}

func (GreaterThan) isExpr() {}

// Logical Operators
type And struct {
	LHS Expr
	RHS Expr
}

func (And) isExpr() {}

type Or struct {
	LHS Expr
	RHS Expr
}

func (Or) isExpr() {}

type Not struct {
	Expr
}

func (Not) isExpr() {}

type If struct {
	Condition Expr
	Then      Expr
	Else      Expr
}

func (If) isExpr() {}

// Arithmetic Operators

type Negative struct {
	Expr
}

func (Negative) isExpr() {}

type Add struct {
	LHS Expr
	RHS Expr
}

func (Add) isExpr() {}

type Subtract struct {
	LHS Expr
	RHS Expr
}

func (Subtract) isExpr() {}

type Multiply struct {
	LHS Expr
	RHS Expr
}

func (Multiply) isExpr() {}

// Hierarchy and membership operators

type In struct {
	LHS Expr
	RHS Expr
}

func (In) isExpr() {}

type HasAttribute struct {
	LHS  Expr
	Name string
}

func (HasAttribute) isExpr() {}

type Is struct {
	LHS        Expr
	EntityType EntityType
	In         Expr // can be nil
}

func (Is) isExpr() {}

type Contains struct {
	LHS Expr
	Arg Expr
}

func (Contains) isExpr() {}

type ContainsAll struct {
	LHS Expr
	Arg Expr
}

func (ContainsAll) isExpr() {}

type ContainsAny struct {
	LHS Expr
	Arg Expr
}

func (ContainsAny) isExpr() {}

type GetAttribute struct {
	LHS  Expr
	Name string
}

func (GetAttribute) isExpr() {}

// Extension types

// TODO: Should this be some kind of fixed-width decimal type? Cedar only allows four digits below the decimal point.
// NB: We don't have to worry about `decimal(<expr>)` because the Cedar docs say that's invalid. Only string literals
// are allowed.
type Decimal float64

func (Decimal) isExpr() {}

// TODO: Put the .greaterThan()/.lessThan() operators here? Or just have users put a Decimal in the Long exprs?

// TODO: What about foo.bar.isIPv4()? That can't currently be expressed.
// NB: We don't have to worry about `ip(<expr>)` because the Cedar docs say that's invalid. Only string literals
// are allowed. Note that the ipaddr type can hold ranges or single IP addresses, so we use Prefix to represent
// it.
type IpAddr netip.Prefix

func (IpAddr) isExpr() {}

type IsIPv4 struct {
	IpAddr IpAddr
}

func (IsIPv4) isExpr() {}

func (i IpAddr) isIpv4() IsIPv4 {
	return IsIPv4{i}
}

type IsIPv6 struct {
	IpAddr IpAddr
}

func (IsIPv6) isExpr() {}

func (i IpAddr) isIpv6() IsIPv6 {
	return IsIPv6{i}
}

type IsLoopback struct {
	IpAddr IpAddr
}

func (IsLoopback) isExpr() {}

func (i IpAddr) isLoopback() IsLoopback {
	return IsLoopback{i}
}

type IsMulticast struct {
	IpAddr IpAddr
}

func (IsMulticast) isExpr() {}

func (i IpAddr) isMulticast() IsMulticast {
	return IsMulticast{i}
}

type IsInRange struct {
	IpAddr IpAddr
	Range  IpAddr
}

func (IsInRange) isExpr() {}

func (i IpAddr) isInRange(ipRange IpAddr) IsInRange {
	return IsInRange{i, ipRange}
}
