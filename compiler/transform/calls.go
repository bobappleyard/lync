package transform

import "github.com/bobappleyard/lync/compiler/ast"

func transformFunctionCalls(p ast.Program) ast.Program {
	calls := &functionCallTransformer{}
	calls.fallbackTransformer = fallbackTransformer{calls}

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
		return ast.Call{
			Method: ast.MemberAccess{
				Object: ast.Unit{},
				Member: "call_function",
			},
			Args: append([]ast.Expr{e.Method}, e.Args...),
		}

	default:
		return t.fallbackTransformer.transformExpr(e)
	}
}
