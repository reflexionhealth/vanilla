package sql

import (
	"testing"

	"github.com/reflexionhealth/vanilla/expect"
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
	expected = `CREATE TABLE "testers" ("name" text NOT NULL, "experience" integer DEFAULT 0, "pet_name" text)`
	expect.Equal(t, tbl.Create().Sql(), expected)
	expected = `CREATE TABLE IF NOT EXISTS "testers" ("name" text NOT NULL, "experience" integer DEFAULT 0, "pet_name" text)`
	expect.Equal(t, tbl.Create().IfNotExists().Sql(), expected)

	expect.Equal(t, len(tbl.Create().Args()), 0)
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
	expected = `ALTER TABLE "testers" ADD COLUMN "age" integer NOT NULL`
	expect.Equal(t, tbl.Alter().AddColumn(Column{"age", "integer", []string{"NOT NULL"}}).Sql(), expected)
	expect.Equal(t, len(tbl.Columns), 4) // should add the column to table

	expected = `ALTER TABLE "testers" DROP COLUMN "experience", DROP COLUMN "pet_name"`
	expect.Equal(t, tbl.Alter().DropColumn("experience").DropColumn("pet_name").Sql(), expected)
	expect.Equal(t, len(tbl.Columns), 2) // should remove the columns from table

	expected = `ALTER TABLE "testers" NO INHERIT foo, INHERIT bar`
	expect.Equal(t, tbl.Alter().Action("NO INHERIT foo").Action("INHERIT bar").Sql(), expected)

	expect.Equal(t, len(tbl.Alter().Args()), 0)
}

func TestSnakecase(t *testing.T) {
	examples := []struct {
		Input  string
		Output string
	}{
		{Input: "snake_case", Output: "snake_case"}, // NOTE: expected input is camelCase or pascalCase
		{Input: "camelCase", Output: "camel_case"},
		{Input: "PascalCase", Output: "pascal_case"},
		{Input: "exampleID", Output: "example_id"},
		{Input: "HTTPPost", Output: "http_post"},
		{Input: "HostURL", Output: "host_url"},
		{Input: "XMLHttpRequest", Output: "xml_http_request"},
	}
	for _, ex := range examples {
		expect.Equal(t, snakecase(ex.Input), ex.Output)
	}
}
