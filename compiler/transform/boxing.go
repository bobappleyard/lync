package transform

import (
	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/data"
)

func transformBoxing(p ast.Program) ast.Program {
	boxes := withFallbackTransformer(&boxing{boxed: newVarSet()})
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
	var res []ast.Stmt
	declared := blockVars(stmts)

	for _, v := range needBoxes.Items() {
		if b.args.Contains(v) {
			res = append(res, ast.Assign{
				Name:  v,
				Value: unitMethodCall("create_box", ast.VariableRef{Var: v}),
			})
		} else if declared.Contains(v) {
			res = append(res, ast.Variable{
				Name:  v,
				Value: unitMethodCall("create_undefined_box", ast.Name{Name: v}),
			})
		}
	}
	needBoxes.AddSet(b.boxed)
	inner := withFallbackTransformer(&boxing{boxed: needBoxes})
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
		args.AddSlice(data.MapSlice(expr.Args, argName))
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
	tracking := withFallbackAnalzyer(&boxScopeAnalyzer{
		referred: newVarSet(),
		captured: newVarSet(),
		boxed:    newVarSet(),
	})

	for _, stmt := range stmts {
		tracking.analyzeStmt(stmt)
	}

	return tracking.boxed
}

func (b *boxing) call(varName, methodName string, args ...ast.Expr) ast.Expr {
	return ast.Call{
		Method: ast.MemberAccess{
			Object: ast.VariableRef{Var: varName},
			Member: methodName,
		},
		Args: args,
	}
}

func (t *boxScopeAnalyzer) analyzeStmt(stmt ast.Stmt) {
	switch stmt := stmt.(type) {

	case ast.Variable:
		t.analyzeStmt(stmt.Value)

		if t.locals.Contains(stmt.Name) {
			return
		}
		if t.referred.Contains(stmt.Name) || t.captured.Contains(stmt.Name) {
			t.boxed.Put(stmt.Name)
		}

	case ast.Assign:
		t.analyzeStmt(stmt.Object)
		t.analyzeStmt(stmt.Value)

		if stmt.Object != nil || t.locals.Contains(stmt.Name) {
			return
		}
		if t.inClosure || t.captured.Contains(stmt.Name) {
			t.boxed.Put(stmt.Name)
		}

	default:
		t.fallbackAnalyzer.analyzeStmt(stmt)
	}
}

func (t *boxScopeAnalyzer) analyzeExpr(expr ast.Expr) {
	switch expr := expr.(type) {

	case ast.VariableRef:
		if t.locals.Contains(expr.Var) {
			return
		}
		t.referred.Put(expr.Var)
		if t.inClosure {
			t.captured.Put(expr.Var)
		}

	case ast.Function:
		locals := newVarSet()
		locals.AddSet(t.locals)
		locals.AddSlice(data.MapSlice(expr.Args, argName))

		inner := withFallbackAnalzyer(&boxScopeAnalyzer{
			inClosure: true,
			locals:    locals,
			referred:  t.referred,
			captured:  t.captured,
			boxed:     t.boxed,
		})
		inner.analyzeBlock(expr.Body)

	default:
		t.fallbackAnalyzer.analyzeExpr(expr)
	}
}

func (t *boxScopeAnalyzer) analyzeBlock(stmts []ast.Stmt) {
	locals := newVarSet()
	locals.AddSet(t.locals)
	locals.AddSet(blockVars(stmts))
	inner := withFallbackAnalzyer(&boxScopeAnalyzer{
		inClosure: t.inClosure,
		locals:    locals,
		referred:  t.referred,
		captured:  t.captured,
		boxed:     t.boxed,
	})
	for _, stmt := range stmts {
		inner.analyzeStmt(stmt)
	}
}
