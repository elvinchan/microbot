package db

import (
	"testing"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestGetTables(t *testing.T) {
	assert.NoError(t, prepareDialect())
	ts, err := dialect.Tables()
	assert.NoError(t, err)
	assert.Len(t, ts, 2)
}
