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

type captureAnalyzer struct {
	fallbackAnalyzer
	inScope  *data.Set[string]
	locals   *data.Set[string]
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
	locals.AddSlice(mapSlice(f.Args, argName))
	analyzer := &captureAnalyzer{
		inScope:  c.captured,
		locals:   locals,
		captured: newVarSet(),
	}
	analyzer.fallbackAnalyzer = fallbackAnalyzer{analyzer}
	analyzer.analyzeBlock(f.Body)
	return analyzer.captured
}

func (c *captureAnalyzer) withLocals(locals *data.Set[string]) *captureAnalyzer {
	inner := &captureAnalyzer{
		inScope:  c.inScope,
		locals:   locals,
		captured: c.captured,
	}
	inner.fallbackAnalyzer = fallbackAnalyzer{inner}
	return inner
}

func (c *captureAnalyzer) analyzeBlock(stmts []ast.Stmt) {
	locals := newVarSet()
	locals.AddSet(c.locals)
	locals.AddSet(blockVars(stmts))

	inner := c.withLocals(locals)
	for _, s := range stmts {
		inner.analyzeStmt(s)
	}
}

func (c *captureAnalyzer) analyzeExpr(e ast.Expr) {
	switch s := e.(type) {
	case ast.VariableRef:
		if c.inScope.Contains(s.Var) && !c.locals.Contains(s.Var) {
			c.captured.Add(s.Var)
		}

	case ast.Function:
		locals := newVarSet()
		locals.AddSet(c.locals)
		for _, a := range s.Args {
			locals.Add(a.Name)
		}
		inner := c.withLocals(locals)
		inner.analyzeBlock(s.Body)

	default:
		c.fallbackAnalyzer.analyzeExpr(e)
	}
}

func (c *closures) blockScope(base *data.Set[string], ss []ast.Stmt) *data.Set[string] {
	inner := newVarSet()
	inner.AddSet(base)
	inner.AddSet(blockVars(ss))
	return inner
}

func (c *closures) createClosure(f ast.Function, closure *data.Set[string]) ast.Expr {
	captured := newVarSet()
	captured.AddSlice(mapSlice(f.Args, argName))
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
