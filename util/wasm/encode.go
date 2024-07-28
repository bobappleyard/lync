package wasm

type WasmAppender interface {
	AppendWasm(buf []byte) []byte
}

type Index uint32

func (i Index) AppendWasm(buf []byte) []byte {
	return appendUint32(buf, uint32(i))
}

func appendUint32(buf []byte, x uint32) []byte {
	for x >= 0x80 {
		buf = append(buf, byte(x|0x80))
		x >>= 7
	}
	buf = append(buf, byte(x))
	return buf
}

func appendVector[T WasmAppender](buf []byte, xs []T) []byte {
	buf = appendUint32(buf, uint32(len(xs)))
	for _, x := range xs {
		buf = x.AppendWasm(buf)
	}
	return buf
}

func appendBytes(buf, bs []byte) []byte {
	buf = appendUint32(buf, uint32(len(bs)))
	buf = append(buf, bs...)
	return buf
}

func appendString(buf []byte, s string) []byte {
	return appendBytes(buf, []byte(s))
}

func appendSection[T WasmAppender](buf []byte, id byte, sec []T) []byte {
	if len(sec) == 0 {
		return buf
	}
	buf = append(buf, id)
	buf = appendBytes(buf, appendVector(nil, sec))
	return buf
}
