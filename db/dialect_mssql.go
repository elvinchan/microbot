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
	s := `SELECT s.name, i.rows, ep.value AS comment
FROM sys.sysobjects AS s
JOIN sys.sysindexes AS i ON s.id = i.id AND i.indid IN (0, 1)
LEFT JOIN sys.extended_properties ep ON s.id = ep.major_id AND ep.minor_id = 0 AND ep.name = 'MS_Description'
WHERE s.type = 'U'`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		if err = rows.Scan(&table.Name, &table.Rows, &table.Comment); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mssql) Columns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := `SELECT c.name AS name, t.name AS ctype, c.max_length, c.precision, c.scale, c.is_nullable AS nullable,
REPLACE(REPLACE(ISNULL(s.text, ''), '(', ''), ')', '') AS vdefault,
ISNULL(i.is_primary_key, 0) AS is_primary_key, ep.value AS comment
FROM sys.columns c
LEFT JOIN sys.types t ON c.user_type_id = t.user_type_id
LEFT JOIN sys.syscomments s ON c.default_object_id = s.id
LEFT JOIN sys.index_columns ic ON c.object_id = ic.object_id AND c.column_id = ic.column_id
LEFT JOIN sys.indexes i ON ic.object_id = i.object_id AND ic.index_id = i.index_id
LEFT JOIN sys.extended_properties ep ON c.object_id = ep.major_id AND c.column_id = ep.minor_id AND ep.name = 'MS_Description'
WHERE c.object_id = object_id(?)`
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
		var comment *string
		if err = rows.Scan(&name, &ctype, &maxLen, &precision, &scale, &nullable, &vdefault, &isPK, &comment); err != nil {
			return nil, err
		}
		var col Column
		col.Name = strings.Trim(name, "` ")
		col.Nullable = nullable
		col.Default = &vdefault
		col.IsPrimaryKey = isPK
		col.Comment = comment

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
	s := `SELECT i.name AS index_name, c.name AS column_name, i.is_unique AS is_unique
FROM sys.indexes i
JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
JOIN sys.columns c ON i.object_id = c.object_id AND ic.column_id = c.column_id
WHERE i.type_desc = 'NONCLUSTERED' AND OBJECT_NAME(i.object_id) = ?`
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
