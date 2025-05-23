package generators

import (
	"go/ast"
	"go/token"
)

// AddImports adds imports to an AST file
func AddImports(file *ast.File, imports []string) {
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

// AddNamedImport adds a named import to an AST file
func AddNamedImport(file *ast.File, name, path string) {
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

	// Check if this named import already exists
	for _, spec := range importDecl.Specs {
		importSpec, ok := spec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		// If the path matches and the name matches (or both are nil), it already exists
		if importSpec.Path.Value == "\""+path+"\"" &&
			(importSpec.Name == nil && name == "" ||
				importSpec.Name != nil && importSpec.Name.Name == name) {
			return
		}
	}

	// Add the named import
	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + path + "\"",
		},
	}

	// Only set the name if it's not empty
	if name != "" {
		importSpec.Name = ast.NewIdent(name)
	}

	importDecl.Specs = append(importDecl.Specs, importSpec)
}

// CreateStructType creates a new struct type declaration
func CreateStructType(name string, fields []*ast.Field) *ast.GenDecl {
	return &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(name),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: fields,
					},
				},
			},
		},
	}
}

// CreateFuncDecl creates a function declaration
func CreateFuncDecl(name string, params, results []*ast.Field, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results},
		},
		Body: body,
	}
}

// CreateMethodDecl creates a method declaration
func CreateMethodDecl(recv *ast.Field, name string, params, results []*ast.Field, body *ast.BlockStmt) *ast.FuncDecl {
	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{recv},
		},
		Name: ast.NewIdent(name),
		Type: &ast.FuncType{
			Params:  &ast.FieldList{List: params},
			Results: &ast.FieldList{List: results},
		},
		Body: body,
	}
}
