package db

import (
	"testing"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestGetTables(t *testing.T) {
	assert.NoError(t, prepareDialect())
	_, err := dialect.Tables()
	assert.NoError(t, err)
	// assert.Len(t, table, 2)
}
