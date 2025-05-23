# Go Init Generator

A code generation system for creating Go microservice templates based on user-provided configuration.

## Directory Structure

- `/engine` - Core generation engine
- `/templates` - Template implementations for different project types
- `/preprocessors` - Input validators and normalizers
- `/postprocessors` - Output processors like archivers

## Environment Configuration

The generator uses these environment variables:

- `STATIC_FILES_ARRAY` - JSON array of static files to always include in the template
- `DYNAMIC_FILES_ARRAY` - JSON array of dynamic files to conditionally generate
- `GENERATOR_SAVE_ARCHIVE_LOCALLY` - Save generated archives for debugging

## Guide: How to Modify or Add Files

### 1. Modifying an Existing Template File

To modify how an existing file is generated:

1. **Locate the template file**:
   ```
   /templates/microservice/static/            # Static templates (always included)
   /templates/microservice/dynamic/           # Dynamic templates (conditionally included)
   ```

2. **Edit the template content**:
   - Templates use Go's `text/template` syntax
   - Use `{{ .variableName }}` to insert dynamic content
   - Common variables:
     - `{{ .projectName }}` - Name of the service
     - `{{ .serviceDescription }}` - Description of the service
     - `{{ if .features.hasGRPC }}...{{ end }}` - Conditional content based on features

3. **Template functions**:
   - `{{ .serviceName | camelCase }}` - Convert to camelCase
   - `{{ .serviceName | snakeCase }}` - Convert to snake_case

### 2. Adding a New File to Generate

