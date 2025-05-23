package app

import (
	"fmt"
	"go/ast"
	"go/token"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
	"go-init-gen/internal/generator/engine/generators/features"
)

// Generator implements the app.go code generation
type Generator struct{}

// NewGenerator creates a new app generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the app.go code based on the input data
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Use the feature detector to identify enabled features
	fs := features.DetectFeatures(data)

	// Add imports if needed
	var importsToAdd []string

	// Basic imports - only add what's actually needed
	importsToAdd = append(importsToAdd,
		"context",
		"time",
		data.Name+"/config",
		data.Name+"/internal/service",
		"gitlab.com/go-init/go-init-common/default/closer",
		"gitlab.com/go-init/go-init-common/default/logger",
	)

	// Add feature-specific imports
	if fs.HasGRPC || fs.HasGraphQL {
		// These are needed for server functionality
		importsToAdd = append(importsToAdd,
			"fmt",
			"sync",
		)
	}

	if fs.HasGRPC {
		importsToAdd = append(importsToAdd,
			"net",
			data.Name+"/internal/grpc",
			data.Name+"/pkg/api/grpc",
			"google.golang.org/grpc/reflection",
		)

		// Add named import for grpcserver
		generators.AddNamedImport(file, "grpcserver", "google.golang.org/grpc")
	}

	if fs.HasGraphQL {
		importsToAdd = append(importsToAdd,
			"net/http",
			data.Name+"/internal/graphql",
			data.Name+"/pkg/api/graphql",
			"gitlab.com/go-init/go-init-common/default/http/server",
		)

		// Add named import for myhttp
		generators.AddNamedImport(file, "myhttp", "gitlab.com/go-init/go-init-common/default/http")
	}

	if fs.HasDatabase {
		importsToAdd = append(importsToAdd,
			data.Name+"/internal/database",
			data.Name+"/internal/database/models",
			"gitlab.com/go-init/go-init-common/default/db/pg/orm",
		)
	}

	// Remove any existing imports and add only the ones we need
	g.removeAllImports(file)

	// Add imports
	if len(importsToAdd) > 0 {
		generators.AddImports(file, importsToAdd)
	}

	// Modify App struct
	g.modifyAppStruct(file, fs.HasGRPC, fs.HasGraphQL, fs.HasDatabase)

	// Modify init dependencies
	g.modifyInitDeps(file, fs.HasGRPC, fs.HasGraphQL, fs.HasDatabase)

	// Add service initialization method
	g.addInitServicesMethod(file, fs.HasDatabase)

	// Modify Run method
	g.modifyRunMethod(file, fs.HasGRPC, fs.HasGraphQL, fs.HasDatabase)

	// Add initialization methods for features
	if fs.HasGRPC {
		g.addInitGRPCMethod(file)
	}

	if fs.HasGraphQL {
		g.addInitGraphQLMethod(file)
	}

	if fs.HasDatabase {
		g.addInitDatabaseMethod(file, fs.DatabaseType)
	}

	return nil
}

// removeAllImports removes all import declarations from the file
func (g *Generator) removeAllImports(file *ast.File) {
	var nonImportDecls []ast.Decl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); !ok || genDecl.Tok != token.IMPORT {
			nonImportDecls = append(nonImportDecls, decl)
		}
	}
	file.Decls = nonImportDecls
}

// modifyAppStruct updates the App struct based on enabled features
func (g *Generator) modifyAppStruct(file *ast.File, hasGRPC, hasGraphQL, hasDatabase bool) {
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == "App" {
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						// Create a new list of fields
						var newFields []*ast.Field

						// Add common fields
						newFields = append(newFields, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent("cfg")},
							Type:  &ast.StarExpr{X: ast.NewIdent("config.AppConfig")},
						})

						newFields = append(newFields, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent("log")},
							Type:  &ast.StarExpr{X: ast.NewIdent("logger.Logger")},
						})

						// Add database fields if enabled
						if hasDatabase {
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("db")},
								Type:  &ast.StarExpr{X: ast.NewIdent("orm.AgentImpl")},
							})
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("repo")},
								Type:  ast.NewIdent("database.DefaultTemplateRepository"),
							})
						}

						// Add service field
						newFields = append(newFields, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent("service")},
							Type:  &ast.StarExpr{X: ast.NewIdent("service.Service")},
						})

						// Add gRPC fields
						if hasGRPC {
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("grpcService")},
								Type:  &ast.StarExpr{X: ast.NewIdent("grpc.GRPCService")},
							})
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("lis")},
								Type:  ast.NewIdent("net.Listener"),
							})
						}

						// Add GraphQL fields
						if hasGraphQL {
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("graphqlService")},
								Type:  &ast.StarExpr{X: ast.NewIdent("graphql.GQLService")},
							})
							newFields = append(newFields, &ast.Field{
								Names: []*ast.Ident{ast.NewIdent("srv")},
								Type:  &ast.StarExpr{X: ast.NewIdent("http.Server")},
							})
						}

						// Update the struct fields
						structType.Fields.List = newFields
					}
				}
			}
		}
	}
}

// modifyInitDeps updates the initDeps method based on enabled features
func (g *Generator) modifyInitDeps(file *ast.File, hasGRPC, hasGraphQL, hasDatabase bool) {
	// Create the function list based on enabled features
	initFuncs := []string{
		"a.initConfig",
		"a.initLogger",
		"a.initCloser",
	}

	if hasDatabase {
		initFuncs = append(initFuncs, "a.initDB", "a.initRepo")
	}

	initFuncs = append(initFuncs, "a.initServices")

	if hasGraphQL {
		initFuncs = append(initFuncs, "a.initHttpServer")
	}

	if hasGRPC {
		initFuncs = append(initFuncs, "a.initGrpcServer")
	}

	// Find the initDeps method and replace its init functions
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != "initDeps" {
			continue
		}

		// Find the assignment statement for the inits slice
		for _, stmt := range funcDecl.Body.List {
			assignStmt, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assignStmt.Lhs) == 0 || len(assignStmt.Rhs) == 0 {
				continue
			}

			// Check if this is the inits assignment
			lhsIdent, ok := assignStmt.Lhs[0].(*ast.Ident)
			if !ok || lhsIdent.Name != "inits" {
				continue
			}

			// Create the new composite literal for the init functions
			compLit := &ast.CompositeLit{
				Type: &ast.ArrayType{
					Elt: &ast.FuncType{
						Params: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: ast.NewIdent("context.Context"),
								},
							},
						},
						Results: &ast.FieldList{
							List: []*ast.Field{
								{
									Type: ast.NewIdent("error"),
								},
							},
						},
					},
				},
				Elts: []ast.Expr{},
			}

			// Add all init functions
			for _, funcName := range initFuncs {
				compLit.Elts = append(compLit.Elts, ast.NewIdent(funcName))
			}

			// Replace the right-hand side with our new composite literal
			assignStmt.Rhs[0] = compLit
			break
		}
	}
}

// modifyRunMethod modifies the Run method to start the appropriate servers
func (g *Generator) modifyRunMethod(file *ast.File, hasGRPC, hasGraphQL, hasDatabase bool) {
	// Find the method Run
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "Run" {
			// Create the function body
			bodyStmts := []ast.Stmt{
				// Add the defer statement for closer
				&ast.DeferStmt{
					Call: &ast.CallExpr{
						Fun: &ast.FuncLit{
							Type: &ast.FuncType{
								Params: &ast.FieldList{},
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.ExprStmt{
										X: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("closer"),
												Sel: ast.NewIdent("CloseAll"),
											},
										},
									},
									&ast.ExprStmt{
										X: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("closer"),
												Sel: ast.NewIdent("Wait"),
											},
										},
									},
								},
							},
						},
					},
				},
			}

			// Add wait group if we have any servers
			if hasGRPC || hasGraphQL {
				// Create wait group
				wgStmt := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("wg")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CompositeLit{Type: &ast.SelectorExpr{X: ast.NewIdent("sync"), Sel: ast.NewIdent("WaitGroup")}}},
				}
				bodyStmts = append(bodyStmts, wgStmt)

				// Add to wait group
				wgAddStmt := &ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("wg"),
							Sel: ast.NewIdent("Add"),
						},
						Args: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: fmt.Sprintf("%d", boolToInt(hasGRPC)+boolToInt(hasGraphQL))}},
					},
				}
				bodyStmts = append(bodyStmts, wgAddStmt)
			}

			// Add HTTP server goroutine
			if hasGraphQL {
				httpServerStmt := &ast.GoStmt{
					Call: &ast.CallExpr{
						Fun: &ast.FuncLit{
							Type: &ast.FuncType{
								Params: &ast.FieldList{},
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.DeferStmt{
										Call: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("wg"),
												Sel: ast.NewIdent("Done"),
											},
										},
									},
									&ast.AssignStmt{
										Lhs: []ast.Expr{ast.NewIdent("err")},
										Tok: token.DEFINE,
										Rhs: []ast.Expr{
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X:   ast.NewIdent("a"),
													Sel: ast.NewIdent("runHttpServer"),
												},
											},
										},
									},
									&ast.IfStmt{
										Cond: &ast.BinaryExpr{
											X:  ast.NewIdent("err"),
											Op: token.NEQ,
											Y:  ast.NewIdent("nil"),
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												&ast.ExprStmt{
													X: &ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X: &ast.SelectorExpr{
																X:   ast.NewIdent("a"),
																Sel: ast.NewIdent("log"),
															},
															Sel: ast.NewIdent("Error"),
														},
														Args: []ast.Expr{
															&ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   ast.NewIdent("fmt"),
																	Sel: ast.NewIdent("Sprintf"),
																},
																Args: []ast.Expr{
																	&ast.BasicLit{
																		Kind:  token.STRING,
																		Value: `"Failed to start http server: %v"`,
																	},
																	ast.NewIdent("err"),
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
					},
				}
				bodyStmts = append(bodyStmts, httpServerStmt)
			}

			// Add gRPC server goroutine
			if hasGRPC {
				grpcServerStmt := &ast.GoStmt{
					Call: &ast.CallExpr{
						Fun: &ast.FuncLit{
							Type: &ast.FuncType{
								Params: &ast.FieldList{},
							},
							Body: &ast.BlockStmt{
								List: []ast.Stmt{
									&ast.DeferStmt{
										Call: &ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("wg"),
												Sel: ast.NewIdent("Done"),
											},
										},
									},
									&ast.AssignStmt{
										Lhs: []ast.Expr{ast.NewIdent("err")},
										Tok: token.DEFINE,
										Rhs: []ast.Expr{
											&ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X:   ast.NewIdent("a"),
													Sel: ast.NewIdent("runGrpcServer"),
												},
											},
										},
									},
									&ast.IfStmt{
										Cond: &ast.BinaryExpr{
											X:  ast.NewIdent("err"),
											Op: token.NEQ,
											Y:  ast.NewIdent("nil"),
										},
										Body: &ast.BlockStmt{
											List: []ast.Stmt{
												&ast.ExprStmt{
													X: &ast.CallExpr{
														Fun: &ast.SelectorExpr{
															X: &ast.SelectorExpr{
																X:   ast.NewIdent("a"),
																Sel: ast.NewIdent("log"),
															},
															Sel: ast.NewIdent("Error"),
														},
														Args: []ast.Expr{
															&ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   ast.NewIdent("fmt"),
																	Sel: ast.NewIdent("Sprintf"),
																},
																Args: []ast.Expr{
																	&ast.BasicLit{
																		Kind:  token.STRING,
																		Value: `"Failed to start grpc server: %v"`,
																	},
																	ast.NewIdent("err"),
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
					},
				}
				bodyStmts = append(bodyStmts, grpcServerStmt)
			}

			// If we have any servers, wait for them to complete
			if hasGRPC || hasGraphQL {
				waitStmt := &ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("wg"),
							Sel: ast.NewIdent("Wait"),
						},
					},
				}
				bodyStmts = append(bodyStmts, waitStmt)
			}

			// Add return nil
			bodyStmts = append(bodyStmts, &ast.ReturnStmt{
				Results: []ast.Expr{ast.NewIdent("nil")},
			})

			// Replace the function body
			funcDecl.Body = &ast.BlockStmt{
				List: bodyStmts,
			}
		}
	}
}

// boolToInt converts a boolean to an integer (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// addInitGRPCMethod adds the gRPC initialization method
func (g *Generator) addInitGRPCMethod(file *ast.File) {
	// Check if method already exists
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "initGrpcServer" {
			// Method already exists, no need to add it
			return
		}
	}

	// Create the initGrpcServer method
	method := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("initGrpcServer"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("_")},
						Type:  ast.NewIdent("context.Context"),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// Create gRPC server
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("server")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("grpcserver"),
								Sel: ast.NewIdent("NewServer"),
							},
						},
					},
				},

				// Register services
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("pb"),
							Sel: ast.NewIdent("RegisterUserServiceServer"),
						},
						Args: []ast.Expr{
							ast.NewIdent("server"),
							&ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("grpcService"),
							},
						},
					},
				},
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("reflection"),
							Sel: ast.NewIdent("Register"),
						},
						Args: []ast.Expr{
							ast.NewIdent("server"),
						},
					},
				},

				// Save server
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   ast.NewIdent("a"),
							Sel: ast.NewIdent("grpcServer"),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent("server")},
				},

				// Create TCP listener
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("port")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("a"),
									Sel: ast.NewIdent("cfg"),
								},
								Sel: ast.NewIdent("GrpcServ"),
							},
							Sel: ast.NewIdent("Port"),
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("listener"), ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("net"),
								Sel: ast.NewIdent("Listen"),
							},
							Args: []ast.Expr{
								&ast.BasicLit{Kind: token.STRING, Value: `"tcp"`},
								&ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X:   ast.NewIdent("fmt"),
										Sel: ast.NewIdent("Sprintf"),
									},
									Args: []ast.Expr{
										&ast.BasicLit{Kind: token.STRING, Value: `":%s"`},
										ast.NewIdent("port"),
									},
								},
							},
						},
					},
				},

				// Check for errors
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ReturnStmt{
								Results: []ast.Expr{
									&ast.CallExpr{
										Fun: &ast.SelectorExpr{
											X:   ast.NewIdent("fmt"),
											Sel: ast.NewIdent("Errorf"),
										},
										Args: []ast.Expr{
											&ast.BasicLit{
												Kind:  token.STRING,
												Value: `"failed to create listener on port %s: %w"`,
											},
											ast.NewIdent("port"),
											ast.NewIdent("err"),
										},
									},
								},
							},
						},
					},
				},

				// Save listener
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   ast.NewIdent("a"),
							Sel: ast.NewIdent("lis"),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{ast.NewIdent("listener")},
				},

				// Return nil
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("nil")},
				},
			},
		},
	}

	// Add runGrpcServer method
	runMethod := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("runGrpcServer"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// Log server start
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							Sel: ast.NewIdent("Info"),
						},
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("fmt"),
									Sel: ast.NewIdent("Sprintf"),
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: `"Запуск gRPC сервера на %s"`,
									},
									&ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X: &ast.SelectorExpr{
												X:   ast.NewIdent("a"),
												Sel: ast.NewIdent("cfg"),
											},
											Sel: ast.NewIdent("GrpcServ"),
										},
										Sel: ast.NewIdent("Port"),
									},
								},
							},
						},
					},
				},

				// Start gRPC server
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("a"),
									Sel: ast.NewIdent("grpcServer"),
								},
								Sel: ast.NewIdent("Serve"),
							},
							Args: []ast.Expr{
								&ast.SelectorExpr{
									X:   ast.NewIdent("a"),
									Sel: ast.NewIdent("lis"),
								},
							},
						},
					},
				},

				// Check for errors
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("log"),
										},
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("fmt"),
												Sel: ast.NewIdent("Sprintf"),
											},
											Args: []ast.Expr{
												&ast.BasicLit{
													Kind:  token.STRING,
													Value: `"Ошибка gRPC сервера: %v"`,
												},
												ast.NewIdent("err"),
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{ast.NewIdent("err")},
							},
						},
					},
				},

				// Return nil
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("nil")},
				},
			},
		},
	}

	file.Decls = append(file.Decls, method)
	file.Decls = append(file.Decls, runMethod)
}

// addInitGraphQLMethod adds the GraphQL initialization method
func (g *Generator) addInitGraphQLMethod(file *ast.File) {
	// Check if method already exists
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "initHttpServer" {
			// Method already exists, no need to add it
			return
		}
	}

	// Create the initHttpServer method
	method := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("initHttpServer"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("ctx")},
						Type:  ast.NewIdent("context.Context"),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("schema")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("gen_graphql"),
								Sel: ast.NewIdent("NewExecutableSchema"),
							},
							Args: []ast.Expr{
								&ast.CompositeLit{
									Type: &ast.SelectorExpr{
										X:   ast.NewIdent("gen_graphql"),
										Sel: ast.NewIdent("Config"),
									},
									Elts: []ast.Expr{
										&ast.KeyValueExpr{
											Key: ast.NewIdent("Resolvers"),
											Value: &ast.UnaryExpr{
												Op: token.AND,
												X: &ast.CompositeLit{
													Type: &ast.SelectorExpr{
														X:   ast.NewIdent("gen_graphql"),
														Sel: ast.NewIdent("Resolver"),
													},
													Elts: []ast.Expr{
														&ast.KeyValueExpr{
															Key:   ast.NewIdent("Service"),
															Value: &ast.SelectorExpr{X: ast.NewIdent("a"), Sel: ast.NewIdent("graphqlService")},
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
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("gqlHandler")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("server"),
								Sel: ast.NewIdent("NewGraphQLServer"),
							},
							Args: []ast.Expr{ast.NewIdent("schema")},
						},
					},
				},

				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("metricsHandler")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("NotFoundHandler"),
							},
						},
					},
				},
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("middlewares")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("myhttp"),
								Sel: ast.NewIdent("CollectHandlers"),
							},
						},
					},
				},

				// Create server
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("s")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("server"),
								Sel: ast.NewIdent("NewServer"),
							},
							Args: []ast.Expr{
								&ast.UnaryExpr{
									Op: token.AND,
									X: &ast.SelectorExpr{
										X:   ast.NewIdent("a.cfg"),
										Sel: ast.NewIdent("HttpServ"),
									},
								},
								ast.NewIdent("gqlHandler"),
								ast.NewIdent("metricsHandler"),
								ast.NewIdent("middlewares"),
							},
						},
					},
				},

				// Add closer
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("closer"),
							Sel: ast.NewIdent("Add"),
						},
						Args: []ast.Expr{
							&ast.FuncLit{
								Type: &ast.FuncType{
									Params:  &ast.FieldList{},
									Results: &ast.FieldList{List: []*ast.Field{{Type: ast.NewIdent("error")}}},
								},
								Body: &ast.BlockStmt{
									List: []ast.Stmt{
										&ast.AssignStmt{
											Lhs: []ast.Expr{
												ast.NewIdent("cancelCtx"),
												ast.NewIdent("cancel"),
											},
											Tok: token.DEFINE,
											Rhs: []ast.Expr{
												&ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   ast.NewIdent("context"),
														Sel: ast.NewIdent("WithTimeout"),
													},
													Args: []ast.Expr{
														ast.NewIdent("ctx"),
														ast.NewIdent("shutDownTimeOut"),
													},
												},
											},
										},
										&ast.DeferStmt{
											Call: &ast.CallExpr{
												Fun: ast.NewIdent("cancel"),
											},
										},
										&ast.IfStmt{
											Cond: &ast.BinaryExpr{
												X: &ast.CallExpr{
													Fun: &ast.SelectorExpr{
														X:   ast.NewIdent("s"),
														Sel: ast.NewIdent("Shutdown"),
													},
													Args: []ast.Expr{ast.NewIdent("cancelCtx")},
												},
												Op: token.NEQ,
												Y:  ast.NewIdent("nil"),
											},
											Body: &ast.BlockStmt{
												List: []ast.Stmt{
													&ast.ReturnStmt{
														Results: []ast.Expr{
															&ast.CallExpr{
																Fun: &ast.SelectorExpr{
																	X:   ast.NewIdent("fmt"),
																	Sel: ast.NewIdent("Errorf"),
																},
																Args: []ast.Expr{
																	&ast.BasicLit{
																		Kind:  token.STRING,
																		Value: `"failed to stop server: %w"`,
																	},
																	ast.NewIdent("err"),
																},
															},
														},
													},
												},
											},
										},
										&ast.ExprStmt{
											X: &ast.CallExpr{
												Fun: &ast.SelectorExpr{
													X: &ast.SelectorExpr{
														X:   ast.NewIdent("a"),
														Sel: ast.NewIdent("log"),
													},
													Sel: ast.NewIdent("Info"),
												},
												Args: []ast.Expr{
													&ast.BasicLit{
														Kind:  token.STRING,
														Value: `"Http server stopped"`,
													},
												},
											},
										},
										&ast.ReturnStmt{
											Results: []ast.Expr{ast.NewIdent("nil")},
										},
									},
								},
							},
						},
					},
				},

				// Set server
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   ast.NewIdent("a"),
							Sel: ast.NewIdent("srv"),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("s"),
								Sel: ast.NewIdent("Server"),
							},
						},
					},
				},

				// Return nil
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("nil")},
				},
			},
		},
	}

	// Add method to run HTTP server
	runHttpMethod := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("runHttpServer"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// Log server start
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							Sel: ast.NewIdent("Info"),
						},
						Args: []ast.Expr{
							&ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   ast.NewIdent("fmt"),
									Sel: ast.NewIdent("Sprintf"),
								},
								Args: []ast.Expr{
									&ast.BasicLit{
										Kind:  token.STRING,
										Value: `"Запуск HTTP сервера на %s"`,
									},
									&ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("srv"),
										},
										Sel: ast.NewIdent("Addr"),
									},
								},
							},
						},
					},
				},

				// Start server
				&ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("err")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("a"),
									Sel: ast.NewIdent("srv"),
								},
								Sel: ast.NewIdent("ListenAndServe"),
							},
						},
					},
				},

				// Check for errors
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y:  ast.NewIdent("nil"),
						},
						Op: token.LAND,
						Y: &ast.BinaryExpr{
							X:  ast.NewIdent("err"),
							Op: token.NEQ,
							Y: &ast.SelectorExpr{
								X:   ast.NewIdent("http"),
								Sel: ast.NewIdent("ErrServerClosed"),
							},
						},
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("log"),
										},
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										&ast.CallExpr{
											Fun: &ast.SelectorExpr{
												X:   ast.NewIdent("fmt"),
												Sel: ast.NewIdent("Sprintf"),
											},
											Args: []ast.Expr{
												&ast.BasicLit{
													Kind:  token.STRING,
													Value: `"Ошибка HTTP сервера: %v"`,
												},
												ast.NewIdent("err"),
											},
										},
									},
								},
							},
							&ast.ReturnStmt{
								Results: []ast.Expr{ast.NewIdent("err")},
							},
						},
					},
				},

				// Return nil
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("nil")},
				},
			},
		},
	}

	file.Decls = append(file.Decls, method)
	file.Decls = append(file.Decls, runHttpMethod)
}

// addInitDatabaseMethod adds the database initialization method
func (g *Generator) addInitDatabaseMethod(file *ast.File, dbType string) {
	// Check if method already exists
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "initDatabase" {
			// Method already exists, no need to add it
			return
		}
	}

	// Create the initDatabase method
	method := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("initDatabase"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				// a.log.Info("Initializing database connection")
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							Sel: ast.NewIdent("Info"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: `"Initializing database connection"`,
							},
						},
					},
				},
				// Connect to database
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						ast.NewIdent("var"),
						ast.NewIdent("err"),
					},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("sql"),
								Sel: ast.NewIdent("Open"),
							},
							Args: []ast.Expr{
								&ast.BasicLit{
									Kind:  token.STRING,
									Value: `"postgres"`,
								},
								&ast.SelectorExpr{
									X: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("cfg"),
										},
										Sel: ast.NewIdent("Database"),
									},
									Sel: ast.NewIdent("URL"),
								},
							},
						},
					},
				},
				// Check for errors
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X:  ast.NewIdent("err"),
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("log"),
										},
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"Failed to connect to database"`,
										},
										ast.NewIdent("err"),
									},
								},
							},
							&ast.ReturnStmt{},
						},
					},
				},
				// Set the database connection
				&ast.AssignStmt{
					Lhs: []ast.Expr{
						&ast.SelectorExpr{
							X:   ast.NewIdent("a"),
							Sel: ast.NewIdent("db"),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []ast.Expr{
						ast.NewIdent("db"),
					},
				},
				// Initialize the database if needed
				&ast.IfStmt{
					Cond: &ast.BinaryExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   ast.NewIdent("db"),
								Sel: ast.NewIdent("Ping"),
							},
						},
						Op: token.NEQ,
						Y:  ast.NewIdent("nil"),
					},
					Body: &ast.BlockStmt{
						List: []ast.Stmt{
							&ast.ExprStmt{
								X: &ast.CallExpr{
									Fun: &ast.SelectorExpr{
										X: &ast.SelectorExpr{
											X:   ast.NewIdent("a"),
											Sel: ast.NewIdent("log"),
										},
										Sel: ast.NewIdent("Error"),
									},
									Args: []ast.Expr{
										&ast.BasicLit{
											Kind:  token.STRING,
											Value: `"Failed to ping database"`,
										},
										ast.NewIdent("err"),
									},
								},
							},
							&ast.ReturnStmt{},
						},
					},
				},
				// a.log.Info("Connected to database")
				&ast.ExprStmt{
					X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							Sel: ast.NewIdent("Info"),
						},
						Args: []ast.Expr{
							&ast.BasicLit{
								Kind:  token.STRING,
								Value: `"Connected to database"`,
							},
						},
					},
				},
			},
		},
	}

	file.Decls = append(file.Decls, method)
}

// Add the initServices method to handle both database and no-database cases
func (g *Generator) addInitServicesMethod(file *ast.File, hasDatabase bool) {
	// Check if method already exists
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "initServices" {
			// Method already exists, no need to add it
			return
		}
	}

	// Create the initServices method with appropriate parameters based on database presence
	var methodBody []ast.Stmt
	if hasDatabase {
		// With database, initialize service with repo and DB agent
		methodBody = []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   ast.NewIdent("a"),
						Sel: ast.NewIdent("service"),
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("service"),
							Sel: ast.NewIdent("New"),
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							ast.NewIdent("serviceName"),
							&ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("repo"),
							},
							&ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("db"),
							},
						},
					},
				},
			},
		}
	} else {
		// Without database, initialize service with just logger and service name
		methodBody = []ast.Stmt{
			&ast.AssignStmt{
				Lhs: []ast.Expr{
					&ast.SelectorExpr{
						X:   ast.NewIdent("a"),
						Sel: ast.NewIdent("service"),
					},
				},
				Tok: token.ASSIGN,
				Rhs: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent("service"),
							Sel: ast.NewIdent("New"),
						},
						Args: []ast.Expr{
							&ast.SelectorExpr{
								X:   ast.NewIdent("a"),
								Sel: ast.NewIdent("log"),
							},
							ast.NewIdent("serviceName"),
						},
					},
				},
			},
		}
	}

	// Add initialization for gRPC and GraphQL services
	methodBody = append(methodBody, &ast.ReturnStmt{
		Results: []ast.Expr{ast.NewIdent("nil")},
	})

	method := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("a")},
					Type:  &ast.StarExpr{X: ast.NewIdent("App")},
				},
			},
		},
		Name: ast.NewIdent("initServices"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("_")},
						Type:  ast.NewIdent("context.Context"),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("error"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: methodBody,
		},
	}

	file.Decls = append(file.Decls, method)
}
