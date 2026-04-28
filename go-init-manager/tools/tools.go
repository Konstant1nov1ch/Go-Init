//go:build tools

package tools

// Пустые импорты держат gqlgen в графе модулей для go generate / codegen.
// Корень github.com/99designs/gqlgen — исполняемый пакет, его нельзя импортировать как библиотеку.
import (
	_ "github.com/99designs/gqlgen/graphql/introspection"
)
