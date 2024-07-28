package asm

import (
	"errors"
	"fmt"

	"github.com/bobappleyard/lync"
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

var (
	ErrUnsupported = errors.New("unsupported")
)

func AssembleProgram(p ast.Program) (lync.Unit, error) {
	a := assembler{
		enc: &wasmEncoder{},
	}

	a.assembleBlock(block{stmts: p.Stmts, enc: a.enc.Block()})

	for a.pending.Ready() {
		b := a.pending.Dequeue()
		a.assembleBlock(b)
	}

	return a.result(byte(requiredRegisters(p.Stmts)))
}

type moduleEncoder interface {
	Block() blockEncoder
	Bytes() []byte
}

type blockEncoder interface {
	ID() uint32

	Unit()
	Name(value lync.Symbol)
	String(value string)
	Int(value int)
	Float(value float64)
	Block(argc, varc byte, id uint32)

	Load(from lync.Register)
	Store(into lync.Register)

	Call(method lync.Symbol, argc byte)
	CallTail(method lync.Symbol, argc byte)
	Return()
}

type assembler struct {
	err     error
	enc     moduleEncoder
	pending data.Queue[block]
	methods []string
}

type block struct {
	enc   blockEncoder
	vars  []string
	args  []string
	regc  int
	stmts []ast.Stmt
}

func (a *assembler) result(regs byte) (lync.Unit, error) {
	if a.err != nil {
		return lync.Unit{}, a.err
	}
	return lync.Unit{
		Registers: regs,
		Code:      a.enc.Bytes(),
		Symbols:   a.methods,
	}, nil
}

func (a *assembler) assembleBlock(b block) {
	for _, stmt := range b.stmts {
		a.assembleStmt(b, stmt)
		if _, ok := stmt.(ast.Return); ok {
			return
		}
	}
}

func (a *assembler) assembleStmt(b block, s ast.Stmt) {
	if a.err != nil {
		return
	}

	switch s := s.(type) {
	case ast.Return:
		if e, ok := s.Value.(ast.Call); ok {
			a.assembleCall(b, e, blockEncoder.CallTail)
		} else {
			a.assembleExpr(b, s.Value)
			if a.err != nil {
				return
			}
			b.enc.Return()
		}

	case ast.Variable:
		a.assembleExpr(b, s.Value)
		if a.err != nil {
			return
		}
		a.assembleDefineVariable(b, s.Name)

	case ast.Expr:
		a.assembleExpr(b, s)

	default:
		a.err = fmt.Errorf("%T: %w", s, ErrUnsupported)
	}
}

func (a *assembler) assembleDefineVariable(b block, name string) {
	off := b.variableOffset(name)
	if off == -1 {
		a.err = fmt.Errorf("non-block variable %s: %w", name, ErrUnsupported)
		return
	}
	b.enc.Store(lync.Register(off))
}

func (a *assembler) assembleExpr(b block, e ast.Expr) {
	if a.err != nil {
		return
	}

	switch e := e.(type) {
	case ast.Unit:
		b.enc.Unit()

	case ast.Name:
		b.enc.Name(a.methodID(e.Name))

	case ast.StringConstant:
		b.enc.String(e.Value)

	case ast.IntConstant:
		b.enc.Int(e.Value)

	case ast.FltConstant:
		b.enc.Float(e.Value)

	case ast.VariableRef:
		off := b.variableOffset(e.Var)
		if off == -1 {
			a.err = fmt.Errorf("non-block variable %s: %w", e.Var, ErrUnsupported)
			return
		}
		b.enc.Load(lync.Register(off))

	case ast.Call:
		a.assembleCall(b, e, blockEncoder.Call)

	case ast.Function:
		args := getArgs(e.Args)
		vars := bindings(e.Body)
		regc := requiredRegisters(e.Body)
		enc := a.enc.Block()
		a.pending.Enqueue(block{
			enc:   enc,
			args:  args,
			vars:  vars,
			regc:  regc,
			stmts: e.Body,
		})
		b.enc.Block(byte(len(args)), byte(len(vars)+regc), enc.ID())

	default:
		a.err = fmt.Errorf("%T: %w", e, ErrUnsupported)
	}
}

func (a *assembler) assembleCall(b block, e ast.Call, write func(blockEncoder, lync.Symbol, byte)) {
	for i, x := range e.Args {
		a.assembleExpr(b, x)
		if a.err != nil {
			return
		}
		b.enc.Store(lync.Register(i))
	}

	m, ok := e.Method.(ast.MemberAccess)
	if !ok {
		a.err = fmt.Errorf("calling objects as functions: %w", ErrUnsupported)
		return
	}

	a.assembleExpr(b, m.Object)
	if a.err != nil {
		return
	}
	write(b.enc, a.methodID(m.Member), byte(len(e.Args)))
}

const frameWidth = 2

func (b block) variableOffset(name string) int {
	for i, v := range b.vars {
		if name == v {
			return i + b.regc
		}
	}
	for i, v := range b.args {
		if name == v {
			return i + b.regc + len(b.vars) + frameWidth
		}
	}
	return -1
}

func (a *assembler) methodID(name string) lync.Symbol {
	for i, m := range a.methods {
		if name == m {
			return lync.Symbol(i)
		}
	}
	ret := len(a.methods)
	a.methods = append(a.methods, name)
	return lync.Symbol(ret)
}

func bindings(stmts []ast.Stmt) []string {
	var names []string

	for _, s := range stmts {
		if s, ok := s.(ast.Variable); ok {
			names = append(names, s.Name)
		}
	}

	return names
}

func getArgs(args []ast.Arg) []string {
	res := make([]string, len(args))
	for i, a := range args {
		res[i] = a.Name
	}
	return res
}

func requiredRegisters(stmts []ast.Stmt) int {
	regs := 0

	for _, s := range stmts {
		switch s := s.(type) {
		case ast.Expr:
			regs = max(regs, requiredRegistersInExpr(s))

		case ast.Return:
			regs = max(regs, requiredRegistersInExpr(s.Value))

		case ast.Variable:
			regs = max(regs, requiredRegistersInExpr(s.Value))

		case ast.If:
			regs = max(regs, requiredRegistersInExpr(s.Cond))
			regs = max(regs, requiredRegisters(s.Then))
			regs = max(regs, requiredRegisters(s.Else))
		}
	}

	return regs
}

func requiredRegistersInExpr(e ast.Expr) int {
	switch e := e.(type) {
	case ast.MemberAccess:
		return requiredRegistersInExpr(e.Object)

	case ast.Call:
		regs := len(e.Args)
		regs = max(regs, requiredRegistersInExpr(e.Method))
		for _, x := range e.Args {
			regs = max(regs, requiredRegistersInExpr(x))
		}
		return regs
	}

	return 0
}
