package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func transformGlobals(p ast.Program) ast.Program {
	globals := withFallbackTransformer(&globalsTransformer{})
	return ast.Program{Stmts: globals.transformToplevel(p.Stmts)}
}

type globalsTransformer struct {
	fallbackTransformer
	nonGlobal *data.Set[string]
}

func (t *globalsTransformer) transformToplevel(stmts []ast.Stmt) []ast.Stmt {
	return t.fallbackTransformer.transformBlock(stmts)
}

func (t *globalsTransformer) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	locals := newVarSet()
	locals.AddSet(t.nonGlobal)
	locals.AddSet(blockVars(stmts))
	inner := withFallbackTransformer(&globalsTransformer{nonGlobal: locals})

	return mapSlice(stmts, inner.transformStmt)
}

func (t *globalsTransformer) transformStmt(stmt ast.Stmt) ast.Stmt {
	switch stmt := stmt.(type) {

	case ast.Assign:
		if stmt.Object != nil || t.nonGlobal.Contains(stmt.Name) {
			return t.fallbackTransformer.transformStmt(stmt)
		}
		return unitMethodCall("global_set", ast.Name{Name: stmt.Name}, t.transformExpr(stmt.Value))

	case ast.Variable:
		if t.nonGlobal.Contains(stmt.Name) {
			return t.fallbackTransformer.transformStmt(stmt)
		}
		return unitMethodCall("global_define", ast.Name{Name: stmt.Name}, t.transformExpr(stmt.Value))

	default:
		return t.fallbackTransformer.transformStmt(stmt)
	}
}

func (t *globalsTransformer) transformExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {

	case ast.VariableRef:
		if t.nonGlobal.Contains(expr.Var) {
			return expr
		}
		return unitMethodCall("global_get", ast.Name{Name: expr.Var})

	case ast.Function:
		locals := newVarSet()
		locals.AddSet(t.nonGlobal)
		locals.AddSlice(mapSlice(expr.Args, argName))
		inner := withFallbackTransformer(&globalsTransformer{nonGlobal: locals})
		return ast.Function{
			Name: expr.Name,
			Args: expr.Args,
			Body: inner.transformBlock(expr.Body),
		}

	default:
		return t.fallbackTransformer.transformExpr(expr)

	}
}
