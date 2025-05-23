package service

import (
	"go/ast"
	"go/token"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
)

// Generator implements the service code generation
type Generator struct{}

// NewGenerator creates a new service generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the service code
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Check if a Service type already exists in the file
	serviceExists := false
	for _, decl := range file.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.Name == "Service" {
						serviceExists = true
						break
					}
				}
			}
			if serviceExists {
				break
			}
		}
	}

	// If a Service type already exists, don't add our generated service
	if serviceExists {
		return nil
	}

	// Add imports
	generators.AddImports(file, []string{
		"context",
	})

	// Create service interface
	serviceInterface := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("Service"),
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("GetStatus")},
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
											{Type: ast.NewIdent("string")},
											{Type: ast.NewIdent("error")},
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
	file.Decls = append(file.Decls, serviceInterface)

	// Create service struct
	serviceStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("ServiceImpl"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("repo")},
								Type:  ast.NewIdent("*Repository"),
							},
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, serviceStruct)

	// Create constructor
	constructorFunc := &ast.FuncDecl{
		Name: ast.NewIdent("NewService"),
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("repo")},
						Type:  ast.NewIdent("*Repository"),
					},
				},
			},
			Results: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: ast.NewIdent("Service"),
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.UnaryExpr{
							Op: token.AND,
							X: &ast.CompositeLit{
								Type: ast.NewIdent("ServiceImpl"),
								Elts: []ast.Expr{
									&ast.KeyValueExpr{
										Key:   ast.NewIdent("repo"),
										Value: ast.NewIdent("repo"),
									},
								},
							},
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, constructorFunc)

	// Add GetStatus method
	getStatusMethod := &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("s")},
					Type:  ast.NewIdent("*ServiceImpl"),
				},
			},
		},
		Name: ast.NewIdent("GetStatus"),
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
					{Type: ast.NewIdent("string")},
					{Type: ast.NewIdent("error")},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.BasicLit{
							Kind:  token.STRING,
							Value: "\"OK\"",
						},
						ast.NewIdent("nil"),
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, getStatusMethod)

	return nil
}
