package transform

import "github.com/bobappleyard/lync/compiler/ast"

func transformClasses(p ast.Program) ast.Program {
	classes := withFallbackTransformer(&classTransformer{})

	return ast.Program{Stmts: classes.transformBlock(p.Stmts)}
}

type classTransformer struct {
	fallbackTransformer
}

var classInit = ast.Variable{
	Name: "@",
	Value: ast.Call{
		Method: ast.MemberAccess{
			Object: ast.Unit{},
			Member: "create_class",
		},
	},
}

func (t *classTransformer) transformExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {

	case ast.Class:
		var body []ast.Stmt
		body = append(body, classInit)
		body = append(body, mapSlice(expr.Members, t.implementMember)...)
		body = append(body, ast.Return{Value: ast.VariableRef{Var: "@"}})

		return ast.Call{Method: ast.Function{Body: body}}

	default:
		return t.fallbackTransformer.transformExpr(expr)
	}
}

func (t *classTransformer) implementMember(member ast.Member) ast.Stmt {
	switch member := member.(type) {
	case ast.Method:
		return ast.Assign{
			Object: ast.VariableRef{Var: "@"},
			Name:   member.Name,
			Value: ast.Call{
				Method: ast.MemberAccess{
					Object: ast.Unit{},
					Member: "create_method",
				},
				Args: []ast.Expr{
					ast.Function{
						Args: append([]ast.Arg{{Name: "this"}}, member.Args...),
						Body: t.transformBlock(member.Body),
					},
				},
			},
		}

	default:
		panic("unimplemented")
	}
}
