package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/parser"
	"github.com/bobappleyard/lync/util/assert"
)

func TestPipeline(t *testing.T) {
	src := `

	import "coll"

	class Router {
		handle(req) {
			return coll.filter(this.routes, func(r) {
				analyze_route(class {
					valid(t) {
						r = req
					}
				})
			})
		}
	}

	`

	prog, err := parser.Parse([]byte(src))
	assert.Nil(t, err)

	core := Program(prog)

	ast2.Print(nil, core)
	t.Fail()

}
