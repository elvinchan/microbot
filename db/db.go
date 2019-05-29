package db

import (
	"database/sql"
	"os"

	"github.com/pangpanglabs/microbot/common"
)

type DBType string

const (
	POSTGRES = "postgres"
	SQLITE   = "sqlite3"
	MYSQL    = "mysql"
	MSSQL    = "mssql"
	ORACLE   = "oracle"
)

const (
	IndexType = iota + 1
	UniqueType
)

type Dialect interface {
	Init(*sql.DB, DBType, string)
	DB() *sql.DB
	DBType() DBType
	Tables() ([]Table, error)
	// not use map in result because we need sequence of Columns
	Columns(tableName string) ([]Column, error)
	Indexes(tableName string) (map[string]*Index, error)
	SetSchema(string)
	SetLogger(logger common.Logger)
	ShowSQL(show ...bool)
	IsShowSQL() bool
}

type Base struct {
	db      *sql.DB
	dbType  DBType
	name    string
	logger  common.Logger
	showSQL bool
}

type Table struct {
	Name    string   `json:"name"`
	Rows    int64    `json:"rows"`             // TODO: Not support SQLite currently (maybe SELECT COUNT(1) ... ?)
	Engine  string   `json:"engine,omitempty"` // Only available for MySQL currently
	Indexes []Index  `json:"indexes"`
	Columns []Column `json:"columns"`
	Comment *string  `json:"comment"` // Not available for SQLite
}

type Column struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Nullable        bool     `json:"nullable"`
	Default         *string  `json:"default"`
	Indexes         []string `json:"indexes"`
	IsPrimaryKey    bool     `json:"isPrimaryKey"`
	IsAutoIncrement bool     `json:"isAutoIncrement"` // Not available for Oracle
	Comment         *string  `json:"comment"`         // Not available for SQLite
}

type Index struct {
	Name     string   `json:"name"`
	IsUnique bool     `json:"isUnique"`
	Columns  []string `json:"columns"`
}

func (b *Base) Init(d *sql.DB, dbType DBType, dbName string) {
	b.db = d
	b.dbType = dbType
	b.name = dbName
	logger := common.NewDefaultLogger(os.Stdout)
	b.SetLogger(logger)
}

func (b *Base) SetSchema(schema string) {}

func (b *Base) SetLogger(logger common.Logger) {
	b.logger = logger
}

func (b *Base) DB() *sql.DB {
	return b.db
}

func (b *Base) DBType() DBType {
	return b.dbType
}

func (b *Base) LogSQL(sql string, args []interface{}) {
	if b.logger != nil && b.IsShowSQL() {
		if len(args) > 0 {
			b.logger.Infof("[SQL] %v %v", sql, args)
		} else {
			b.logger.Infof("[SQL] %v", sql)
		}
	}
}

// ShowSQL
func (b *Base) ShowSQL(show ...bool) {
	if len(show) == 0 {
		b.showSQL = true
		return
	}
	b.showSQL = show[0]
}

// IsShowSQL
func (b *Base) IsShowSQL() bool {
	return b.showSQL
}

func (table *Table) Column(name string) *Column {
	for i, c := range table.Columns {
		if c.Name == name {
			return &table.Columns[i]
		}
	}
	return nil
}
