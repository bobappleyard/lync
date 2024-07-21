package wasm

type Import interface {
	WasmAppender
	imprt()
}

type Export interface {
	WasmAppender
	export()
}

type FuncExport struct {
	Name string
	Func Index
}

func (FuncExport) export() {}

func (e FuncExport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Name)
	buf = append(buf, 0)
	buf = e.Func.AppendWasm(buf)
	return buf
}

type FuncImport struct {
	Module string
	Name   string
	Type   Index
}

func (FuncImport) imprt() {}

func (e FuncImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	buf = append(buf, 0)
	buf = e.Type.AppendWasm(buf)
	return buf
}
