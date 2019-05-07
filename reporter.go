package microbot

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pangpanglabs/microbot/db"
	"github.com/pangpanglabs/microbot/utils"
)

type DBPingResult struct {
	dbType   db.DBType
	duration int64
	err      error
}

func PingDB() []DBPingResult {
	var results []DBPingResult
	var wg sync.WaitGroup
	for _, d := range dialects {
		wg.Add(1)
		go func(dt db.Dialect) {
			defer wg.Done()
			start := time.Now()
			err := dt.DB().Ping()
			duration := time.Since(start).Nanoseconds()
			results = append(results, DBPingResult{
				dbType:   dt.DBType(),
				duration: duration,
				err:      err,
			})
		}(d)
	}
	wg.Wait()
	return results
}

type TableInfo struct {
	DBType db.DBType  `json:"dbType"`
	Tables []db.Table `json:"tables"`
}

func TableInfoController() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, err := GetTableInfo()
		if err != nil {
			utils.RenderErrorJson(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.RenderDataJson(w, http.StatusOK, v)
	})
}

func GetTableInfo() ([]TableInfo, error) {
	var tableInfos []TableInfo
	for _, d := range dialects {
		tables, err := d.Tables()
		if err != nil {
			return nil, err
		}

		for i := range tables {
			cols, err := d.Columns(tables[i].Name)
			if err != nil {
				return nil, err
			}

			tables[i].Columns = cols
			indexes, err := d.Indexes(tables[i].Name)
			if err != nil {
				return nil, err
			}
			for _, index := range indexes {
				tables[i].Indexes = append(tables[i].Indexes, *index)
				for _, name := range index.Columns {
					if col := tables[i].Column(name); col != nil {
						col.Indexes = append(col.Indexes, index.Name)
					} else {
						return nil, fmt.Errorf("Unknown col %s in index %v of table %v", name, index.Name, tables[i].Name)
					}
				}
			}
		}
		tableInfos = append(tableInfos, TableInfo{
			Tables: tables,
			DBType: d.DBType(),
		})
	}

	return tableInfos, nil
}
