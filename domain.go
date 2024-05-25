//go:generate go run github.com/bobappleyard/lync/util/bytecode -src Bytecode
package lync

type Unit struct {
	Registers byte
	Code      []byte
	Methods   []string
}

type Object interface {
	Invoke(p Process) Action
}

type MethodID int

type Action struct{}

type Process interface {
	MethodID() MethodID
	Argc() int
	Arg(n int) Object

	Return(x Object) Action
	Error(e Object) Action

	Method(x Object, id MethodID) CallBuilder
}

type CallBuilder interface {
	WithArg(x Object) CallBuilder
	CallTail() Action
	Call() Object
}

type Register byte
type CodeRef uint32

type Bytecode interface {
	Load(r Register)
	Store(r Register)

	Int(value int)
	String(value string)
	Float(value float64)
	Name(id MethodID)
	Block(argc, varc byte, entry CodeRef)

	Unit()

	Branch(ref CodeRef)
	Jump(ref CodeRef)
	Call(method MethodID, argc byte)
	CallTail(method MethodID, argc byte)
	Return()
}
