package wasm

import (
	"testing"

	"github.com/bobappleyard/lync/util/assert"
	"github.com/wasmerio/wasmer-go/wasmer"
)

func TestTables(t *testing.T) {
	var m1 Module
	m1.Tables = []Table{FuncTable}
	m1.Exports = []Export{TableExport{Name: "table"}}

	var m2 Module
	m2.Imports = []Import{TableImport{Module: "m1", Name: "table"}}

	f := m2.AddExportedFunc("test", []Type{Int32}, []Type{Int32})
	f.I32Const(2)
	f.End()

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)
	imp := wasmer.NewImportObject()

	mod1, err := wasmer.NewModule(store, m1.AppendWasm(nil))
	if err != nil {
		t.Error(err)
		return
	}
	inst1, err := wasmer.NewInstance(mod1, imp)
	if err != nil {
		t.Error(err)
		return
	}

	defs, err := exportsToNamespace(mod1, inst1)
	if err != nil {
		t.Error(err)
		return
	}

	imp.Register("m1", defs)

	mod2, err := wasmer.NewModule(store, m2.AppendWasm(nil))
	if err != nil {
		t.Error(err)
		return
	}

	inst2, err := wasmer.NewInstance(mod2, imp)
	if err != nil {
		t.Error(err)
		return
	}

	test, err := inst2.Exports.GetFunction("test")
	if err != nil {
		t.Error(err)
		return
	}

	res, err := test(int32(1))
	if err != nil {
		t.Error(err)
		return
	}

	assert.Equal(t, res.(int32), 2)
}

func exportsToNamespace(module *wasmer.Module, instance *wasmer.Instance) (map[string]wasmer.IntoExtern, error) {
	res := map[string]wasmer.IntoExtern{}
	for _, e := range module.Exports() {
		def, err := instance.Exports.Get(e.Name())
		if err != nil {
			return nil, err
		}
		res[e.Name()] = def
	}
	return res, nil
}
