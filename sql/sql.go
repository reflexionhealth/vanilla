// Package sql provides a collection of utilities to make working with sql
// in Golang more natural. The primary use is building dynamic queries from
// from in-memory data structures and other data sources.
//
// Non-goals:
//  * Wrapping or abstracting the existing SQL library
//  * Replacing simple hard-coded queries, eg:
//      DON'T: Select({"username"}).From("users").Sql()
//      DO: `SELECT username FROM users`
//
package sql

import (
	"bytes"
	"reflect"
	"unicode"
)

type Sqler interface {
	Sql() string
	Args() []interface{}
}

type Column struct {
	Name        string
	Type        string
	Constraints []string
}

func (c *Column) WriteSql(buf *bytes.Buffer, dct *Dialect) {
	dct.WriteIdentifier(buf, c.Name)
	buf.WriteString(" ")
	buf.WriteString(c.Type)
	for _, con := range c.Constraints {
		buf.WriteString(" ")
		buf.WriteString(con)
	}
}

type Table struct {
	Name        string
	Columns     []Column
	Constraints []string
}

// CREATE TABLE table_name ( ... )
type CreateTableStmt struct {
	dialect     *Dialect
	table       *Table
	ifNotExists bool
}

// TODO: Tests for CreateTable
func CreateTable(name string) *CreateTableStmt {
	return (&Table{Name: name}).Create()
}

func (t *Table) Create() *CreateTableStmt {
	return &CreateTableStmt{nil, t, false}
}

func (ct *CreateTableStmt) IfNotExists() *CreateTableStmt {
	ct.ifNotExists = true
	return ct
}

func (ct *CreateTableStmt) Column(col Column) *CreateTableStmt {
	ct.table.Columns = append(ct.table.Columns, col)
	return ct
}

// TODO: Tests for (CreateTableStmt).Constraint
func (ct *CreateTableStmt) Constraint(cons string) *CreateTableStmt {
	ct.table.Constraints = append(ct.table.Constraints, cons)
	return ct
}

// TODO: Tests for (CreateTableStmt).Dialect
func (ct *CreateTableStmt) Dialect(dialect *Dialect) *CreateTableStmt {
	ct.dialect = dialect
	return ct
}

func (ct *CreateTableStmt) Sql() string {
	dct := useDialect(ct.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("CREATE TABLE ")
	if ct.ifNotExists {
		qry.WriteString("IF NOT EXISTS ")
	}
	dct.WriteIdentifier(&qry, ct.table.Name)
	qry.WriteString(" (")

	exprs := 0
	for _, col := range ct.table.Columns {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		col.WriteSql(&qry, dct)
	}

	for _, con := range ct.table.Constraints {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString(" ")
		qry.WriteString(con)
	}

	qry.WriteString(")")
	return qry.String()
}

func (ct *CreateTableStmt) Args() []interface{} {
	return nil
}

// ALTER TABLE table_name ...
type AlterTableStmt struct {
	dialect *Dialect
	table   *Table
	adds    []Column
	drops   []string
	actions []string
}

// TODO: Tests for AlterTable
func AlterTable(name string) *AlterTableStmt {
	return (&Table{Name: name}).Alter()
}

func (t *Table) Alter() *AlterTableStmt {
	return &AlterTableStmt{nil, t, nil, nil, nil}
}

func (at *AlterTableStmt) Action(action string) *AlterTableStmt {
	at.actions = append(at.actions, action)
	return at
}

func (at *AlterTableStmt) AddColumn(col Column) *AlterTableStmt {
	at.table.Columns = append(at.table.Columns, col)
	at.adds = append(at.adds, col)
	return at
}

func (at *AlterTableStmt) DropColumn(name string) *AlterTableStmt {
	at.drops = append(at.drops, name)
	for i, col := range at.table.Columns {
		if col.Name == name {
			at.table.Columns = append(at.table.Columns[:i], at.table.Columns[i+1:]...)
			break
		}
	}
	return at
}

// TODO: Tests for (AlterTableStmt).Dialect
func (at *AlterTableStmt) Dialect(dialect *Dialect) *AlterTableStmt {
	at.dialect = dialect
	return at
}

func (at *AlterTableStmt) Sql() string {
	dct := useDialect(at.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("ALTER TABLE ")
	dct.WriteIdentifier(&qry, at.table.Name)
	qry.WriteString(" ")

	exprs := 0

	for _, col := range at.adds {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString("ADD COLUMN ")
		col.WriteSql(&qry, dct)
	}

	for _, name := range at.drops {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString("DROP COLUMN ")
		dct.WriteIdentifier(&qry, name)
	}

	for _, action := range at.actions {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString(action)
	}

	return qry.String()
}

func (at *AlterTableStmt) Args() []interface{} {
	return nil
}

// SELECT columns ...
// TODO: Tests for SelectStmt et al.
// TODO: Having, GroupBy, OrderBy, Limit, Offset
type SelectStmt struct {
	dialect    *Dialect
	table      string
	selection  string
	columns    []Column
	conditions []string
	arguments  []interface{}
}

func Select(columns string) *SelectStmt {
	return &SelectStmt{nil, "", columns, nil, nil, nil}
}

func SelectColumns(columns []Column) *SelectStmt {
	return &SelectStmt{nil, "", "", columns, nil, nil}
}

func (ss *SelectStmt) Dialect(dialect *Dialect) *SelectStmt {
	ss.dialect = dialect
	return ss
}

func (ss *SelectStmt) From(table string) *SelectStmt {
	ss.table = table
	return ss
}

func (ss *SelectStmt) FromTable(table Table) *SelectStmt {
	ss.table = table.Name
	return ss
}

func (ss *SelectStmt) Where(condition string, args ...interface{}) *SelectStmt {
	ss.conditions = append(ss.conditions, condition)
	ss.arguments = append(ss.arguments, args...)
	return ss
}

func (ss *SelectStmt) Sql() string {
	dct := useDialect(ss.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("SELECT ")
	if len(ss.columns) > 0 {
		for i, col := range ss.columns {
			if i > 0 {
				qry.WriteString(", ")
			}
			dct.WriteIdentifier(&qry, col.Name)
		}
	} else {
		qry.WriteString(ss.selection)
	}

	qry.WriteString(" FROM ")
	qry.WriteString(ss.table)

	if len(ss.conditions) > 0 {
		qry.WriteString(" WHERE ")
		for i, cond := range ss.conditions {
			if i > 0 {
				qry.WriteString(", ")
			}
			qry.WriteString(cond)
		}
	}

	return qry.String()
}

func (ss *SelectStmt) Args() []interface{} {
	return ss.arguments
}

type ColumnsFlag int

const (
	ColumnNamesSnakecase ColumnsFlag = 1 << iota
	// ColumnNamesLowercase
	// ColumnNamesCamelcase
	// ColumnNamesPascalcase
	ColumnsOnlyExported
	// ColumnsOnlyTagged
)

// TODO: Tests for Columns
func Columns(structValue interface{}, flags ColumnsFlag) ([]Column, error) {
	typ := reflect.TypeOf(structValue)
	if typ.Kind() != reflect.Struct {
		// needless runtime sacrifice to the gods of type safety
		return nil, &reflect.ValueError{"ColumnsFor", typ.Kind()}
	}

	var columns []Column
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if flags&ColumnsOnlyExported != 0 && len(fld.PkgPath) > 0 {
			continue
		}

		if flags&ColumnNamesSnakecase != 0 {
			columns = append(columns, Column{Name: snakecase(fld.Name)})
		} else {
			columns = append(columns, Column{Name: fld.Name})
		}
	}

	return columns, nil
}

func snakecase(input string) string {
	var output bytes.Buffer
	for i, char := range input {
		if unicode.IsUpper(char) {
			if i > 0 {
				output.WriteRune('_')
			}
			output.WriteRune(unicode.ToLower(char))
		} else {
			output.WriteRune(char)
		}
	}

	return output.String()
}
