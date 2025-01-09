package routes

import (
	"be20250107/internal/app"
	"be20250107/internal/controllers/public"

	"github.com/go-chi/chi/v5"
)

func RegisterAccountRoutes(root chi.Router, app *app.Registry) {
	root.Route("/accounts", func(r chi.Router) {
		accountController := public.NewAccountController(app)

		r.Get("/check-account-duplicate", accountController.CheckDuplicate)
	})
}
