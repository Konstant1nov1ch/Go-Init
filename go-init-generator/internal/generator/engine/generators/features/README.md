# Features Package

This package centralizes the feature detection logic for the code generators. Instead of each generator implementing its own logic for determining which features to enable, this package provides a common interface for all generators.

## Usage

```go
import "go-init-gen/internal/generator/engine/generators/features"

func (g *MyGenerator) Generate(file *ast.File, data *eventdata.TemplateEventData) error {
    // Detect features from template data
    fs := features.DetectFeatures(data)
    
    // Check for specific features
    if fs.HasGRPC {
        // Add gRPC-specific code
    }
    
    if fs.HasHTTP {
        // Add HTTP server code
    }
    
    if fs.HasDatabase {
        // Add database-specific code
        
        // Check database type
        if fs.HasPostgres() {
            // PostgreSQL-specific code
        } else if fs.HasMySQL() {
            // MySQL-specific code
        }
    }
    
    return nil
}
```

## Constants

The package defines constants for protocol types, database types, and feature names to ensure consistency across all generators.

## Benefits

1. **Consistency**: All generators use the same logic to detect features.
2. **Maintainability**: Changes to feature detection logic only need to be made in one place.
3. **Extensibility**: New features can be added easily without modifying multiple generators.
4. **Testability**: Feature detection logic can be tested independently of generators.

## Adding New Features

To add a new feature:

1. Add appropriate constants to `constants.go`
2. Update the `FeatureSet` struct in `detector.go`
3. Update the `DetectFeatures` function to detect the new feature
4. Add any helper methods if needed 