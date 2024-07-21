package asm

import (
	"context"
	"testing"

	"github.com/bobappleyard/lync/util/wasm"
	"github.com/tetratelabs/wazero"
)

func TestMemory(t *testing.T) {
	var c wasm.Code
	c.I32Const(1)
	c.I32Const(4)
	c.StoreInt32(2, 0)
	c.I32Const(0)
	c.LoadInt32(2, 0)
	c.End()

	m := (&wasm.Module{
		Memories: []wasm.Memory{wasm.MinMemory{Min: 4096}},
		Types:    []wasm.Type{wasm.FuncType{Out: []wasm.Type{wasm.Int32}}},
		Exports:  []wasm.Export{wasm.FuncExport{Name: "example", Func: 0}},
		Funcs:    []wasm.Index{0},
		Codes:    []*wasm.Code{&c},
	}).AppendWasm(nil)

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	mod, err := r.Instantiate(ctx, m)
	if err != nil {
		t.Error(err)
		return
	}

	res, err := mod.ExportedFunction("example").Call(ctx)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(res)
	t.Fail()
}
