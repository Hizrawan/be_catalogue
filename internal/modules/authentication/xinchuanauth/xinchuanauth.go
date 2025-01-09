package xinchuanauth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	options Options
}

type Options struct {
	BaseURL      string
	RedirectURI  string
	ClientID     int
	ClientSecret string
}

type Token struct {
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Account struct {
	ID            int        `json:"id"`
	RoleID        *int       `json:"role_id"`
	Name          string     `json:"name"`
	Username      string     `json:"username"`
	Status        string     `json:"status"`
	ActivatedAt   *time.Time `json:"activated_at"`
	DeactivatedAt *time.Time `json:"deactivated_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

var (
	ErrInvalidClient       = errors.New("invalid client")
	ErrUnprocessableEntity = errors.New("unprocessable entity")
	ErrServerUnavailable   = errors.New("server unavailable")
)

func NewClient(option Options) *Client {
	if option.BaseURL == "" {
		option.BaseURL = "https://auth.xinchuan.tw/"
	}
	return &Client{options: option}
}

func (c *Client) Authenticate(code string) (*Token, error) {
	u, err := url.Parse(c.options.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/oauth/token"

	payload, err := json.Marshal(struct {
		GrantType    string `json:"grant_type"`
		ClientID     int    `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		Scope        string `json:"scope"`
		RedirectURI  string `json:"redirect_uri"`
		Code         string `json:"code"`
	}{
		GrantType:    "authorization_code",
		ClientID:     c.options.ClientID,
		ClientSecret: c.options.ClientSecret,
		Scope:        "*",
		RedirectURI:  c.options.RedirectURI,
		Code:         code,
	})
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		switch resp.StatusCode {
		case 400, 422:
			return nil, ErrUnprocessableEntity
		case 401:
			return nil, ErrInvalidClient
		default:
			return nil, ErrServerUnavailable
		}
	}

	var token Token
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func parseTime(t string) *time.Time {
	if t == "" {
		return nil
	}

	var c time.Time
	c, err := time.Parse(time.RFC3339, t)
	if err == nil {
		return &c
	}
	oldLaravelFormat := "2006-01-02 15:04:05"
	c, err = time.Parse(oldLaravelFormat, t)
	if err == nil {
		return &c
	}
	return nil
}

func (c *Client) FetchAccount(token *Token) (*Account, error) {
	u, err := url.Parse(c.options.BaseURL)
	if err != nil {
		return nil, err
	}
	u.Path = "/auth/user"

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{
		"Authorization": {fmt.Sprintf("Bearer %s", token.AccessToken)},
		"Accept":        {"application/json"},
	}

	hc := http.Client{}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, ErrServerUnavailable
	}

	var rawResponse struct {
		ID            int     `json:"id"`
		RoleID        *int    `json:"role_id"`
		Name          string  `json:"name"`
		Username      string  `json:"username"`
		Status        string  `json:"status"`
		ActivatedAt   *string `json:"activated_at"`
		DeactivatedAt *string `json:"deactivated_at"`
		CreatedAt     string  `json:"created_at"`
		UpdatedAt     string  `json:"updated_at"`
	}
	err = json.NewDecoder(resp.Body).Decode(&rawResponse)
	if err != nil {
		return nil, err
	}

	account := Account{
		ID:       rawResponse.ID,
		RoleID:   rawResponse.RoleID,
		Name:     rawResponse.Name,
		Username: rawResponse.Username,
		Status:   rawResponse.Status,
	}
	if rawResponse.ActivatedAt != nil {
		account.ActivatedAt = parseTime(*rawResponse.ActivatedAt)
	}
	if rawResponse.DeactivatedAt != nil {
		account.DeactivatedAt = parseTime(*rawResponse.DeactivatedAt)
	}
	if rawResponse.CreatedAt != "" {
		account.CreatedAt = *parseTime(rawResponse.CreatedAt)
	}
	if rawResponse.UpdatedAt != "" {
		account.UpdatedAt = *parseTime(rawResponse.UpdatedAt)
	}
	return &account, nil
}
