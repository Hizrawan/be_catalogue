package public

import (
	"net/http"

	"be20250107/internal/responses"

	"github.com/lestrrat-go/jwx/v2/jwk"

	"be20250107/internal/app"
	"be20250107/internal/controllers"
)

type KeysController struct {
	controllers.Controller
}

func NewKeysController(app *app.Registry) *KeysController {
	return &KeysController{controllers.Controller{App: app}}
}

func (c *KeysController) Keys(w http.ResponseWriter, r *http.Request) {
	publicKey := c.App.VerifyKey
	err := responses.JSON(w, 200, struct {
		Keys []jwk.Key `json:"keys"`
	}{
		Keys: []jwk.Key{publicKey},
	})
	if err != nil {
		panic(err)
	}
}
