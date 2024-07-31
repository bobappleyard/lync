package wasm

func (c *Code) AppendWasm(buf []byte) []byte {
	var tmp []byte
	tmp = appendVector(tmp, c.Locals)
	tmp = append(tmp, c.Instructions...)

	buf = appendBytes(buf, tmp)
	return buf
}

func (c LocalDecl) AppendWasm(buf []byte) []byte {
	buf = appendUint32(buf, c.Count)
	buf = c.Type.AppendWasm(buf)

	return buf
}

func (c *Code) op(code byte, args ...uint32) {
	c.Instructions = append(c.Instructions, code)
	for _, arg := range args {
		c.Instructions = appendUint32(c.Instructions, arg)
	}
}

func (c *Code) Call(idx uint32)         { c.op(0x10, idx) }
func (c *Code) CallIndirect(idx uint32) { c.op(0x11, idx, 0) }
func (c *Code) Loop()                   { c.op(0x03, 0x40) }
func (c *Code) If()                     { c.op(0x04, 0x40) }
func (c *Code) Else()                   { c.op(0x05) }
func (c *Code) End()                    { c.op(0x0b) }
func (c *Code) Br(depth uint32)         { c.op(0x0c, depth) }
func (c *Code) BrIf(depth uint32)       { c.op(0x0d, depth) }

func (c *Code) Drop()                { c.op(0x1a) }
func (c *Code) LocalGet(idx uint32)  { c.op(0x20, idx) }
func (c *Code) LocalSet(idx uint32)  { c.op(0x21, idx) }
func (c *Code) GlobalGet(idx uint32) { c.op(0x23, idx) }
func (c *Code) GlobalSet(idx uint32) { c.op(0x24, idx) }

func (c *Code) NullFunc() { c.op(0xd0, 0x70) }

func (c *Code) TableInit(elem, table uint32) { c.op(0xfc, 0x0c, elem, table) }
func (c *Code) TableGrow(table uint32)       { c.op(0xfc, 0xf, table) }
func (c *Code) TableGet(table uint32)        { c.op(0x25, table) }

func (c *Code) I32Load(align, offset uint32)  { c.op(0x28, align, offset) }
func (c *Code) I32Store(align, offset uint32) { c.op(0x36, align, offset) }
func (c *Code) MemGrow()                      { c.op(0x40, 0) }

func (c *Code) I32Const(x uint32) { c.op(0x41, x) }
func (c *Code) I32Eqz()           { c.op(0x45) }
func (c *Code) I32Eq()            { c.op(0x46) }
func (c *Code) I32Ne()            { c.op(0x47) }
func (c *Code) I32Lts()           { c.op(0x48) }
func (c *Code) I32Gts()           { c.op(0x4a) }
func (c *Code) I32Les()           { c.op(0x4c) }
func (c *Code) I32Ges()           { c.op(0x4e) }
func (c *Code) I32Add()           { c.op(0x6a) }
func (c *Code) I32Sub()           { c.op(0x6b) }
func (c *Code) I32Mul()           { c.op(0x6c) }
func (c *Code) I32Div()           { c.op(0x6d) }
func (c *Code) I32And()           { c.op(0x71) }
func (c *Code) I32Or()            { c.op(0x72) }
func (c *Code) I32Shl()           { c.op(0x74) }
func (c *Code) I32Shr()           { c.op(0x76) }
