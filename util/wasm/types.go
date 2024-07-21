package wasm

type Type interface {
	WasmAppender
	typ()
}

type FuncType struct {
	In, Out []Type
}

func (FuncType) typ() {}

func (t FuncType) AppendWasm(buf []byte) []byte {
	buf = append(buf, 0x60)
	buf = appendVector(buf, t.In)
	buf = appendVector(buf, t.Out)
	return buf
}

type NumberType byte

const (
	Int32 NumberType = 0x7f - iota
	Int64
	Float32
	Float64
)

func (NumberType) typ() {}

func (t NumberType) AppendWasm(buf []byte) []byte {
	return append(buf, byte(t))
}
