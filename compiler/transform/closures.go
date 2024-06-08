package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

// assumes boxes have been handled
func transformClosures(p ast.Program) ast.Program {
	c := new(closures)
	c.fallbackTransformer = fallbackTransformer{c}

	return ast.Program{
		Stmts: c.transformBlock(p.Stmts),
	}
}

type closures struct {
	fallbackTransformer
	captured *data.Set[string]
}

func (c *closures) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	inner := &closures{captured: c.blockScope(c.captured, stmts)}
	inner.fallbackTransformer = fallbackTransformer{inner}
	return mapSlice(stmts, inner.transformStmt)
}

func (c *closures) transformExpr(e ast.Expr) ast.Expr {
	switch e := e.(type) {

	case ast.Function:
		return c.createClosure(e, c.capturedVariables(e))

	default:
		return c.fallbackTransformer.transformExpr(e)
	}
}

func (c *closures) capturedVariables(f ast.Function) *data.Set[string] {
	locals := newVarSet()
	for _, x := range f.Args {
		locals.Add(x.Name)
	}
	return c.capturedInBlock(locals, f.Body)
}

func (c *closures) capturedInBlock(locals *data.Set[string], ss []ast.Stmt) *data.Set[string] {
	inner := c.blockScope(locals, ss)
	captured := newVarSet()
	for _, s := range ss {
		captured.AddSet(c.capturedInStmt(inner, s))
	}
	return captured
}

func (c *closures) capturedInStmt(locals *data.Set[string], s ast.Stmt) *data.Set[string] {
	captured := newVarSet()
	switch s := s.(type) {
	case ast.VariableRef:
		if c.captured.Contains(s.Var) && !locals.Contains(s.Var) {
			captured.Add(s.Var)
		}

	case ast.MemberAccess:
		captured = c.capturedInStmt(locals, s.Object)

	case ast.Call:
		captured.AddSet(c.capturedInStmt(locals, s.Method))
		for _, x := range s.Args {
			captured.AddSet(c.capturedInStmt(locals, x))
		}

	case ast.Function:
		inner := newVarSet()
		inner.AddSet(locals)
		for _, a := range s.Args {
			inner.Add(a.Name)
		}
		return c.capturedInBlock(inner, s.Body)

	case ast.Return:
		captured = c.capturedInStmt(locals, s.Value)

	case ast.Variable:
		captured = c.capturedInStmt(locals, s.Value)

	case ast.If:
		captured.AddSet(c.capturedInStmt(locals, s.Cond))
		captured.AddSet(c.capturedInBlock(locals, s.Then))
		captured.AddSet(c.capturedInBlock(locals, s.Else))

	default:
		return nil
	}

	return captured
}

func (c *closures) blockScope(base *data.Set[string], ss []ast.Stmt) *data.Set[string] {
	inner := newVarSet()
	inner.AddSet(base)
	inner.AddSet(blockVars(ss))
	return inner
}

func (c *closures) createClosure(f ast.Function, closure *data.Set[string]) ast.Expr {
	captured := newVarSet()
	for _, a := range f.Args {
		captured.Add(a.Name)
	}
	captured.AddSet(closure)
	inner := &closures{captured: captured}
	inner.fallbackTransformer = fallbackTransformer{inner}

	lifted := ast.Function{
		Name: f.Name,
		Args: append(mapSlice(closure.Items(), namedArg), f.Args...),
		Body: inner.transformBlock(f.Body),
	}
	if closure.Empty() {
		return lifted
	}

	return ast.NodeAt(f.Start(), ast.Call{
		Method: ast.MemberAccess{
			Object: ast.Unit{},
			Member: "create_closure",
		},
		Args: append([]ast.Expr{lifted}, mapSlice(closure.Items(), func(name string) ast.Expr {
			return ast.VariableRef{Var: name}
		})...),
	})
}
