package config

import (
	"go/ast"
	"go/token"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
	"go-init-gen/internal/generator/engine/generators/features"
)

// Generator implements the config.go code generation
type Generator struct{}

// NewGenerator creates a new config generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the config.go code based on the input data
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Use the feature detector to identify enabled features
	fs := features.DetectFeatures(data)

	// Clear existing imports and add the new ones
	g.setupImports(file, fs.HasPostgres(), fs.HasGRPC, fs.HasHTTP)

	// Replace or create AppConfig struct
	g.createAppConfigStruct(file, fs.HasPostgres(), fs.HasGRPC, fs.HasHTTP)

	// Create GetConfig function
	g.createGetConfigFunc(file, fs.HasPostgres(), fs.HasGRPC, fs.HasHTTP)

	return nil
}

// setupImports sets up the imports for the config file
func (g *Generator) setupImports(file *ast.File, hasPostgres, hasGRPC, hasHTTP bool) {
	// Remove all existing imports
	var nonImportDecls []ast.Decl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); !ok || genDecl.Tok != token.IMPORT {
			nonImportDecls = append(nonImportDecls, decl)
		}
	}
	file.Decls = nonImportDecls

	// Create new import declaration
	importDecl := &ast.GenDecl{
		Tok:    token.IMPORT,
		Lparen: token.Pos(1), // Enable multi-line import
		Specs:  []ast.Spec{},
	}

	// Add required imports
	requiredImports := []string{
		"github.com/mcuadros/go-defaults",
		"gitlab.com/go-init/go-init-common/default/config",
		"gitlab.com/go-init/go-init-common/default/logger",
	}

	// Add optional imports based on features
	if hasPostgres {
		requiredImports = append(requiredImports, "gitlab.com/go-init/go-init-common/default/db/pg")
	}
	if hasGRPC {
		requiredImports = append(requiredImports, "gitlab.com/go-init/go-init-common/default/grpcpkg")
	}
	if hasHTTP {
		requiredImports = append(requiredImports, "gitlab.com/go-init/go-init-common/default/http/server")
	}

	// Add all imports to the declaration
	for _, importPath := range requiredImports {
		var importName *ast.Ident

		// Add alias for config package
		if importPath == "gitlab.com/go-init/go-init-common/default/config" {
			importName = ast.NewIdent("c")
		}

		importSpec := &ast.ImportSpec{
			Name: importName,
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + importPath + `"`,
			},
		}
		importDecl.Specs = append(importDecl.Specs, importSpec)
	}

	// Add import declaration to the file
	file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
}

// createAppConfigStruct creates the AppConfig struct
func (g *Generator) createAppConfigStruct(file *ast.File, hasPostgres, hasGRPC, hasHTTP bool) {
	// Create the fields for the AppConfig struct
	fields := []*ast.Field{
		{
			Names: []*ast.Ident{ast.NewIdent("Logger")},
			Type:  ast.NewIdent("logger.Config"),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`yaml:\"logger\"`"},
		},
	}

	// Add optional fields based on features
	if hasPostgres {
		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("Database")},
			Type:  ast.NewIdent("pg.Config"),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`yaml:\"postgres_db\"`"},
		})
	}

	if hasHTTP {
		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("HttpServ")},
			Type:  ast.NewIdent("server.Config"),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`yaml:\"http_server\"`"},
		})
	}

	if hasGRPC {
		fields = append(fields, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("GrpcServ")},
			Type:  ast.NewIdent("grpcpkg.ServerConfig"),
			Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`yaml:\"grpc_server\"`"},
		})
	}

	// Create the AppConfig struct
	appConfigStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("AppConfig"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}

	// Remove any existing AppConfig struct
	var filteredDecls []ast.Decl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			keepDecl := true
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.Name == "AppConfig" {
					keepDecl = false
					break
				}
			}
			if keepDecl {
				filteredDecls = append(filteredDecls, decl)
			}
		} else {
			filteredDecls = append(filteredDecls, decl)
		}
	}
	file.Decls = filteredDecls

	// Add the new AppConfig struct
	file.Decls = append(file.Decls, appConfigStruct)
}

// createGetConfigFunc creates the GetConfig function
func (g *Generator) createGetConfigFunc(file *ast.File, hasPostgres, hasGRPC, hasHTTP bool) {
	// Create the statements for the function body
	bodyStmts := []ast.Stmt{
		// config := &AppConfig{}
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("config")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{
				&ast.UnaryExpr{
					Op: token.AND,
					X:  &ast.CompositeLit{Type: ast.NewIdent("AppConfig")},
				},
			},
		},
		// c.OpenConfig(&config)
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("c"),
					Sel: ast.NewIdent("OpenConfig"),
				},
				Args: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X:  ast.NewIdent("config"),
					},
				},
			},
		},
		// defaults.SetDefaults(&config.Logger)
		&ast.ExprStmt{
			X: &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   ast.NewIdent("defaults"),
					Sel: ast.NewIdent("SetDefaults"),
				},
				Args: []ast.Expr{
					&ast.UnaryExpr{
						Op: token.AND,
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("config"),
							Sel: ast.NewIdent("Logger"),
						},
					},
				},
			},
		},
		// return config
		&ast.ReturnStmt{
			Results: []ast.Expr{ast.NewIdent("config")},
		},
	}

	// Create the GetConfig function
	getConfigFunc := &ast.FuncDecl{
		Name: ast.NewIdent("GetConfig"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{X: ast.NewIdent("AppConfig")},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: bodyStmts,
		},
	}

	// Remove any existing GetConfig function
	var filteredDecls []ast.Decl
	for _, decl := range file.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "GetConfig" {
			continue
		}
		filteredDecls = append(filteredDecls, decl)
	}
	file.Decls = filteredDecls

	// Add the new GetConfig function
	file.Decls = append(file.Decls, getConfigFunc)
}
