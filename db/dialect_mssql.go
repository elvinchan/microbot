package db

import (
	"strconv"
	"strings"
)

type mssql struct {
	Base
}

func (db *mssql) GetTables() ([]Table, error) {
	args := []interface{}{}
	s := `SELECT name FROM sysobjects WHERE xtype = 'U'`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		table := NewTable()
		var name string
		if err = rows.Scan(&name); err != nil {
			return nil, err
		}
		table.Name = strings.Trim(name, "` ")
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mssql) GetColumns(tableName string) ([]Column, error) {
	args := []interface{}{db.name, tableName}
	s := `SELECT a.name AS name, b.name AS ctype, a.max_length, a.precision, a.scale, a.is_nullable AS nullable,
	REPLACE(REPLACE(ISNULL(c.text, ''), '(', ''), ')', '') AS vdefault,
	ISNULL(i.is_primary_key, 0)
	FROM sys.columns a
	LEFT JOIN sys.types b ON a.user_type_id = b.user_type_id
	LEFT JOIN sys.syscomments c ON a.default_object_id = c.id
	LEFT OUTER JOIN sys.index_columns ic ON ic.object_id = a.object_id AND ic.column_id = a.column_id
	LEFT OUTER JOIN sys.indexes i ON ic.object_id = i.object_id AND ic.index_id = i.index_id 
	WHERE a.object_id = object_id('` + tableName + `')`
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
		col := new(Column)
		col.Indexes = make(map[string]int)
		col.Name = strings.Trim(name, "` ")
		col.Nullable = nullable
		col.Default = vdefault
		col.IsPrimaryKey = isPK
		// TODO col.Length
		cols = append(cols, *col)
	}
	return cols, nil
}

func (db *mssql) GetIndexes(tableName string) (map[string]Index, error) {
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

	indexes := make(map[string]Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, isUnique string

		if err = rows.Scan(&indexName, &colName, &isUnique); err != nil {
			return nil, err
		}

		i, err := strconv.ParseBool(isUnique)
		if err != nil {
			return nil, err
		}

		if i {
			indexType = UniqueType
		} else {
			indexType = IndexType
		}

		colName = strings.Trim(colName, "` ")

		var index Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index.Type = indexType
			index.Name = indexName
			indexes[indexName] = index
		}
		index.AddColumn(colName)
	}
	return indexes, nil
}
