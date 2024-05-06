package parser

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

type argList[T ast.Node] struct {
	start int
	items []T
}

type argItem[T ast.Node] struct {
	value T
}

type argParser[T ast.Node] struct{}

func (argList[T]) Parser() argParser[T] {
	return argParser[T]{}
}

func (argParser[T]) ParseEmpty(op openPTok, _ closePTok) argList[T] {
	return argList[T]{
		start: op.start(),
	}
}

func (argParser[T]) ParseNonEmpty(op openPTok, arg0 T, rest []argItem[T], _ closePTok) argList[T] {
	items := make([]T, len(rest)+1)
	items[0] = arg0
	for i, x := range rest {
		items[i+1] = x.value
	}
	return argList[T]{
		start: op.start(),
		items: items,
	}
}

func (argParser[T]) ParseItem(_ commaTok, arg T) argItem[T] {
	return argItem[T]{value: arg}
}

type block[T ast.Node] struct {
	stmts []T
}

type blockList[T ast.Node] struct {
	stmts []T
}

type blockItem[T ast.Node] struct {
	stmt T
}

type blockParser[T ast.Node] struct{}

type blockListParser[T ast.Node] struct{}

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

func (blockParser[T]) ParseBlock(_ openBTok, stmts blockList[T], _ closeBTok) block[T] {
	return block[T]{stmts: stmts.stmts}
}

func (blockList[T]) Parser() blockListParser[T] {
	return blockListParser[T]{}
}

func (blockListParser[T]) ParseEmpty() blockList[T] {
	return blockList[T]{}
}

func (blockListParser[T]) ParseSingle(_ optionalNewline, x T) blockList[T] {
	return blockList[T]{
		stmts: []T{x},
	}
}

func (blockListParser[T]) ParseSingleTrailingNewline(_ optionalNewline, x T, _ newlineTok) blockList[T] {
	return blockList[T]{
		stmts: []T{x},
	}
}

func (blockListParser[T]) ParseNonEmpty(_ optionalNewline, stmt0 T, rest []blockItem[T]) blockList[T] {
	stmts := make([]T, len(rest)+1)
	stmts[0] = stmt0
	for i, s := range rest {
		stmts[i+1] = s.stmt
	}

	return blockList[T]{
		stmts: stmts,
	}
}

func (blockListParser[T]) ParseNonEmptyTrailingNewline(_ optionalNewline, stmt0 T, rest []blockItem[T], _ newlineTok) blockList[T] {
	stmts := make([]T, len(rest)+1)
	stmts[0] = stmt0
	for i, s := range rest {
		stmts[i+1] = s.stmt
	}

	return blockList[T]{
		stmts: stmts,
	}
}

func (blockListParser[T]) ParseStmt(_ newlineTok, stmt T) blockItem[T] {
	return blockItem[T]{stmt: stmt}
}
