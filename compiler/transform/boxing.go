package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func introduceBoxing(p ast.Program) ast.Program {
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

type boxScopeTracking struct {
	locals   *data.Set[string]
	referred *data.Set[string]
	captured *data.Set[string]
	boxed    *data.Set[string]
}

func (b *boxing) transformBlock(stmts []ast.Stmt) []ast.Stmt {
	needBoxes := b.needBoxes(stmts)
	res := mapSlice(needBoxes.Items(), func(v string) ast.Stmt {
		var value ast.Expr
		if b.args.Contains(v) {
			value = b.call("", "create_box", ast.VariableRef{Var: v})
		} else {
			value = b.call("", "create_undefined_box", ast.Name{Name: v})
		}
		return ast.Variable{Name: v, Value: value}
	})
	for _, b := range b.boxed.Items() {
		needBoxes.Add(b)
	}
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
	tracking := boxScopeTracking{
		referred: newVarSet(),
		captured: newVarSet(),
		boxed:    newVarSet(),
	}

	for _, stmt := range stmts {
		tracking.analyzeStmt(stmt, false)
	}

	return tracking.boxed
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

func (t boxScopeTracking) analyzeStmt(stmt ast.Stmt, inClosure bool) {
	switch stmt := stmt.(type) {

	case ast.VariableRef:
		if t.locals.Contains(stmt.Var) {
			return
		}
		t.referred.Add(stmt.Var)
		if inClosure {
			t.captured.Add(stmt.Var)
		}

	case ast.Variable:
		t.analyzeStmt(stmt.Value, inClosure)

		if t.locals.Contains(stmt.Name) {
			return
		}
		if t.referred.Contains(stmt.Name) || t.captured.Contains(stmt.Name) {
			t.boxed.Add(stmt.Name)
		}

	case ast.Assign:
		t.analyzeStmt(stmt.Object, inClosure)
		t.analyzeStmt(stmt.Value, inClosure)

		if stmt.Object != nil || t.locals.Contains(stmt.Name) {
			return
		}
		if inClosure || t.captured.Contains(stmt.Name) {
			t.boxed.Add(stmt.Name)
		}

	case ast.Function:
		locals := newVarSet()
		locals.AddSet(t.locals)
		locals.AddSlice(mapSlice(stmt.Args, argName))

		inner := boxScopeTracking{
			locals:   locals,
			referred: t.referred,
			captured: t.captured,
			boxed:    t.boxed,
		}
		inner.analyzeBlock(stmt.Body, true)

	case ast.If:
		t.analyzeStmt(stmt.Cond, inClosure)
		t.analyzeBlock(stmt.Then, inClosure)
		t.analyzeBlock(stmt.Else, inClosure)

	case ast.Call:
		t.analyzeStmt(stmt.Method, inClosure)
		for _, a := range stmt.Args {
			t.analyzeStmt(a, inClosure)
		}

	case ast.MemberAccess:
		t.analyzeStmt(stmt.Object, inClosure)
	}
}

func (t boxScopeTracking) analyzeBlock(stmts []ast.Stmt, inClosure bool) {
	locals := newVarSet()
	locals.AddSet(t.locals)
	locals.AddSet(blockVars(stmts))
	inner := boxScopeTracking{
		locals:   locals,
		referred: t.referred,
		captured: t.captured,
		boxed:    t.boxed,
	}
	for _, stmt := range stmts {
		inner.analyzeStmt(stmt, inClosure)
	}
}
