package transform

import (
	ast2 "go/ast"
	"testing"

	"github.com/bobappleyard/lync/compiler/ast"
	"github.com/bobappleyard/lync/util/assert"
)

func TestBoxing(t *testing.T) {
	for _, test := range []struct {
		name    string
		in, out ast.Program
	}{
		{
			name: "UsePrecedesDeclaration",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.VariableRef{Var: "x"},
				ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Call{
					Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
					Args:   []ast.Expr{ast.Name{Name: "x"}},
				}},
				ast.Call{Method: ast.MemberAccess{
					Object: ast.VariableRef{Var: "x"},
					Member: "get",
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "define",
					},
					Args: []ast.Expr{ast.IntConstant{Value: 1}},
				},
			}},
		},
		{
			name: "DeclarationIncludesUse",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.VariableRef{Var: "x"}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Call{
					Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
					Args:   []ast.Expr{ast.Name{Name: "x"}},
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "define",
					},
					Args: []ast.Expr{ast.Call{Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "get",
					}}},
				},
			}},
		},
		{
			name: "DeclarationIncludesDeferredUse",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Function{Body: []ast.Stmt{ast.VariableRef{Var: "x"}}}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Call{
					Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
					Args:   []ast.Expr{ast.Name{Name: "x"}},
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "define",
					},
					Args: []ast.Expr{ast.Function{Body: []ast.Stmt{ast.Call{Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "get",
					}}}}},
				},
			}},
		},
		{
			name: "DeclarationIncludesShadowedUse",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{ast.VariableRef{Var: "x"}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{ast.VariableRef{Var: "x"}},
				}},
			}},
		},
		{
			name: "AssignedToplevelReferredInside",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
				ast.Function{Body: []ast.Stmt{ast.VariableRef{Var: "x"}}},
				ast.Assign{Name: "x", Value: ast.IntConstant{Value: 2}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Call{
					Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
					Args:   []ast.Expr{ast.Name{Name: "x"}},
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "define",
					},
					Args: []ast.Expr{ast.IntConstant{Value: 1}},
				},
				ast.Function{Body: []ast.Stmt{
					ast.Call{Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "get",
					}},
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "set",
					},
					Args: []ast.Expr{ast.IntConstant{Value: 2}},
				},
			}},
		},
		{
			name: "AssignedInside",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
				ast.Function{Body: []ast.Stmt{
					ast.Assign{Name: "x", Value: ast.IntConstant{Value: 2}},
				}},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.Call{
					Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
					Args:   []ast.Expr{ast.Name{Name: "x"}},
				}},
				ast.Call{
					Method: ast.MemberAccess{
						Object: ast.VariableRef{Var: "x"},
						Member: "define",
					},
					Args: []ast.Expr{ast.IntConstant{Value: 1}},
				},
				ast.Function{Body: []ast.Stmt{
					ast.Call{
						Method: ast.MemberAccess{
							Object: ast.VariableRef{Var: "x"},
							Member: "set",
						},
						Args: []ast.Expr{ast.IntConstant{Value: 2}},
					},
				}},
			}},
		},
		{
			name: "ShadowedAssignment",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{ast.Assign{
						Name:  "x",
						Value: ast.IntConstant{Value: 3},
					}},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{ast.Assign{
						Name:  "x",
						Value: ast.IntConstant{Value: 3},
					}},
				},
			}},
		},
		{
			name: "NestedVariable",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{},
					Body: []ast.Stmt{
						ast.VariableRef{Var: "x"},
						ast.Variable{Name: "x", Value: ast.IntConstant{Value: 1}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{},
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.Call{
							Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
							Args:   []ast.Expr{ast.Name{Name: "x"}},
						}},
						ast.Call{Method: ast.MemberAccess{
							Object: ast.VariableRef{Var: "x"},
							Member: "get",
						}},
						ast.Call{
							Method: ast.MemberAccess{
								Object: ast.VariableRef{Var: "x"},
								Member: "define",
							},
							Args: []ast.Expr{ast.IntConstant{Value: 1}},
						},
					},
				},
			}},
		},
		{
			name: "BoxedArgument",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Function{Body: []ast.Stmt{ast.VariableRef{Var: "x"}}},
						ast.Assign{Name: "x", Value: ast.IntConstant{Value: 1}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.Variable{Name: "x", Value: ast.Call{
							Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_box"},
							Args:   []ast.Expr{ast.VariableRef{Var: "x"}},
						}},
						ast.Function{Body: []ast.Stmt{
							ast.Call{Method: ast.MemberAccess{
								Object: ast.VariableRef{Var: "x"},
								Member: "get",
							}},
						}},
						ast.Call{
							Method: ast.MemberAccess{
								Object: ast.VariableRef{Var: "x"},
								Member: "set",
							},
							Args: []ast.Expr{ast.IntConstant{Value: 1}},
						},
					},
				},
			}},
		},
		{
			name: "BoxedShadowedArgument",
			in: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.If{Then: []ast.Stmt{
							ast.Variable{Name: "x", Value: ast.IntConstant{Value: 2}},
							ast.Function{Body: []ast.Stmt{ast.VariableRef{Var: "x"}}},
							ast.Assign{Name: "x", Value: ast.IntConstant{Value: 1}},
						}},
					},
				},
			}},
			out: ast.Program{Stmts: []ast.Stmt{
				ast.Function{
					Args: []ast.Arg{{Name: "x"}},
					Body: []ast.Stmt{
						ast.If{Then: []ast.Stmt{
							ast.Variable{Name: "x", Value: ast.Call{
								Method: ast.MemberAccess{Object: ast.Unit{}, Member: "create_undefined_box"},
								Args:   []ast.Expr{ast.Name{Name: "x"}},
							}},
							ast.Call{
								Method: ast.MemberAccess{
									Object: ast.VariableRef{Var: "x"},
									Member: "define",
								},
								Args: []ast.Expr{ast.IntConstant{Value: 2}},
							},
							ast.Function{Body: []ast.Stmt{
								ast.Call{Method: ast.MemberAccess{
									Object: ast.VariableRef{Var: "x"},
									Member: "get",
								}},
							}},
							ast.Call{
								Method: ast.MemberAccess{
									Object: ast.VariableRef{Var: "x"},
									Member: "set",
								},
								Args: []ast.Expr{ast.IntConstant{Value: 1}},
							},
						}},
					},
				},
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			out := introduceBoxing(test.in)
			assert.Equal(t, out, test.out)
			if t.Failed() {
				ast2.Print(nil, out)
			}
		})
	}

}
