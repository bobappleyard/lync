package parser

import (
	"slices"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
	"github.com/r3labs/diff"
)

func TestSyntax(t *testing.T) {
	for _, test := range []struct {
		name string
		in   string
		out  ast.Program
	}{
		{
			name: "Import",
			in:   `import "path"`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{
						Path: "path",
					},
				},
			},
		},
		{
			name: "ImportLeadingNewline",
			in: `
			import "path"`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{
						Path: "path",
					},
				},
			},
		},
		{
			name: "ImportTwice",
			in: `
			import "path"
			import "path"`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{Path: "path"},
					ast.Import{Path: "path"},
				},
			},
		},
		{
			name: "ImportTwiceNewline",
			in: `
			import "path"
			import "path"
			`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{Path: "path"},
					ast.Import{Path: "path"},
				},
			},
		},
		{
			name: "VarDecl",
			in:   `var a = 1`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Variable{
						Name:  "a",
						Value: ast.IntConstant{Value: 1},
					},
				},
			},
		},
		{
			name: "VarRef",
			in:   `a`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.VariableRef{
						Var: "a",
					},
				},
			},
		},
		{
			name: "StringConstant",
			in:   `"a"`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.StringConstant{
						Value: "a",
					},
				},
			},
		},
		{
			name: "IntConstant",
			in:   `1000`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.IntConstant{
						Value: 1000,
					},
				},
			},
		},
		{
			name: "FloatConstant",
			in:   `1.234`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.FltConstant{
						Value: 1.234,
					},
				},
			},
		},
		{
			name: "FunctionCall",
			in:   `f(1, 2)`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Call{
						Method: ast.VariableRef{Var: "f"},
						Args: []ast.Expr{
							ast.IntConstant{Value: 1},
							ast.IntConstant{Value: 2},
						},
					},
				},
			},
		},
		{
			name: "MethodCall",
			in:   `object.method(1, 2)`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Call{
						Method: ast.MemberAccess{
							Object: ast.VariableRef{Var: "object"},
							Member: "method",
						},
						Args: []ast.Expr{
							ast.IntConstant{Value: 1},
							ast.IntConstant{Value: 2},
						},
					},
				},
			},
		},
		{
			name: "HOFCall",
			in:   `object.method(func(x) { return x })`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Call{
						Method: ast.MemberAccess{
							Object: ast.VariableRef{Var: "object"},
							Member: "method",
						},
						Args: []ast.Expr{
							ast.Function{
								Args: []ast.Arg{{Name: "x"}},
								Body: []ast.Stmt{
									ast.Return{Value: ast.VariableRef{Var: "x"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "BasicFunction",
			in:   `func f(a, b, c) { return a }`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "a"}, {Name: "b"}, {Name: "c"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.VariableRef{Var: "a"}},
						},
					},
				},
			},
		},
		{
			name: "BasicFunctionNoArgs",
			in:   `func f() { return a }`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Body: []ast.Stmt{
							ast.Return{Value: ast.VariableRef{Var: "a"}},
						},
					},
				},
			},
		},
		{
			name: "MultiStmtFunction",
			in: `func f(x) {
				var y = 1
				return a
			}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.Variable{
								Name:  "y",
								Value: ast.IntConstant{Value: 1},
							},
							ast.Return{Value: ast.VariableRef{Var: "a"}},
						},
					},
				},
			},
		},
		{
			name: "FunctionMultilineArgs",
			in: `func f(
				x,
				y
			) {
				return x
			}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "x"}, {Name: "y"}},
						Body: []ast.Stmt{
							ast.Return{Value: ast.VariableRef{Var: "x"}},
						},
					},
				},
			},
		},
		{
			name: "FunctionNestedMultiline",
			in: `func f(x) {
				func g() {
					var y = 1
					return y
				}
				return x
			}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Function{
						Name: "f",
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.Function{
								Name: "g",
								Body: []ast.Stmt{
									ast.Variable{
										Name:  "y",
										Value: ast.IntConstant{Value: 1},
									},
									ast.Return{Value: ast.VariableRef{Var: "y"}},
								},
							},
							ast.Return{Value: ast.VariableRef{Var: "x"}},
						},
					},
				},
			},
		},
		{
			name: "If",
			in: `if x {
				return 2
			}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.If{
						Cond: ast.VariableRef{Var: "x"},
						Then: []ast.Stmt{
							ast.Return{
								Value: ast.IntConstant{Value: 2},
							},
						},
					},
				},
			},
		},
		{
			name: "EmptyClass",
			in:   `class A {}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Class{
						Name: "A",
					},
				},
			},
		},
		{
			name: "ClassWithMethod",
			in: `class A {
				name() {
					return "A"
				}
			}`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Class{
						Name: "A",
						Members: []ast.Member{
							ast.Method{
								Name: "name",
								Body: []ast.Stmt{
									ast.Return{Value: ast.StringConstant{Value: "A"}},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "SemiRealProgram",
			in: `
				import "array"

				func loop(f) {
					func step(a, i) {
						if i.gt(a.size) {
							return
						}
						f(a.get(i))
						return step(a, i.plus(1))
					}
					return func(a) {
						return step(a, 0)
					}
				}
			`,
			out: ast.Program{
				Stmts: []ast.Stmt{
					ast.Import{Path: "array"},
					ast.Function{
						Name: "loop",
						Args: []ast.Arg{{Name: "f"}},
						Body: []ast.Stmt{
							ast.Function{
								Name: "step",
								Args: []ast.Arg{{Name: "a"}, {Name: "i"}},
								Body: []ast.Stmt{
									ast.If{
										Cond: ast.Call{
											Method: ast.MemberAccess{
												Object: ast.VariableRef{Var: "i"},
												Member: "gt",
											},
											Args: []ast.Expr{
												ast.MemberAccess{
													Object: ast.VariableRef{Var: "a"},
													Member: "size",
												},
											},
										},
										Then: []ast.Stmt{
											ast.Return{
												Value: ast.VariableRef{Var: "void"},
											},
										},
									},
									ast.Call{
										Method: ast.VariableRef{Var: "f"},
										Args: []ast.Expr{
											ast.Call{
												Method: ast.MemberAccess{
													Object: ast.VariableRef{Var: "a"},
													Member: "get",
												},
												Args: []ast.Expr{
													ast.VariableRef{Var: "i"},
												},
											},
										},
									},
									ast.Return{
										Value: ast.Call{
											Method: ast.VariableRef{Var: "step"},
											Args: []ast.Expr{
												ast.VariableRef{Var: "a"},
												ast.Call{
													Method: ast.MemberAccess{
														Object: ast.VariableRef{Var: "i"},
														Member: "plus",
													},
													Args: []ast.Expr{
														ast.IntConstant{Value: 1},
													},
												},
											},
										},
									},
								},
							},
							ast.Return{
								Value: ast.Function{
									Args: []ast.Arg{{Name: "a"}},
									Body: []ast.Stmt{
										ast.Return{
											Value: ast.Call{
												Method: ast.VariableRef{Var: "step"},
												Args: []ast.Expr{
													ast.VariableRef{Var: "a"},
													ast.IntConstant{Value: 0},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			prog, err := Parse([]byte(test.in))

			assert.Nil(t, err)

			cl, _ := diff.Diff(test.out, prog)
			for _, c := range cl {
				if c.Type == "update" && len(c.Path) > 2 &&
					slices.Equal([]string{"astNodeData", "s"}, c.Path[len(c.Path)-2:]) {
					continue
				}
				if c.Type == "update" {
					t.Errorf("at %v: %v -> %v", c.Path, c.From, c.To)
				} else {
					t.Error(c)
				}
			}
			t.Log(prog)
		})
	}

}
