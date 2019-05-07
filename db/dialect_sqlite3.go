package db

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

type sqlite3 struct {
	Base
}

func (db *sqlite3) Tables() ([]Table, error) {
	args := []interface{}{}
	s := "SELECT name FROM sqlite_master WHERE type='table'"
	db.LogSQL(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var table Table
		err = rows.Scan(&table.Name)
		if err != nil {
			return nil, err
		}
		if table.Name == "sqlite_sequence" {
			continue
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func (db *sqlite3) Columns(tableName string) ([]Column, error) {
	args := []interface{}{tableName}
	s := "SELECT sql FROM sqlite_master WHERE type='table' and name = ?"
	// db.LogSQL(s, args)
	fmt.Println(s, args)
	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var name string
	for rows.Next() {
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
		break
	}

	if name == "" {
		return nil, errors.New("microbot: no table named " + tableName)
	}

	nStart := strings.Index(name, "(")
	nEnd := strings.LastIndex(name, ")")
	reg := regexp.MustCompile(`[^\(,\)]*(\([^\(]*\))?`)
	colCreates := reg.FindAllString(name[nStart+1:nEnd], -1)
	var cols []Column
	// colSeq := make([]string, 0)
	for _, colStr := range colCreates {
		reg = regexp.MustCompile(`,\s`)
		colStr = reg.ReplaceAllString(colStr, ",")
		if strings.HasPrefix(strings.TrimSpace(colStr), "PRIMARY KEY") {
			parts := strings.Split(strings.TrimSpace(colStr), "(")
			if len(parts) == 2 {
				pkCols := strings.Split(strings.TrimRight(strings.TrimSpace(parts[1]), ")"), ",")
				for _, pk := range pkCols {
					pk := strings.Trim(strings.TrimSpace(pk), "`")
					for i := range cols {
						if cols[i].Name == pk {
							cols[i].IsPrimaryKey = true
						}
					}
				}
			}
			continue
		}

		fields := strings.Fields(strings.TrimSpace(colStr))
		var col Column
		col.Nullable = true

		for idx, field := range fields {
			if idx == 0 {
				col.Name = strings.Trim(strings.Trim(field, "`[] "), `"`)
				continue
			} else if idx == 1 {
				// col.SQLType = core.SQLType{Name: field, DefaultLength: 0, DefaultLength2: 0}
			}
			switch field {
			case "PRIMARY":
				col.IsPrimaryKey = true
			case "AUTOINCREMENT":
				col.IsAutoIncrement = true
			case "NULL":
				if fields[idx-1] == "NOT" {
					col.Nullable = false
				} else {
					col.Nullable = true
				}
			case "DEFAULT":
				col.Default = &fields[idx+1]
			}
		}
		// if !col.SQLType.IsNumeric() && !col.DefaultIsEmpty {
		// 	col.Default = "'" + col.Default + "'"
		// }
		cols = append(cols, col)
		// colSeq = append(colSeq, col.Name)
	}
	return cols, nil
}

func (db *sqlite3) Indexes(tableName string) (map[string]*Index, error) {
	args := []interface{}{tableName}
	s := "SELECT sql FROM sqlite_master WHERE type='index' and tbl_name = ?"
	db.LogSQL(s, args)
	fmt.Println(s, args)

	rows, err := db.DB().Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexes := make(map[string]*Index)
	for rows.Next() {
		var tmpSQL sql.NullString
		err = rows.Scan(&tmpSQL)
		if err != nil {
			return nil, err
		}

		if !tmpSQL.Valid {
			continue
		}
		sql := tmpSQL.String

		index := new(Index)
		nNStart := strings.Index(sql, "INDEX")
		nNEnd := strings.Index(sql, "ON")
		if nNStart == -1 || nNEnd == -1 {
			continue
		}

		indexName := strings.Trim(sql[nNStart+6:nNEnd], "` []")
		index.Name = indexName

		if strings.HasPrefix(sql, "CREATE UNIQUE INDEX") {
			index.IsUnique = true
		}

		nStart := strings.Index(sql, "(")
		nEnd := strings.Index(sql, ")")
		colIndexes := strings.Split(sql[nStart+1:nEnd], ",")

		for _, col := range colIndexes {
			index.Columns = append(index.Columns, strings.Trim(col, "` []"))
		}
		indexes[index.Name] = index
	}

	return indexes, nil
}
