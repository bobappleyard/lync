package data

import "slices"

type Set[T any] struct {
	items []T
	cmp   func(T, T) int
}

func NewSet[T any](cmp func(T, T) int) *Set[T] {
	return &Set[T]{cmp: cmp}
}

func (s *Set[T]) Empty() bool {
	return s == nil || len(s.items) == 0
}

func (s *Set[T]) Add(x T) {
	idx, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	if ok {
		return
	}
	if idx >= len(s.items) {
		s.items = append(s.items, x)
		return
	}
	var zero T
	s.items = append(s.items, zero)
	copy(s.items[idx+1:], s.items[idx:])
	s.items[idx] = x
}

func (s *Set[T]) AddSlice(xs []T) {
	for _, x := range xs {
		s.Add(x)
	}
}

func (s *Set[T]) AddSet(xs *Set[T]) {
	if xs == nil {
		return
	}
	s.AddSlice(xs.items)
}

func (s *Set[T]) Contains(x T) bool {
	if s == nil {
		return false
	}
	_, ok := slices.BinarySearchFunc(s.items, x, s.cmp)
	return ok
}

func (s *Set[T]) Items() []T {
	if s == nil {
		return nil
	}
	return s.items
}
