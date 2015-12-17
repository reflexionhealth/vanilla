// Package sql provides a collection of utilities to make working with sql
// in Golang more natural. The primary use is building dynamic queries from
// from in-memory data structures and other data sources.
//
// The package is not intended to wrap or abstract Go's existing sql package.
// Additionaly, it is not intended to replace simple hard-coded queries.
// If you know what you want before compile-time, just type it!
//
//   DO: `SELECT username FROM users`
//   DONT: Select({"username"}).From("users").Sql()
//
package sql

import (
	"bytes"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Sqler is the interface for SQL expression builders
type Sqler interface {
	// Sql should build and return the SQL string representation of the expression
	Sql() string

	// Args returns the literal values provided to the builder as a slice that
	// can be passed to db.Exec or db.Query with the spread "..." operator
	Args() []interface{}
}

// Column is a Go representation of a single column in a table
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

// Table is a Go representation of a single table in a database
type Table struct {
	Name        string
	Columns     []Column
	Constraints []string
}

// CreateTableStmt is an expression builder for statements of the form:
//
//   CREATE TABLE table_name ( ... )"
//
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

// AlterTableStmt is an expression builder for statements of the form:
//
//   ALTER TABLE table_name ...
//
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


// SelectStmt is an expression builder for statements of the form:
//
//   SELECT columns FROM table ...
//
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
	dct.WriteIdentifier(&qry, ss.table)

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

// A ColumnsFlag is a flag which controls how Columns and ColumnNames interpret
// struct fields as columns.
type ColumnsFlag int

// If no ColumnNames flag is passed, column names will match the field names
const (
	// ColumnNamesCamelcase downcases the first character of the field name
	ColumnNamesCamelcase ColumnsFlag = 1 << iota

	// ColumnNamesLowercase downcases all characters of the field name
	ColumnNamesLowercase

	// ColumnNamesPascalcase upcases the first character of the field name
	ColumnNamesPascalcase

	// ColumnNamesSnakecase downcases all characters of the field name, inserting
	// an underscore before each word (other than first).
	// It does not try to detect common initialisms
	ColumnNamesSnakecase

	// ColumnsOnlyExported skips fields that are unexported (first character uppercase)
	ColumnsOnlyExported

	// ColumnsOnlyExported only outputs columns for fields with the "sql" tag
	//ColumnsOnlyTagged
)

// Columns uses the reflect package to inspect a struct value and returns
// a slice of columns for the struct's fields. It accepts the ColumnsFlag
// flags to control what fields and names are returned.
//
// For a plain []string, see the ColumnNames function
func Columns(structValue interface{}, flags ColumnsFlag) ([]Column, error) {
	typ := reflect.TypeOf(structValue)
	if typ.Kind() != reflect.Struct {
		// needless runtime sacrifice to the gods of type safety
		return nil, &reflect.ValueError{"Columns", typ.Kind()}
	}

	var columns []Column
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if flags&ColumnsOnlyExported != 0 && len(fld.PkgPath) > 0 {
			continue
		}

		columns = append(columns, Column{Name: inflect(fld.Name, flags)})
	}

	return columns, nil
}

// ColumnNames uses the reflect package to inspect a struct value and returns
// a slice of the column names for the struct's fields. It accepts the ColumnsFlag
// flags to control what fields and names are returned.
func ColumnNames(structValue interface{}, flags ColumnsFlag) ([]string, error) {
	typ := reflect.TypeOf(structValue)
	if typ.Kind() != reflect.Struct {
		// needless runtime sacrifice to the gods of type safety
		return nil, &reflect.ValueError{"ColumnNames", typ.Kind()}
	}

	var columns []string
	for i := 0; i < typ.NumField(); i++ {
		fld := typ.Field(i)
		if flags&ColumnsOnlyExported != 0 && len(fld.PkgPath) > 0 {
			continue
		}

		columns = append(columns, inflect(fld.Name, flags))
	}

	return columns, nil
}

func inflect(input string, flags ColumnsFlag) string {
	switch {
	case flags&ColumnNamesCamelcase != 0:
		return camelcase(input)
	case flags&ColumnNamesLowercase != 0:
		return strings.ToLower(input)
	case flags&ColumnNamesPascalcase != 0:
		return pascalcase(input)
	case flags&ColumnNamesSnakecase != 0:
		return snakecase(input)
	default:
		return input
	}
}

func camelcase(input string) string {
	r, size := utf8.DecodeRuneInString(input)
	if unicode.IsUpper(r) {
		return string(unicode.ToLower(r)) + input[size:]
	} else {
		return input
	}
}

func pascalcase(input string) string {
	r, size := utf8.DecodeRuneInString(input)
	if unicode.IsLower(r) {
		return string(unicode.ToUpper(r)) + input[size:]
	} else {
		return input
	}
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
