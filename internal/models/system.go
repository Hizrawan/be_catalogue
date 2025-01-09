package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"be20250107/internal/constants"
	"be20250107/utils/database"

	"be20250107/utils/random"

	"gopkg.in/guregu/null.v4"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/oklog/ulid/v2"
)

type System struct {
	Model
	Name           string      `json:"name"`
	URL            string      `json:"url"`
	SecretKey      null.String `json:"secret" db:"secret_key"`
	PlainSecretKey null.String `json:"-" db:"-"`
}

func (s *System) Insert(db database.Queryer) error {
	id := ulid.Make()
	s.ID = "systems:" + id.String()

	now := time.Now().Unix()
	s.Model.CreatedAt = now
	s.Model.UpdatedAt = now

	if s.PlainSecretKey.Valid {
		h := sha256.New()
		h.Write([]byte(s.PlainSecretKey.ValueOrZero()))
		hashed := hex.EncodeToString(h.Sum(nil))
		s.SecretKey = null.StringFrom(hashed)
		s.PlainSecretKey = null.String{}
	}

	q := "INSERT INTO systems (id, name, url, secret_key, created_at, updated_at) " +
		"VALUES (:id, :name, :url, :secret_key, :created_at, :updated_at)"
	_, err := db.NamedExec(q, s)
	if err != nil {
		return fmt.Errorf("[s.Insert][NamedExec]%w", err)
	}
	return nil
}

func (s *System) Update(db database.Queryer) error {
	s.UpdatedAt = time.Now().Unix()

	if s.PlainSecretKey.Valid {
		h := sha256.New()
		h.Write([]byte(s.PlainSecretKey.ValueOrZero()))
		hashed := hex.EncodeToString(h.Sum(nil))
		s.SecretKey = null.StringFrom(hashed)
		s.PlainSecretKey = null.String{}
	}

	q := `
		UPDATE systems SET
			name = :name,
			url = :url,
			secret_key = :secret_key,
			created_at = :created_at,
			updated_at = :updated_at
		WHERE id = :id
	`
	_, err := db.NamedExec(q, s)
	if err != nil {
		return fmt.Errorf("[s.Update][NamedExec]%w", err)
	}
	return nil
}

func GenerateSecretKey() string {
	return random.GenerateString(
		64,
		random.UppercaseAlphabeticCharset+random.LowercaseAlphabeticCharset+random.NumericCharset,
	)
}

func (s *System) IssueAccessToken(exp time.Time) (jwt.Token, error) {
	id := ulid.Make()

	token, err := jwt.
		NewBuilder().
		Issuer(constants.TokenIssuer).
		Expiration(exp).
		IssuedAt(time.Now()).
		JwtID("system_access_tokens:"+id.String()).
		Subject(s.ID).
		Claim("act", "system").
		Build()
	if err != nil {
		return nil, fmt.Errorf("[s.IssueAccessToken][Build]%w", err)
	}

	return token, nil
}

type SystemAccessToken struct {
	Model
	SystemID  string    `json:"system_id" db:"system_id"`
	RevokedAt null.Time `json:"revoked_at" db:"revoked_at"`
	ExpiredAt null.Time `json:"expired_at" db:"expired_at"`
}

func (sat *SystemAccessToken) Insert(db database.Queryer) error {
	sat.BeforeInsert("system_access_tokens")

	q := "INSERT INTO system_access_tokens (id, system_id, expired_at, revoked_at, created_at, updated_at) " +
		"VALUES (:id, :system_id, :expired_at, :revoked_at, :created_at, :updated_at)"
	_, err := db.NamedExec(q, sat)
	if err != nil {
		return fmt.Errorf("[sat.Insert][NamedExec]%w", err)
	}
	return nil
}

func (sat *SystemAccessToken) Update(db database.Queryer) error {
	sat.BeforeUpdate()

	q := "UPDATE system_access_tokens " +
		"SET system_id = :system_id," +
		"expired_at = :expired_at," +
		"revoked_at = :revoked_at," +
		"created_at = :created_at," +
		"updated_at = :updated_at" +
		" WHERE id = :id;"
	_, err := db.NamedExec(q, sat)
	if err != nil {
		return fmt.Errorf("[sat.Update][NamedExec]%w", err)
	}
	return nil
}

type UserClient struct {
	Model
	Name      string   `json:"name"`
	Secret    string   `json:"secret"`
	DeletedAt null.Int `json:"deleted_at" db:"deleted_at"`
}

func (uc *UserClient) Insert(db database.Queryer) error {
	uc.BeforeInsert("user_clients")

	q := `
		INSERT INTO user_clients
		(id, name, secret, created_at, updated_at, deleted_at)
		VALUES
		(:id, :name, :secret, :created_at, :updated_at, :deleted_at);
	`

	_, err := db.NamedExec(q, uc)
	if err != nil {
		return fmt.Errorf("[uc.Insert][NamedExec]%w", err)
	}
	return nil
}

func (uc *UserClient) Update(db database.Queryer) error {
	uc.BeforeUpdate()

	q := `
		UPDATE user_clients SET
			name = :name,
			secret = :secret,
			updated_at = :updated_at,
			deleted_at = :deleted_at 
		WHERE id = :id;
	`
	_, err := db.NamedExec(q, uc)
	if err != nil {
		return fmt.Errorf("[uc.Update][NamedExec]%w", err)
	}
	return nil
}

func GetUserClient(db database.Queryer, id string, secret string) (*UserClient, error) {
	var uc UserClient
	err := db.Get(&uc, "SELECT * FROM user_clients WHERE id = ? AND secret = ? AND deleted_at is NULL", id, secret)

	if err != nil {
		return nil, fmt.Errorf("[GetUserClient][Get]%w", err)
	}
	return &uc, nil
}
