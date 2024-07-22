package wasm

type Code struct {
	locals []LocalDecl
	code   []byte
}

type LocalDecl struct {
	Count uint32
	Type  Type
}

func (c *Code) AppendWasm(buf []byte) []byte {
	var tmp []byte
	tmp = appendVector(tmp, c.locals)
	tmp = append(tmp, c.code...)

	buf = appendBytes(buf, tmp)
	return buf
}

func (c LocalDecl) AppendWasm(buf []byte) []byte {
	buf = appendUint32(buf, c.Count)
	buf = c.Type.AppendWasm(buf)

	return buf
}

func (c *Code) op(code byte, args ...uint32) {
	c.code = append(c.code, code)
	for _, arg := range args {
		c.code = appendUint32(c.code, arg)
	}
}

func (c *Code) I32Const(x uint32)               { c.op(0x41, x) }
func (c *Code) Call(idx uint32)                 { c.op(0x10, idx) }
func (c *Code) If()                             { c.op(0x04, 0x40) }
func (c *Code) Else()                           { c.op(0x05) }
func (c *Code) End()                            { c.op(0x0b) }
func (c *Code) LoadInt32(align, offset uint32)  { c.op(0x28, align, offset) }
func (c *Code) StoreInt32(align, offset uint32) { c.op(0x36, align, offset) }
func (c *Code) LocalGet(idx uint32)             { c.op(0x20, idx) }
func (c *Code) LocalSet(idx uint32)             { c.op(0x21, idx) }
func (c *Code) GlobalGet(idx uint32)            { c.op(0x23, idx) }
func (c *Code) GlobalSet(idx uint32)            { c.op(0x24, idx) }
