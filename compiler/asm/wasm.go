package asm

import (
	"github.com/bobappleyard/lync/util/wasm"
)

type wasmEncoder struct {
	m wasm.Module
}

type wasmBlockEncoder struct {
	id wasm.Index
}

func (e *wasmEncoder) init() *wasmEncoder {
	e.m.Types = []wasm.Type{
		wasm.FuncType{In: []wasm.Type{wasm.Int64, wasm.Int64}, Out: []wasm.Type{wasm.Int64}},
	}
	e.m.Imports = []wasm.Import{
		wasm.FuncImport{Module: "runtime", Name: "lookup", Type: 0},
	}
	return e
}

func (e *wasmEncoder) Block() blockEncoder {
	panic("unsupported")
}

func (e *wasmEncoder) Bytes() []byte {
	panic("unsupported")
}
