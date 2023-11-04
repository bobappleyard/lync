package sparse

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
	row := len(m.rows)
	offset := m.findOffset(elements)
	m.ensureEntries(offset)
	m.insertEntries(elements, offset)
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

func (m *Matrix[T]) findOffset(elements []Element[T]) int {

}

func (m *Matrix[T]) ensureEntries(offset int) {

}

func (m *Matrix[T]) insertEntries(elements []Element[T], offset int) {

}
