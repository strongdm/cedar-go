package types

type PolicySet struct {
	Policies []Policy
}

type Policy struct {
	Effect     Effect
	Principal  PrincipalConstraint
	Action     ActionConstraint
	Resource   ResourceConstraint
	Conditions []Condition
}

type Effect uint8

const (
	PermitEffect Effect = iota
	ForbidEffect
)

type Condition interface {
	isCondition()
}

type WhenCondition struct {
	Expr
}

func (WhenCondition) isCondition() {}

type UnlessCondition struct {
	Expr
}

func (UnlessCondition) isCondition() {}
