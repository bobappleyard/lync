package data

import (
	"slices"
)

type Set[T any] struct {
	items []T
	cmp   func(T, T) int
}

func NewSet[T any](cmp func(T, T) int) *Set[T] {
	return &Set[T]{cmp: cmp}
}

func (s *Set[T]) Copy() *Set[T] {
	return &Set[T]{
		items: slices.Clone(s.items),
		cmp:   s.cmp,
	}
}

func (s *Set[T]) Empty() bool {
	return s.Size() == 0
}

func (s *Set[T]) Size() int {
	if s == nil {
		return 0
	}
	return len(s.items)
}

func (s *Set[T]) Put(x T) {
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if ok {
		return
	}
	var zero T
	s.items = append(s.items, zero)
	copy(s.items[idx+1:], s.items[idx:])
	s.items[idx] = x
}

func (s *Set[T]) Get(x T) (T, bool) {
	var zero T
	if s == nil {
		return zero, false
	}
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if !ok {
		return zero, false
	}
	return s.items[idx], true
}

func (s *Set[T]) AddSlice(xs []T) {
	for _, x := range xs {
		s.Put(x)
	}
}

func (s *Set[T]) AddSet(xs *Set[T]) {
	if xs == nil {
		return
	}
	s.AddSlice(xs.items)
}

func (s *Set[T]) Contains(x T) bool {
	_, ok := s.Get(x)
	return ok
}

func (s *Set[T]) Items() []T {
	if s == nil {
		return nil
	}
	return s.items
}
