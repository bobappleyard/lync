package wasm

type Module struct {
	Types    []Type
	Imports  []Import
	Funcs    []Index
	Tables   []Table
	Exports  []Export
	Codes    []*Code
	Elements []Element
}

func (m *Module) AppendWasm(mod []byte) []byte {
	mod = m.wasmHeader(mod)
	mod = appendSection(mod, 1, m.Types)
	mod = appendSection(mod, 2, m.Imports)
	mod = appendSection(mod, 3, m.Funcs)
	mod = appendSection(mod, 4, m.Tables)
	mod = appendSection(mod, 7, m.Exports)
	mod = appendSection(mod, 9, m.Elements)
	mod = appendSection(mod, 10, m.Codes)
	return mod
}

func (m *Module) Func(in []Type, out []Type, code *Code) {
	typeId := -1
	for i, t := range m.Types {
		if !t.Matches(FuncType{In: in, Out: out}) {
			continue
		}
		typeId = i
		break
	}
	if typeId == -1 {
		typeId = len(m.Types)
		m.Types = append(m.Types, FuncType{In: in, Out: out})
	}
	m.Funcs = append(m.Funcs, Index(typeId))
	m.Codes = append(m.Codes, code)
}

func (m *Module) wasmHeader(buf []byte) []byte {
	buf = append(buf, 0)
	buf = append(buf, []byte("asm")...)
	buf = append(buf, 1, 0, 0, 0)
	return buf
}
