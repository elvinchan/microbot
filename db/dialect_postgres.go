package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type postgres struct {
	Base
	schema string
}

// DefaultPostgresSchema default postgres schema
const DefaultPostgresSchema = "public"

func (db *postgres) Init(d *sql.DB, dbType DBType, dbName string) {
	db.Base.Init(d, dbType, dbName)
	db.schema = DefaultPostgresSchema
}

func (db *postgres) Tables() ([]Table, error) {
	args := []interface{}{}
	s := `SELECT t.tablename, c.reltuples::bigint AS rows, obj_description(c.oid)
FROM pg_tables t
JOIN pg_class c ON t.tablename = c.relname
JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = t.schemaname`
	if len(db.schema) != 0 {
		args = append(args, db.schema)
		s = s + "\nWHERE t.schemaname = $1"
	}
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make([]Table, 0)
	for rows.Next() {
		var name string
		var comment *string
		var tableRows int64
		err = rows.Scan(&name, &tableRows, &comment)
		if err != nil {
			return nil, err
		}
		var table Table
		table.Name = name
		table.Rows = tableRows
		table.Comment = comment
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *postgres) Columns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := `SELECT column_name, column_default, is_nullable, data_type,
coalesce(character_maximum_length, numeric_precision) AS num_length,
numeric_scale AS num_scale,
CASE WHEN p.contype = 'p' THEN true ELSE false END AS primarykey,
CASE WHEN p.contype = 'u' THEN true ELSE false END AS uniquekey,
col_description(f.attrelid, f.attnum) AS column_comment
FROM INFORMATION_SCHEMA.COLUMNS s
JOIN pg_attribute f ON f.attname = s.column_name
JOIN pg_class c ON c.oid = f.attrelid AND c.relname = s.table_name
JOIN pg_namespace n ON n.oid = c.relnamespace AND n.nspname = s.table_schema
LEFT JOIN pg_constraint p ON p.conrelid = c.oid AND f.attnum = ANY (p.conkey)
WHERE c.relkind = 'r'::char AND s.table_name = $1%s AND f.attnum > 0
ORDER BY f.attnum`
	var f string
	if len(db.schema) != 0 {
		args = append(args, db.schema)
		f = " AND s.table_schema = $2"
	}
	s = fmt.Sprintf(s, f)
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var colName, isNullable, dataType string
		var colDefault, maxLength, numScale, comment *string
		var isPK, isUnique bool
		if err = rows.Scan(&colName, &colDefault, &isNullable, &dataType, &maxLength, &numScale, &isPK, &isUnique, &comment); err != nil {
			return nil, err
		}

		var col Column
		col.Name = strings.Trim(colName, `" `)
		if colDefault != nil || isPK {
			if isPK {
				col.IsPrimaryKey = true
			} else {
				col.Default = colDefault
			}
		}
		if colDefault != nil && strings.HasPrefix(*colDefault, "nextval(") {
			col.IsAutoIncrement = true
		}
		switch dataType {
		case "character varying":
			col.Type = "varchar"
		case "character":
			col.Type = "char"
		case "bit varying":
			col.Type = "varbit"
		case "timestamp without time zone":
			col.Type = "timestamp"
		case "timestamp with time zone":
			col.Type = "timestamptz"
		case "time without time zone":
			col.Type = "time"
		case "time with time zone":
			col.Type = "timetz"
		case "double precision":
			col.Type = "float8"
		case "boolean":
			col.Type = "bool"
		case "oid":
			col.Type = "bigint"
		case "bigserial", "smallserial", "serial":
			col.IsAutoIncrement = true
			col.Nullable = false
			col.Type = dataType
		default:
			col.Type = dataType
		}
		if maxLength != nil && numScale != nil && *numScale != "0" {
			col.Type += "(" + *maxLength + "," + *numScale + ")"
		} else if maxLength != nil {
			col.Type += "(" + *maxLength + ")"
		}
		col.Nullable = (isNullable == "YES")
		col.Comment = comment
		cols = append(cols, col)
	}

	return cols, nil
}

func (db *postgres) Indexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename=$1")
	if len(db.schema) != 0 {
		args = append(args, db.schema)
		s = s + " AND schemaname=$2"
	}
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index)
	for rows.Next() {
		var indexName, indexdef string
		var colNames []string
		err = rows.Scan(&indexName, &indexdef)
		if err != nil {
			return nil, err
		}
		indexName = strings.Trim(indexName, `" `)
		if strings.HasSuffix(indexName, "_pkey") {
			continue
		}
		var isUnique bool
		if strings.HasPrefix(indexdef, "CREATE UNIQUE INDEX") {
			isUnique = true
		}
		colNames = getIndexColName(indexdef)
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			newIdxName := indexName[5+len(tableName):]
			if newIdxName != "" {
				indexName = newIdxName
			}
		}

		var index *Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index = new(Index)
			index.IsUnique = isUnique
			index.Name = indexName
			indexes[indexName] = index
		}
		index.Columns = append(index.Columns, colNames...)
	}
	return indexes, nil
}

func (b *postgres) SetSchema(schema string) {
	b.schema = schema
}

func getIndexColName(indexdef string) []string {
	var colNames []string

	cs := strings.Split(indexdef, "(")
	for _, v := range strings.Split(strings.Split(cs[1], ")")[0], ",") {
		colNames = append(colNames, strings.Split(strings.TrimLeft(v, " "), " ")[0])
	}

	return colNames
}
