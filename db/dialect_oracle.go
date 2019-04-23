package db

import "strings"

type oracle struct {
	Base
}

func (db *oracle) GetTables() ([]Table, error) {
	args := []interface{}{}
	s := "SELECT table_name FROM user_tables"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		table := NewTable()
		err = rows.Scan(&table.Name)
		if err != nil {
			return nil, err
		}

		tables = append(tables, table)
	}
	return tables, nil
}

func (db *oracle) GetColumns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := "SELECT column_name, data_default, data_type, data_length, data_precision, data_scale," +
		"nullable FROM USER_TAB_COLUMNS WHERE table_name = :1"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var colName, colDefault, nullable, dataType, dataPrecision, dataScale *string
		var dataLen int
		if err = rows.Scan(&colName, &colDefault, &dataType, &dataLen, &dataPrecision, &dataScale, &nullable); err != nil {
			return nil, err
		}
		col := new(Column)
		col.Indexes = make(map[string]int)
		col.Name = strings.Trim(*colName, `" `)
		col.Default = *colDefault
		col.Type = strings.ToLower(*dataType)
		if *nullable == "Y" {
			col.Nullable = true
		}
		cols = append(cols, *col)
	}
	return cols, nil
}

func (db *oracle) GetIndexes(tableName string) (map[string]Index, error) {
	args := []interface{}{tableName}
	s := "SELECT t.column_name, i.uniqueness, i.index_name FROM user_ind_columns t, user_indexes i " +
		"WHERE t.index_name = i.index_name AND t.table_name = i.table_name AND t.table_name =:1"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]Index, 0)
	for rows.Next() {
		var indexType int
		var indexName, colName, uniqueness string
		err = rows.Scan(&colName, &uniqueness, &indexName)
		if err != nil {
			return nil, err
		}

		indexName = strings.Trim(indexName, "` ")
		if err != nil {
			return nil, err
		}

		if uniqueness == "UNIQUE" {
			indexType = UniqueType
		} else {
			indexType = IndexType
		}

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
