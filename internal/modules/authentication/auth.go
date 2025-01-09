package authentication

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"gopkg.in/guregu/null.v4"

	"be20250107/internal/models"
	mobilebe "be20250107/internal/modules/authentication/mobile_be"
	"be20250107/internal/modules/authentication/xinchuanauth"
	"be20250107/internal/modules/cache"
	"be20250107/utils/database"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pkg/errors"
)

type Auth struct {
	XinchuanAuth *xinchuanauth.Client
	MobileBEAuth *mobilebe.Client
	cache        cache.Cache
	db           database.Queryer
	verifyKey    jwk.RSAPublicKey
}

var (
	ErrInvalidTokenString   = errors.New("failed to parse token string or verify its signature")
	ErrUnprocessableRequest = errors.New("failed to parse request or verify its signature")
	ErrTokenRevoked         = errors.New("access token has been revoked")
	ErrIncorrectAccountType = errors.New("token is created for other account type")
)

func (a *Auth) Init(db database.Queryer, cache cache.Cache, verifyKey jwk.RSAPublicKey) *Auth {
	a.db = db
	a.cache = cache
	a.verifyKey = verifyKey
	return a
}

type tokenToRevoke struct {
	ID        string    `db:"id"`
	ExpiredAt null.Time `db:"expired_at"`
}

func (a *Auth) GetTokenInfo(r *http.Request) (jwt.Token, error) {
	opts := []jwt.ParseOption{
		jwt.WithHeaderKey("Authorization"),
		jwt.WithFormKey("x_access_token"),
		jwt.WithKey(jwa.RS256, a.verifyKey),
	}

	t, err := jwt.ParseRequest(r, opts...)
	if err != nil {
		return nil, ErrUnprocessableRequest
	}

	return t, nil
}

func (a *Auth) LoadRevocationList() error {
	tables := []string{"system_access_tokens", "admin_access_tokens"}
	for _, table := range tables {
		q := fmt.Sprintf(`
			SELECT id, expired_at 
			FROM %s 
			WHERE
				revoked_at IS NOT NULL AND 
				(expired_at IS NULL OR expired_at > NOW());
		`, table)
		rows, err := a.db.Queryx(q)
		if errors.Is(sql.ErrNoRows, err) {
			continue
		} else if err != nil {
			return err
		}
		for rows.Next() {
			var token tokenToRevoke
			if err = rows.StructScan(&token); err != nil {
				return err
			}
			opt := cache.Options{}
			if token.ExpiredAt.Valid {
				opt.Expiration = time.Until(token.ExpiredAt.Time) + time.Hour
			}
			err := a.cache.PutValue(fmt.Sprintf("auth:revoked_%s", token.ID), true, &opt)
			if err != nil {
				panic(err)
			}
		}
	}
	return nil
}

func (a *Auth) Verify(tokenStr string, accountType string) (jwt.Token, error) {
	opts := []jwt.ParseOption{
		jwt.WithKey(jwa.RS256, a.verifyKey),
	}

	t, err := jwt.Parse([]byte(tokenStr), opts...)
	if err != nil {
		return nil, ErrInvalidTokenString
	}
	return a.verifyTokenContent(t, accountType)
}

func (a *Auth) VerifyRequest(r *http.Request, accountType ...string) (jwt.Token, error) {
	t, err := a.GetTokenInfo(r)
	if err != nil {
		return nil, err
	}

	return a.verifyTokenContent(t, accountType...)
}

func (a *Auth) verifyTokenContent(token jwt.Token, accountTypes ...string) (jwt.Token, error) {
	if revoked, err := a.cache.Has(fmt.Sprintf("auth:revoked_%s", token.JwtID())); revoked {
		return nil, ErrTokenRevoked
	} else if err != nil {
		return nil, err
	}

	act, _ := token.Get("act")
	actVal := ""
	if val, ok := act.(string); ok {
		actVal = val
	}

	for _, accType := range accountTypes {
		if actVal == accType {
			return token, nil
		}
	}

	return nil, ErrIncorrectAccountType
}

func (a *Auth) Revoke(token any) error {
	var tokenID string
	var expiration null.Time
	assertionError := errors.New("Revoke expects  models.SystemAccessToken, models.AdminAccessToken or reference to those instances")

	switch t := token.(type) {
	case models.AdminAccessToken:
		token = &t
	case models.SystemAccessToken:
		token = &t
	default:
		return assertionError
	}

	now := null.TimeFrom(time.Now())

	switch t := token.(type) {

	case *models.AdminAccessToken:
		t.RevokedAt = now
		if err := t.Update(a.db); err != nil {
			return err
		}
		tokenID = t.ID
		expiration = t.ExpiredAt
	case *models.SystemAccessToken:
		t.RevokedAt = now
		if err := t.Update(a.db); err != nil {
			return err
		}
		tokenID = t.ID
		expiration = t.ExpiredAt

	default:
		return assertionError
	}

	opt := cache.Options{}
	if expiration.Valid {
		opt.Expiration = time.Until(expiration.Time) + time.Hour
	}
	return a.cache.PutValue(fmt.Sprintf("auth:revoked_%s", tokenID), true, &opt)
}
