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

func appendExport(buf []byte, name string, id byte, index Index) []byte {
	buf = appendString(buf, name)
	buf = append(buf, id)
	buf = index.AppendWasm(buf)

	return buf
}

func (FuncExport) export() {}

func (e FuncExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 0, e.Func)
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

type MemoryExport struct {
	Name string
	Mem  Index
}

func (MemoryExport) export() {}

func (e MemoryExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 2, e.Mem)
}

type MemoryImport struct {
	Module string
	Name   string
	Type   Memory
}

func (MemoryImport) imprt() {}

func (e MemoryImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	buf = append(buf, 2)
	buf = e.Type.AppendWasm(buf)
	return buf
}

type TableExport struct {
	Name  string
	Table Index
}

func (TableExport) export() {}

func (e TableExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 1, e.Table)
}

type TableImport struct {
	Module string
	Name   string
}

func (TableImport) imprt() {}

func (e TableImport) AppendWasm(buf []byte) []byte {
	buf = appendString(buf, e.Module)
	buf = appendString(buf, e.Name)
	// just do functions with no further requirements
	buf = append(buf, 1, 0x70, 0, 0)
	return buf
}

type GlobalExport struct {
	Name   string
	Global Index
}

func (GlobalExport) export() {}

func (e GlobalExport) AppendWasm(buf []byte) []byte {
	return appendExport(buf, e.Name, 3, e.Global)
}
