package sqltest

import (
	"database/sql"
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
)

func init() {
	sql.Register("sqltest", &Driver{})
}

func TestDriverUsage(t *testing.T) {
	db, err := sql.Open("sqltest", "")
	expect.Nil(t, err)

	rows, err := db.Query("SELECT * FROM examples")
	expect.Nil(t, err)

	total := 0
	for rows.Next() {
		err := rows.Scan()
		expect.Nil(t, err)
		total += 1
	}
	err = rows.Close()
	expect.Nil(t, err)
	expect.Equal(t, total, 0)
}

func TestSqlParseError(t *testing.T) {
	db, err := sql.Open("sqltest", "")
	expect.Nil(t, err)

	_, err = db.Query("SELECT * FROM")
	expect.NotNil(t, err)
	expect.Equal(t, err.Error(), "sql:1:14: expected 'a table name' but received 'End of statement'")
}
