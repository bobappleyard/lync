package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestDeclarators(t *testing.T) {
	for _, test := range []struct {
		name string
		in   ast.Program
		out  ast.Program
	}{
		{
			name: "Import",
			in: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{
						Name: "os",
						Path: "os",
					},
				},
			},
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Variable{
						Name: "os",
						Value: ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "import_package",
							},
							Args: []ast.Expr{
								ast.StringConstant{Value: "os"},
							},
						},
					},
				},
			},
		},
		{
			name: "Func",
			in: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.VariableRef{Var: "x"}},
						},
					},
				},
			},
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Variable{
						Name: "f",
						Value: ast.Function{
							Args: []ast.Arg{{Name: "x"}},
							Body: []ast.Stmt{
								ast.Return{Value: ast.VariableRef{Var: "x"}},
							},
						},
					},
				},
			},
		},
		{
			name: "Class",
			in: ast.Program{
				Stmts: []ast.Stmt{
					ast.Class{
						Name: "Pair",
					},
				},
			},
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Variable{
						Name:  "Pair",
						Value: ast.Class{},
					},
				},
			},
		},
		{
			name: "NestedImport",
			in: ast.Program{
				Stmts: []ast.Stmt{
					ast.Call{
						Method: ast.Function{
							Body: []ast.Stmt{
								ast.Import{Name: "x", Path: "woop"},
							},
						},
					},
				},
			},
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Call{
						Method: ast.Function{
							Body: []ast.Stmt{
								ast.Variable{Name: "x", Value: ast.Call{
									Method: ast.MemberAccess{
										Object: ast.Unit{},
										Member: "import_package",
									},
									Args: []ast.Expr{
										ast.StringConstant{Value: "woop"},
									},
								}},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := transformDeclarators(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}
}
