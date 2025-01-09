package database

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type Queryer interface {
	sqlx.Ext
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	Select(dest interface{}, query string, args ...interface{}) error
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	MustBegin() *sqlx.Tx
}

type TxQueryer interface {
	Get(dest interface{}, query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) *sql.Row
	Select(dest interface{}, query string, args ...interface{}) error
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type LimitOffsetCursor struct {
	Limit  int
	Offset int
}

type IDTimestampCursor struct {
	CreatedAt time.Time
	ID        string
	Limit     int
	Ascending bool
}
 
func (c *LimitOffsetCursor) Apply(query string, args []any) (string, []any) {
	if c.Limit != 0 {
		query += " LIMIT ?"
		args = append(args, c.Limit)
	}

	if c.Offset != 0 {
		query += " FFSET ?"
		args = append(args, c.Offset)
	}

	return query, args
}

const UUIDMaxValue = "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzz"
const UUIDMinValue = "a"

func (c *IDTimestampCursor) Apply(query string, args []any) (string, []any) {
	var newArgs []any
	var newQuery string
	directionOrderBy := "ASC"
	directionWhere := ">"

	if !c.Ascending {
		directionOrderBy = "DESC"
		directionWhere = "<"
	}

	if c.ID == "" {
		if c.Ascending {
			c.ID = UUIDMinValue
		} else {
			c.ID = UUIDMaxValue
		}
	}

	limitQuery := ""
	OrderByQuery := ""
	WhereQuery := ""

	// Splitting by LIMIT
	splitByLimit := strings.SplitN(query, "LIMIT", 2)
	baseQuery := splitByLimit[0]
	limitQuery = fmt.Sprintf("LIMIT %v", c.Limit)

	// Splitting by ORDER BY
	splitByOrderBy := strings.SplitN(baseQuery, "ORDER BY", 2)
	OrderByQuery = fmt.Sprintf("ORDER BY created_at %v, id %v", directionOrderBy, directionOrderBy)

	// Splitting by WHERE
	splitByWhere := strings.SplitN(splitByOrderBy[0], "WHERE", 2)
	WhereQuery = fmt.Sprintf("WHERE created_at %v ? AND (created_at %v ? OR id %v ?)", directionWhere, directionWhere, directionWhere)
	newArgs = append(newArgs, c.CreatedAt.Unix(), c.CreatedAt.Unix(), c.ID)

	if len(splitByWhere) > 1 {
		WhereQuery += " AND " + splitByWhere[1]
		newArgs = append(newArgs, args...)
	}

	// Combining the components to form the new query
	newQuery = fmt.Sprintf("%v %v %v %v", splitByWhere[0], WhereQuery, OrderByQuery, limitQuery)

	return newQuery, newArgs
}

func DecodeCursor(encodedCursor string) (res time.Time, uuid string, err error) {
	byt, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		err = fmt.Errorf("[DecodeCursor]: %w", err)
		return
	}

	arrStr := strings.Split(string(byt), ",")
	if len(arrStr) != 2 {
		err = errors.New("[DecodeCursor]: cursor is invalid")
		return
	}

	res, err = time.Parse(time.RFC3339Nano, arrStr[0])
	if err != nil {
		return
	}
	uuid = arrStr[1]
	return
}

func EncodeCursor(t time.Time, uuid string) string {
	key := fmt.Sprintf("%s,%s", t.Format(time.RFC3339Nano), uuid)
	return base64.StdEncoding.EncodeToString([]byte(key))
}

// Get ID without prefix table name
// Ex: "stores:ABCDEFG" will return "ABCDEFG"
func GetIDPostfix(id string) string {
	splittedID := strings.SplitN(id, ":", 2)
	if len(splittedID) != 2 {
		return id
	}

	return splittedID[1]
}
