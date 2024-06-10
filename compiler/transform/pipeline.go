package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

func Program(p ast.Program) ast.Program {
	p = transformDeclarators(p)
	p = transformClasses(p)
	p = transformMemberAccess(p)
	p = transformGlobals(p)
	p = transformBoxing(p)
	p = transformClosures(p)
	p = transformFunctionCalls(p)

	return p
}
