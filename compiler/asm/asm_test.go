package asm

import (
	"fmt"
	"testing"

	"github.com/bobappleyard/lync"
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestAssemble(t *testing.T) {
	p := ast.Program{
		Stmts: []ast.Stmt{
			ast.Call{
				Method: ast.MemberAccess{
					Object: ast.Unit{},
					Member: "set_global",
				},
				Args: []ast.Expr{
					ast.Name{Name: "f"},
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "x"}, {Name: "y"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.Call{
								Method: ast.MemberAccess{
									Object: ast.VariableRef{Var: "x"},
									Member: "method",
								},
								Args: []ast.Expr{
									ast.VariableRef{Var: "y"},
								},
							}},
						},
					},
				},
			},
			ast.Return{Value: ast.Call{
				Method: ast.MemberAccess{
					Object: ast.Unit{},
					Member: "null",
				},
			}},
		},
	}

	buf, err := AssembleProgram(p)

	assert.Nil(t, err)

	dec := lync.BytecodeDecoder{
		Code: buf.Code,
		Impl: &testProcessor{},
	}

	for dec.Pos < len(dec.Code) {
		dec.Step()
	}

	t.Log(buf)
	t.Fail()
}

type testProcessor struct {
}

// Block implements lync.Bytecode.
func (t *testProcessor) Block(argc byte, varc byte, entry lync.CodeRef) {
	fmt.Println("Block", argc, varc, entry)
}

// Branch implements lync.Bytecode.
func (t *testProcessor) Branch(ref lync.CodeRef) {
	fmt.Println("Branch", ref)
}

// Call implements lync.Bytecode.
func (t *testProcessor) Call(method lync.MethodID, argc byte) {
	fmt.Println("Call", method, argc)
}

// CallTail implements lync.Bytecode.
func (t *testProcessor) CallTail(method lync.MethodID, argc byte) {
	fmt.Println("CallTail", method, argc)
}

// Float implements lync.Bytecode.
func (t *testProcessor) Float(value float64) {
	fmt.Println("Float", value)
}

// Int implements lync.Bytecode.
func (t *testProcessor) Int(value int) {
	fmt.Println("Int", value)
}

// Jump implements lync.Bytecode.
func (t *testProcessor) Jump(ref lync.CodeRef) {
	fmt.Println("Jump", ref)
}

// Load implements lync.Bytecode.
func (t *testProcessor) Load(r lync.Register) {
	fmt.Println("Load", r)
}

// Name implements lync.Bytecode.
func (t *testProcessor) Name(id lync.MethodID) {
	fmt.Println("Name", id)
}

// Return implements lync.Bytecode.
func (t *testProcessor) Return() {
	fmt.Println("Return")
}

// Store implements lync.Bytecode.
func (t *testProcessor) Store(r lync.Register) {
	fmt.Println("Store", r)
}

// String implements lync.Bytecode.
func (t *testProcessor) String(value string) {
	fmt.Println("String", value)
}

// Unit implements lync.Bytecode.
func (t *testProcessor) Unit() {
	fmt.Println("Unit")
}

var _ lync.Bytecode = &testProcessor{}
