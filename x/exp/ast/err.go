package ast

import (
	"cmp"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/names"
	"github.com/cedar-policy/cedar-go/types"
)

func (n Node) Err() error {
	return n.err
}

func (n Node) Must() Node {
	if n.err != nil {
		panic(n.err)
	}
	return n
}

func (p *Policy) Err() error {
	return p.err
}

func (p *Policy) Must() *Policy {
	if p.err != nil {
		panic(p.err)
	}
	return p
}

func newNode(v IsNode, errs ...error) Node {
	return Node{v: v, err: cmp.Or(errs...)}
}

func nodesErr(nodes []Node) error {
	for _, n := range nodes {
		if n.err != nil {
			return n.err
		}
	}
	return nil
}

func checkEntityType(t types.EntityType) error {
	if !names.IsName(string(t)) {
		return fmt.Errorf("invalid entity type: %q", t)
	}
	return nil
}

func checkEntityUIDs(es ...types.EntityUID) error {
	for _, e := range es {
		if err := checkEntityType(e.Type); err != nil {
			return err
		}
	}
	return nil
}

func checkValue(v types.Value) error {
	switch v := v.(type) {
	case types.EntityUID:
		return checkEntityType(v.Type)
	case types.Set:
		for e := range v.All() {
			if err := checkValue(e); err != nil {
				return err
			}
		}
	case types.Record:
		for _, e := range v.All() {
			if err := checkValue(e); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkAnnotationKey(k types.Ident) error {
	if !names.IsAnyIdent(string(k)) {
		return fmt.Errorf("invalid annotation key: %q", k)
	}
	return nil
}

func checkExtensionName(name types.Path) error {
	if !names.IsName(string(name)) {
		return fmt.Errorf("invalid extension function name: %q", name)
	}
	return nil
}
