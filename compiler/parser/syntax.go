package parser

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/text"
)

func Parse(src []byte) (ast.Program, error) {
	toks, err := tokenize(src)
	if err != nil {
		return ast.Program{}, err
	}
	return text.Parse[token, ast.Program](syntax{}, toks)
}

type syntax struct {
}

func (syntax) ParseProgram(stmts []ast.Stmt) ast.Program {
	return ast.Program{
		Stmts: stripNoOps(stmts),
	}
}

func (syntax) ParseNoOp(nl newlineTok) ast.Stmt {
	return ast.NodeAt(nl.start(), ast.NoOp{})
}

func stripNoOps(stmts []ast.Stmt) []ast.Stmt {
	var res []ast.Stmt
	for _, s := range stmts {
		if _, ok := s.(ast.NoOp); ok {
			continue
		}
		res = append(res, s)
	}
	return res
}

func (syntax) ParseImport(imp importTok, path stringTok) ast.Stmt {
	return ast.NodeAt(imp.start(), ast.Import{
		Path: path.value(),
	})
}

func (syntax) ParseFunctionStmt(fn funcTok, name idTok, args argList[ast.Arg], _ openBTok, stmts []ast.Stmt, _ closeBTok) ast.Stmt {
	return ast.NodeAt(fn.start(), ast.Method{
		Name: name.text(),
		Args: args.items,
		Body: stripNoOps(stmts),
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

func (syntax) ParseVarDecl(v varTok, name idTok, _ eqTok, value ast.Expr) ast.Stmt {
	return ast.NodeAt(v.start(), ast.Variable{
		Name:  name.text(),
		Value: value,
	})
}

func (syntax) ParseString(s stringTok) ast.Expr {
	return ast.NodeAt(s.start(), ast.StringConstant{
		Value: s.value(),
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
