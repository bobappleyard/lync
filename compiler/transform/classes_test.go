package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestClasses(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{
		{
			name: "EmptyClass",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Class{Members: []ast.Member{}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{Method: ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "@", Value: ast.Call{
						Method: ast.MemberAccess{
							Object: ast.Unit{},
							Member: "create_class",
						},
					}},
					ast.Return{Value: ast.VariableRef{Var: "@"}},
				}}},
			}},
		},
		{
			name: "ClassWithMethod",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Class{Members: []ast.Member{
					ast.Method{
						Name: "example",
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.VariableRef{Var: "x"}},
						},
					},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{Method: ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "@", Value: ast.Call{
						Method: ast.MemberAccess{
							Object: ast.Unit{},
							Member: "create_class",
						},
					}},
					ast.Assign{
						Object: ast.VariableRef{Var: "@"},
						Name:   "example",
						Value: ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "create_method",
							},
							Args: []ast.Expr{
								ast.Function{
									Args: []ast.Arg{{Name: "this"}, {Name: "x"}},
									Body: []ast.Stmt{
										ast.Return{Value: ast.VariableRef{Var: "x"}},
									},
								},
							},
						},
					},
					ast.Return{Value: ast.VariableRef{Var: "@"}},
				}}},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := transformClasses(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}

}
