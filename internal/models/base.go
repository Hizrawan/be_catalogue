package models

import (
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/oklog/ulid/v2"
)

type Model struct {
	ID        string `json:"id" mapstructure:"id"`
	Timestamp `mapstructure:",squash"`
}

type Timestamp struct {
	CreatedAt int64 `json:"created_at" db:"created_at" mapstructure:"created_at"`
	UpdatedAt int64 `json:"updated_at" db:"updated_at" mapstructure:"updated_at"`
}

//type SoftDelete struct {
//	DeletedAt soft_delete.DeletedAt `json:"deleted_at"`
//}

type JWTAuthenticatable interface {
	IssueAccessToken(exp time.Time) (jwt.Token, error)
}

func (m *Model) BeforeInsert(table string) {
	id := ulid.Make()
	now := time.Now().Unix()

	m.ID = table + ":" + id.String()
	m.CreatedAt = now
	m.UpdatedAt = now
}

func (m *Model) BeforeUpdate() {
	m.UpdatedAt = time.Now().Unix()
}
