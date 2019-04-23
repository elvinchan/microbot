package db

import (
	"strings"
)

type mysql struct {
	Base
}

func (db *mysql) GetTables() ([]Table, error) {
	args := []interface{}{db.name}
	s := "SELECT `TABLE_NAME`, `ENGINE`, `TABLE_ROWS`, `AUTO_INCREMENT`, `TABLE_COMMENT` FROM " +
		"`INFORMATION_SCHEMA`.`TABLES` WHERE `TABLE_SCHEMA` = ? AND (`ENGINE`='MyISAM' OR `ENGINE` = 'InnoDB' OR `ENGINE` = 'TokuDB')"
	db.LogSQL(s, db.name)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		table := NewTable()
		var name, engine, comment string
		var tableRows int64
		var autoIncr *string
		err = rows.Scan(&name, &engine, &tableRows, &autoIncr, &comment)
		if err != nil {
			return nil, err
		}

		table.Name = name
		table.Rows = tableRows
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *mysql) GetColumns(tableName string) ([]Column, error) {
	args := []interface{}{db.name, tableName}
	s := "SELECT `COLUMN_NAME`, `IS_NULLABLE`, `COLUMN_DEFAULT`, `COLUMN_TYPE`," +
		" `COLUMN_KEY`, `EXTRA`,`COLUMN_COMMENT` FROM `INFORMATION_SCHEMA`.`COLUMNS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
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
		col := new(Column)
		col.Indexes = make(map[string]int)
		// colName := strings.Trim(columnName, "` ")
		col.Comment = comment
		if isNullable == "YES" {
			col.Nullable = true
		}

		col.Default = *colDefault
		col.Type = strings.ToLower(colType)

		if colKey == "PRI" {
			col.IsPrimaryKey = true
		}

		if extra == "auto_increment" {
			col.IsAutoIncrement = true
		}
		cols = append(cols, *col)
	}
	return cols, nil
}

func (db *mysql) GetIndexes(tableName string) (map[string]Index, error) {
	args := []interface{}{db.name, tableName}
	s := "SELECT `INDEX_NAME`, `NON_UNIQUE`, `COLUMN_NAME` FROM `INFORMATION_SCHEMA`.`STATISTICS` WHERE `TABLE_SCHEMA` = ? AND `TABLE_NAME` = ?"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, nonUnique string
		err = rows.Scan(&indexName, &nonUnique, &colName)
		if err != nil {
			return nil, err
		}

		if indexName == "PRIMARY" {
			continue
		}

		if nonUnique == "YES" || nonUnique == "1" {
			indexType = IndexType
		} else {
			indexType = UniqueType
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
