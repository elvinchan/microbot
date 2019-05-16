package db

import (
	"strings"
)

type mysql struct {
	Base
}

func (db *mysql) Tables() ([]Table, error) {
	args := []interface{}{db.name}
	s := "SELECT `TABLE_NAME`, `ENGINE`, `TABLE_ROWS`, `TABLE_COMMENT` FROM " +
		"`INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA` = ? AND (`ENGINE` = 'MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var name, engine, comment string
		var tableRows int64
		err = rows.Scan(&name, &engine, &tableRows, &comment)
		if err != nil {
			return nil, err
		}

		var table Table
		table.Name = name
		table.Engine = engine
		table.Comment = &comment
		table.Rows = tableRows
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mysql) Columns(tableName string) ([]Column, error) {
	args := []interface{}{db.name, tableName}
	s := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `COLUMN_TYPE`, `COLUMN_KEY`, `EXTRA`, `COLUMN_COMMENT`" +
		" FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var columnName, isNullable, colType, colKey, extra, comment string
		var colDefault *string
		if err = rows.Scan(&columnName, &isNullable, &colDefault, &colType, &colKey, &extra, &comment); err != nil {
			return nil, err
		}
		var col Column
		col.Name = strings.Trim(columnName, "` ")
		col.Comment = &comment
		col.Nullable = (isNullable == "YES")
		col.Default = colDefault
		col.Type = strings.ToLower(colType)
		col.IsPrimaryKey = (colKey == "PRI")
		col.IsAutoIncrement = (extra == "auto_increment")
		cols = append(cols, col)
	}
	return cols, nil
}

func (db *mysql) Indexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{db.name, tableName}
	s := "SELECT `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index)
	for rows.Next() {
		var indexName, colName, nonUnique string
		err = rows.Scan(&indexName, &nonUnique, &colName)
		if err != nil {
			return nil, err
		}

		if indexName == "PRIMARY" {
			continue
		}

		isUnique := true
		if nonUnique == "YES" || nonUnique == "1" {
			isUnique = false
		}

		colName = strings.Trim(colName, "` ")

		var index *Index
		var ok bool
		if index, ok = indexes[indexName]; !ok {
			index = new(Index)
			index.IsUnique = isUnique
			index.Name = indexName
			indexes[indexName] = index
		}
		index.Columns = append(index.Columns, colName)
	}
	return indexes, nil
}
