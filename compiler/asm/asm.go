package asm

import (
	"encoding/binary"
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
	var a assembler

	a.assembleBlock(block{stmts: p.Stmts})

	for a.pending.Ready() {
		b := a.pending.Dequeue()
		binary.LittleEndian.PutUint32(a.enc.Buf[b.usedAt+3:], uint32(len(a.enc.Buf)))
		a.assembleBlock(b)
	}

	return a.result(byte(requiredRegisters(p.Stmts)))
}

type assembler struct {
	err     error
	enc     lync.BytecodeEncoder
	pending data.Queue[block]
	methods []string
}

type block struct {
	usedAt uint32
	vars   []string
	args   []string
	regc   int
	stmts  []ast.Stmt
}

func (a *assembler) result(regs byte) (lync.Unit, error) {
	if a.err != nil {
		return lync.Unit{}, a.err
	}
	return lync.Unit{
		Registers: regs,
		Code:      a.enc.Buf,
		Methods:   a.methods,
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
			a.assembleCall(b, e, (*lync.BytecodeEncoder).CallTail)
		} else {
			a.assembleExpr(b, s.Value)
			if a.err != nil {
				return
			}
			a.err = a.enc.Return()
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
	a.err = a.enc.Store(lync.Register(off))
}

func (a *assembler) assembleExpr(b block, e ast.Expr) {
	if a.err != nil {
		return
	}

	switch e := e.(type) {
	case ast.Unit:
		a.err = a.enc.Unit()

	case ast.Name:
		a.err = a.enc.Name(a.methodID(e.Name))

	case ast.StringConstant:
		a.err = a.enc.String(e.Value)

	case ast.IntConstant:
		a.err = a.enc.Int(e.Value)

	case ast.FltConstant:
		a.err = a.enc.Float(e.Value)

	case ast.VariableRef:
		off := b.variableOffset(e.Var)
		if off == -1 {
			a.err = fmt.Errorf("non-block variable %s: %w", e.Var, ErrUnsupported)
			return
		}
		a.err = a.enc.Load(lync.Register(off))

	case ast.Call:
		a.assembleCall(b, e, (*lync.BytecodeEncoder).Call)

	case ast.Function:
		args := getArgs(e.Args)
		vars := bindings(e.Body)
		regc := requiredRegisters(e.Body)
		a.pending.Enqueue(block{
			usedAt: uint32(len(a.enc.Buf)),
			args:   args,
			vars:   vars,
			regc:   regc,
			stmts:  e.Body,
		})
		a.err = a.enc.Block(byte(len(args)), byte(len(vars)+regc), 0)

	default:
		a.err = fmt.Errorf("%T: %w", e, ErrUnsupported)
	}
}

func (a *assembler) assembleCall(b block, e ast.Call, write func(*lync.BytecodeEncoder, lync.MethodID, byte) error) {
	for i, x := range e.Args {
		a.assembleExpr(b, x)
		if a.err != nil {
			return
		}
		a.err = a.enc.Store(lync.Register(i))
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
	a.err = write(&a.enc, a.methodID(m.Member), byte(len(e.Args)))
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

func (a *assembler) methodID(name string) lync.MethodID {
	for i, m := range a.methods {
		if name == m {
			return lync.MethodID(i)
		}
	}
	ret := len(a.methods)
	a.methods = append(a.methods, name)
	return lync.MethodID(ret)
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
