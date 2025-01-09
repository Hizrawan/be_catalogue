package routes

import (
	controller "be20250107/internal/controllers/catalogue"

	"be20250107/internal/app"
	"be20250107/internal/middlewares"

	"github.com/go-chi/chi/v5"
)

func RegisterCatalogueRoutes(root chi.Router, app *app.Registry) {
	CatalogueController := controller.NewCatalogueController(app)

	root.Route("/catalogues", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middlewares.AuthMiddleware(app))
			r.Post("/", CatalogueController.CreateCatalogue)
			r.Get("/{CatalogueID}", CatalogueController.GetCatalogue)
			r.Patch("/{CatalogueID}", CatalogueController.UpdateCatalogue)
			r.Delete("/{CatalogueID}", CatalogueController.DeleteCatalogue)
			r.Get("/", CatalogueController.GetCatalogues)
		})
	})
}
