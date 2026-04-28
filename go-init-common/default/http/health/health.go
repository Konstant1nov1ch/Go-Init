package health

import (
	"net/http"

	"github.com/go-chi/render"
)

// Response представляет JSON-ответ о статусе сервиса.
type Response struct {
	Status string `json:"status"`
}

// HealthHandler возвращает обработчик статуса сервиса.
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, Response{Status: "UP"})
	}
}
