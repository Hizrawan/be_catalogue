package routes

import (
	"be20250107/internal/app"
	"be20250107/internal/controllers/auth"
	"be20250107/internal/middlewares"

	"github.com/go-chi/chi/v5"
)

func RegisterAuthRoutes(root chi.Router, app *app.Registry) {
	root.Route("/auth", func(r chi.Router) {
		r.Mount("/admin", AdminAuthRoutes(app))
	})
}

func AdminAuthRoutes(app *app.Registry) chi.Router {
	controller := auth.NewAuthAdminController(app)
	r := chi.NewRouter()

	r.Post("/", controller.LoginByXinchuanAuth)

	r.Group(func(r chi.Router) {
		r.Use(middlewares.AdminAuthMiddleware(app))
		r.Get("/", controller.Me)
		r.Delete("/", controller.Logout)
	})

	return r
}
