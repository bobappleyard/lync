package wasm

type Code struct {
	locals []WasmAppender
	code   []byte
}

func (c *Code) AppendWasm(buf []byte) []byte {
	var tmp []byte
	tmp = appendVector(tmp, c.locals)
	tmp = append(tmp, c.code...)

	buf = appendBytes(buf, tmp)
	return buf
}

func (c *Code) I32Const(x uint32) {
	c.code = append(c.code, 0x41)
	c.code = appendUint32(c.code, x)
}

func (c *Code) Call(idx uint32) {
	c.code = append(c.code, 0x10)
	c.code = appendUint32(c.code, idx)
}

func (c *Code) End() {
	c.code = append(c.code, 0x0b)
}

func (c *Code) LoadInt32(align, offset uint32) {
	c.code = append(c.code, 0x28)
	c.code = appendUint32(c.code, align)
	c.code = appendUint32(c.code, offset)
}

func (c *Code) StoreInt32(align, offset uint32) {
	c.code = append(c.code, 0x36)
	c.code = appendUint32(c.code, align)
	c.code = appendUint32(c.code, offset)
}
