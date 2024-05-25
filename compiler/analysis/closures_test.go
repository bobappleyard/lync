package analysis

import (
	"testing"

	ast2 "go/ast"

	"github.com/bobappleyard/lync/compiler/ast"
)

func TestClosures(t *testing.T) {
	p := ast.Program{
		Stmts: []ast.Stmt{
			ast.Function{
				Name: "f",
				Args: []ast.Arg{{Name: "y"}, {Name: "x"}},
				Body: []ast.Stmt{
					ast.Return{Value: ast.Function{
						Args: []ast.Arg{{Name: "a"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.Call{
								Method: ast.VariableRef{Var: "g"},
								Args: []ast.Expr{
									ast.VariableRef{Var: "x"},
									ast.Function{
										Body: []ast.Stmt{
											ast.Return{Value: ast.Call{
												Method: ast.VariableRef{Var: "g"},
												Args: []ast.Expr{
													ast.VariableRef{Var: "y"},
													ast.VariableRef{Var: "x"},
													ast.VariableRef{Var: "a"},
												},
											}},
										},
									},
								},
							}},
						},
					}},
				},
			},
		},
	}
	q := EliminateClosures(p)
	ast2.Print(nil, q)
	t.Fail()
}
