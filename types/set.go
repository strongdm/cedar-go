package types

import (
	"bytes"
	"encoding/json"

	"github.com/cedar-policy/cedar-go/internal/sets"
)

// A Set is an immutable collection of elements that can be of the same or different types.
type Set struct {
	sets.ImmutableHashSet[Value]
	hashVal uint64
}

// NewSet returns an immutable Set given a Go slice of Values. Duplicates are removed and order is not preserved.
func NewSet(v []Value) Set {
	set := sets.NewImmutableHashSet(v)

	// Special case hashVal for empty set to 0 so that the return value of Value.Hash() of Set{} and NewSet([]Value{})
	// are the same
	var hashVal uint64
	set.Iterate(func(v Value) bool {
		hashVal += v.Hash()
		return true
	})

	return Set{ImmutableHashSet: set, hashVal: hashVal}
}

// Equal returns true if the sets are Equal.
func (as Set) Equal(bi Value) bool {
	bs, ok := bi.(Set)
	if !ok {
		return false
	}

	return as.ImmutableHashSet.Equal(bs.ImmutableHashSet)
}

// UnmarshalJSON parses a JSON-encoded Cedar set literal into a Set
func (v *Set) UnmarshalJSON(b []byte) error {
	var res []explicitValue
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	vals := make([]Value, len(res))
	for i, vv := range res {
		vals[i] = vv.Value
	}

	*v = NewSet(vals)
	return nil
}

// String produces a string representation of the Set, e.g. `[1,2,3]`.
func (v Set) String() string { return string(v.MarshalCedar()) }

// MarshalCedar produces a valid MarshalCedar language representation of the Set, e.g. `[1,2,3]`.
// Set elements are rendered in a non-deterministic order.
func (v Set) MarshalCedar() []byte {
	var sb bytes.Buffer
	sb.WriteRune('[')
	var i int
	v.Iterate(func(vv Value) bool {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.Write(vv.MarshalCedar())
		i++
		return true
	})
	sb.WriteRune(']')
	return sb.Bytes()
}

func (v Set) Hash() uint64 {
	return v.hashVal
}
