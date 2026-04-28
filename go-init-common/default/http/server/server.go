package server

import (
	"net"
	"net/http"

	http2 "gitlab.com/go-init/go-init-common/default/http"
	"gitlab.com/go-init/go-init-common/default/http/health"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
)

// HttpServer представляет HTTP-сервер.
type HttpServer struct {
	http.Server
	port string
}

// NewServer создает новый HTTP-сервер.
func NewServer(config *Config, server, metricsHandler http.Handler, handlers *http2.MiddleWareHandlers) *HttpServer {
	router := chi.NewRouter()

	for _, handler := range handlers.GetHandlers() {
		router.Use(handler)
	}
	router.Use(middleware.Recoverer)

	if server != nil {
		router.Mount("/graphql", server)
		router.Mount("/playground", playground.Handler("GraphQL Playground", "/graphql"))
	}

	router.Handle("/metrics", metricsHandler)
	router.Handle("/health", health.HealthHandler())

	return &HttpServer{
		Server: http.Server{
			Addr:         ":" + config.Port,
			Handler:      router,
			ReadTimeout:  config.Timeout,
			WriteTimeout: config.Timeout,
			IdleTimeout:  config.IdleTimeout,
		},
		port: config.Port,
	}
}

// ListenAndServe запускает сервер.
func (s *HttpServer) ListenAndServe() error {
	lsn, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}
	return s.Serve(lsn)
}
