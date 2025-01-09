package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"be20250107/internal/constants"
	"be20250107/utils/database"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/oklog/ulid/v2"
	"gopkg.in/guregu/null.v4"
)

type Admin struct {
	Model         `mapstructure:",squash"`
	Name          string      `json:"name" db:"name" mapstructure:"name"`
	Username      string      `json:"username" db:"username" mapstructure:"username"`
	Provider      string      `json:"provider" db:"provider" mapstructure:"provider"`
	RoleID        null.String `json:"role_id" db:"role_id" mapstructure:"role_id"`
	Role          *Role       `json:"role,omitempty" mapstructure:"role"`
	ProviderID    string      `json:"provider_id" db:"provider_id" mapstructure:"provider_id"`
	DeactivatedAt null.Time   `json:"deactivated_at" db:"deactivated_at" mapstructure:"deactivated_at"`
}

func (a *Admin) Insert(db database.Queryer) error {
	a.BeforeInsert("admins")

	q := `
		INSERT INTO admins
		(id, name, username, provider, provider_id, deactivated_at, created_at, updated_at)
		VALUES
		(:id, :name, :username, :provider, :provider_id, :deactivated_at, :created_at, :updated_at)
	`
	_, err := db.NamedExec(q, a)
	if err != nil {
		return fmt.Errorf("[a.Insert][NamedExec]%w", err)
	}
	return nil
}

func (a *Admin) Update(db database.Queryer) error {
	a.BeforeUpdate()
	q := `
		UPDATE admins SET
			name = :name,
			username = :username,
			role_id = :role_id,
			updated_at = :updated_at,
			deactivated_at = :deactivated_at
		WHERE id = :id
	`

	_, err := db.NamedExec(q, a)
	if err != nil {
		return fmt.Errorf("[a.Update][NamedExec]%w", err)
	}
	return nil
}

func (a *Admin) ChangeRole(db database.Queryer, r *Role) error {
	a.RoleID = null.StringFrom(r.ID)
	err := a.Update(db)
	if err != nil {
		return fmt.Errorf("[a.ChangeRole]%w", err)
	}
	return nil
}

func (a *Admin) LoadRole(db database.Queryer, withPermission bool) error {
	if a.RoleID.ValueOrZero() == "" {
		return nil
	}

	role, exist, err := GetRoleByID(db, a.RoleID.ValueOrZero())
	if err != nil {
		return fmt.Errorf("[a.LoadRole]%w", err)
	}

	if !exist {
		return nil
	}

	a.Role = role

	if withPermission {
		err := a.Role.LoadPermissions(db)
		if err != nil {
			return fmt.Errorf("[a.LoadRole]%w", err)
		}

	}
	return nil
}

func (a *Admin) IssueAccessToken(exp time.Time) (jwt.Token, error) {
	id := ulid.Make()

	token, err := jwt.
		NewBuilder().
		Issuer(constants.TokenIssuer).
		Expiration(exp).
		IssuedAt(time.Now()).
		JwtID("admin_access_tokens:"+id.String()).
		Subject(a.ID).
		Claim("act", "admin").
		Build()
	if err != nil {
		return nil, fmt.Errorf("[a.IssueAccessToken][Build]%w", err)
	}

	return token, nil
}

type AdminAccessToken struct {
	Model
	AdminID   string    `json:"admin_id" db:"admin_id"`
	RevokedAt null.Time `json:"revoked_at" db:"revoked_at"`
	ExpiredAt null.Time `json:"expired_at" db:"expired_at"`
}

func (aat *AdminAccessToken) Insert(db database.Queryer) error {
	now := time.Now().Unix()
	aat.Model.CreatedAt = now
	aat.Model.UpdatedAt = now

	q := "INSERT INTO admin_access_tokens (id,admin_id,expired_at,revoked_at,created_at,updated_at) " +
		"VALUES (:id,:admin_id,:expired_at,:revoked_at,:created_at,:updated_at)"
	_, err := db.NamedExec(q, aat)
	if err != nil {
		return fmt.Errorf("[aat.Insert][NamedExec]%w", err)
	}
	return nil
}

func (aat *AdminAccessToken) Update(db database.Queryer) error {
	aat.Model.UpdatedAt = time.Now().Unix()

	q := "UPDATE admin_access_tokens" +
		"SET admin_id = :admin_id," +
		"expired_at = :expired_at," +
		"revoked_at = :revoked_at," +
		"created_at = :created_at," +
		"updated_at = :updated_at" +
		" WHERE id = :id;"
	_, err := db.NamedExec(q, aat)
	if err != nil {
		return fmt.Errorf("[aat.Update][NamedExec]%w", err)
	}
	return nil
}
func GetAdminAccessTokenByID(db database.Queryer, id string) (*AdminAccessToken, bool, error) {
	var mat AdminAccessToken

	err := db.Get(&mat, "SELECT * FROM admin_access_tokens WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("[GetAdminAccessTokenByID][Get]%w", err)
	}

	return &mat, true, nil
}

func GetAdminBatched(db database.Queryer, lastID string, lastCreatedAt int64, limit int) ([]Admin, error) {
	q := `
	SELECT * FROM admins 
	WHERE 
		created_at <= ? AND
        (
            created_at < ? OR id < ?
        )
        
	ORDER BY 
		created_at DESC, id DESC
	LIMIT ?
	`

	var admins []Admin
	err := db.Select(&admins, q, lastCreatedAt, lastCreatedAt, lastID, limit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("[GetAdminBatched][Select]%w", err)
	}

	return admins, nil
}

func GetAdminByID(db database.Queryer, id string) (*Admin, bool, error) {
	var admin Admin
	err := db.Get(&admin, "SELECT * FROM admins WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("[GetAdminByID][Get]%w", err)
	}
	return &admin, true, nil
}

func GetAdminsByRoleID(db database.Queryer, role_id string) ([]Admin, bool, error) {
	var admins []Admin
	err := db.Select(&admins, "SELECT * FROM admins WHERE role_id = ?", role_id)
	if err != nil {
		return nil, false, fmt.Errorf("[GetAdminsByRoleID][Select]%w", err)
	} else if len(admins) == 0 {
		return admins, false, nil
	}
	return admins, true, nil
}

func GetAdminByProviderAndProviderID(db database.Queryer, providerID int, provider string) (*Admin, bool, error) {
	var admin Admin
	err := db.Get(&admin, "SELECT * FROM admins WHERE provider_id = ? AND provider = ?", providerID, provider)
	if err == sql.ErrNoRows {
		return nil, false, nil
	} else if err != nil {
		return nil, false, fmt.Errorf("[GetAdminByProviderAndProviderID][Get]%w", err)
	}
	return &admin, true, nil
}

func CountExistingAdminUsername(db database.Queryer, username string) (int, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM admins WHERE username = ?", username)
	if err != nil {
		return 0, fmt.Errorf("[CountExistingAdminUsername][Get]%w", err)
	}
	return count, nil
}

func (a *Admin) IsAdminAuthorized(db database.Queryer, identifierPermission string) bool {
	if a.RoleID.Valid {
		err := a.LoadRole(db, true)
		if err != nil {
			return false
		}

		for _, permission := range a.Role.Permissions {
			if permission.Identifier == identifierPermission {
				return true
			}
		}

	}
	return false
}
