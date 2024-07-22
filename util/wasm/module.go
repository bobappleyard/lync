package wasm

type Module struct {
	Types    []Type
	Imports  []Import
	Funcs    []Index
	Tables   []Table
	Memories []Memory
	Exports  []Export
	Codes    []*Code
}

func (m *Module) AppendWasm(mod []byte) []byte {
	mod = m.wasmHeader(mod)
	mod = m.typeSection(mod)
	mod = m.importSection(mod)
	mod = m.funcSection(mod)
	mod = m.tableSection(mod)
	mod = m.memorySection(mod)
	mod = m.exportSection(mod)
	mod = m.codeSection(mod)
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

func (m *Module) typeSection(mod []byte) []byte {
	return appendSection(mod, 1, appendVector(nil, m.Types))
}

func (m *Module) importSection(mod []byte) []byte {
	return appendSection(mod, 2, appendVector(nil, m.Imports))
}

func (m *Module) funcSection(mod []byte) []byte {
	return appendSection(mod, 3, appendVector(nil, m.Funcs))
}

func (m *Module) tableSection(mod []byte) []byte {
	return appendSection(mod, 4, appendVector(nil, m.Tables))
}

func (m *Module) memorySection(mod []byte) []byte {
	return appendSection(mod, 5, appendVector(nil, m.Memories))
}

func (m *Module) exportSection(mod []byte) []byte {
	return appendSection(mod, 7, appendVector(nil, m.Exports))
}

func (m *Module) codeSection(mod []byte) []byte {
	return appendSection(mod, 10, appendVector(nil, m.Codes))
}
