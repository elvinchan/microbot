package db

import (
	"fmt"
	"strconv"
	"strings"
)

type mssql struct {
	Base
}

func (db *mssql) Version() string {
	return ""
}

func (db *mssql) Tables() ([]Table, error) {
	args := []interface{}{}
	s := "SELECT a.name, b.rows FROM sys.sysobjects AS a " +
		"INNER JOIN sys.sysindexes AS b ON a.id = b.id WHERE b.indid IN (0, 1) AND a.type = 'U'"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err = rows.Scan(&table.Name, &table.Rows); err != nil {
			return nil, err
		}
		fmt.Printf("----%+v---", table.Name)
		table.Name = strings.Trim(table.Name, "` ")
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mssql) Columns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := `SELECT a.name AS name, b.name AS ctype, a.max_length, a.precision, a.scale, a.is_nullable AS nullable,
REPLACE(REPLACE(ISNULL(c.text, ''), '(', ''), ')', '') AS vdefault,
ISNULL(i.is_primary_key, 0) AS is_primary_key
FROM sys.columns a
LEFT JOIN sys.types b ON a.user_type_id = b.user_type_id
LEFT JOIN sys.syscomments c ON a.default_object_id = c.id
LEFT OUTER JOIN sys.index_columns ic ON ic.object_id = a.object_id AND ic.column_id = a.column_id
LEFT OUTER JOIN sys.indexes i ON ic.object_id = i.object_id AND ic.index_id = i.index_id 
WHERE a.object_id = object_id(?)`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var name, ctype, vdefault string
		var maxLen, precision, scale int
		var nullable, isPK bool
		if err = rows.Scan(&name, &ctype, &maxLen, &precision, &scale, &nullable, &vdefault, &isPK); err != nil {
			return nil, err
		}
		var col Column
		col.Name = strings.Trim(name, "` ")
		col.Nullable = nullable
		col.Default = &vdefault
		col.IsPrimaryKey = isPK

		switch ctype {
		case "decimal", "numeric":
			col.Type = fmt.Sprintf("%s(%d,%d)", ctype, precision, scale)
		case "binary", "char", "varbinary", "varchar":
			if maxLen > 0 {
				col.Type = fmt.Sprintf("%s(%d)", ctype, maxLen)
			} else if maxLen == -1 {
				col.Type = ctype + "(max)"
			} else {
				col.Type = ctype
			}
		case "nchar", "nvarchar":
			if maxLen > 0 {
				col.Type = fmt.Sprintf("%s(%d)", ctype, maxLen/2)
			} else if maxLen == -1 {
				col.Type = ctype + "(max)"
			} else {
				col.Type = ctype
			}
		default:
			col.Type = ctype
		}
		cols = append(cols, col)
	}
	return cols, nil
}

func (db *mssql) Indexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := `SELECT IXS.NAME AS [INDEX_NAME], C.NAME AS [COLUMN_NAME], IXS.is_unique AS [IS_UNIQUE] 
FROM SYS.INDEXES IXS
INNER JOIN SYS.INDEX_COLUMNS IXCS ON IXS.OBJECT_ID = IXCS.OBJECT_ID AND IXS.INDEX_ID = IXCS.INDEX_ID
INNER JOIN SYS.COLUMNS C ON IXS.OBJECT_ID = C.OBJECT_ID AND IXCS.COLUMN_ID= C.COLUMN_ID 
WHERE IXS.TYPE_DESC= 'NONCLUSTERED'
AND OBJECT_NAME(IXS.OBJECT_ID) = ?`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index)
	for rows.Next() {
		var indexName, colName, isUnique string

		if err = rows.Scan(&indexName, &colName, &isUnique); err != nil {
			return nil, err
		}

		i, err := strconv.ParseBool(isUnique)
		if err != nil {
			return nil, err
		}

		colName = strings.Trim(colName, "` ")

		var index *Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index = new(Index)
			index.IsUnique = i
			index.Name = indexName
			indexes[indexName] = index
		}
		index.Columns = append(index.Columns, colName)
	}
	return indexes, nil
}
