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

// 交集 - 返回两个集合中都存在的元素
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// 差集 - 返回在当前集合中但不在另一个集合中的元素
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := NewSet[T]()
	for item := range s {
		if !other.Contains(item) {
			result.Add(item)
		}
	}
	return result
}

// 并集 - 返回两个集合中所有的元素
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := NewSet[T]()

	// 添加当前集合的所有元素
	for item := range s {
		result.Add(item)
	}

	// 添加另一个集合的所有元素
	for item := range other {
		result.Add(item)
	}

	return result
}

// 对称差集 - 返回只在其中一个集合中存在的元素
func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	union := s.Union(other)
	intersection := s.Intersect(other)
	return union.Difference(intersection)
}

// 判断是否是另一个集合的子集
func (s Set[T]) IsSubset(other Set[T]) bool {
	for item := range s {
		if !other.Contains(item) {
			return false
		}
	}
	return true
}

// 判断是否是另一个集合的超集
func (s Set[T]) IsSuperset(other Set[T]) bool {
	return other.IsSubset(s)
}
