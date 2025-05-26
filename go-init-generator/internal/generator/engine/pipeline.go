package engine

import (
	"context"
	"fmt"

	"go-init-gen/internal/eventdata"
)

// GenerationPipeline orchestrates the code generation process
type GenerationPipeline struct {
	templateLoader   *TemplateLoader
	fileFilter       *FeatureBasedFileFilter
	contentGenerator *ContentGenerator
	archiver         *Archiver
}

// NewGenerationPipeline creates a new generation pipeline
func NewGenerationPipeline(templateDir string, debugArchives bool, debugDir string) *GenerationPipeline {
	return &GenerationPipeline{
		templateLoader:   NewTemplateLoader(templateDir),
		fileFilter:       NewFileFilter(),
		contentGenerator: NewContentGenerator(templateDir),
		archiver:         NewArchiver(debugArchives, debugDir),
	}
}

// Execute runs the complete generation pipeline
func (p *GenerationPipeline) Execute(ctx context.Context, template *eventdata.ProcessTemplate) ([]byte, error) {
	// Step 1: Prepare template variables
	variables, err := p.prepareTemplateVariables(&template.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare template variables: %w", err)
	}

	// Step 2: Load template files
	files, err := p.templateLoader.LoadTemplateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load template files: %w", err)
	}

	// Step 3: Filter files based on features
	filesToGenerate := p.fileFilter.FilterFiles(files, &template.Data)

	// Step 4: Generate file content
	generatedFiles, err := p.contentGenerator.GenerateFiles(filesToGenerate, &template.Data, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to generate files: %w", err)
	}

	// Step 5: Create archive
	archiveBytes, err := p.archiver.CreateArchive(generatedFiles, template.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	return archiveBytes, nil
}

// prepareTemplateVariables prepares variables for template rendering
func (p *GenerationPipeline) prepareTemplateVariables(data *eventdata.TemplateEventData) (map[string]interface{}, error) {
	customizer := NewCustomizer()
	customizer.ProcessInput(data)

	variables := customizer.GetVariables()
	features := customizer.GetFeatureFlags()

	// Add features to variables for template usage
	variables["features"] = features

	return variables, nil
}
