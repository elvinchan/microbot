package db

import (
	"fmt"
	"strings"
)

type oracle struct {
	Base
}

func (db *oracle) Tables() ([]Table, error) {
	args := []interface{}{}
	s := "SELECT T.TABLE_NAME, T.NUM_ROWS, C.COMMENTS FROM USER_TABLES T " +
		"LEFT JOIN USER_TAB_COMMENTS C ON T.TABLE_NAME = C.TABLE_NAME"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		var tableRows *int64
		if err = rows.Scan(&table.Name, &tableRows, &table.Comment); err != nil {
			return nil, err
		}
		if tableRows != nil {
			table.Rows = *tableRows
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *oracle) Columns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := `SELECT T.COLUMN_NAME, T.DATA_DEFAULT, T.DATA_TYPE, T.DATA_LENGTH, T.DATA_PRECISION, T.DATA_SCALE, T.NULLABLE, 
(
	SELECT COUNT(1) FROM USER_CONS_COLUMNS CS
	JOIN USER_CONSTRAINTS CC ON CS.CONSTRAINT_NAME = CC.CONSTRAINT_NAME AND CC.CONSTRAINT_TYPE = 'P'
	WHERE CS.TABLE_NAME = T.TABLE_NAME AND T.COLUMN_NAME = CS.COLUMN_NAME
) IS_PRIMARY_KEY, C.COMMENTS
FROM USER_TAB_COLUMNS T
LEFT JOIN USER_COL_COMMENTS C ON T.TABLE_NAME = C.TABLE_NAME AND T.COLUMN_NAME = C.COLUMN_NAME
WHERE T.TABLE_NAME = :1`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var colName string
		var colDefault, nullable, dataType, comment *string
		var dataLen, isPrimaryKey int
		var dataPrecision, dataScale *int
		if err = rows.Scan(&colName, &colDefault, &dataType, &dataLen, &dataPrecision, &dataScale, &nullable, &isPrimaryKey, &comment); err != nil {
			return nil, err
		}
		if dataType == nil {
			continue
		}
		var col Column
		col.Name = colName
		col.Type = *dataType
		switch col.Type {
		case "CHAR", "NCHAR", "VARCHAR2", "NVARCHAR2":
			col.Type += fmt.Sprintf("(%d)", dataLen)
		case "NUMBER":
			if dataPrecision != nil && dataScale != nil {
				col.Type += fmt.Sprintf("(%d,%d)", *dataPrecision, *dataScale)
			} else if dataPrecision != nil {
				col.Type += fmt.Sprintf("(%d)", *dataPrecision)
			}
		case "FLOAT":
			if dataPrecision != nil {
				col.Type += fmt.Sprintf("(%d)", *dataPrecision)
			}
		}
		if nullable != nil && *nullable == "Y" {
			col.Nullable = true
		}
		if colDefault != nil {
			colDef := strings.Trim(*colDefault, `' `)
			col.Default = &colDef
		}
		col.IsPrimaryKey = (isPrimaryKey > 0)
		col.Comment = comment
		cols = append(cols, col)
	}
	return cols, nil
}

func (db *oracle) Indexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := `SELECT T.COLUMN_NAME, I.UNIQUENESS, I.INDEX_NAME FROM USER_IND_COLUMNS T, USER_INDEXES I
WHERE T.INDEX_NAME = I.INDEX_NAME AND T.TABLE_NAME = I.TABLE_NAME
AND NOT EXISTS (
	SELECT 1 FROM USER_CONSTRAINTS WHERE CONSTRAINT_NAME = T.INDEX_NAME
)
AND T.TABLE_NAME = :1`
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index)
	for rows.Next() {
		var indexName, colName string
		var uniqueness *string
		if err = rows.Scan(&colName, &uniqueness, &indexName); err != nil {
			return nil, err
		}

		var isUnique bool
		if uniqueness != nil && *uniqueness == "UNIQUE" {
			isUnique = true
		}

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
