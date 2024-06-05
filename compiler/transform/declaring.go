package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

func transformDeclarators(p ast.Program) ast.Program {
	d := declarators{}
	return ast.Program{
		Stmts: d.transformBlock(p.Stmts),
	}
}

type declarators struct{}

func (d declarators) transformBlock(ss []ast.Stmt) []ast.Stmt {
	return mapSlice(ss, d.transformStmt)
}

func (d declarators) transformStmt(s ast.Stmt) ast.Stmt {
	switch s := s.(type) {

	case ast.Class:
		transformed := ast.Class{Members: mapSlice(s.Members, d.transformMember)}
		if s.Name == "" {
			return transformed
		}
		return ast.Variable{
			Name:  s.Name,
			Value: transformed,
		}

	case ast.Function:
		transformed := ast.Function{Args: s.Args, Body: d.transformBlock(s.Body)}
		if s.Name == "" {
			return transformed
		}
		return ast.Variable{Name: s.Name, Value: transformed}

	case ast.Import:
		return ast.Variable{Name: s.Name, Value: ast.Call{
			Method: ast.MemberAccess{
				Object: ast.Unit{},
				Member: "import_pkg",
			},
			Args: []ast.Expr{
				ast.StringConstant{Value: s.Path},
			},
		}}

	case ast.If:
		return ast.If{
			Cond: d.transformExpr(s.Cond),
			Then: d.transformBlock(s.Then),
			Else: d.transformBlock(s.Else),
		}

	case ast.Return:
		return ast.Return{
			Value: d.transformExpr(s.Value),
		}

	case ast.Expr:
		return d.transformExpr(s)

	default:
		return s
	}
}

func (d declarators) transformMember(m ast.Member) ast.Member {
	switch m := m.(type) {

	case ast.Method:
		m.Body = d.transformBlock(m.Body)
		return m

	default:
		return m
	}
}

func (d declarators) transformExpr(e ast.Expr) ast.Expr {
	switch e := e.(type) {

	case ast.MemberAccess:
		e.Object = d.transformExpr(e.Object)
		return e

	case ast.Call:
		return ast.Call{
			Method: d.transformExpr(e.Method),
			Args:   mapSlice(e.Args, d.transformExpr),
		}

	case ast.Class:
		e.Members = mapSlice(e.Members, d.transformMember)
		return e

	case ast.Function:
		e.Body = d.transformBlock(e.Body)
		return e

	default:
		return e
	}

}
