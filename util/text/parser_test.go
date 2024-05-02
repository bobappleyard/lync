package text

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
)

type testTok interface {
	testTok()
}

type intTok struct {
	value int
}

type plusTok struct {
}

func (intTok) testTok()  {}
func (plusTok) testTok() {}

type testExpr interface {
	testExpr()
}

type add struct {
	left, right testExpr
}

type intVal struct {
	value int
}

type intList struct {
	vals []int
}

func (add) testExpr()    {}
func (intVal) testExpr() {}

type ruleset struct {
}

func (ruleset) ParseExprInt(val intTok) intVal {
	return intVal(val)
}

func (ruleset) ParseExprAdd(left testExpr, op plusTok, right testExpr) add {
	return add{left: left, right: right}
}

func TestGrammar(t *testing.T) {
	toks := []testTok{
		intTok{1},
		plusTok{},
		intTok{2},
		plusTok{},
		intTok{3},
	}

	expr, err := Parse[testTok, testExpr](ruleset{}, toks)
	assert.Nil(t, err)
	assert.Equal[testExpr](t, expr, add{
		left:  add{left: intVal{value: 1}, right: intVal{value: 2}},
		right: intVal{value: 3},
	})
}

type nullableRuleset struct {
}

func (nullableRuleset) ParseInt(left intList, val intTok) intList {
	return intList{append(left.vals, val.value)}
}

func (nullableRuleset) ParseNull() intList {
	return intList{}
}

func TestNullableGrammar(t *testing.T) {
	toks := []testTok{
		intTok{1},
		intTok{2},
		intTok{3},
	}

	expr, err := Parse[testTok, intList](nullableRuleset{}, toks)
	assert.Nil(t, err)
	assert.Equal(t, intList{[]int{1, 2, 3}}, expr)
}

type sliceRuleset struct {
}

func (sliceRuleset) ParseInts(ints []intTok) intList {
	vals := make([]int, len(ints))
	for i, t := range ints {
		vals[i] = t.value
	}
	return intList{vals: vals}
}

func TestSliceGrammar(t *testing.T) {
	toks := []testTok{
		intTok{1},
		intTok{2},
		intTok{3},
	}

	expr, err := Parse[testTok, intList](sliceRuleset{}, toks)
	assert.Nil(t, err)
	assert.Equal(t, intList{[]int{1, 2, 3}}, expr)
}

type polymorphicRuleset struct {
}

type delim[T, D any] struct {
	items []T
}

type delimItem[T, D any] struct {
	item T
}

type delimParser[T, D any] struct {
}

func (d delim[T, D]) Parser() delimParser[T, D] {
	return delimParser[T, D]{}
}

func (delimParser[T, D]) ParseEmpty() delim[T, D] {
	return delim[T, D]{}
}

func (delimParser[T, D]) ParseNonEmpty(item0 T, rest []delimItem[T, D]) delim[T, D] {
	items := make([]T, len(rest)+1)
	items[0] = item0
	for i, item := range rest {
		items[i+1] = item.item
	}
	return delim[T, D]{items: items}
}

func (delimParser[T, D]) ParseItem(_ D, item T) delimItem[T, D] {
	return delimItem[T, D]{item: item}
}

func (polymorphicRuleset) ParseIntList(items delim[int, plusTok]) intList {
	return intList{vals: items.items}
}

func (polymorphicRuleset) ParseInt(tok intTok) int {
	return tok.value
}

func TestReusableParser(t *testing.T) {
	toks := []testTok{
		intTok{1},
		plusTok{},
		intTok{2},
		plusTok{},
		intTok{3},
	}

	expr, err := Parse[testTok, intList](polymorphicRuleset{}, toks)
	assert.Nil(t, err)
	assert.Equal(t, intList{[]int{1, 2, 3}}, expr)

}
