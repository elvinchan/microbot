package db

import (
	"database/sql"
	"os"
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
	SetLogger(logger ILogger)
	DB() *sql.DB
	DBType() DBType
	Tables() ([]Table, error)
	// not use map in result because we need sequence of Columns
	Columns(tableName string) ([]Column, error)
	Indexes(tableName string) (map[string]*Index, error)
}

type Base struct {
	db     *sql.DB
	dbType DBType
	name   string
	logger ILogger
}

type Table struct {
	Name    string   `json:"name"`
	Rows    int64    `json:"rows"`
	Engine  string   `json:"engine,omitempty"` // Only available for MySQL currently
	Indexes []Index  `json:"indexes"`
	Columns []Column `json:"columns"`
	Comment string   `json:"comment"`
}

type Column struct {
	Name            string   `json:"name"`
	Type            string   `json:"type"`
	Nullable        bool     `json:"nullable"`
	Default         *string  `json:"default"`
	Indexes         []string `json:"indexes"`
	IsPrimaryKey    bool     `json:"isPrimaryKey"`
	IsAutoIncrement bool     `json:"isAutoIncrement"`
	Comment         string   `json:"comment"`
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
	logger := NewDefaultLogger(os.Stdout)
	b.SetLogger(logger)
}

func (b *Base) SetLogger(logger ILogger) {
	b.logger = logger
}

func (b *Base) DB() *sql.DB {
	return b.db
}

func (b *Base) DBType() DBType {
	return b.dbType
}

func (b *Base) LogSQL(sql string, args ...interface{}) {
	if b.logger != nil && b.logger.IsShowSQL() {
		if len(args) > 0 {
			b.logger.Infof("[SQL] %v %v", sql, args)
		} else {
			b.logger.Infof("[SQL] %v", sql)
		}
	}
}

func (table *Table) Column(name string) *Column {
	for i, c := range table.Columns {
		if c.Name == name {
			return &table.Columns[i]
		}
	}
	return nil
}
