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
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type SortOrder bool

const (
	ASC  SortOrder = false
	DESC SortOrder = true
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

// A ValueCountErrors is thrown while building an Insert or Update if the number
// of values doesn't match the number of columns.
type ValueCountError struct {
	Builder Sqler
	Columns []string
	Values  []interface{}
}

func (e *ValueCountError) Error() string {
	builder := reflect.TypeOf(e.Builder).Elem().Name
	return fmt.Sprintf("in %v.Values(...) expected %v values but received %v", builder, len(e.Columns), len(e.Values))
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
	orderBy    []string
	orderDesc  []SortOrder
	limit      int
}

func Select(columns string) *SelectStmt {
	return &SelectStmt{nil, "", columns, nil, nil, nil, nil, nil, 0}
}

func SelectColumns(columns []Column) *SelectStmt {
	return &SelectStmt{nil, "", "", columns, nil, nil, nil, nil, 0}
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

func (ss *SelectStmt) OrderBy(column string, isDesc SortOrder) *SelectStmt {
	ss.orderBy = append(ss.orderBy, column)
	ss.orderDesc = append(ss.orderDesc, isDesc)
	return ss
}

func (ss *SelectStmt) Limit(num int) *SelectStmt {
	ss.limit = num
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
				qry.WriteString(" AND ")
			}
			qry.WriteString(cond)
		}
	}

	if len(ss.orderBy) > 0 {
		qry.WriteString(" ORDER BY ")
		for i, col := range ss.orderBy {
			if i > 0 {
				qry.WriteString(", ")
			}
			qry.WriteString(col)

			if ss.orderDesc[i] {
				qry.WriteString(" DESC")
			} else {
				qry.WriteString(" ASC")
			}
		}
	}

	if ss.limit > 0 {
		qry.WriteString(fmt.Sprintf(" LIMIT %d", ss.limit))
	}

	return qry.String()
}

func (ss *SelectStmt) Args() []interface{} {
	return ss.arguments
}

// InsertStmt is an expression builder for statements of the form:
//
//   INSERT INTO table (columns) VALUES (values)
//
// TODO: Tests for InsertStmt et al.
type InsertStmt struct {
	dialect   *Dialect
	table     string
	insertion string
	columns   []Column
	arguments []interface{}

	values  int
	records int
}

func Insert(columns string) *InsertStmt {
	values := strings.Count(columns, ",") + 1
	return &InsertStmt{nil, "", columns, nil, nil, values, 0}
}

func InsertColumns(columns []Column) *InsertStmt {
	return &InsertStmt{nil, "", "", columns, nil, len(columns), 0}
}

func (is *InsertStmt) Dialect(dialect *Dialect) *InsertStmt {
	is.dialect = dialect
	return is
}

func (is *InsertStmt) Into(table string) *InsertStmt {
	is.table = table
	return is
}

func (is *InsertStmt) IntoTable(table Table) *InsertStmt {
	is.table = table.Name
	return is
}

// Values will panic with ValueCountError if the number of arguments doesn't
// match the number of columns provided in a previous call to "columns"
func (is *InsertStmt) Values(args ...interface{}) *InsertStmt {
	if len(args) != is.values {
		if len(is.columns) > 0 {
			panic(&ValueCountError{is, ColumnsToNames(is.columns), args})
		} else {
			panic(&ValueCountError{is, strings.Split(is.insertion, ","), args})
		}
	}
	is.arguments = append(is.arguments, args...)
	is.records += 1
	return is
}

func (is *InsertStmt) Sql() string {
	dct := useDialect(is.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("INSERT INTO ")
	dct.WriteIdentifier(&qry, is.table)
	qry.WriteString(" (")
	if len(is.columns) > 0 {
		for i, col := range is.columns {
			if i > 0 {
				qry.WriteString(", ")
			}
			dct.WriteIdentifier(&qry, col.Name)
		}
	} else {
		qry.WriteString(is.insertion)
	}
	qry.WriteString(")")
	if is.records > 0 {
		argn := 0
		qry.WriteString(" VALUES ")
		for r := 0; r < is.records; r++ {
			if r > 0 {
				qry.WriteString(", (")
			} else {
				qry.WriteString("(")
			}
			for v := 0; v < is.values; v++ {
				if v > 0 {
					qry.WriteString(", ")
				}
				argn += 1
				qry.WriteString(dct.Placeholder(argn))
			}
			qry.WriteString(")")
		}
	}

	return qry.String()
}

func (is *InsertStmt) Args() []interface{} {
	return is.arguments
}

// UpdateStmt is an expression builder for statements of the form:
//
//   UPDATE table SET columns ...
//
// TODO: Tests for UpdateStmt et al.
type UpdateStmt struct {
	dialect         *Dialect
	table           string
	columns         []string
	columnValues    []interface{}
	conditions      []string
	conditionValues []interface{}
}

func Update(name string) *UpdateStmt {
	return &UpdateStmt{nil, name, nil, nil, nil, nil}
}

func UpdateTable(table Table) *UpdateStmt {
	return Update(table.Name)
}

func (us *UpdateStmt) Dialect(dialect *Dialect) *UpdateStmt {
	us.dialect = dialect
	return us
}

func (us *UpdateStmt) Set(name string, value interface{}) *UpdateStmt {
	us.columns = append(us.columns, name)
	us.columnValues = append(us.columnValues, value)
	return us
}

func (us *UpdateStmt) Where(condition string, args ...interface{}) *UpdateStmt {
	us.conditions = append(us.conditions, condition)
	us.conditionValues = append(us.conditionValues, args...)
	return us
}

func (us *UpdateStmt) Sql() string {
	dct := useDialect(us.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("UPDATE ")
	dct.WriteIdentifier(&qry, us.table)
	qry.WriteString(" SET ")
	argn := 0

	for i, col := range us.columns {
		if i > 0 {
			qry.WriteString(", ")
		}
		dct.WriteIdentifier(&qry, col)
		qry.WriteString(" = ")
		argn += 1
		qry.WriteString(dct.Placeholder(argn))
	}
	if len(us.conditions) > 0 {
		qry.WriteString(" WHERE ")

		for i, cond := range us.conditions {
			if i > 0 {
				qry.WriteString(", ")
			}
			qry.WriteString(cond)
		}

	}
	return qry.String()
}

func (us *UpdateStmt) Args() []interface{} {
	return append(us.columnValues, us.conditionValues...)
}

// DeleteStmt is an expression builder for statements of the form:
//
//   DELETE FROM table WHERE ...
//
// TODO: Tests for DeleteStmt et al.
type DeleteStmt struct {
	dialect         *Dialect
	table           string
	conditions      []string
	conditionValues []interface{}
}

func Delete(name string) *DeleteStmt {
	return &DeleteStmt{nil, name, nil, nil}
}

func (ds *DeleteStmt) Dialect(dialect *Dialect) *DeleteStmt {
	ds.dialect = dialect
	return ds
}

func (ds *DeleteStmt) From(table string) *DeleteStmt {
	ds.table = table
	return ds
}

func (ds *DeleteStmt) Where(condition string, args ...interface{}) *DeleteStmt {
	ds.conditions = append(ds.conditions, condition)
	ds.conditionValues = append(ds.conditionValues, args...)
	return ds
}

func (ds *DeleteStmt) Args() []interface{} {
	return ds.conditionValues
}

func (ds *DeleteStmt) Sql() string {
	dct := useDialect(ds.dialect)
	qry := bytes.Buffer{}
	qry.WriteString("DELETE FROM ")
	dct.WriteIdentifier(&qry, ds.table)

	if len(ds.conditions) > 0 {
		qry.WriteString(" WHERE ")

		for i, cond := range ds.conditions {
			if i > 0 {
				qry.WriteString(", ")
			}
			qry.WriteString(cond)
		}

	}
	return qry.String()
}

// TODO: Better documentation and tests for InCondition
// e.g. qry.Where(sql.InCondition("thing", len(things), len(qry.Args()), Mysql), things...)
func InCondition(what string, optionCount int, argOffset int, dct *Dialect) string {
	dct = useDialect(dct)
	cond := bytes.Buffer{}
	cond.WriteString(what)
	cond.WriteString(" IN (")
	for i := 0; i < optionCount; i++ {
		if i > 0 {
			cond.WriteString(", ")
		}
		cond.WriteString(dct.Placeholder(i + argOffset))
	}
	cond.WriteString(")")
	return cond.String()
}

// TODO: Better documentation and tests for InCondition
// e.g. qry.Where(sql.NotInCondition("thing", len(things), len(qry.Args()), Mysql), things...)
func NotInCondition(what string, optionCount int, argOffset int, dct *Dialect) string {
	dct = useDialect(dct)
	cond := bytes.Buffer{}
	cond.WriteString(what)
	cond.WriteString(" NOT IN (")
	for i := 0; i < optionCount; i++ {
		if i > 0 {
			cond.WriteString(", ")
		}
		cond.WriteString(dct.Placeholder(i + argOffset))
	}
	cond.WriteString(")")
	return cond.String()
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
// Anonymous struct fields are treated as if their inner fields were in
// the outer struct. Currently, if multiple fields with the same name multiple
// columns with that name will appear multiple times in the response.
//
// For a plain []string, see the ColumnNames function
func Columns(structValue interface{}, flags ColumnsFlag) ([]Column, error) {
	val := reflect.ValueOf(structValue)
	typ := val.Type()
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

		if fld.Anonymous {
			cols, err := Columns(val.Field(i).Interface(), flags)
			if err != nil {
				return nil, err
			}
			columns = append(columns, cols...)
		} else {
			columns = append(columns, Column{Name: inflect(fld.Name, flags)})
		}
	}

	return columns, nil
}

// ColumnNames uses the reflect package to inspect a struct value and returns
// a slice of the column names for the struct's fields. It accepts the ColumnsFlag
// flags to control what fields and names are returned.
//
// Anonymous struct fields are treated as if their inner fields were in
// the outer struct. Currently, if multiple fields with the same name multiple
// columns with that name will appear multiple times in the response.
func ColumnNames(structValue interface{}, flags ColumnsFlag) ([]string, error) {
	val := reflect.ValueOf(structValue)
	typ := val.Type()
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

		if fld.Anonymous {
			cols, err := ColumnNames(val.Field(i).Interface(), flags)
			if err != nil {
				return nil, err
			}
			columns = append(columns, cols...)
		} else {
			columns = append(columns, inflect(fld.Name, flags))
		}
	}

	return columns, nil
}

// ColumnsToNames maps an array of columns to an array of column names
func ColumnsToNames(columns []Column) []string {
	names := make([]string, len(columns))
	for i, col := range columns {
		names[i] = col.Name
	}
	return names
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
