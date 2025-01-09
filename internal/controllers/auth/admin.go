package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"be20250107/internal/app"
	"be20250107/internal/controllers"
	"be20250107/internal/middlewares"
	"be20250107/internal/models"
	"be20250107/internal/modules/cache"
	"be20250107/internal/modules/mq"
	"be20250107/internal/responses"

	"github.com/oklog/ulid/v2"
	"gopkg.in/guregu/null.v4"
)

type AuthAdminController struct {
	controllers.Controller
}

func NewAuthAdminController(app *app.Registry) *AuthAdminController {
	return &AuthAdminController{controllers.Controller{App: app}}
}

func (c *AuthAdminController) LoginByXinchuanAuth(w http.ResponseWriter, r *http.Request) {
	var req controllers.LoginByXinchuanAuthRequest
	err := c.Validate(&req, r)
	if err != nil {
		panic(err)
	}

	auth := c.App.Auth.XinchuanAuth
	oToken, err := auth.Authenticate(req.Code)
	if err != nil {
		panic(err)
	}

	ssoUser, err := auth.FetchAccount(oToken)
	if err != nil {
		panic(err)
	}

	admin, exist, err := models.GetAdminByProviderAndProviderID(c.App.DB, ssoUser.ID, "xinchuan-auth")
	if err != nil {
		panic(err)
	}

	if !exist {
		loop := 0
		maxLoop := 5
		username := ssoUser.Username

		for {
			if loop >= maxLoop {
				panic(errors.New("maximum number of tries reached when generating unique username"))
			}
			loop += 1

			count, err := models.CountExistingAdminUsername(c.App.DB, username)
			if err != nil {
				panic(err)
			}

			if count == 0 {
				break
			} else {
				uid := ulid.Make()
				username = ssoUser.Username + "-" + uid.String()[:4]
			}
		}

		admin = &models.Admin{
			Name:          ssoUser.Name,
			Username:      username,
			Provider:      "xinchuan-auth",
			ProviderID:    fmt.Sprintf("%d", ssoUser.ID),
			DeactivatedAt: null.Time{},
		}
		err = admin.Insert(c.App.DB)
		if err != nil {
			panic(err)
		}
	}

	err = mq.PublishMessage(c.App.MessageProducer, mq.AdminUpdatedTopic, mq.AdminUpdatedMsg{
		AdminID: admin.ID,
	})
	if err != nil {
		log.Println("[Admin.LoginByXinchuanAuth] publish updated:", err)
	}

	resp := controllers.GenerateAccessToken(c.App, admin, 24*time.Hour, controllers.AuthTokenContext{
		AuthProvider:       "xinchuan-auth",
		RequestFingerprint: controllers.GetRequestFingerprint(r),
	}, map[string]any{
		"via": "xinchuan-auth",
		"as":  "admin",
	})
	err = responses.JSON(w, 200, resp)
	if err != nil {
		panic(err)
	}
}

func (c *AuthAdminController) Me(w http.ResponseWriter, r *http.Request) {
	auth := c.RequestContext(r).Auth

	if u, err := auth.User(); err != nil {
		panic(err)
	} else if user, ok := u.(*models.Admin); ok {
		if err := user.LoadRole(c.App.DB, true); err != nil {
			panic(err)
		}
		if err := responses.JSON(w, 200, user); err != nil {
			panic(err)
		}
	} else {
		c.Forbidden()
	}
}

func (c *AuthAdminController) Logout(w http.ResponseWriter, r *http.Request) {
	auth := r.Context().Value(middlewares.ContextAuth).(*middlewares.AdminAuthInformation)

	id := auth.TokenID()
	var expiredAt time.Time
	if val, exists := auth.Expiration(); exists {
		expiredAt = val
	}

	token, exist, err := models.GetAdminAccessTokenByID(c.App.DB, id)
	if err != nil {
		panic(err)
	}

	if !exist {
		panic(fmt.Errorf("access token not found"))
	}

	token.RevokedAt = null.TimeFrom(time.Now())
	err = token.Update(c.App.DB)
	if err != nil {
		panic(err)
	}
	id = token.ID
	if token.ExpiredAt.Valid {
		expiredAt = token.ExpiredAt.Time
	}

	opt := cache.Options{}
	if !expiredAt.IsZero() {
		duration := time.Until(expiredAt)
		if duration > 0 {
			opt.Expiration = duration + time.Hour
		}
	}
	err = c.App.Cache.PutValue(fmt.Sprintf("auth:revoked_%s", id), true, &opt)
	if err != nil {
		panic(err)
	}

	err = responses.JSON(w, 200, struct {
		OK bool `json:"ok"`
	}{
		OK: true,
	})
	if err != nil {
		panic(err)
	}

}
