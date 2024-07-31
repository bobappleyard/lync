package wasm

type Module struct {
	Types    []Type
	Imports  []Import
	Funcs    []Index
	Tables   []Table
	Memories []Memory
	Exports  []Export
	Codes    []*Code
	Elements []Element
}

type Code struct {
	Func         Index
	Locals       []LocalDecl
	Instructions []byte
}

type LocalDecl struct {
	Count uint32
	Type  Type
}

func (m *Module) AppendWasm(mod []byte) []byte {
	mod = m.wasmHeader(mod)
	mod = appendSection(mod, 1, m.Types)
	mod = appendSection(mod, 2, m.Imports)
	mod = appendSection(mod, 3, m.Funcs)
	mod = appendSection(mod, 4, m.Tables)
	mod = appendSection(mod, 5, m.Memories)
	mod = appendSection(mod, 7, m.Exports)
	mod = appendSection(mod, 9, m.Elements)
	mod = appendSection(mod, 10, m.Codes)
	return mod
}

func (m *Module) EnsureType(t Type) Index {
	typeID := -1
	for i, u := range m.Types {
		if !t.Matches(u) {
			continue
		}
		typeID = i
		break
	}
	if typeID == -1 {
		typeID = len(m.Types)
		m.Types = append(m.Types, t)
	}
	return Index(typeID)
}

func (m *Module) AddFunc(in []Type, out []Type) *Code {
	typeID := m.EnsureType(FuncType{In: in, Out: out})
	idx := len(m.Funcs)
	res := &Code{Func: Index(idx)}
	m.Funcs = append(m.Funcs, typeID)
	m.Codes = append(m.Codes, res)
	return res
}

func (m *Module) AddExportedFunc(name string, in []Type, out []Type) *Code {
	c := m.AddFunc(in, out)
	m.Exports = append(m.Exports, FuncExport{Name: name, Func: c.Func})
	return c
}

func (m *Module) wasmHeader(buf []byte) []byte {
	buf = append(buf, 0)
	buf = append(buf, []byte("asm")...)
	buf = append(buf, 1, 0, 0, 0)
	return buf
}
