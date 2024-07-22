package wasm

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func TestConst(t *testing.T) {
	var c Code
	c.I32Const(12)
	c.End()

	var m Module
	m.Func([]Type{Int32}, []Type{Int32}, &c)
	m.Exports = []Export{FuncExport{Name: "test", Func: 0}}

	testModule(t, m, 0, 12)
}

func TestLogic(t *testing.T) {
	var c Code
	c.locals = []LocalDecl{{1, Int32}}
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

	var m Module
	m.Func([]Type{Int32}, []Type{Int32}, &c)
	m.Exports = []Export{FuncExport{Name: "test", Func: 0}}

	testModule(t, m, 0, 1)
	testModule(t, m, 1, 21)
}

func testModule(t *testing.T, m Module, in, out int32) {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	imp := wasmer.NewImportObject()

	mbytes := m.AppendWasm(nil)
	// err := os.WriteFile(t.Name()+".wasm", mbytes, 0666)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }

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

func TestMemory(t *testing.T) {
	var c Code
	c.I32Const(1)
	c.If()
	c.I32Const(0)
	c.I32Const(4)
	c.StoreInt32(2, 0)
	c.End()
	c.I32Const(0)
	c.LoadInt32(2, 0)
	c.End()

	var m Module
	m.Func([]Type{Int32}, []Type{Int32}, &c)
	m.Memories = []Memory{MinMemory{Min: 4096}}
	m.Exports = []Export{FuncExport{Name: "test", Func: 0}}

	testModule(t, m, 0, 4)

}
