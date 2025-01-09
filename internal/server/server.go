package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"be20250107/internal/app"
	"be20250107/internal/config"
	"be20250107/internal/middlewares"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Server struct {
	TLS    bool
	App    *app.Registry
	Router *chi.Mux
	Http   *http.Server
	Log    log.Logger
}

type RouteRegister func(root chi.Router, app *app.Registry)

func New() *Server {
	cfg := config.NewConfig(config.DefaultConfigName, config.DefaultConfigLocation)
	return NewWithConfig(cfg)
}

func NewWithConfig(cfg *config.Config) *Server {
	registry := app.NewRegistry(cfg, "main")

	router := chi.NewRouter()
	router.Use(middlewares.MetricsMiddleware(cfg.Public.PrometheusAPIJobName))
	router.NotFound(middlewares.NotFound(registry))
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-Geolocation"},
	}))
	router.Use(middleware.Logger)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.CleanPath)
	router.Use(middlewares.Recover(registry))

	server := Server{
		App:    registry,
		Router: router,
		Http: &http.Server{
			Handler: router,
			Addr:    fmt.Sprintf("%s:%d", cfg.Public.Listen.Host, cfg.Public.Listen.Port),
		},
		TLS: cfg.Public.Listen.EnableTLS,
	}

	return &server
}

func (s *Server) bootstrap() {
	s.BeforeStart()
	for _, route := range s.RegisterRoutes() {
		route(s.Router, s.App)
	}
}

func (s *Server) postStart() {
	s.AfterStart()
}

func (s *Server) Start() error {
	s.bootstrap()

	go func() {
		log.Printf("Server started and listening on %s\n", s.Http.Addr)
		if err := s.Http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v\n", err)
		}
	}()

	s.postStart()
	return nil
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if s.Http != nil {
		if err := s.Http.Shutdown(ctx); err != nil {
			log.Fatalf("Server shutdown failed: %+v", err)
		}
	}
}
