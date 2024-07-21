package transform

import (
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

type testTransformer struct {
	fallbackTransformer
}

func TestWithTransformer(t *testing.T) {
	tr := withFallbackTransformer(new(testTransformer))

	in := ast.Class{Members: []ast.Member{
		ast.Method{
			Name: "test",
			Body: []ast.Stmt{
				ast.Return{Value: ast.IntConstant{Value: 1}},
			},
		},
	}}

	assert.Equal(t, tr.transformStmt(in), ast.Stmt(in))
}

func TestFallbackIdentity(t *testing.T) {

}
