package runtime

import "github.com/bobappleyard/lync/util/wasm"

type runtimeScope struct {
	actionType    uint32
	pendingAction uint32
}

func mainLoop(s runtimeScope) *wasm.Code {
	var c wasm.Code

	c.Loop()

	c.GlobalGet(s.pendingAction)
	c.I32Eqz()
	c.BrIf(1)

	c.GlobalGet(s.pendingAction)
	c.CallIndirect(s.actionType)
	c.Br(0)

	c.End()
	c.End()

	return &c
}
