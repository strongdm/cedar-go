package sets

import (
	"bytes"
	"encoding/json"

	"golang.org/x/exp/maps"
)

type item[T any] interface {
	Equal(T) bool
	Hash() uint64
}

// A ImmutableHashSet is an immutable collection of hashable elements that are themselves immutable
type ImmutableHashSet[T item[T]] struct {
	s map[uint64]T
}

// NewImmutableHashSet returns an immutable ImmutableHashSet given a Go slice of Values. Duplicates are removed and
// order is not preserved.
func NewImmutableHashSet[T item[T]](i []T) ImmutableHashSet[T] {
	var set map[uint64]T
	if len(i) > 0 {
		set = make(map[uint64]T)
	}
	for _, ii := range i {
		hash := ii.Hash()

		// Insert the item into the map. Deal with collisions via open addressing by simply incrementing the hash
		// value. This method is safe so long as ImmutableHashSet is immutable because nothing can be removed from the
		// map.
		for {
			existing, ok := set[hash]
			if !ok {
				set[hash] = ii
				break
			} else if ii.Equal(existing) {
				// found duplicate in slice
				break
			}
			hash++
		}
	}

	return ImmutableHashSet[T]{s: set}
}

// Len returns the number of unique items in the ImmutableHashSet
func (s ImmutableHashSet[T]) Len() int {
	return len(s.s)
}

// Iterate calls iter for each item in the set. Returning false from the iter function causes iteration to cease.
// Iteration order is non-deterministic.
func (s ImmutableHashSet[T]) Iterate(iter func(i T) bool) {
	for _, v := range s.s {
		if !iter(v) {
			break
		}
	}
}

// Contains returns true if the item i is present in the set
func (s ImmutableHashSet[T]) Contains(i item[T]) bool {
	hash := i.Hash()

	for {
		existing, ok := s.s[hash]
		if !ok {
			return false
		} else if i.Equal(existing) {
			return true
		}
		hash++
	}
}

// Slice returns a slice of the items in the ImmutableHashSet which is safe to mutate. The order of the values is
// non-deterministic.
func (s ImmutableHashSet[T]) Slice() []T {
	if s.s == nil {
		return nil
	}
	return maps.Values(s.s)
}

// Equal returns true if the ImmutableHashSets are equal.
func (as ImmutableHashSet[T]) Equal(bs ImmutableHashSet[T]) bool {
	if len(as.s) != len(bs.s) {
		return false
	}

	for _, v := range as.s {
		if !bs.Contains(v) {
			return false
		}
	}
	return true
}

// UnmarshalJSON parses a JSON array into an ImmutableHashSet
func (s *ImmutableHashSet[T]) UnmarshalJSON(b []byte) error {
	var res []T
	err := json.Unmarshal(b, &res)
	if err != nil {
		return err
	}

	*s = NewImmutableHashSet(res)
	return nil
}

// MarshalJSON marshals the ImmutableHashSet into a JSON array.
// ImmutableHashSet elements are rendered in a non-deterministic order.
func (s ImmutableHashSet[T]) MarshalJSON() ([]byte, error) {
	w := &bytes.Buffer{}
	w.WriteByte('[')
	var i int
	for _, v := range s.s {
		if i != 0 {
			w.WriteByte(',')
		}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		w.Write(b)
		i++
	}
	w.WriteByte(']')
	return w.Bytes(), nil
}
