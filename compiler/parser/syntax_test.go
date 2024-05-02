package parser

import (
	"slices"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
	"github.com/r3labs/diff"
)

func TestSyntax(t *testing.T) {
	src := []byte(`

import "test"

func a(x, y) {
	"a"
	var z = y.field
	return x
}

`)

	expected := ast.Program{
		Stmts: []ast.Stmt{
			ast.Import{Path: "test"},
			ast.NodeAt(17, ast.Method{
				Name: "a",
				Args: []ast.Arg{
					{Name: "x"},
					{Name: "y"},
				},
				Body: []ast.Stmt{
					ast.StringConstant{
						Value: "a",
					},
					ast.Variable{
						Name: "z",
						Value: ast.MemberAccess{
							Object: ast.VariableRef{
								Var: "y",
							},
							Member: "field",
						},
					},
					ast.Return{
						Value: ast.VariableRef{
							Var: "x",
						},
					},
				},
			}),
		},
	}

	prog, err := Parse(src)

	assert.Nil(t, err)

	cl, _ := diff.Diff(expected, prog)
	for _, c := range cl {
		if c.Type == "update" && len(c.Path) > 2 && slices.Equal([]string{"astNodeData", "s"}, c.Path[len(c.Path)-2:]) {
			continue
		}
		t.Error(c)
	}
	t.Log(prog)
}
