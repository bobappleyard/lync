package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func transformBoxing(p ast.Program) ast.Program {
	boxes := &boxing{boxed: newVarSet()}
	boxes.fallbackTransformer = fallbackTransformer{boxes}
	return ast.Program{
		Stmts: boxes.transformBlock(p.Stmts),
	}
}

type boxing struct {
	fallbackTransformer
	boxed *data.Set[string]
	args  *data.Set[string]
}

type boxScopeAnalyzer struct {
	fallbackAnalyzer
	inClosure bool
	locals    *data.Set[string]
	referred  *data.Set[string]
	captured  *data.Set[string]
	boxed     *data.Set[string]
}

func (b *boxing) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	needBoxes := b.needBoxes(stmts)
	res := mapSlice(needBoxes.Items(), b.createBox)
	needBoxes.AddSet(b.boxed)
	inner := &boxing{
		boxed: needBoxes,
	}
	inner.fallbackTransformer = fallbackTransformer{inner}
	for _, s := range stmts {
		res = append(res, inner.transformStmt(s))
	}
	return res
}

func (b *boxing) transformStmt(stmt ast.Stmt) ast.Stmt {
	switch stmt := stmt.(type) {

	case ast.Variable:
		if b.boxed.Contains(stmt.Name) {
			return b.call(stmt.Name, "define", b.transformExpr(stmt.Value))
		}
		return ast.Variable{
			Name:  stmt.Name,
			Value: b.transformExpr(stmt.Value),
		}

	case ast.Assign:
		if b.boxed.Contains(stmt.Name) {
			return b.call(stmt.Name, "set", b.transformExpr(stmt.Value))
		}
		return ast.Assign{
			Object: b.transformExpr(stmt.Object),
			Name:   stmt.Name,
			Value:  b.transformExpr(stmt.Value),
		}

	default:
		return b.fallbackTransformer.transformStmt(stmt)
	}
}

func (b *boxing) transformExpr(expr ast.Expr) ast.Expr {
	switch expr := expr.(type) {

	case ast.VariableRef:
		if b.boxed.Contains(expr.Var) {
			return b.call(expr.Var, "get")
		}
		return expr

	case ast.Function:
		args := newVarSet()
		args.AddSlice(mapSlice(expr.Args, argName))
		inner := &boxing{
			boxed: b.boxed,
			args:  args,
		}
		return ast.Function{
			Args: expr.Args,
			Body: inner.transformBlock(expr.Body),
		}

	default:
		return b.fallbackTransformer.transformExpr(expr)
	}
}

func (b *boxing) needBoxes(stmts []ast.Stmt) *data.Set[string] {
	tracking := &boxScopeAnalyzer{
		referred: newVarSet(),
		captured: newVarSet(),
		boxed:    newVarSet(),
	}
	tracking.fallbackAnalyzer = fallbackAnalyzer{tracking}

	for _, stmt := range stmts {
		tracking.analyzeStmt(stmt)
	}

	return tracking.boxed
}

func (b *boxing) createBox(v string) ast.Stmt {
	var value ast.Expr
	if b.args.Contains(v) {
		value = b.call("", "create_box", ast.VariableRef{Var: v})
	} else {
		value = b.call("", "create_undefined_box", ast.Name{Name: v})
	}
	return ast.Variable{Name: v, Value: value}
}

func (b *boxing) call(varName, methodName string, args ...ast.Expr) ast.Expr {
	var target ast.Expr = ast.Unit{}
	if varName != "" {
		target = ast.VariableRef{Var: varName}
	}
	return ast.Call{
		Method: ast.MemberAccess{
			Object: target,
			Member: methodName,
		},
		Args: args,
	}
}

func (t *boxScopeAnalyzer) analyzeStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {

	case ast.VariableRef:
		if t.locals.Contains(stmt.Var) {
			return
		}
		t.referred.Add(stmt.Var)
		if t.inClosure {
			t.captured.Add(stmt.Var)
		}

	case ast.Variable:
		t.analyzeStmt(stmt.Value)

		if t.locals.Contains(stmt.Name) {
			return
		}
		if t.referred.Contains(stmt.Name) || t.captured.Contains(stmt.Name) {
			t.boxed.Add(stmt.Name)
		}

	case ast.Assign:
		t.analyzeStmt(stmt.Object)
		t.analyzeStmt(stmt.Value)

		if stmt.Object != nil || t.locals.Contains(stmt.Name) {
			return
		}
		if t.inClosure || t.captured.Contains(stmt.Name) {
			t.boxed.Add(stmt.Name)
		}

	case ast.Function:
		locals := newVarSet()
		locals.AddSet(t.locals)
		locals.AddSlice(mapSlice(stmt.Args, argName))

		inner := boxScopeAnalyzer{
			inClosure: true,
			locals:    locals,
			referred:  t.referred,
			captured:  t.captured,
			boxed:     t.boxed,
		}
		inner.analyzeBlock(stmt.Body)

	default:
		t.fallbackAnalyzer.analyzeStmt(stmt)
	}
}

func (t *boxScopeAnalyzer) analyzeBlock(stmts []ast.Stmt) {
	locals := newVarSet()
	locals.AddSet(t.locals)
	locals.AddSet(blockVars(stmts))
	inner := &boxScopeAnalyzer{
		inClosure: t.inClosure,
		locals:    locals,
		referred:  t.referred,
		captured:  t.captured,
		boxed:     t.boxed,
	}
	inner.fallbackAnalyzer = fallbackAnalyzer{inner}
	for _, stmt := range stmts {
		inner.analyzeStmt(stmt)
	}
}
