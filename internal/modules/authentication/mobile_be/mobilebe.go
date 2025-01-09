package mobilebe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"gopkg.in/guregu/null.v4"
)

type UserInformation struct {
	ID                 string             `json:"id"`
	LegacyID           string             `json:"legacy_id"`
	MemberID           string             `json:"member_id"`
	DisplayName        string             `json:"display_name"`
	ActualName         string             `json:"actual_name"`
	LanguagePreference string             `json:"language_preference"`
	Accounts           []ConnectedAccount `json:"accounts"`
	Contacts           []VerifiedContact  `json:"contacts"`
}

type VerifiedContact struct {
	UserID    string    `json:"user_id" db:"user_id"`
	Medium    string    `json:"medium" db:"medium"`
	Value     string    `json:"value" db:"value"`
	Type      string    `json:"type" db:"type"`
	ExpiredAt null.Time `json:"expired_at" db:"expired_at"`
	DeletedAt null.Int  `json:"deleted_at" db:"deleted_at"`
}

type ConnectedAccount struct {
	UserID     string    `json:"user_id" db:"user_id"`
	Provider   string    `json:"provider" db:"provider"`
	ForeignID  string    `json:"foreign_id" db:"foreign_id"`
	DisplayAs  string    `json:"display_as" db:"display_as"`
	PublicData string    `json:"public_data" db:"public_data"`
	RevokedAt  null.Time `json:"revoked_at" db:"revoked_at"`
	DeletedAt  null.Int  `json:"deleted_at" db:"deleted_at"`
}

type Client struct {
	options Options
}

type Options struct {
	BaseURL string
	Secret  string
}

type VerifyAccessResponse struct {
	Ok   bool   `json:"ok"`
	Data string `json:"data"`
}

var (
	ErrInvalidClient       = errors.New("invalid client")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrServerUnavailable   = errors.New("server unavailable")
)

func NewClient(option Options) *Client {
	if option.BaseURL == "" {
		option.BaseURL = "website mobile-be"
	}
	return &Client{options: option}
}

func (c *Client) VerifyAccessKeyAndGetCustomerData(ack string) (user *UserInformation, err error) {
	verifyLink := fmt.Sprintf("%v/system/auth/verify-access-key", c.options.BaseURL)

	payload := struct {
		AccessKey string `json:"access_key"`
		Secret    string `json:"secret"`
	}{
		AccessKey: ack,
		Secret:    c.options.Secret,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	body := bytes.NewBuffer(jsonPayload)

	client := http.Client{}
	request, err := http.NewRequest(http.MethodPost, verifyLink, body)
	if err != nil {
		return nil, err
	}

	r, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	bodyResp, err := io.ReadAll(r.Body)
	if err != nil {
		return user, err
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		switch r.StatusCode {
		case 400, 422:
			return user, ErrUnprocessableEntity
		case 401:
			return user, ErrInvalidClient
		default:
			return user, ErrServerUnavailable
		}
	}

	var resp VerifyAccessResponse

	err = json.Unmarshal(bodyResp, &resp)
	if err != nil {
		return user, err
	}

	fetchJWTKey := fmt.Sprintf("%v/keys", c.options.BaseURL)
	set, err := jwk.Fetch(context.Background(), fetchJWTKey)
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseString(resp.Data, jwt.WithKeySet(set))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired()) {
			return nil, ErrUnprocessableEntity
		}
		return nil, err
	}

	userMapData, err := token.AsMap(context.Background())
	if err != nil {
		return nil, nil
	}

	var tempUser UserInformation
	tempMarshal, err := json.Marshal(userMapData["user"])
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tempMarshal, &tempUser)
	if err != nil {
		return nil, err
	}

	return &tempUser, err
}
