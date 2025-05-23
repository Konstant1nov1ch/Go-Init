package model

import (
	"go/ast"
	"go/token"
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
)

// Generator implements the model code generation
type Generator struct{}

// NewGenerator creates a new model generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the model code
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Add imports for time fields
	generators.AddImports(file, []string{"time"})

	// Create base model type
	baseModel := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent("BaseModel"),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("ID")},
								Type:  ast.NewIdent("int64"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"id\"`"},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("CreatedAt")},
								Type:  ast.NewIdent("time.Time"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"created_at\"`"},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("UpdatedAt")},
								Type:  ast.NewIdent("time.Time"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"updated_at\"`"},
							},
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, baseModel)

	// Create specific model types based on the service name
	modelName := strings.Title(data.Name) + "Model"
	modelStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(modelName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("BaseModel")},
								Type:  ast.NewIdent("BaseModel"),
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Name")},
								Type:  ast.NewIdent("string"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"name\"`"},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Description")},
								Type:  ast.NewIdent("string"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"description\"`"},
							},
							// Add more fields based on what's available in the template data
						},
					},
				},
			},
		},
	}
	file.Decls = append(file.Decls, modelStruct)

	return nil
}
