package transform

import "github.com/bobappleyard/lync/compiler/ast"

// assumes class syntax has been resolved
func transformMemberAccess(p ast.Program) ast.Program {
	members := new(memberAccessTransformer)
	members.fallbackTransformer = fallbackTransformer{members}

	return ast.Program{Stmts: members.transformBlock(p.Stmts)}
}

type memberAccessTransformer struct {
	fallbackTransformer
}

func (m *memberAccessTransformer) transformStmt(stmt ast.Stmt) ast.Stmt {
	switch stmt := stmt.(type) {

	case ast.Assign:
		if stmt.Object != nil {
			return ast.Call{
				Method: ast.MemberAccess{
					Object: ast.Unit{},
					Member: "property_set",
				},
				Args: []ast.Expr{
					m.transformExpr(stmt.Object),
					ast.Name{Name: stmt.Name},
					m.transformExpr(stmt.Value),
				},
			}
		}
		return ast.Assign{
			Name:  stmt.Name,
			Value: m.transformExpr(stmt.Value),
		}

	default:
		return m.fallbackTransformer.transformStmt(stmt)
	}

}

func (m *memberAccessTransformer) transformExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {
	case ast.MemberAccess:
		return ast.Call{
			Method: ast.MemberAccess{
				Object: ast.Unit{},
				Member: "property_get",
			},
			Args: []ast.Expr{
				m.transformExpr(expr.Object),
				ast.Name{Name: expr.Member},
			},
		}

	case ast.Call:
		if method, ok := expr.Method.(ast.MemberAccess); ok {
			return ast.Call{
				Method: ast.MemberAccess{
					Object: m.transformExpr(method.Object),
					Member: method.Member,
				},
				Args: mapSlice(expr.Args, m.transformExpr),
			}
		}
		return m.fallbackTransformer.transformExpr(expr)

	default:
		return m.fallbackTransformer.transformExpr(expr)

	}
}
