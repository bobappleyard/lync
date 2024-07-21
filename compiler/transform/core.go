package transform

import (
	"unsafe"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
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

type transformer interface {
	transformBlock(stmts []ast.Stmt) []ast.Stmt
	transformStmt(stmt ast.Stmt) ast.Stmt
	transformExpr(expr ast.Expr) ast.Expr
	transformMember(member ast.Member) ast.Member
}

type fallbackTransformer struct {
	impl transformer
}

func withFallbackTransformer[T any, PT interface {
	transformer
	*T
}](tr PT) PT {
	fbt := (*fallbackTransformer)(unsafe.Pointer(tr))
	fbt.impl = tr
	return tr
}

func (t fallbackTransformer) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	return data.MapSlice(stmts, t.impl.transformStmt)
}

func (t fallbackTransformer) transformStmt(stmt ast.Stmt) ast.Stmt {
	switch stmt := stmt.(type) {

	case ast.Assign:
		return ast.Assign{
			Object: t.impl.transformExpr(stmt.Object),
			Name:   stmt.Name,
			Value:  t.impl.transformExpr(stmt.Value),
		}

	case ast.Return:
		return ast.Return{Value: t.impl.transformExpr(stmt.Value)}

	case ast.Variable:
		return ast.Variable{
			Name:  stmt.Name,
			Value: t.impl.transformExpr(stmt.Value),
		}

	case ast.If:
		return ast.If{
			Cond: t.impl.transformExpr(stmt.Cond),
			Then: t.impl.transformBlock(stmt.Then),
			Else: t.impl.transformBlock(stmt.Else),
		}

	case ast.Expr:
		return t.impl.transformExpr(stmt)

	default:
		return stmt
	}
}

func (t fallbackTransformer) transformExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {

	case ast.MemberAccess:
		return ast.MemberAccess{
			Object: t.impl.transformExpr(expr.Object),
			Member: expr.Member,
		}

	case ast.Call:
		return ast.Call{
			Method: t.impl.transformExpr(expr.Method),
			Args:   data.MapSlice(expr.Args, t.impl.transformExpr),
		}

	case ast.Class:
		return ast.Class{
			Name:    expr.Name,
			Members: data.MapSlice(expr.Members, t.impl.transformMember),
		}

	case ast.Function:
		return ast.Function{
			Name: expr.Name,
			Args: expr.Args,
			Body: t.impl.transformBlock(expr.Body),
		}

	default:
		return expr
	}
}

func (t fallbackTransformer) transformMember(member ast.Member) ast.Member {
	switch member := member.(type) {

	case ast.Method:
		return ast.Method{
			Name: member.Name,
			Args: member.Args,
			Body: t.impl.transformBlock(member.Body),
		}

	default:
		return member
	}
}

type analyzer interface {
	analyzeBlock(stmts []ast.Stmt)
	analyzeStmt(stmt ast.Stmt)
	analyzeExpr(expr ast.Expr)
	analyzeMember(member ast.Member)
}

type fallbackAnalyzer struct {
	impl analyzer
}

func withFallbackAnalzyer[T any, PT interface {
	analyzer
	*T
}](tr PT) PT {
	fbt := (*fallbackAnalyzer)(unsafe.Pointer(tr))
	fbt.impl = tr
	return tr
}

func (a fallbackAnalyzer) analyzeBlock(stmts []ast.Stmt) {
	for _, s := range stmts {
		a.impl.analyzeStmt(s)
	}
}

func (a fallbackAnalyzer) analyzeStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {

	case ast.Assign:
		a.impl.analyzeExpr(stmt.Object)
		a.impl.analyzeExpr(stmt.Value)

	case ast.Return:
		a.impl.analyzeExpr(stmt.Value)

	case ast.Variable:
		a.impl.analyzeExpr(stmt.Value)

	case ast.If:
		a.impl.analyzeExpr(stmt.Cond)
		a.impl.analyzeBlock(stmt.Then)
		a.impl.analyzeBlock(stmt.Else)

	case ast.Expr:
		a.impl.analyzeExpr(stmt)

	}
}

func (a fallbackAnalyzer) analyzeExpr(expr ast.Expr) {
	switch expr := expr.(type) {

	case ast.MemberAccess:
		a.impl.analyzeExpr(expr.Object)

	case ast.Call:
		a.impl.analyzeExpr(expr.Method)
		for _, x := range expr.Args {
			a.impl.analyzeExpr(x)
		}

	case ast.Class:
		for _, m := range expr.Members {
			a.impl.analyzeMember(m)
		}

	case ast.Function:
		a.impl.analyzeBlock(expr.Body)

	}
}

func (a fallbackAnalyzer) analyzeMember(member ast.Member) {
	switch member := member.(type) {

	case ast.Method:
		a.impl.analyzeBlock(member.Body)

	}
}
