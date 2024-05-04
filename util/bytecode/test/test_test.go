package test

import "testing"

func TestGeneratedCode(t *testing.T) {
	e := ExampleEncoder{}
	e.SetValue(12)
	e.Add(10)

	t.Log(e.Buf)

	d := ExampleDecoder{Code: e.Buf}
	for d.Pos < len(d.Code) {
		d.Step()
	}

	t.Log(d.Impl.value)
	if d.Impl.value != 22 {
		t.Fail()
	}
}
