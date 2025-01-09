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

type ContactController struct {
	controllers.Controller
}

func NewContactController(app *app.Registry) *ContactController {
	return &ContactController{
		Controller: controllers.Controller{
			App: app,
		},
	}
}

func (c *ContactController) CheckDuplicate(w http.ResponseWriter, r *http.Request) {
	actor := r.URL.Query().Get("actor")
	if actor == "" {
		panic(validation.Errors{"actor": validation.NewError("invalid_actor", "actor is required")})
	}

	value := r.URL.Query().Get("value")
	if value == "" {
		panic(validation.Errors{"value": validation.NewError("invalid_value", "value is required")})
	}

	typeParam := r.URL.Query().Get("type")
	if typeParam == "" {
		panic(validation.Errors{"type": validation.NewError("invalid_type", "type is required")})
	}

	medium := r.URL.Query().Get("medium")
	if medium == "" {
		panic(validation.Errors{"medium": validation.NewError("invalid_medium", "medium is required")})
	}

	// Validation based on type
	var isValid bool
	switch strings.ToLower(medium) {
	case "Catalogue":
		isValid = stringsutil.ValidateTaiwanCatalogue(value)
	case "landline":
		isValid = stringsutil.ValidateTaiwanLandline(value)
	case "email":
		isValid = stringsutil.ValidateEmail(value)
	case "line_id":
		isValid = stringsutil.ValidateLineID(value)
	default:
		panic(httperr.NewErrUnprocessableEntity(
			"invalid_medium",
			"unsupported medium", nil,
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
