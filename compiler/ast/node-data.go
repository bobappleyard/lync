package ast

import "unsafe"

func NodeAt[T Node](start int, n T) T {
	d := (*astNodeData)(unsafe.Pointer(&n))
	d.s = start
	return n
}

type astNodeData struct {
	s int
}

func (d astNodeData) Start() int {
	return d.s
}

func (d astNodeData) node() {}
