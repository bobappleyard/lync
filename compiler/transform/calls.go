package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func transformFunctionCalls(p ast.Program) ast.Program {
	calls := withFallbackTransformer(&functionCallTransformer{})

	return ast.Program{Stmts: calls.transformBlock(p.Stmts)}
}

type functionCallTransformer struct {
	fallbackTransformer
}

func (t *functionCallTransformer) transformExpr(e ast.Expr) ast.Expr {
	switch e := e.(type) {
	case ast.Call:
		if _, ok := e.Method.(ast.MemberAccess); ok {
			return t.fallbackTransformer.transformExpr(e)
		}
		method := t.transformExpr(e.Method)
		args := data.MapSlice(e.Args, t.transformExpr)
		return ast.Call{
			Method: ast.MemberAccess{
				Object: ast.Unit{},
				Member: "call_function",
			},
			Args: append([]ast.Expr{method}, args...),
		}

	default:
		return t.fallbackTransformer.transformExpr(e)
	}
}
