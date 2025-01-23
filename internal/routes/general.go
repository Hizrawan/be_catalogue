package routes

import (
	"be20250107/internal/app"
	"be20250107/internal/controllers/public"

	"github.com/go-chi/chi/v5"
)

func RegisterGeneralRoutes(root chi.Router, app *app.Registry) {
	root.Route("/public", func(r chi.Router) {})
	root.Mount("/", KeysRoutes(app))
}

func KeysRoutes(app *app.Registry) chi.Router {
	controller := public.NewKeysController(app)

	r := chi.NewRouter()
	r.Get("/keys", controller.Keys)

	return r
}
