package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"be20250107/internal/app"
	"be20250107/internal/models"
	"be20250107/internal/reqdata"
	"be20250107/internal/responses"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"gopkg.in/guregu/null.v4"
)

type LoginByXinchuanAuthRequest struct {
	Code string `json:"code"`
}

func (r LoginByXinchuanAuthRequest) Authorized(_ *reqdata.Context) bool {
	return true
}

func (r LoginByXinchuanAuthRequest) Validate(_ *reqdata.Context) error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Code, validation.Required))
}

type AuthTokenContext struct {
	AuthProvider       string
	ClientID           string
	DeviceID           string
	Fingerprint        string
	RequestFingerprint RequestFingerprint
}

type LoginContext struct {
	AuthProvider       string             `json:"auth_provider"`
	Fingerprint        string             `json:"fingerprint"`
	RequestFingerprint RequestFingerprint `json:"request_fingerprint"`
}

type RequestFingerprint struct {
	UserAgent    string `json:"user_agent"`
	Forwarded    string `json:"forward"`
	ForwardedFor string `json:"forwarded_for"`
	RealIP       string `json:"real_ip"`
}

func GetRequestFingerprint(r *http.Request) RequestFingerprint {
	return RequestFingerprint{
		UserAgent:    r.UserAgent(),
		Forwarded:    r.Header.Get("Forwarded"),
		ForwardedFor: r.Header.Get("X-Forwarded-For"),
		RealIP:       r.Header.Get("X-Real-IP"),
	}
}

type LoginLog struct {
	LoginContext
	ClientID          string         `json:"client_id"`
	DeviceID          string         `json:"device_id"`
	UserAccessTokenID string         `json:"user_access_token_id"`
	AdditionalInfo    map[string]any `json:"additional_info"`
}

func GenerateAccessToken(app *app.Registry, user models.JWTAuthenticatable, d time.Duration, authCtx AuthTokenContext, info map[string]any) responses.AuthToken {
	token, err := user.IssueAccessToken(time.Now().Add(d))
	if err != nil {
		panic(err)
	}
	err = token.Set("role", "system")
	if err != nil {
		panic(err)
	}
	sign, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, app.SigningKey))
	if err != nil {
		panic(err)
	}

	tx := app.DB

	go func() {
		expiredAt := null.TimeFrom(time.Now().Add(d))
		additionalInfo := map[string]any{
			"token_id":   token.JwtID(),
			"expiration": expiredAt.Time,
		}
		for k, v := range info {
			additionalInfo[k] = v
		}

		switch u := user.(type) {
		case *models.Admin:
			t := models.AdminAccessToken{
				Model: models.Model{
					ID: token.JwtID(),
				},
				AdminID:   u.ID,
				ExpiredAt: expiredAt,
			}
			err := t.Insert(tx)
			if err != nil {
				panic(fmt.Errorf("failed to save access token: %s", err.Error()))
			}
		case *models.System:
			t := models.SystemAccessToken{
				Model: models.Model{
					ID: token.JwtID(),
				},
				SystemID:  u.ID,
				ExpiredAt: expiredAt,
			}
			err := t.Insert(tx)
			if err != nil {
				panic(fmt.Errorf("failed to save access token: %s", err.Error()))
			}

		default:
			panic(fmt.Errorf("GenerateAccessToken expects user to be*models.Admin or *models.System"))
		}
	}()

	return responses.AuthToken{
		AccessToken: string(sign),
		TokenType:   "Bearer",
		ExpiresIn:   int(d.Seconds()),
	}
}

// func SendBindingCode(app *app.Registry, actor string, name string, medium string, destination string, linkCode string, language string) error {
// 	if actor != "merchant" && actor != "driver" {
// 		return fmt.Errorf("invalid actor")
// 	}

// 	u, err := url.Parse(app.Config.AppURL)
// 	if err != nil {
// 		panic("invalid url")
// 	}

// 	u.Path = path.Join(u.Path, "contact-verification-page", actor, linkCode)
// 	verificationLink := u.String()

// 	switch medium {

// 	// case models.ContactMediumEmail:
// 	// 	if _, err := app.Messaging.Email.SendVerificationEmail(destination, name, verificationLink, language); err != nil {
// 	// 		return err
// 	// 	}
// 	// }
// 	return nil
// }

func GetAdminFromAuth(auth reqdata.AuthInformation) (*models.Admin, error) {
	if auth == nil {
		return nil, errors.New("auth info not found")
	}

	user, err := auth.User()
	if err != nil {
		return nil, err
	}

	admin, ok := user.(*models.Admin)
	if !ok {
		return nil, errors.New("[GetAdminFromAuth]: user unparsable as admin")
	}

	return admin, nil
}

type ReviewSummaryResponse struct {
	AverageRate float64
	ReviewCount map[int64]int64
}

type PaginationDetail struct {
	NextPageCursor string `json:"next_page_cursor,omitempty"`
	PerPage        int    `json:"per_page"`
	Asc            bool   `json:"asc"`
	HasNext        bool   `json:"has_next"`
}
