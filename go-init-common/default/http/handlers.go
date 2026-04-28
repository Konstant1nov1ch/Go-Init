package http

import "net/http"

// MiddleWareHandlers управляет списком middleware обработчиков.
type MiddleWareHandlers struct {
	handlers []func(http.Handler) http.Handler
}

// AddHandler добавляет middleware обработчики.
func (m *MiddleWareHandlers) AddHandler(handlers ...func(http.Handler) http.Handler) {
	m.handlers = append(m.handlers, handlers...)
}

// GetHandlers возвращает список middleware обработчиков.
func (m *MiddleWareHandlers) GetHandlers() []func(http.Handler) http.Handler {
	return m.handlers
}

// CollectHandlers собирает middleware обработчики в структуру.
func CollectHandlers(handlers ...func(http.Handler) http.Handler) *MiddleWareHandlers {
	return &MiddleWareHandlers{handlers: handlers}
}
