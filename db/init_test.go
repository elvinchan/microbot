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

	dbType = flag.String("db_type", "postgres", "the tested DB type")
	// connStr = flag.String("conn_str", "./test.db", "test DB connection string")
	connStr = flag.String("conn_str", "host=127.0.0.1 port=5432 user=postgres password=123456 dbname=postgres sslmode=disable", "test DB connection string")
	schema  = flag.String("schema", "", "test DB schema")
	dbName  = flag.String("db_name", "postgres", "test DB name")
)

func prepareDialect() error {
	d, err := sql.Open(*dbType, *connStr)
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
PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS "user";
CREATE TABLE "user" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "desc" text,
  "income" real(9,3),
  "attrs" blob
);

-- ----------------------------
-- Indexes structure for table user
-- ----------------------------
CREATE UNIQUE INDEX "main"."IDX_attrs"
ON "user" (
  "attrs" ASC
);

PRAGMA foreign_keys = true;
`
