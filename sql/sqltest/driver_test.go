package sqltest

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	sql.Register("sqltest", &Driver{})
}

func TestDriverUsage(t *testing.T) {
	db, err := sql.Open("sqltest", "")
	assert.Nil(t, err)

	rows, err := db.Query("SELECT * FROM examples")
	assert.Nil(t, err)

	total := 0
	for rows.Next() {
		err := rows.Scan()
		assert.Nil(t, err)
		total += 1
	}
	err = rows.Close()
	assert.Nil(t, err)
	assert.Zero(t, total)
}

func TestSqlParseError(t *testing.T) {
	db, err := sql.Open("sqltest", "")
	assert.Nil(t, err)

	_, err = db.Query("SELECT * FROM")
	assert.NotNil(t, err)
	assert.Equal(t, "sql:1:14: expected 'a table name' but received 'End of statement'", err.Error())
}
