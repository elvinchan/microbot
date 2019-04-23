package microbot

import (
	"database/sql"
	"errors"

	"github.com/pangpanglabs/microbot/db"
)

var dialects []db.Dialect

func RegisterDB(d *sql.DB, dbType db.DBType) error {
	if d == nil {
		return errors.New("microbot: nil DB")
	}
	dialect := db.QueryDialect(dbType)
	if dialect == nil {
		return errors.New("microbot: Unsupported DBType")
	}
	dialect.Init(d, dbType)
	dialects = append(dialects, dialect)
	return nil
}
