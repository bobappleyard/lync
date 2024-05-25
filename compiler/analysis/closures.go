package analysis

import (
	"slices"

	"github.com/bobappleyard/lync/compiler/ast"
)

func EliminateClosures(p ast.Program) ast.Program {
	var c closures

	return ast.Program{
		Stmts: c.analyzeBlock(p.Stmts),
	}
}

type closures struct {
	vars []string
}

func (c *closures) analyzeBlock(stmts []ast.Stmt) []ast.Stmt {
	return mapSlice(stmts, c.analyzeStmt)
}

func (c *closures) analyzeStmt(s ast.Stmt) ast.Stmt {
	switch s := s.(type) {
	case ast.Expr:
		return c.analyzeExpr(s)

	case ast.Return:
		return ast.NodeAt(s.Start(), ast.Return{
			Value: c.analyzeExpr(s.Value),
		})

	case ast.Variable:
		return ast.NodeAt(s.Start(), ast.Variable{
			Name:  s.Name,
			Value: c.analyzeExpr(s.Value),
		})

	case ast.If:
		return ast.NodeAt(s.Start(), ast.If{
			Cond: c.analyzeExpr(s.Cond),
			Then: c.analyzeBlock(s.Then),
			Else: c.analyzeBlock(s.Else),
		})

	default:
		return s
	}
}

func (c *closures) analyzeExpr(e ast.Expr) ast.Expr {
	switch e := e.(type) {
	case ast.MemberAccess:
		return ast.NodeAt(e.Start(), ast.MemberAccess{
			Object: c.analyzeExpr(e.Object),
			Member: e.Member,
		})

	case ast.Call:
		return ast.NodeAt(e.Start(), ast.Call{
			Method: c.analyzeExpr(e.Method),
			Args:   mapSlice(e.Args, c.analyzeExpr),
		})

	case ast.Function:
		return c.createClosure(e, c.capturedVariables(e))

	default:
		return e
	}
}

func (c *closures) capturedVariables(f ast.Function) []string {
	locals := c.localVariables(f)
	return c.capturedInBlock(locals, f.Body)
}

func (c *closures) localVariables(f ast.Function) []string {
	// just do fn args for now, do var decls later
	vars := mapSlice(f.Args, argName)
	slices.Sort(vars)
	return vars
}

func (c *closures) capturedInBlock(locals []string, ss []ast.Stmt) []string {
	var vars []string
	for _, s := range ss {
		vars = append(vars, c.capturedInStmt(locals, s)...)
	}
	slices.Sort(vars)
	return uniq(vars)
}

func (c *closures) capturedInStmt(locals []string, s ast.Stmt) []string {
	switch s := s.(type) {
	case ast.VariableRef:
		_, local := slices.BinarySearch(locals, s.Var)
		_, captured := slices.BinarySearch(c.vars, s.Var)
		if captured && !local {
			return []string{s.Var}
		}
		return nil

	case ast.MemberAccess:
		return c.capturedInStmt(locals, s.Object)

	case ast.Call:
		vars := c.capturedInStmt(locals, s.Method)
		for _, x := range s.Args {
			vars = append(vars, c.capturedInStmt(locals, x)...)
		}
		slices.Sort(vars)
		return uniq(vars)

	case ast.Function:
		inner := append([]string{}, locals...)
		inner = append(inner, c.localVariables(s)...)
		return c.capturedInBlock(inner, s.Body)

	case ast.Return:
		return c.capturedInStmt(locals, s.Value)

	case ast.Variable:
		return c.capturedInStmt(locals, s.Value)

	case ast.If:
		vars := c.capturedInStmt(locals, s.Cond)
		vars = append(vars, c.capturedInBlock(locals, s.Then)...)
		vars = append(vars, c.capturedInBlock(locals, s.Else)...)
		slices.Sort(vars)
		return uniq(vars)

	default:
		return nil
	}
}

func (c *closures) createClosure(f ast.Function, closure []string) ast.Expr {
	inner := closures{vars: append(mapSlice(f.Args, argName), closure...)}
	slices.Sort(inner.vars)

	lifted := ast.Function{
		Name: f.Name,
		Args: append(mapSlice(closure, func(name string) ast.Arg { return ast.Arg{Name: name} }), f.Args...),
		Body: inner.analyzeBlock(f.Body),
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

func uniq[T comparable](xs []T) []T {
	if len(xs) == 0 {
		return xs
	}

	needsNew := false
	for i, x := range xs[1:] {
		if xs[i] == x {
			needsNew = true
			break
		}
	}
	if !needsNew {
		return xs
	}

	res := []T{xs[0]}
	for i, x := range xs[1:] {
		if xs[i] != x {
			res = append(res, x)
		}
	}
	return res
}
