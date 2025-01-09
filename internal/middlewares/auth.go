package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"be20250107/internal/app"
	httperr "be20250107/internal/errors"
	"be20250107/internal/models"
	"be20250107/utils/database"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AuthInformation struct {
	tokenID     string
	userID      string
	accountType string
	user        models.JWTAuthenticatable
	db          database.Queryer
	token       jwt.Token
}

func (auth *AuthInformation) IsLoggedIn() bool {
	if strings.TrimSpace(auth.userID) == "" {
		return false
	}
	if t, canExpire := auth.Expiration(); canExpire && t.Before(time.Now()) {
		return false
	}
	return true
}

func (auth *AuthInformation) TokenID() string {
	return auth.tokenID
}

func (auth *AuthInformation) UserID() string {
	return auth.userID
}

func (auth *AuthInformation) AccountType() string {
	return auth.accountType
}

func (auth *AuthInformation) Expiration() (time.Time, bool) {
	if auth.Token() == nil {
		return time.Time{}, true
	}
	if _, exist := auth.Token().Get(jwt.ExpirationKey); !exist {
		return time.Time{}, false
	}
	return auth.token.Expiration(), true
}

func (auth *AuthInformation) Token() jwt.Token {
	return auth.token
}

func (auth *AuthInformation) User() (any, error) {
	if auth.user == nil {
		err := auth.RefreshUser()
		if err != nil {
			return nil, err
		}
	}
	return auth.user, nil
}

func (auth *AuthInformation) RefreshUser() error {
	if auth.accountType == models.AccountTypeAdmin {
		user := models.Admin{}
		err := auth.db.Get(&user, "SELECT * FROM admins WHERE id=?", auth.userID)
		if err != nil {
			return err
		}
		auth.user = &user
	} else {
		return fmt.Errorf("account type must only be models.Admin, models.System")
	}

	return nil
}

func AuthMiddleware(app *app.Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			opts := []jwt.ParseOption{
				jwt.WithHeaderKey("Authorization"),
				jwt.WithFormKey("x_access_token"),
				jwt.WithKey(jwa.RS256, app.VerifyKey),
			}

			t, err := jwt.ParseRequest(r, opts...)
			if err != nil {
				//	Add log to the access log
				panic(httperr.ErrUnauthenticated)
			}

			if revoked, err := app.Cache.Has(fmt.Sprintf("auth:revoked_%s", t.JwtID())); revoked {
				panic(httperr.ErrUnauthenticated)
			} else if err != nil {
				panic(err)
			}

			act, _ := t.Get("act")
			accountType := ""
			if val, ok := act.(string); ok {
				accountType = val
			}

			auth := AuthInformation{
				tokenID:     t.JwtID(),
				userID:      t.Subject(),
				accountType: accountType,
				db:          app.DB,
				token:       t,
			}
			ctx := context.WithValue(r.Context(), ContextAuth, &auth)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
