package sparse

import "math"

type Matrix[T any] struct {
	entries []matrixEntry[T]
	rows    []matrixRow
}

type Element[T any] struct {
	Col   int
	Value T
}

type matrixEntry[T any] struct {
	value T
	row   int // the ID of the row this belongs to
	next  int // the next column, or -1
	delta int // distance to the next available entry
}

type matrixRow struct {
	offset int
	start  int
}

func (m *Matrix[T]) AddRow(elements []Element[T]) int {
	start, end := rowBounds(elements)
	row := len(m.rows)

	offset := m.findOffset(elements, start)
	m.ensureEntries(offset + end + 1)
	m.insertEntries(elements, row, offset)

	m.rows = append(m.rows, matrixRow{
		offset: offset,
		start:  elements[0].Col,
	})

	return row
}

func (m *Matrix[T]) LookupValue(row, col int) (T, bool) {
	var zero T

	if row < 0 || row >= len(m.rows) {
		return zero, false
	}

	pos := m.rows[row].offset + col
	if pos < 0 || pos >= len(m.entries) {
		return zero, false
	}

	return m.entries[pos].value, true
}

func (m *Matrix[T]) LookupRow(row int) []Element[T] {
	var res []Element[T]

	if row < 0 || row >= len(m.rows) {
		return res
	}

	info := m.rows[row]

	for cur := info.start; cur != -1; cur = m.entries[info.offset+cur].next {
		res = append(res, Element[T]{
			Col:   cur,
			Value: m.entries[info.offset+cur].value,
		})
	}

	return res
}

func rowBounds[T any](elements []Element[T]) (start, end int) {
	start = math.MaxInt
	for _, e := range elements {
		start = min(start, e.Col)
		end = max(end, e.Col)
	}
	return start, end
}

func (m *Matrix[T]) findOffset(elements []Element[T], start int) int {
	offset := -start

	for {
		delta := 0

		for _, e := range elements {
			if e.Col+offset >= len(m.entries) {
				continue
			}
			delta = max(delta, m.entries[e.Col+offset].delta)
		}

		if delta == 0 {
			break
		}

		offset += delta
	}

	return offset
}

func (m *Matrix[T]) ensureEntries(n int) {
	if n <= len(m.entries) {
		return
	}

	toAdd := make([]matrixEntry[T], n-len(m.entries))
	for i := range toAdd {
		toAdd[i] = matrixEntry[T]{
			row:  -1,
			next: -1,
		}
	}

	m.entries = append(m.entries, toAdd...)
}

func (m *Matrix[T]) insertEntries(elements []Element[T], row, offset int) {
	for i := len(elements) - 1; i >= 0; i-- {
		e := elements[i]

		next := -1
		if i < len(elements)-1 {
			next = elements[i+1].Col
		}

		delta := 1
		if offset+e.Col < len(m.entries)-1 {
			delta = m.entries[offset+e.Col+1].delta + 1
		}
		for j := offset + e.Col - 1; j >= 0 && m.entries[j].row != -1; j-- {
			m.entries[j].delta = m.entries[j+1].delta + 1
		}

		m.entries[e.Col+offset] = matrixEntry[T]{
			value: e.Value,
			next:  next,
			row:   row,
			delta: delta,
		}
	}
}
