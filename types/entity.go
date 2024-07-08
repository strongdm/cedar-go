package types

type EntityType struct {
	ID         string
	Namespaces []string
}

type EntityUID struct {
	Type EntityType
	ID   string
}
