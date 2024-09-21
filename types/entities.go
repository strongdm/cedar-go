package types

import (
	"encoding/json"
	"slices"
	"strings"

	"golang.org/x/exp/maps"
)

// An Entities is a collection of all the Entities that are needed to evaluate
// authorization requests.  The key is an EntityUID which uniquely identifies
// the Entity (it must be the same as the UID within the Entity itself.)
type Entities map[EntityUID]*Entity

// An Entity defines the parents and attributes for an EntityUID.
type Entity struct {
	UID        EntityUID   `json:"uid"`
	Parents    []EntityUID `json:"parents"`
	Attributes Record      `json:"attrs"`
}

// MarshalJSON serializes Entity as a JSON object, using the implicit form of EntityUID encoding to match the Rust
// SDK's behavior.
func (e Entity) MarshalJSON() ([]byte, error) {
	parents := make([]ImplicitlyMarshaledEntityUID, len(e.Parents))
	for i, p := range e.Parents {
		parents[i] = ImplicitlyMarshaledEntityUID(p)
	}

	m := struct {
		UID        ImplicitlyMarshaledEntityUID   `json:"uid"`
		Parents    []ImplicitlyMarshaledEntityUID `json:"parents"`
		Attributes Record                         `json:"attrs"`
	}{
		ImplicitlyMarshaledEntityUID(e.UID),
		parents,
		e.Attributes,
	}
	return json.Marshal(m)
}

func (e Entities) MarshalJSON() ([]byte, error) {
	s := maps.Values(e)
	slices.SortFunc(s, func(a, b *Entity) int {
		return strings.Compare(a.UID.String(), b.UID.String())
	})
	return json.Marshal(s)
}

func (e *Entities) UnmarshalJSON(b []byte) error {
	var s []*Entity
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var res = Entities{}
	for _, e := range s {
		res[e.UID] = e
	}
	*e = res
	return nil
}

func (e Entities) Clone() Entities {
	return maps.Clone(e)
}
