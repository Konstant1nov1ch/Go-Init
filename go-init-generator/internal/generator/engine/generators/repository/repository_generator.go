package repository

import (
	"go/ast"
	"go/token"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
	"go-init-gen/internal/generator/engine/generators/features"
)

// Generator implements the repository code generation
type Generator struct{}

// NewGenerator creates a new repository generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the repository code
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Use the feature detector to identify enabled features
	fs := features.DetectFeatures(data)

	// Add database-specific imports
	g.addImports(file, []string{
		"database/sql",
	})

	switch fs.DatabaseType {
	case features.DatabaseTypePostgres, features.DatabaseTypePostgresql:
		g.addImports(file, []string{"github.com/lib/pq"})
	case features.DatabaseTypeMysql:
		g.addImports(file, []string{"github.com/go-sql-driver/mysql"})
	}

	// Add repository struct
	repoStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("Repository"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("db")},
								Type:  ast.NewIdent("*sql.DB"),
							},
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, repoStruct)

	// Add constructor
	constructor := g.createConstructor("NewRepository",
		[]*ast.Field{{Type: ast.NewIdent("*sql.DB")}},
		[]*ast.Field{{Type: ast.NewIdent("*Repository")}},
		&ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: ast.NewIdent("Repository"),
								Elts: []ast.Expr{
									&ast.KeyValueExpr{
										Key:   ast.NewIdent("db"),
										Value: ast.NewIdent("db"),
									},
								},
							},
						},
					},
				},
			},
		},
	)
	file.Decls = append(file.Decls, constructor)

	return nil
}

// addImports adds imports to the file
func (g *Generator) addImports(file *ast.File, imports []string) {
	// Check if the file already has an import declaration
	var importDecl *ast.GenDecl
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			importDecl = genDecl
			break
		}
	}

	// Create a new import declaration if none exists
	if importDecl == nil {
		importDecl = &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: 1, // This enables multiline imports with parentheses
			Specs:  []ast.Spec{},
		}
		file.Decls = append([]ast.Decl{importDecl}, file.Decls...)
	}

	// Add each import
	for _, pkg := range imports {
		// Check if import already exists
		alreadyExists := false
		for _, spec := range importDecl.Specs {
			importSpec, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			if importSpec.Path.Value == "\""+pkg+"\"" {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			importDecl.Specs = append(importDecl.Specs, &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"" + pkg + "\"",
				},
			})
		}
	}
}

// createConstructor creates a constructor function
func (g *Generator) createConstructor(name string, params, results []*ast.Field, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results},
		},
		Body: body,
	}
}
