//go:generate go run github.com/bobappleyard/lync/util/bytecode -src Example
package test

type Example struct {
	value int
}

func (e *Example) SetValue(value int) {
	e.value = value
}

func (e *Example) Add(delta int) {
	e.value += delta
}
