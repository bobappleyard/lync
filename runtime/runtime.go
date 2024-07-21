package runtime

import (
	"context"

	"github.com/bobappleyard/lync/util/wasm"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Construct a WASM module representing the runtime.
// Contract is primarily with asm package - it generates WASM modules for the lync packages.
func runtimeModule(ctx context.Context, r wazero.Runtime) (api.Module, error) {
	var m wasm.Module
	cfg := wazero.NewModuleConfig().WithName("runtime")
	return r.InstantiateWithConfig(ctx, m.AppendWasm(nil), cfg)
}
