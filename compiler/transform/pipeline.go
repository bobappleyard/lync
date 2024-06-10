package transform

import (
	"strings"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func Program(p ast.Program) ast.Program {
	p = transformDeclarators(p)
	p = transformClasses(p)
	p = transformMemberAccess(p)
	p = transformBoxing(p)
	p = transformClosures(p)
	p = transformFunctionCalls(p)

	return p
}

func newVarSet() *data.Set[string] {
	return data.NewSet(strings.Compare)
}

func mapSlice[T, U any](xs []T, f func(T) U) []U {
	res := make([]U, len(xs))
	for i, x := range xs {
		res[i] = f(x)
	}
	return res
}

func argName(x ast.Arg) string {
	return x.Name
}

func namedArg(name string) ast.Arg {
	return ast.Arg{Name: name}
}

func blockVars(ss []ast.Stmt) *data.Set[string] {
	vars := newVarSet()
	for _, s := range ss {
		if s, ok := s.(ast.Variable); ok {
			vars.Add(s.Name)
		}
	}
	return vars
}
