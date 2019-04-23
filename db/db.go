package db

import (
	"database/sql"
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
	Init(*sql.DB, DBType)
	DB() *sql.DB
	DBType() DBType
	GetTables() ([]Table, error)
	GetColumns(tableName string) ([]Column, error)
	GetIndexes(tableName string) (map[string]Index, error)
}

type Base struct {
	db     *sql.DB
	dbType DBType
	name   string
	logger ILogger
}

type Table struct {
	Name    string           `json:"name"`
	Rows    int64            `json:"rows"`
	Indexes map[string]Index `json:"indexes"`
	Columns []Column         `json:"column"`
}

type Column struct {
	Name            string         `json:"name"`
	Type            string         `json:"type"`
	Nullable        bool           `json:"nullable"`
	Default         string         `json:"default"`
	Indexes         map[string]int `json:"indexes"`
	IsPrimaryKey    bool           `json:"isPrimaryKey"`
	IsAutoIncrement bool           `json:"isAutoIncrement"`
	Comment         string         `json:"comment"`
}

type Index struct {
	Name string   `json:"name"`
	Type int      `json:"type"`
	Cols []string `json:"cols"`
}

func (b *Base) Init(d *sql.DB, dbType DBType) {
	b.db = d
	b.dbType = dbType
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

func NewTable() Table {
	return Table{
		Indexes: make(map[string]Index),
	}
}

// add columns which will be composite index
func (index *Index) AddColumn(cols ...string) {
	for _, col := range cols {
		index.Cols = append(index.Cols, col)
	}
}

func (table *Table) GetColumn(name string) *Column {
	for _, c := range table.Columns {
		if c.Name == name {
			return &c
		}
	}
	return nil
}
