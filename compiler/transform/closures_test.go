package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestClosures(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{
		{
			name: "Toplevel",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Return{Value: ast.Call{
						Method: ast.VariableRef{Var: "f"},
						Args:   []ast.Expr{ast.VariableRef{Var: "x"}},
					}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Return{Value: ast.Call{
						Method: ast.VariableRef{Var: "f"},
						Args:   []ast.Expr{ast.VariableRef{Var: "x"}},
					}},
				}},
			}},
		},
		{
			name: "InnerNoCapture",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Function{Body: []ast.Stmt{
						ast.VariableRef{Var: "x"},
					}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Function{Body: []ast.Stmt{
						ast.VariableRef{Var: "x"},
					}},
				}},
			}},
		},
		{
			name: "InnerCaptureVar",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Function{Body: []ast.Stmt{
							ast.VariableRef{Var: "x"},
						}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "create_closure",
							},
							Args: []ast.Expr{
								ast.Function{
									Args: []ast.Arg{{Name: "x"}},
									Body: []ast.Stmt{
										ast.VariableRef{Var: "x"},
									},
								},
								ast.VariableRef{Var: "x"},
							},
						},
					},
				},
			}},
		},
		{
			name: "InnerShadowVar",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Function{
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.VariableRef{Var: "x"},
						},
					},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Function{
						Args: []ast.Arg{{Name: "x"}},
						Body: []ast.Stmt{
							ast.VariableRef{Var: "x"},
						},
					},
				}},
			}},
		},
		{
			name: "CaptureAndShadow",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}, {Name: "y"}},
					Body: []ast.Stmt{
						ast.Function{
							Args: []ast.Arg{{Name: "x"}},
							Body: []ast.Stmt{
								ast.VariableRef{Var: "x"},
								ast.VariableRef{Var: "y"},
							},
						},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}, {Name: "y"}},
					Body: []ast.Stmt{
						ast.Call{
							Method: ast.MemberAccess{
								Object: ast.Unit{},
								Member: "create_closure",
							},
							Args: []ast.Expr{
								ast.Function{
									Args: []ast.Arg{{Name: "y"}, {Name: "x"}},
									Body: []ast.Stmt{
										ast.VariableRef{Var: "x"},
										ast.VariableRef{Var: "y"},
									},
								},
								ast.VariableRef{Var: "y"},
							},
						},
					},
				},
			}},
		},
		{
			name: "VarDecl",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Function{Body: []ast.Stmt{
						ast.VariableRef{Var: "x"},
					}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Call{
						Method: ast.MemberAccess{
							Object: ast.Unit{},
							Member: "create_closure",
						},
						Args: []ast.Expr{
							ast.Function{
								Args: []ast.Arg{{Name: "x"}},
								Body: []ast.Stmt{
									ast.VariableRef{Var: "x"},
								},
							},
							ast.VariableRef{Var: "x"},
						},
					},
				},
				}},
			}},
		{
			name: "NestedVarDecl",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Variable{Name: "y", Value: ast.IntConstant{Value: 2}},
					ast.Function{Body: []ast.Stmt{
						ast.Variable{Name: "y", Value: ast.IntConstant{Value: 2}},
						ast.VariableRef{Var: "x"},
						ast.VariableRef{Var: "y"},
					}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Variable{Name: "y", Value: ast.IntConstant{Value: 2}},
					ast.Call{
						Method: ast.MemberAccess{
							Object: ast.Unit{},
							Member: "create_closure",
						},
						Args: []ast.Expr{
							ast.Function{
								Args: []ast.Arg{{Name: "x"}},
								Body: []ast.Stmt{
									ast.Variable{Name: "y", Value: ast.IntConstant{Value: 2}},
									ast.VariableRef{Var: "x"},
									ast.VariableRef{Var: "y"},
								},
							},
							ast.VariableRef{Var: "x"},
						},
					},
				}},
			}},
		},
		{
			name: "NonFuncBlock",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
						ast.If{
							Then: []ast.Stmt{
								ast.VariableRef{Var: "x"},
							},
						},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
						ast.If{
							Then: []ast.Stmt{
								ast.VariableRef{Var: "x"},
							},
						},
					},
				},
			}},
		},
		{
			name: "ClosureArg",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Call{
						Method: ast.VariableRef{Var: "f"},
						Args: []ast.Expr{
							ast.Function{Body: []ast.Stmt{
								ast.VariableRef{Var: "x"},
							}},
						},
					},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{Body: []ast.Stmt{
					ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
					ast.Call{
						Method: ast.VariableRef{Var: "f"},
						Args: []ast.Expr{
							ast.Call{
								Method: ast.MemberAccess{
									Object: ast.Unit{},
									Member: "create_closure",
								},
								Args: []ast.Expr{
									ast.Function{
										Args: []ast.Arg{{Name: "x"}},
										Body: []ast.Stmt{
											ast.VariableRef{Var: "x"},
										},
									},
									ast.VariableRef{Var: "x"},
								},
							},
						},
					},
				}},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := transformClosures(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}
}
