package server

import (
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/gorilla/websocket"
)

// NewGraphQLServer возвращает http.Handler, который обслуживает GraphQL-запросы.
// Сюда можно добавлять нужные транспорты, middleware (метрики, трейсинг и т.д.)
// NewGraphQLServer создает GraphQL сервер с поддержкой WebSocket.
func NewGraphQLServer(es graphql.ExecutableSchema) http.Handler {
	srv := handler.New(es)

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		KeepAlivePingInterval: 10 * time.Second,
	})

	srv.Use(extension.Introspection{})

	return srv
}
