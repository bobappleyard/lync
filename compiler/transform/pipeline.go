package transform

import "github.com/bobappleyard/lync/compiler/ast"

func Program(p ast.Program) (ast.Program, error) {
	p = transformDeclarators(p)
	p = transformClosures(p)

	return p, nil
}
