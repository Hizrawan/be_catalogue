package middlewares

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"be20250107/internal/app"
	httperr "be20250107/internal/errors"
	"be20250107/internal/models"
	"be20250107/internal/modules/authentication"
	"be20250107/utils/database"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AdminAuthInformation struct {
	tokenID     string
	userID      string
	accountType string
	user        *models.Admin
	db          database.Queryer
	token       jwt.Token
}

func (auth *AdminAuthInformation) IsLoggedIn() bool {
	if strings.TrimSpace(auth.userID) == "" {
		return false
	}
	if t, canExpire := auth.Expiration(); canExpire && t.Before(time.Now()) {
		return false
	}
	return true
}

func (auth *AdminAuthInformation) TokenID() string {
	return auth.tokenID
}

func (auth *AdminAuthInformation) UserID() string {
	return auth.userID
}

func (auth *AdminAuthInformation) AccountType() string {
	return auth.accountType
}

func (auth *AdminAuthInformation) Expiration() (time.Time, bool) {
	if auth.Token() == nil {
		return time.Time{}, true
	}
	if _, exist := auth.Token().Get(jwt.ExpirationKey); !exist {
		return time.Time{}, false
	}
	return auth.token.Expiration(), true
}

func (auth *AdminAuthInformation) Token() jwt.Token {
	return auth.token
}

func (auth *AdminAuthInformation) User() (any, error) {
	if auth.user == nil {
		err := auth.RefreshUser()
		if err != nil {
			return nil, err
		}
	}
	return auth.user, nil
}

func (auth *AdminAuthInformation) RefreshUser() error {
	user := models.Admin{}
	err := auth.db.Get(&user, "SELECT * FROM admins WHERE id=?", auth.userID)
	if err != nil {
		return err
	}
	auth.user = &user
	return nil
}

func AdminAuthMiddleware(app *app.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t, err := app.Auth.VerifyRequest(r, models.AccountTypeAdmin)
			if errors.Is(authentication.ErrIncorrectAccountType, err) {
				panic(httperr.ErrForbidden)
			} else if errors.Is(authentication.ErrTokenRevoked, err) || errors.Is(authentication.ErrUnprocessableRequest, err) {
				panic(httperr.ErrUnauthenticated)
			} else if err != nil {
				panic(err)
			}

			auth := AdminAuthInformation{
				tokenID:     t.JwtID(),
				userID:      t.Subject(),
				accountType: models.AccountTypeAdmin,
				db:          app.DB,
				token:       t,
			}
			ctx := context.WithValue(r.Context(), ContextAuth, &auth)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
