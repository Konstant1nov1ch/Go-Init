package entity

import (
	"go/ast"
	"go/token"
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators"
	"go-init-gen/internal/generator/engine/generators/features"
)

// Generator implements the entity code generation
type Generator struct{}

// NewGenerator creates a new entity generator
func NewGenerator() generators.ASTComponentGenerator {
	return &Generator{}
}

// Generate generates the entity code
func (g *Generator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
	// Use the feature detector to identify enabled features
	fs := features.DetectFeatures(data)

	// Add imports for potential external types
	if fs.HasDatabase {
		generators.AddImports(file, []string{"time"})
	}

	// Create entity struct based on service name
	entityName := strings.Title(data.Name)
	entityStruct := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(entityName),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("ID")},
								Type:  ast.NewIdent("int64"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"id\"`"},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Name")},
								Type:  ast.NewIdent("string"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"name\"`"},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Active")},
								Type:  ast.NewIdent("bool"),
								Tag:   &ast.BasicLit{Kind: token.STRING, Value: "`json:\"active\"`"},
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
	file.Decls = append(file.Decls, entityStruct)

	// Add repository interface for the entity
	repoInterface := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(entityName + "Repository"),
				Type: &ast.InterfaceType{
					Methods: &ast.FieldList{
						List: []*ast.Field{
							{
								Names: []*ast.Ident{ast.NewIdent("Get")},
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: []*ast.Field{
											{
												Names: []*ast.Ident{ast.NewIdent("id")},
												Type:  ast.NewIdent("int64"),
											},
										},
									},
									Results: &ast.FieldList{
										List: []*ast.Field{
											{Type: &ast.StarExpr{X: ast.NewIdent(entityName)}},
											{Type: ast.NewIdent("error")},
										},
									},
								},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Create")},
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: []*ast.Field{
											{
												Names: []*ast.Ident{ast.NewIdent("entity")},
												Type:  &ast.StarExpr{X: ast.NewIdent(entityName)},
											},
										},
									},
									Results: &ast.FieldList{
										List: []*ast.Field{
											{Type: ast.NewIdent("error")},
										},
									},
								},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Update")},
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: []*ast.Field{
											{
												Names: []*ast.Ident{ast.NewIdent("entity")},
												Type:  &ast.StarExpr{X: ast.NewIdent(entityName)},
											},
										},
									},
									Results: &ast.FieldList{
										List: []*ast.Field{
											{Type: ast.NewIdent("error")},
										},
									},
								},
							},
							{
								Names: []*ast.Ident{ast.NewIdent("Delete")},
								Type: &ast.FuncType{
									Params: &ast.FieldList{
										List: []*ast.Field{
											{
												Names: []*ast.Ident{ast.NewIdent("id")},
												Type:  ast.NewIdent("int64"),
											},
										},
									},
									Results: &ast.FieldList{
										List: []*ast.Field{
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
	file.Decls = append(file.Decls, repoInterface)

	return nil
}
