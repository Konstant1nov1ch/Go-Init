package engine

import (
	"context"

	"go-init-gen/internal/eventdata"
)

// Preprocessor defines the interface for preprocessing input data
type Preprocessor interface {
	// Process preprocesses the input data
	Process(ctx context.Context, data *eventdata.TemplateEventData) error
}

// CodeGenerator defines the interface for code generators
type CodeGenerator interface {
	// SupportedStrategies returns a list of supported code generation strategies
	SupportedStrategies() []CodeGenStrategy

	// GenerateCode generates code based on a template file and data
	GenerateCode(ctx context.Context, file TemplateFile, data *eventdata.TemplateEventData) (string, error)
}

// Postprocessor defines the interface for post-processing generated code
type Postprocessor interface {
	// Process post-processes the generated code
	Process(ctx context.Context, code []byte) ([]byte, error)
}

// TemplateGenerator is the basic templating generator
type templateGenerator struct{}

// NewTemplateGenerator creates a new template-based code generator
func NewTemplateGenerator() CodeGenerator {
	return &templateGenerator{}
}

// SupportedStrategies returns the strategies supported by this generator
func (g *templateGenerator) SupportedStrategies() []CodeGenStrategy {
	return []CodeGenStrategy{StrategyTextTemplate, StrategyRaw}
}

// GenerateCode generates code using text templates
func (g *templateGenerator) GenerateCode(ctx context.Context, file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	// This is a placeholder - in real implementation, it would use text/template
	return file.Content, nil
}

// ASTGenerator is the AST-based code generator
type astGenerator struct{}

// NewASTGenerator creates a new AST-based code generator
func NewASTGenerator() CodeGenerator {
	return &astGenerator{}
}

// SupportedStrategies returns the strategies supported by this generator
func (g *astGenerator) SupportedStrategies() []CodeGenStrategy {
	return []CodeGenStrategy{StrategyASTGeneration}
}

// GenerateCode generates code using AST manipulation
func (g *astGenerator) GenerateCode(ctx context.Context, file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	// This is a placeholder - in real implementation, it would use Go's AST package
	return file.Content, nil
}

// HybridGenerator combines template and AST approaches
type hybridGenerator struct{}

// NewHybridGenerator creates a new hybrid code generator
func NewHybridGenerator() CodeGenerator {
	return &hybridGenerator{}
}

// SupportedStrategies returns the strategies supported by this generator
func (g *hybridGenerator) SupportedStrategies() []CodeGenStrategy {
	return []CodeGenStrategy{StrategyHybrid}
}

// GenerateCode generates code using a hybrid approach
func (g *hybridGenerator) GenerateCode(ctx context.Context, file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	// This is a placeholder - in real implementation, it would combine templating and AST
	return file.Content, nil
}
