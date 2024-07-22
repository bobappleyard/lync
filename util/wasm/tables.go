package wasm

type Table uint32

const (
	FuncTable   Table = 0x70
	ExternTable Table = 0x6f
)

func (t Table) AppendWasm(buf []byte) []byte {
	return appendUint32(buf, uint32(t))
}
