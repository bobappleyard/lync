package transform

import (
	"strings"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func unitMethodCall(name string, args ...ast.Expr) ast.Expr {
	return ast.Call{
		Method: ast.MemberAccess{
			Object: ast.Unit{},
			Member: name,
		},
		Args: args,
	}
}
func newVarSet() *data.Set[string] {
	return data.NewSet(strings.Compare)
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
			vars.Put(s.Name)
		}
	}
	return vars
}
