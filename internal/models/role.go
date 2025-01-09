package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"be20250107/utils/database"
)

const (
	PermissionAdminIndex = "admin::index"
)

type Authorities struct {
	Model
	RoleID       string `db:"role_id" json:"role_id"`
	PermissionID int64  `db:"permission_id" json:"permission_id"`
}

func (a *Authorities) Insert(db database.Queryer) error {
	a.BeforeInsert("authorities")
	q := `
	INSERT INTO authorities
		(id, role_id, permission_id, created_at, updated_at)
		VALUES
		(:id, :role_id, :permission_id, :created_at, :updated_at);
		`
	_, err := db.NamedExec(q, a)
	if err != nil {
		return fmt.Errorf("[a.Insert][NamedExec]%w", err)
	}
	return nil
}

func (a *Authorities) Delete(db database.Queryer) error {
	q := `
	DELETE FROM authorities
	WHERE 
		role_id = :role_id AND 
		permission_id = :permission_id
	`
	_, err := db.NamedExec(q, a)
	if err != nil {
		return fmt.Errorf("[a.Delete][NamedExec]%w", err)
	}
	return nil
}

func GetRoles(db database.Queryer, sortBy string, isAscending bool, filter Role) ([]Role, bool, error) {
	var roles []Role

	query := "SELECT * FROM roles"
	extraFilter := []string{}
	args := []any{}

	if filter.Name != "" {
		extraFilter = append(extraFilter, "name LIKE ?")
		args = append(args, "%"+filter.Name+"%")
	}

	if len(extraFilter) > 0 {
		query += " WHERE " + strings.Join(extraFilter, " AND ")
	}

	sortColumn := "created_at"
	if sortBy != "" {
		sortColumn = sortBy
	}

	sortDir := "DESC"
	if isAscending {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %v %v", sortColumn, sortDir)
	err := db.Select(&roles, query, args...)
	if err != nil {
		return nil, false, fmt.Errorf("[GetRoles][Select]%w", err)
	} else if len(roles) == 0 {
		return nil, false, nil
	}

	return roles, true, nil
}

type Permission struct {
	ID          int64 `db:"id" json:"id" mapstructure:"id"`
	Timestamp   `mapstructure:",squash"`
	Identifier  string `db:"identifier" json:"identifier"`
	Module      string `db:"module" json:"module"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

func (p *Permission) Insert(db database.Queryer) error {
	p.CreatedAt = time.Now().Unix()
	p.UpdatedAt = time.Now().Unix()

	q := `
	INSERT INTO permissions
		(id, identifier, module, name, description, created_at, updated_at)
		VALUES
		(:id, :identifier, :module, :name, :description, :created_at, :updated_at);
		`
	res, err := db.NamedExec(q, p)
	if err != nil {
		return fmt.Errorf("[p.Insert][NamedExec]%w", err)
	}

	id, _ := res.LastInsertId()
	p.ID = id

	return nil
}

func GetAllPermissions(db database.Queryer) ([]Permission, bool, error) {
	var permissions []Permission
	query := "SELECT * FROM permissions"
	err := db.Select(&permissions, query)
	if err != nil {
		return nil, false, fmt.Errorf("[GetAllPermissions][Select]%w", err)
	} else if len(permissions) == 0 {
		return nil, false, nil
	}

	return permissions, true, nil
}

type Role struct {
	Model       `mapstructure:",squash"`
	Name        string       `db:"name" json:"name" mapstructure:"name"`
	Permissions []Permission `json:"permissions,omitempty" mapstructure:"-"`
}

func (r *Role) Insert(db database.Queryer) error {
	r.BeforeInsert("roles")

	q := `
		INSERT INTO roles
		(id, name, created_at, updated_at)
		VALUES
		(:id, :name, :created_at, :updated_at)
	`
	_, err := db.NamedExec(q, r)
	if err != nil {
		return fmt.Errorf("[r.Insert][NamedExec]%w", err)
	}
	return nil
}

func (r *Role) Update(db database.Queryer) error {
	r.BeforeUpdate()
	q := `
		UPDATE roles SET
			name = :name,
			updated_at = :updated_at
		WHERE id = :id
	`

	_, err := db.NamedExec(q, r)
	if err != nil {
		return fmt.Errorf("[r.Update][NamedExec]%w", err)
	}
	return nil
}

func (r *Role) Delete(db database.Queryer) error {
	r.BeforeUpdate()

	q := `DELETE FROM roles WHERE id=:id`
	_, err := db.NamedExec(q, r)
	if err != nil {
		return fmt.Errorf("[r.Delete][NamedExec]%w", err)
	}
	return nil
}

func GetRoleByID(db database.Queryer, id string) (*Role, bool, error) {
	var role Role
	err := db.Get(&role, "SELECT * FROM roles WHERE id = ?", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return &role, true, nil
}

func (r *Role) LoadPermissions(db database.Queryer) error {
	r.Permissions = nil

	var pivots []Authorities
	err := db.Select(&pivots, "SELECT * FROM authorities WHERE role_id = ?", r.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return fmt.Errorf("[r.LoadPermissions][Select]%w", err)
	}

	for _, pivot := range pivots {
		var permission Permission
		err := db.Get(&permission, "SELECT * FROM permissions WHERE id = ?", pivot.PermissionID)
		if err != nil {
			return fmt.Errorf("[r.LoadPermissions][Get]%w", err)
		}

		r.Permissions = append(r.Permissions, permission)
	}

	return nil
}

func (r *Role) AssignPermission(db database.Queryer, permissionID int64) error {
	authorities := Authorities{
		RoleID:       r.ID,
		PermissionID: permissionID,
	}

	return authorities.Insert(db)
}

func (r *Role) RemovePermission(db database.Queryer, permissionID int64) error {
	authorities := Authorities{
		RoleID:       r.ID,
		PermissionID: permissionID,
	}

	return authorities.Delete(db)
}

func (r *Role) AdjustAssignedPermissions(db database.Queryer, newPermissions []int64) error {
	oldPermissions := make(map[int64]bool)
	if err := r.LoadPermissions(db); err != nil {
		return fmt.Errorf("[r.AdjustAssignedPermissions]%w", err)
	}

	for _, permission := range r.Permissions {
		oldPermissions[permission.ID] = true
	}

	newPermissionsMap := make(map[int64]bool)
	for _, permission := range newPermissions {
		newPermissionsMap[permission] = true
	}

	// if newPermissions not exist on oldPermissions
	// means that we need to assign a new permission
	for _, permission := range newPermissions {
		if oldPermissions[permission] {
			continue
		}

		err := r.AssignPermission(db, permission)
		if err != nil {
			return fmt.Errorf("[r.AdjustAssignedPermissions]%w", err)
		}
	}

	// if oldPermissions not exist on newPermissions
	// means that we need to remove the old permission
	for _, permission := range r.Permissions {
		if newPermissionsMap[permission.ID] {
			continue
		}

		err := r.RemovePermission(db, permission.ID)
		if err != nil {
			return fmt.Errorf("[r.AdjustAssignedPermissions]%w", err)
		}
	}

	if err := r.LoadPermissions(db); err != nil {
		return (err)
	}

	return nil
}
