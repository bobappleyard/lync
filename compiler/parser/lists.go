package parser

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

type delimList[T ast.Node, D token] struct {
	items []T
}

type delimItem[T ast.Node, D token] struct {
	value T
}

type delimParser[T ast.Node, D token] struct{}

func (delimList[T, D]) Parser() delimParser[T, D] {
	return delimParser[T, D]{}
}

func (delimParser[T, D]) ParseEmpty() delimList[T, D] {
	return delimList[T, D]{}
}

func (delimParser[T, D]) ParseNonEmpty(item0 T, rest []delimItem[T, D]) delimList[T, D] {
	items := make([]T, len(rest)+1)
	items[0] = item0
	for i, x := range rest {
		items[i+1] = x.value
	}
	return delimList[T, D]{items: items}
}

func (delimParser[T, D]) ParseItem(_ D, item T) delimItem[T, D] {
	return delimItem[T, D]{value: item}
}

type argList[T ast.Node] struct {
	start int
	items []T
}

type argParser[T ast.Node] struct{}

func (argList[T]) Parser() argParser[T] {
	return argParser[T]{}
}

func (argParser[T]) ParseArgs(op openPTok, items delimList[T, commaTok], _ closePTok) argList[T] {
	return argList[T]{
		start: op.start(),
		items: items.items,
	}
}

type block[T ast.Node] struct {
	stmts []T
}

type blockParser[T ast.Node] struct{}

type optionalNewline struct{}

func (syntax) ParseNewline(_ newlineTok) optionalNewline {
	return optionalNewline{}
}

func (syntax) ParseNoNewline() optionalNewline {
	return optionalNewline{}
}

func (block[T]) Parser() blockParser[T] {
	return blockParser[T]{}
}

func (blockParser[T]) ParseBlock(_ openBTok, _ optionalNewline, stmts delimList[T, newlineTok], _ optionalNewline, _ closeBTok) block[T] {
	return block[T]{
		stmts: stmts.items,
	}
}
