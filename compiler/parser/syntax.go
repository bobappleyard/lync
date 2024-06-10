package parser

import (
	"path"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/text"
)

func Parse(src []byte) (ast.Program, error) {
	toks, err := tokenize(src)
	if err != nil {
		return ast.Program{}, err
	}
	return parser.Parse(toks)
}

type syntax struct {
}

var parser = text.NewParser[token, ast.Program](syntax{})

func (syntax) ParseProgram(_ optionalNewline, stmts delimList[ast.Stmt, newlineTok], _ optionalNewline) ast.Program {
	return ast.Program{
		Stmts: stmts.items,
	}
}

func (syntax) ParseImport(imp importTok, importPath stringTok) ast.Stmt {
	name := path.Base(importPath.value())
	return ast.NodeAt(imp.start(), ast.Import{
		Name: name,
		Path: importPath.value(),
	})
}

func (syntax) ParseFunctionStmt(fn funcTok, name idTok, args argList[ast.Arg], stmts block[ast.Stmt]) ast.Stmt {
	return ast.NodeAt(fn.start(), ast.Function{
		Name: name.text(),
		Args: args.items,
		Body: stmts.stmts,
	})
}

func (syntax) ParseClassStmt(class classTok, name idTok, members block[ast.Member]) ast.Stmt {
	return ast.NodeAt(class.start(), ast.Class{
		Name:    name.text(),
		Members: members.stmts,
	})
}

func (syntax) ParseFunctionExpr(fn funcTok, args argList[ast.Arg], stmts block[ast.Stmt]) ast.Expr {
	return ast.NodeAt(fn.start(), ast.Function{
		Name: "",
		Args: args.items,
		Body: stmts.stmts,
	})
}

func (syntax) ParseClassExpr(class classTok, members block[ast.Member]) ast.Expr {
	return ast.NodeAt(class.start(), ast.Class{
		Name:    "",
		Members: members.stmts,
	})
}

func (syntax) ParseArg(arg idTok) ast.Arg {
	return ast.NodeAt(arg.start(), ast.Arg{
		Name: arg.text(),
	})
}

func (syntax) ParseReturn(ret returnTok, value ast.Expr) ast.Stmt {
	return ast.NodeAt(ret.start(), ast.Return{
		Value: value,
	})
}

func (syntax) ParseEmptyReturn(ret returnTok) ast.Stmt {
	return ast.NodeAt(ret.start(), ast.Return{
		Value: ast.VariableRef{Var: "void"},
	})
}

func (syntax) ParseVarDecl(v varTok, name idTok, _ eqTok, value ast.Expr) ast.Stmt {
	return ast.NodeAt(v.start(), ast.Variable{
		Name:  name.text(),
		Value: value,
	})
}

func (syntax) ParseVarAssign(name idTok, _ eqTok, value ast.Expr) ast.Stmt {
	return ast.NodeAt(name.start(), ast.Assign{
		Name:  name.text(),
		Value: value,
	})
}

func (syntax) ParseIf(ifT ifTok, cond ast.Expr, stmts block[ast.Stmt]) ast.Stmt {
	return ast.NodeAt(ifT.start(), ast.If{
		Cond: cond,
		Then: stmts.stmts,
	})
}

func (syntax) ParseString(s stringTok) ast.Expr {
	return ast.NodeAt(s.start(), ast.StringConstant{
		Value: s.value(),
	})
}

func (syntax) ParseInt(i intTok) ast.Expr {
	return ast.NodeAt(i.start(), ast.IntConstant{
		Value: i.value(),
	})
}

func (syntax) ParseFlt(f fltTok) ast.Expr {
	return ast.NodeAt(f.start(), ast.FltConstant{
		Value: f.value(),
	})
}

func (syntax) ParseVarRef(name idTok) ast.Expr {
	return ast.NodeAt(name.start(), ast.VariableRef{
		Var: name.text(),
	})
}

func (syntax) ParseMemberAccess(object ast.Expr, dot dotTok, id idTok) ast.Expr {
	return ast.NodeAt(dot.start(), ast.MemberAccess{
		Object: object,
		Member: id.text(),
	})
}

func (syntax) ParseCall(callable ast.Expr, args argList[ast.Expr]) ast.Expr {
	return ast.NodeAt(args.start, ast.Call{
		Method: callable,
		Args:   args.items,
	})
}

func (syntax) ParseMethod(name idTok, args argList[ast.Arg], body block[ast.Stmt]) ast.Member {
	return ast.NodeAt(name.start(), ast.Method{
		Name: name.text(),
		Args: args.items,
		Body: body.stmts,
	})
}
