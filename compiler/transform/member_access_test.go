package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestMemberAccess(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{
		{
			name: "PropertyGet",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.MemberAccess{
					Object: ast.VariableRef{Var: "x"},
					Member: "y",
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.Unit{},
						Member: "property_get",
					},
					Args: []ast.Expr{
						ast.VariableRef{Var: "x"},
						ast.Name{Name: "y"},
					},
				},
			}},
		},
		{
			name: "PropertySet",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Assign{
					Object: ast.VariableRef{Var: "x"},
					Name:   "y",
					Value:  ast.VariableRef{Var: "z"},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.Unit{},
						Member: "property_set",
					},
					Args: []ast.Expr{
						ast.VariableRef{Var: "x"},
						ast.Name{Name: "y"},
						ast.VariableRef{Var: "z"},
					},
				},
			}},
		},
		{
			name: "MethodCall",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "method",
					},
					Args: []ast.Expr{
						ast.VariableRef{Var: "y"},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "method",
					},
					Args: []ast.Expr{
						ast.VariableRef{Var: "y"},
					},
				},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := transformMemberAccess(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}

}
