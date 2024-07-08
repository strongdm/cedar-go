package types

type PrincipalConstraint interface {
	isPrincipalConstraint()
}

type ActionConstraint interface {
	isActionConstraint()
}

type ResourceConstraint interface {
	isResourceConstraint()
}

type EqScopeConstraint struct {
	Entity EntityUID
}

func (EqScopeConstraint) isPrincipalConstraint() {}
func (EqScopeConstraint) isActionConstraint()    {}
func (EqScopeConstraint) isResourceConstraint()  {}

type IsScopeConstraint EntityType

func (IsScopeConstraint) isPrincipalConstraint() {}
func (IsScopeConstraint) isResourceConstraint()  {}

type InEntityScopeConstraint EntityUID

func (InEntityScopeConstraint) isPrincipalConstraint() {}
func (InEntityScopeConstraint) isActionConstraint()    {}
func (InEntityScopeConstraint) isResourceConstraint()  {}

type InEntitiesScopeConstraint []EntityUID

func (InEntitiesScopeConstraint) isActionConstraint() {}

type IsInScopeConstraint struct {
	EntityType
	EntityUID
}

func (IsInScopeConstraint) isPrincipalConstraint() {}
func (IsInScopeConstraint) isResourceConstraint()  {}
