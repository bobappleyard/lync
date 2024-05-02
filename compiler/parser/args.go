package parser

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

type argList[T ast.Node] struct {
	items []T
}

type argItem[T ast.Node] struct {
	value T
}

type argParser[T ast.Node] struct {
}

func (argList[T]) Parser() argParser[T] {
	return argParser[T]{}
}

func (argParser[T]) ParseEmpty(_ openPTok, _ closePTok) argList[T] {
	return argList[T]{}
}

func (argParser[T]) ParseNonEmpty(_ openPTok, arg0 T, rest []argItem[T], _ closePTok) argList[T] {
	items := make([]T, len(rest)+1)
	items[0] = arg0
	for i, x := range rest {
		items[i+1] = x.value
	}
	return argList[T]{items: items}
}

func (argParser[T]) ParseItem(_ commaTok, arg T) argItem[T] {
	return argItem[T]{value: arg}
}
