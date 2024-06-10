package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestFunctionCall(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{} {
		t.Run(test.name, func(t *testing.T) {
			out := transformFunctionCalls(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}

}
