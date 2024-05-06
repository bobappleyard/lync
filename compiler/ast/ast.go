package ast

type Node interface {
	node()
	Start() int
}

// Toplevel

type Program struct {
	Stmts []Stmt
}

// Expressions

type Expr interface {
	expr()
	Stmt
	Node
}

type StringConstant struct {
	astNodeData

	Value string
}

type IntConstant struct {
	astNodeData

	Value int
}

type FltConstant struct {
	astNodeData

	Value float64
}

type VariableRef struct {
	astNodeData

	Var string
}

type MemberAccess struct {
	astNodeData

	Object Expr
	Member string
}

type Call struct {
	astNodeData

	Method Expr
	Args   []Expr
}

type Class struct {
	astNodeData

	Name    string
	Members []Member
}

type Method struct {
	astNodeData

	Name string
	Args []Arg
	Body []Stmt
}

type Arg struct {
	astNodeData

	Name string
}

func (StringConstant) expr() {}
func (IntConstant) expr()    {}
func (FltConstant) expr()    {}
func (VariableRef) expr()    {}
func (MemberAccess) expr()   {}
func (Call) expr()           {}
func (Class) expr()          {}
func (Method) expr()         {}

// Class Members

type Member interface {
	member()

	Node
}

// Statements

type Stmt interface {
	stmt()
	Node
}

type Return struct {
	astNodeData

	Value Expr
}

type Variable struct {
	astNodeData

	Name  string
	Value Expr
}

type Import struct {
	astNodeData

	Name string
	Path string
}

type NoOp struct {
	astNodeData
}

type If struct {
	astNodeData

	Cond Expr
	Then []Stmt
	Else []Stmt
}

func (Return) stmt()   {}
func (Variable) stmt() {}
func (Import) stmt()   {}
func (NoOp) stmt()     {}
func (If) stmt()       {}

func (StringConstant) stmt() {}
func (IntConstant) stmt()    {}
func (FltConstant) stmt()    {}
func (VariableRef) stmt()    {}
func (MemberAccess) stmt()   {}
func (Call) stmt()           {}
func (Class) stmt()          {}
func (Method) stmt()         {}
