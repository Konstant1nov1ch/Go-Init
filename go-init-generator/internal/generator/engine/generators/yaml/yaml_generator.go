package yaml

import (
	"fmt"
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/features"
)

// Generator implements configuration YAML files content generation
type Generator struct{}

// NewGenerator creates a new YAML generator
func NewGenerator() *Generator {
	return &Generator{}
}

// ProcessConfigYAML модифицирует содержимое config.yml в зависимости от входных параметров
func (g *Generator) ProcessConfigYAML(content string, data *eventdata.TemplateEventData) (string, error) {
	// Используем детектор возможностей для определения доступных функций
	fs := features.DetectFeatures(data)

	// Подробное логирование входных данных и обнаруженных фич
	fmt.Printf("=== YAML config generation debug info ===\n")
	fmt.Printf("Input endpoints: %d\n", len(data.Endpoints))
	for i, endpoint := range data.Endpoints {
		fmt.Printf("Endpoint %d: Protocol=%q, Role=%q\n",
			i, endpoint.Protocol, endpoint.Role)
	}

	fmt.Printf("Database type: %q\n", data.Database.Type)
	fmt.Printf("Detected features: hasGRPC=%v, hasGraphQL=%v, hasHTTP=%v, hasDatabase=%v\n",
		fs.HasGRPC, fs.HasGraphQL, fs.HasHTTP, fs.HasDatabase)

	// Разделяем YAML на строки для более точной обработки
	lines := strings.Split(content, "\n")

	// Результирующие строки
	var resultLines []string

	// Флаги для отслеживания секций
	inHttpServerSection := false
	inGrpcServerSection := false
	inPostgresDbSection := false

	// Обрабатываем строки по одной
	for _, line := range lines {
		// Проверяем начало секций
		trimmedLine := strings.TrimSpace(line)

		// Определяем начало секций верхнего уровня
		if trimmedLine == "http_server:" {
			inHttpServerSection = true
			inGrpcServerSection = false
			inPostgresDbSection = false

			// Пропускаем эту секцию, если не нужна
			if !fs.HasHTTP && !fs.HasGraphQL {
				fmt.Printf("Skipping http_server section (hasHTTP=%v, hasGraphQL=%v)\n",
					fs.HasHTTP, fs.HasGraphQL)
				continue
			}
		} else if trimmedLine == "grpc_server:" {
			inHttpServerSection = false
			inGrpcServerSection = true
			inPostgresDbSection = false

			// Пропускаем эту секцию, если нет gRPC
			if !fs.HasGRPC {
				fmt.Printf("Skipping grpc_server section (hasGRPC=%v)\n", fs.HasGRPC)
				continue
			}
		} else if trimmedLine == "postgres_db:" {
			inHttpServerSection = false
			inGrpcServerSection = false
			inPostgresDbSection = true

			// Пропускаем эту секцию, если нет базы данных
			if !fs.HasDatabase {
				fmt.Printf("Skipping postgres_db section (hasDatabase=%v)\n", fs.HasDatabase)
				continue
			}
		} else if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "#") && !strings.HasPrefix(line, " ") {
			// Это новая секция верхнего уровня (не вложенная, не комментарий, не пустая)
			inHttpServerSection = false
			inGrpcServerSection = false
			inPostgresDbSection = false
		}

		// Пропускаем строки из секций, которые нужно исключить
		if (inHttpServerSection && !fs.HasHTTP && !fs.HasGraphQL) ||
			(inGrpcServerSection && !fs.HasGRPC) ||
			(inPostgresDbSection && !fs.HasDatabase) {
			continue
		}

		// Добавляем строку в результат
		resultLines = append(resultLines, line)
	}

	// Удаляем пустые строки в начале и в конце
	// и убираем повторяющиеся пустые строки между секциями
	cleanedLines := cleanEmptyLines(resultLines)

	// Собираем результат
	result := strings.Join(cleanedLines, "\n")

	// Убеждаемся, что файл заканчивается переводом строки
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	// Логируем, какие секции остались в результате
	hasHttpSection := strings.Contains(result, "http_server:")
	hasGrpcSection := strings.Contains(result, "grpc_server:")
	hasPostgresSection := strings.Contains(result, "postgres_db:")

	fmt.Printf("Result sections: HTTP=%v, gRPC=%v, Postgres=%v\n",
		hasHttpSection, hasGrpcSection, hasPostgresSection)
	fmt.Printf("=== End of YAML generation debug info ===\n")

	return result, nil
}

// cleanEmptyLines удаляет лишние пустые строки
func cleanEmptyLines(lines []string) []string {
	var result []string
	wasEmpty := false

	// Сначала удаляем пустые строки в начале
	startIndex := 0
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			startIndex = i
			break
		}
	}

	// Обрабатываем остальные строки, убирая повторяющиеся пустые строки
	for i := startIndex; i < len(lines); i++ {
		line := lines[i]
		isEmpty := strings.TrimSpace(line) == ""

		if isEmpty && wasEmpty {
			// Пропускаем повторные пустые строки
			continue
		}

		result = append(result, line)
		wasEmpty = isEmpty
	}

	// Удаляем пустые строки в конце
	for len(result) > 0 && strings.TrimSpace(result[len(result)-1]) == "" {
		result = result[:len(result)-1]
	}

	return result
}

// ProcessMakefile модифицирует содержимое Makefile в зависимости от входных параметров
func (g *Generator) ProcessMakefile(content string, data *eventdata.TemplateEventData) (string, error) {
	// Используем детектор возможностей для определения доступных функций
	fs := features.DetectFeatures(data)

	// Логируем информацию о фичах
	fmt.Printf("Makefile generation: hasGRPC=%v, hasGraphQL=%v, hasDatabase=%v, hasHTTP=%v\n",
		fs.HasGRPC, fs.HasGraphQL, fs.HasDatabase, fs.HasHTTP)

	lines := strings.Split(content, "\n")
	var resultLines []string

	skipTarget := false

	for _, line := range lines {
		// Проверяем, является ли строка целевой задачей (target)
		if strings.HasSuffix(line, ":") && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			skipTarget = false
			// Определяем targetName в области видимости этого блока if
			targetName := strings.TrimSpace(strings.TrimSuffix(line, ":"))

			// Определяем, нужно ли пропустить эту задачу
			if strings.Contains(targetName, "grpc") || strings.Contains(targetName, "proto") {
				skipTarget = !fs.HasGRPC
			} else if strings.Contains(targetName, "graphql") || strings.Contains(targetName, "gql") {
				skipTarget = !fs.HasGraphQL
			} else if strings.Contains(targetName, "http") || strings.Contains(targetName, "rest") {
				skipTarget = !fs.HasHTTP && !fs.HasGraphQL && !fs.HasREST
			} else if strings.Contains(targetName, "db") || strings.Contains(targetName, "migrate") {
				skipTarget = !fs.HasDatabase
			}
		} else if line == "" || (len(line) > 0 && line[0] != '\t') {
			// Пустая строка или строка без отступа означает конец цели
			skipTarget = false
		}

		// Добавляем строку, если она не в пропускаемой цели
		if !skipTarget {
			resultLines = append(resultLines, line)
		}
	}

	return strings.Join(resultLines, "\n"), nil
}

// splitYAMLIntoSections разбивает YAML-содержимое на секции с учетом отступов
func (g *Generator) splitYAMLIntoSections(content string) []string {
	lines := strings.Split(content, "\n")
	sections := []string{}

	// Текущая секция первого уровня
	var currentSection strings.Builder

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Пропускаем пустые строки между секциями
		if trimmed == "" && currentSection.Len() == 0 {
			sections = append(sections, line+"\n")
			continue
		}

		// Если это начало новой секции первого уровня (без отступа)
		if len(line) > 0 && line[0] != ' ' && strings.HasSuffix(trimmed, ":") {
			// Если у нас уже есть накопленная секция, добавляем её в результат
			if currentSection.Len() > 0 {
				sections = append(sections, currentSection.String())
				currentSection.Reset()
			}

			// Начинаем новую секцию
			currentSection.WriteString(line + "\n")
		} else {
			// Это строка, относящаяся к текущей секции или пустая строка
			if currentSection.Len() > 0 || trimmed != "" {
				currentSection.WriteString(line + "\n")
			} else {
				// Пустая строка между секциями
				sections = append(sections, line+"\n")
			}
		}

		// Если это последняя строка и у нас есть накопленная секция
		if i == len(lines)-1 && currentSection.Len() > 0 {
			sections = append(sections, currentSection.String())
		}
	}

	return sections
}
