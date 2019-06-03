package db

import (
	"database/sql"
	"flag"
	"fmt"
	"runtime"
)

const DBName = "microbot_test"

var (
	dialect Dialect

	dbType  = flag.String("db_type", "sqlite3", "the tested DB type")
	connStr = flag.String("conn_str", "", "test DB connection string")
	schema  = flag.String("schema", "", "test DB schema")
	dbName  = flag.String("db_name", "xorm_test", "test DB name")
)

func connection(dbType string) (string, error) {
	if *connStr != "" {
		return *connStr, nil
	}
	switch dbType {
	case "sqlite3":
		return ":memory:", nil
	case "postgres":
		return "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=xorm_test sslmode=disable", nil
	case "mysql":
		return "root:@/xorm_test", nil
	case "oci8":
		return "xorm_test/123456@47.103.34.110:1521/xe", nil
	case "mssql":
		return "", nil
	}
	return "", fmt.Errorf("invalid dbType: %s", dbType)
}

func prepareDialect() error {
	conn, err := connection(*dbType)
	if err != nil {
		return err
	}
	d, err := sql.Open(*dbType, conn)
	if err != nil {
		return err
	}
	dialect = QueryDialect(DBType(*dbType))
	dialect.Init(d, DBType(*dbType), *dbName)
	runtime.SetFinalizer(dialect, close)

	var sql string
	switch *dbType {
	case "postgres":
		rows, err := dialect.DB().Query("SELECT 1 FROM pg_database WHERE datname = $1", DBName)
		if err != nil {
			return fmt.Errorf("Query database error: %v", err)
		}
		defer rows.Close()

		if !rows.Next() {
			if _, err = dialect.DB().Exec("CREATE DATABASE " + DBName); err != nil {
				return fmt.Errorf("create database error: %v", err)
			}
		}
		if *schema != "" {
			if _, err = dialect.DB().Exec("CREATE SCHEMA IF NOT EXISTS " + *schema); err != nil {
				return fmt.Errorf("create schema error: %v", err)
			}
		}
	case "sqlite3":
		sql = sqliteSQL
	}
	if _, err = dialect.DB().Exec(sql); err != nil {
		return err
	}
	return nil
}

func close(d Dialect) {
	d.DB().Close()
}

const sqliteSQL = `
DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "desc" text,
  "income" real(9,3),
  "attrs" blob
);

CREATE UNIQUE INDEX "main"."IDX_attrs"
ON "user" (
  "attrs" ASC
);

DROP TABLE IF EXISTS "phone";
CREATE TABLE "phone" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "userId" integer,
  "num" text
);

CREATE UNIQUE INDEX "main"."IDX_phone"
ON "phone" (
  "userId" ASC
);
`
