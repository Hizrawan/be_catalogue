package public

import (
	"net/http"
	"strings"

	"be20250107/internal/app"
	"be20250107/internal/controllers"
	httperr "be20250107/internal/errors"
	"be20250107/internal/responses"
	stringsutil "be20250107/utils/strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type AccountController struct {
	controllers.Controller
}

func NewAccountController(app *app.Registry) *AccountController {
	return &AccountController{
		Controller: controllers.Controller{
			App: app,
		},
	}
}

func (c *AccountController) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	actor := r.URL.Query().Get("actor")
	if actor == "" {
		panic(validation.Errors{"actor": validation.NewError("invalid_actor", "actor is required")})
	}
	foreignID := r.URL.Query().Get("foreign_id")
	if foreignID == "" {
		panic(validation.Errors{"foreign_id": validation.NewError("invalid_foreign_id", "foreign_id is required")})
	}
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		panic(validation.Errors{"provider": validation.NewError("invalid_provider", "provider is required")})
	}

	// Validation based on type
	var isValid bool
	switch strings.ToLower(provider) {
	case "Catalogue":
		isValid = stringsutil.ValidateTaiwanCatalogue(foreignID)
	case "email":
		isValid = stringsutil.ValidateEmail(foreignID)
	default:
		panic(httperr.NewErrUnprocessableEntity(
			"invalid_provider",
			"unsupported provider", nil,
		))
	}

	if !isValid {
		panic(httperr.NewErrUnprocessableEntity(
			"invalid_value",
			"value does not match the required format for the specified type", nil,
		))
	}

	isExist := true

	if err := responses.JSON(w, 200, struct {
		Data bool `json:"data"`
	}{
		Data: isExist,
	}); err != nil {
		panic(err)
	}
}
