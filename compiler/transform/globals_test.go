package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestGlobals(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{
		{
			name: "NoGlobals",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.VariableRef{Var: "x"}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.VariableRef{Var: "x"}},
					},
				},
			}},
		},
		{
			name: "AssignNonGlobal",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
						ast.Assign{Name: "x", Value: ast.IntConstant{Value: 2}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
						ast.Assign{Name: "x", Value: ast.IntConstant{Value: 2}},
					},
				},
			}},
		},
		{
			name: "ImplicitGlobal",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.VariableRef{Var: "y"}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "global_get",
							},
							Args: []ast.Expr{
								ast.Name{Name: "y"},
							},
						}},
					},
				},
			}},
		},
		{
			name: "ImplicitGlobalUpdate",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Assign{
							Name:  "y",
							Value: ast.VariableRef{Var: "x"},
						},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "global_set",
							},
							Args: []ast.Expr{
								ast.Name{Name: "y"},
								ast.VariableRef{Var: "x"},
							},
						},
					},
				},
			}},
		},
		{
			name: "ExplicitGlobal",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{
					Name:  "y",
					Value: ast.IntConstant{Value: 1},
				},
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.VariableRef{Var: "y"}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.Unit{},
						Member: "global_define",
					},
					Args: []ast.Expr{
						ast.Name{Name: "y"},
						ast.IntConstant{Value: 1},
					},
				},
				ast.Function{
					Name: "f",
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Return{Value: ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "global_get",
							},
							Args: []ast.Expr{
								ast.Name{Name: "y"},
							},
						}},
					},
				},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := transformGlobals(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}

}
