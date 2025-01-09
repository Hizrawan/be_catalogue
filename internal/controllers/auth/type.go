package auth

import (
	"database/sql"
	"errors"
	"fmt"

	"be20250107/internal/models"
	"be20250107/internal/reqdata"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type RequestLoginByCodeRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Fingerprint  string `json:"fingerprint"`
	Info         string `json:"info"`
	Medium       string `json:"medium"`
	Credential   string `json:"credential"`
}

func (r *RequestLoginByCodeRequest) Authorized(ctx *reqdata.Context) bool {
	_, err := models.GetUserClient(ctx.App.DB, r.ClientID, r.ClientSecret)
	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			panic(err)
		}
		return false
	}
	return true
}

func (r *RequestLoginByCodeRequest) Validate(_ *reqdata.Context) error {
	return validation.ValidateStruct(r,
		validation.Field(&r.ClientID, validation.Required),
		validation.Field(&r.ClientSecret, validation.Required),

		validation.Field(&r.Credential, validation.Required, validation.By(func(value interface{}) error {

			return nil
		})),
	)
}

type LoginByCodeRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Fingerprint  string `json:"fingerprint"`
	Info         string `json:"info"`
	Medium       string `json:"medium"`
	Credential   string `json:"credential"`
	Code         string `json:"code"`
}

func (r LoginByCodeRequest) Authorized(ctx *reqdata.Context) bool {
	_, err := models.GetUserClient(ctx.App.DB, r.ClientID, r.ClientSecret)
	if err != nil {
		if !errors.Is(sql.ErrNoRows, err) {
			panic(err)
		}
		return false
	}
	return true
}

func (r LoginByCodeRequest) Validate(ctx *reqdata.Context) error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ClientID, validation.Required),
		validation.Field(&r.ClientSecret, validation.Required),

		validation.Field(&r.Credential, validation.Required),
		validation.Field(&r.Code, validation.Required, validation.By(func(value interface{}) error {
			if whitelistedAccounts[r.Credential] != nil && whitelistedAccounts[r.Credential].OTP == r.Code {
				return nil
			}
			key := fmt.Sprintf("auth:code_%s_%s:%s", r.Medium, r.Credential, value.(string))
			if exist, err := ctx.App.Cache.Has(key); err != nil {
				return err
			} else if !exist {
				return validation.Errors{"code": fmt.Errorf("invalid code")}
			}
			return nil
		})),
	)
}
