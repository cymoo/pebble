package types

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T], len(items))
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func (s Set[T]) Add(item T) {
	s[item] = struct{}{}
}

func (s Set[T]) Remove(item T) {
	delete(s, item)
}

func (s Set[T]) Contains(item T) bool {
	_, exists := s[item]
	return exists
}

// Intersect - Returns elements common to both sets
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// Difference - Returns elements in the current set that are not in the other set
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// Union - Returns all elements from both sets
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := NewSet[T]()

	// Add all elements from the current set
	for item := range s {
		result.Add(item)
	}

	// Add all elements from the other set
	for item := range other {
		result.Add(item)
	}

	return result
}

// SymmetricDifference - Returns elements in either set but not in both
func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	union := s.Union(other)
	intersection := s.Intersect(other)
	return union.Difference(intersection)
}

// IsSubset - Checks if the current set is a subset of another set
func (s Set[T]) IsSubset(other Set[T]) bool {
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// IsSuperset - Checks if the current set is a superset of another set
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}
