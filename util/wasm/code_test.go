package wasm

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func TestConst(t *testing.T) {
	var m Module

	c := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})
	c.I32Const(12)
	c.End()

	testModule(t, m, 0, 12)
}

func TestLogic(t *testing.T) {
	var m Module
	c := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})

	c.Locals = []LocalDecl{{1, Int32}}

	c.LocalGet(0)
	c.If()
	c.I32Const(21)
	c.LocalSet(1)
	c.Else()
	c.I32Const(1)
	c.LocalSet(1)
	c.End()
	c.LocalGet(1)
	c.End()

	testModule(t, m, 0, 1)
	testModule(t, m, 1, 21)
}

func TestCall(t *testing.T) {
	var m Module

	f := m.AddFunc([]Type{Int32}, []Type{Int32})
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.End()

	g := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})
	g.LocalGet(0)
	g.Call(0)
	g.End()

	testModule(t, m, 10, 11)
}

func TestCallIndirect(t *testing.T) {
	var m Module

	f := m.AddFunc([]Type{Int32}, []Type{Int32})
	f.LocalGet(0)
	f.I32Const(1)
	f.I32Add()
	f.End()

	g := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})
	g.Locals = []LocalDecl{{1, Int32}}

	// TableGrow: [fillWith, growAmount] -> [oldSize]
	g.NullFunc()
	g.I32Const(10)
	g.TableGrow(0)
	g.Drop()

	// TableGrow: [fillWith, growAmount] -> [oldSize]
	g.NullFunc()
	g.I32Const(1)
	g.TableGrow(0)
	g.LocalSet(1)

	// TableInit: [destPos, srcPos, size] -> []
	g.LocalGet(1)
	g.I32Const(0)
	g.I32Const(1)
	g.TableInit(0, 0)

	g.LocalGet(0)
	g.LocalGet(1)
	g.CallIndirect(0)
	g.End()

	m.Types = []Type{FuncType{In: []Type{Int32}, Out: []Type{Int32}}}
	m.Tables = []Table{FuncTable}
	m.Elements = []Element{&FuncElement{Funcs: []Index{0}}}

	testModule(t, m, 5, 6)
}

func TestLoop(t *testing.T) {
	var m Module
	c := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})

	// var acc = 0
	c.Locals = []LocalDecl{{1, Int32}}

	c.Loop()

	// acc = acc + n
	c.LocalGet(0)
	c.LocalGet(1)
	c.I32Add()
	c.LocalSet(1)

	// n = n - 1
	c.LocalGet(0)
	c.I32Const(1)
	c.I32Sub()
	c.LocalSet(0)

	// for n > 0
	c.LocalGet(0)
	c.BrIf(0)

	c.End()

	c.LocalGet(1)
	c.End()

	testModule(t, m, 3, 6)
}

func TestMemory(t *testing.T) {
	var m Module
	m.Memories = []Memory{MinMemory{0}}

	c := m.AddExportedFunc("test", []Type{Int32}, []Type{Int32})
	c.Locals = []LocalDecl{{1, Int32}}

	// [amt] -> [old]
	c.I32Const(1)
	c.MemGrow()
	c.LocalSet(1)

	// [addr, val] -> []
	c.I32Const(1024)
	c.I32Const(45)
	c.I32Store(2, 0)

	// [addr] -> [val]
	c.I32Const(1024)
	c.I32Load(2, 0)

	c.End()

	testModule(t, m, 0, 45)
}

func testModule(t *testing.T, m Module, in, out int32) {
	t.Helper()

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	imp := wasmer.NewImportObject()

	mbytes := m.AppendWasm(nil)
	mod, err := wasmer.NewModule(store, mbytes)
	if err != nil {
		t.Error(err)
		return
	}

	inst, err := wasmer.NewInstance(mod, imp)
	if err != nil {
		t.Error(err)
		return
	}

	test, err := inst.Exports.GetFunction("test")
	if err != nil {
		t.Error(err)
		return
	}

	res, err := test(in)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, res.(int32), out)
}
