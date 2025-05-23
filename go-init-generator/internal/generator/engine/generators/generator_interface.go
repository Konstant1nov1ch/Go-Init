package generators

import (
	"go/ast"

	"go-init-gen/internal/eventdata"
)

// ASTComponentGenerator defines the interface for AST-based component generators
type ASTComponentGenerator interface {
	// Generate generates code for a specific component
	Generate(file *ast.File, data *eventdata.TemplateEventData) error
}
