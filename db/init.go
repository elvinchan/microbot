package db

import (
	"strings"
)

var dialects = map[string]func() Dialect{}

func init() {
	providedDialects := []struct {
		dbType     DBType
		getDialect func() Dialect
	}{
		{"mssql", func() Dialect { return &mysql{} }},
		{"mysql", func() Dialect { return &mysql{} }},
		{"postgres", func() Dialect { return &postgres{} }},
		{"sqlite3", func() Dialect { return &sqlite3{} }},
		{"oracle", func() Dialect { return &oracle{} }},
	}

	for _, v := range providedDialects {
		RegisterDialect(v.dbType, v.getDialect)
	}
}

// RegisterDialect register database dialect
func RegisterDialect(dbType DBType, dialectFunc func() Dialect) {
	if dialectFunc == nil {
		panic("microbot: nil dialectFunc")
	}
	dialects[strings.ToLower(string(dbType))] = dialectFunc
}

// QueryDialect query database dialect if registed
func QueryDialect(dbType DBType) Dialect {
	if d, ok := dialects[strings.ToLower(string(dbType))]; ok {
		return d()
	}
	return nil
}
