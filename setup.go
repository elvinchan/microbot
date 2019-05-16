package microbot

import (
	"database/sql"
	"errors"

	"github.com/pangpanglabs/microbot/db"
)

var dialects []db.Dialect

func RegisterDB(d *sql.DB, dbType, dbName string) (db.Dialect, error) {
	if d == nil {
		return nil, errors.New("microbot: nil DB")
	}
	dialect := db.QueryDialect(db.DBType(dbType))
	if dialect == nil {
		return nil, errors.New("microbot: Unsupported DBType")
	}
	dialect.Init(d, db.DBType(dbType), dbName)
	dialects = append(dialects, dialect)
	return dialect, nil
}
