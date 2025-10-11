package collections

// Set hash set allows O(1) check for membership
type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(v T)           { s[v] = struct{}{} }
func (s Set[T]) Remove(v T)        { delete(s, v) }
func (s Set[T]) Has(v T) bool      { _, ok := s[v]; return ok }
func (s Set[T]) Clear()            { clear(s) }
func (s Set[T]) Len() int          { return len(s) }
func NewSet[T comparable]() Set[T] { return make(Set[T]) }
