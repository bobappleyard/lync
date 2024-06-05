package transform

import (
	"slices"

	"github.com/bobappleyard/lync/compiler/ast"
)

func transformClosures(p ast.Program) ast.Program {
	var c closures

	return ast.Program{
		Stmts: c.transformBlock(p.Stmts),
	}
}

type closures struct {
	captured vars
}

func (c *closures) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	inner := &closures{c.blockScope(c.captured, stmts)}
	return mapSlice(stmts, inner.transformStmt)
}

func (c *closures) transformStmt(s ast.Stmt) ast.Stmt {
	switch s := s.(type) {
	case ast.Expr:
		return c.transformExpr(s)

	case ast.Return:
		return ast.NodeAt(s.Start(), ast.Return{
			Value: c.transformExpr(s.Value),
		})

	case ast.Variable:
		return ast.NodeAt(s.Start(), ast.Variable{
			Name:  s.Name,
			Value: c.transformExpr(s.Value),
		})

	case ast.If:
		return ast.NodeAt(s.Start(), ast.If{
			Cond: c.transformExpr(s.Cond),
			Then: c.transformBlock(s.Then),
			Else: c.transformBlock(s.Else),
		})

	default:
		return s
	}
}

func (c *closures) transformExpr(e ast.Expr) ast.Expr {
	switch e := e.(type) {
	case ast.MemberAccess:
		return ast.NodeAt(e.Start(), ast.MemberAccess{
			Object: c.transformExpr(e.Object),
			Member: e.Member,
		})

	case ast.Call:
		return ast.NodeAt(e.Start(), ast.Call{
			Method: c.transformExpr(e.Method),
			Args:   mapSlice(e.Args, c.transformExpr),
		})

	case ast.Function:
		return c.createClosure(e, c.capturedVariables(e))

	default:
		return e
	}
}

func (c *closures) capturedVariables(f ast.Function) vars {
	locals := mapSlice(f.Args, argName)
	return c.capturedInBlock(locals, f.Body)
}

func (c *closures) capturedInBlock(locals vars, ss []ast.Stmt) vars {
	inner := c.blockScope(locals, ss)
	var captured vars
	for _, s := range ss {
		captured.addAll(c.capturedInStmt(inner, s))
	}
	return captured
}

func (c *closures) capturedInStmt(locals vars, s ast.Stmt) vars {
	var captured vars
	switch s := s.(type) {
	case ast.VariableRef:
		if c.captured.contains(s.Var) && !locals.contains(s.Var) {
			captured.add(s.Var)
		}

	case ast.MemberAccess:
		captured = c.capturedInStmt(locals, s.Object)

	case ast.Call:
		captured.addAll(c.capturedInStmt(locals, s.Method))
		for _, x := range s.Args {
			captured.addAll(c.capturedInStmt(locals, x))
		}

	case ast.Function:
		var inner vars
		inner.addAll(locals)
		inner.addAll(mapSlice(s.Args, argName))
		return c.capturedInBlock(inner, s.Body)

	case ast.Return:
		captured = c.capturedInStmt(locals, s.Value)

	case ast.Variable:
		captured = c.capturedInStmt(locals, s.Value)

	case ast.If:
		captured.addAll(c.capturedInStmt(locals, s.Cond))
		captured.addAll(c.capturedInBlock(locals, s.Then))
		captured.addAll(c.capturedInBlock(locals, s.Else))

	default:
		return nil
	}

	return captured
}

func (c *closures) blockScope(base vars, ss []ast.Stmt) vars {
	inner := base.clone()
	for _, s := range ss {
		if s, ok := s.(ast.Variable); ok {
			inner.add(s.Name)
		}
	}
	return inner
}

func (c *closures) createClosure(f ast.Function, closure []string) ast.Expr {
	inner := closures{captured: append(mapSlice(f.Args, argName), closure...)}
	slices.Sort(inner.captured)

	capturedArgs := mapSlice(closure, func(name string) ast.Arg {
		return ast.Arg{Name: name}
	})
	lifted := ast.Function{
		Name: f.Name,
		Args: append(capturedArgs, f.Args...),
		Body: inner.transformBlock(f.Body),
	}
	if len(closure) == 0 {
		return lifted
	}

	return ast.NodeAt(f.Start(), ast.Call{
		Method: ast.MemberAccess{
			Object: ast.Unit{},
			Member: "create_closure",
		},
		Args: append([]ast.Expr{lifted}, mapSlice(closure, func(name string) ast.Expr {
			return ast.VariableRef{Var: name}
		})...),
	})
}

func mapSlice[T, U any](xs []T, f func(T) U) []U {
	res := make([]U, len(xs))
	for i, x := range xs {
		res[i] = f(x)
	}
	return res
}

func argName(x ast.Arg) string {
	return x.Name
}
