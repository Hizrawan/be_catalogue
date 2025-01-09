package reqdata

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
)

// AuthInformation represents the authentication state and data associated with a request
type AuthInformation interface {
	IsLoggedIn() bool
	TokenID() string
	Token() jwt.Token
	UserID() string
	User() (any, error)
	AccountType() string
	Expiration() (time.Time, bool)
}
