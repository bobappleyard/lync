package link

type Package struct {
	Code []byte
	Data []byte
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
