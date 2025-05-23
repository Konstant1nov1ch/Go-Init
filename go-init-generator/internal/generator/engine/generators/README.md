# Code Generators

This package contains specialized component generators for the Go code generation system. 
Each generator is responsible for generating a specific type of file, making the codebase more maintainable and extensible.

## Generator Interface

All component generators implement the `ASTComponentGenerator` interface:

```go
type ASTComponentGenerator interface {
    Generate(file *ast.File, data *eventdata.TemplateEventData) error
}
```

## Available Generators

The following generators are available:

1. **Repository Generator** - Generates data access layer code
2. **Model Generator** - Generates data model structures
3. **Handler Generator** - Generates HTTP handlers
4. **Service Generator** - Generates business logic services
5. **Entity Generator** - Generates domain entities

## Utility Functions

The package provides common utility functions for working with AST:

- `AddImports` - Adds imports to an AST file
- `CreateStructType` - Creates a new struct type declaration
- `CreateFuncDecl` - Creates a function declaration
- `CreateMethodDecl` - Creates a method declaration

## Usage

Generators are used by the `ASTGeneratorNew` class to generate code based on the file type:

```go
// Example usage of a component generator
repositoryGenerator := repository.NewGenerator()
repositoryGenerator.Generate(file, templateData)
```

## Extending

To add a new type of generator:

1. Create a new package under `generators/`
2. Implement the `ASTComponentGenerator` interface
3. Register your generator in `ASTGeneratorNew` 