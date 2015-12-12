package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTable(t *testing.T) {
	tbl := Table{
		Name: "testers",
		Columns: []Column{
			{"name", "text", []string{"NOT NULL"}},
			{"experience", "integer", []string{"DEFAULT 0"}},
			{"pet_name", "text", nil},
		},
	}

	var expected string
	expected = `CREATE TABLE testers (name text NOT NULL, experience integer DEFAULT 0, pet_name text)`
	assert.Equal(t, tbl.Create().Sql(), expected)
	expected = `CREATE TABLE IF NOT EXISTS testers (name text NOT NULL, experience integer DEFAULT 0, pet_name text)`
	assert.Equal(t, tbl.Create().IfNotExists().Sql(), expected)

	assert.Equal(t, len(tbl.Create().Args()), 0)
}

func TestAlterTable(t *testing.T) {
	tbl := Table{
		Name: "testers",
		Columns: []Column{
			{"name", "text", []string{"NOT NULL"}},
			{"experience", "integer", []string{"DEFAULT 0"}},
			{"pet_name", "text", nil},
		},
	}

	var expected string
	expected = `ALTER TABLE testers ADD COLUMN age integer NOT NULL`
	assert.Equal(t, tbl.Alter().AddColumn(Column{"age", "integer", []string{"NOT NULL"}}).Sql(), expected)
	assert.Equal(t, len(tbl.Columns), 4) // should add the column to table

	expected = `ALTER TABLE testers DROP COLUMN experience, DROP COLUMN pet_name`
	assert.Equal(t, tbl.Alter().DropColumn("experience").DropColumn("pet_name").Sql(), expected)
	assert.Equal(t, len(tbl.Columns), 2) // should remove the columns from table

	expected = `ALTER TABLE testers NO INHERIT foo, INHERIT bar`
	assert.Equal(t, tbl.Alter().Action("NO INHERIT foo").Action("INHERIT bar").Sql(), expected)

	assert.Equal(t, len(tbl.Alter().Args()), 0)
}
