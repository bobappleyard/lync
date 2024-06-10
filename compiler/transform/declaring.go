package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
)

func transformDeclarators(p ast.Program) ast.Program {
	d := withFallbackTransformer(&declarators{})
	return ast.Program{Stmts: d.transformBlock(p.Stmts)}
}

type declarators struct {
	fallbackTransformer
}

func (d *declarators) transformStmt(s ast.Stmt) ast.Stmt {
	switch s := s.(type) {

	case ast.Class:
		if s.Name == "" {
			return d.fallbackTransformer.transformStmt(s)
		}
		return ast.Variable{
			Name:  s.Name,
			Value: ast.Class{Members: mapSlice(s.Members, d.transformMember)},
		}

	case ast.Function:
		if s.Name == "" {
			return d.fallbackTransformer.transformStmt(s)
		}
		return ast.Variable{
			Name:  s.Name,
			Value: ast.Function{Args: s.Args, Body: d.transformBlock(s.Body)}}

	case ast.Import:
		return ast.Variable{Name: s.Name, Value: ast.Call{
			Method: ast.MemberAccess{
				Object: ast.Unit{},
				Member: "import_package",
			},
			Args: []ast.Expr{
				ast.StringConstant{Value: s.Path},
			},
		}}

	default:
		return d.fallbackTransformer.transformStmt(s)
	}
}
