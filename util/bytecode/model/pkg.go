package model

type Interpreter struct {
	HostPkg string
	Imports []Import
	Opcodes []Opcode
}

type Import struct {
	Rename bool
	Name   string
	Path   string
}

type Opcode struct {
	ID       int
	HasError bool
	Name     string
	Args     []Arg
}

type Arg struct {
	Name string
	Pkg  string
	Type string
}
