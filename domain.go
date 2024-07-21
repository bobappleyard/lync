package lync

type Unit struct {
	Registers byte
	Code      []byte
	Symbols   []string
}

type Symbol uint64

type Register byte
