package db

import (
	"fmt"
	"strings"

	"github.com/go-xorm/core"
)

type postgres struct {
	Base
	Schema string
}

func (db *postgres) GetTables() ([]Table, error) {
	args := []interface{}{}
	s := "SELECT tablename FROM pg_tables"
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
		s = s + " WHERE schemaname = $1"
	}
	db.LogSQL(s, args)
	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tables := make([]Table, 0)
	for rows.Next() {
		table := NewTable()
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		table.Name = name
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *postgres) GetColumns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := `SELECT column_name, column_default, is_nullable, data_type, character_maximum_length, numeric_precision, numeric_precision_radix ,
    CASE WHEN p.contype = 'p' THEN true ELSE false END AS primarykey,
    CASE WHEN p.contype = 'u' THEN true ELSE false END AS uniquekey
FROM pg_attribute f
    JOIN pg_class c ON c.oid = f.attrelid JOIN pg_type t ON t.oid = f.atttypid
    LEFT JOIN pg_attrdef d ON d.adrelid = c.oid AND d.adnum = f.attnum
    LEFT JOIN pg_namespace n ON n.oid = c.relnamespace
    LEFT JOIN pg_constraint p ON p.conrelid = c.oid AND f.attnum = ANY (p.conkey)
    LEFT JOIN pg_class AS g ON p.confrelid = g.oid
    LEFT JOIN INFORMATION_SCHEMA.COLUMNS s ON s.column_name=f.attname AND c.relname=s.table_name
WHERE c.relkind = 'r'::char AND c.relname = $1%s AND f.attnum > 0 ORDER BY f.attnum;`
	var f string
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
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
		col := new(Column)
		col.Indexes = make(map[string]int)

		var colName, isNullable, dataType string
		var maxLenStr, colDefault, numPrecision, numRadix *string
		var isPK, isUnique bool
		err = rows.Scan(&colName, &colDefault, &isNullable, &dataType, &maxLenStr, &numPrecision, &numRadix, &isPK, &isUnique)
		if err != nil {
			return nil, err
		}

		col.Name = strings.Trim(colName, `" `)

		if colDefault != nil || isPK {
			if isPK {
				col.IsPrimaryKey = true
			} else {
				col.Default = *colDefault
			}
		}

		if colDefault != nil && strings.HasPrefix(*colDefault, "nextval(") {
			col.IsAutoIncrement = true
		}

		col.Nullable = (isNullable == "YES")
		cols = append(cols, *col)
	}

	return cols, nil
}

func (db *postgres) GetIndexes(tableName string) (map[string]Index, error) {
	args := []interface{}{tableName}
	s := fmt.Sprintf("SELECT indexname, indexdef FROM pg_indexes WHERE tablename=$1")
	if len(db.Schema) != 0 {
		args = append(args, db.Schema)
		s = s + " AND schemaname=$2"
	}
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]Index, 0)
	for rows.Next() {
		var indexType int
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
		if strings.HasPrefix(indexdef, "CREATE UNIQUE INDEX") {
			indexType = core.UniqueType
		} else {
			indexType = core.IndexType
		}
		colNames = getIndexColName(indexdef)
		if strings.HasPrefix(indexName, "IDX_"+tableName) || strings.HasPrefix(indexName, "UQE_"+tableName) {
			newIdxName := indexName[5+len(tableName):]
			if newIdxName != "" {
				indexName = newIdxName
			}
		}

		var index Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index.Type = indexType
			index.Name = indexName
			indexes[indexName] = index
		}
		index.AddColumn(colNames...)
	}
	return indexes, nil
}

func getIndexColName(indexdef string) []string {
	var colNames []string

	cs := strings.Split(indexdef, "(")
	for _, v := range strings.Split(strings.Split(cs[1], ")")[0], ",") {
		colNames = append(colNames, strings.Split(strings.TrimLeft(v, " "), " ")[0])
	}

	return colNames
}
