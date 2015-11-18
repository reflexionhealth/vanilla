// Package sql provides a collection of statement builders to make writing sql
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

import "bytes"

type Sqler interface {
	Sql() string
	Args() []interface{}
}

type Column struct {
	Name        string
	Type        string
	Constraints []string
}

func (c *Column) WriteSql(buf *bytes.Buffer) {
	buf.WriteString(c.Name)
	buf.WriteString(" ")
	buf.WriteString(c.Type)
	for _, con := range c.Constraints {
		buf.WriteString(" ")
		buf.WriteString(con)
	}
}

type Table struct {
	Name    string
	Columns []Column
}

// CREATE TABLE table_name ( ... )
type CreateTable struct {
	table       *Table
	ifNotExists bool
}

func (t *Table) Create() *CreateTable {
	return &CreateTable{t, false}
}

func (ct *CreateTable) IfNotExists() *CreateTable {
	ct.ifNotExists = true
	return ct
}

func (ct *CreateTable) Sql() string {
	qry := bytes.Buffer{}
	qry.WriteString("CREATE TABLE ")
	if ct.ifNotExists {
		qry.WriteString("IF NOT EXISTS ")
	}
	qry.WriteString(ct.table.Name)
	qry.WriteString(" (")

	exprs := 0
	for _, col := range ct.table.Columns {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		col.WriteSql(&qry)
	}

	// for _, con := range ct.constraints {
	//   if exprs += 1; exprs > 1 {
	//     qry.WriteString(", ")
	//   }
	//   qry.WriteString(" ")
	//   qry.WriteString(con)
	// }

	qry.WriteString(")")
	return qry.String()
}

func (ct *CreateTable) Args() []interface{} {
	return nil
}

// ALTER TABLE table_name ...
type AlterTable struct {
	table   *Table
	adds    []Column
	actions []string
}

func (t *Table) Alter() *AlterTable {
	return &AlterTable{t, nil, nil}
}

func (at *AlterTable) Action(action string) *AlterTable {
	at.actions = append(at.actions, action)
	return at
}

func (at *AlterTable) AddColumn(col Column) *AlterTable {
	at.table.Columns = append(at.table.Columns, col)
	at.adds = append(at.adds, col)
	return at
}

func (at *AlterTable) DropColumn(name string) *AlterTable {
	at.actions = append(at.actions, "DROP COLUMN "+name)
	for i, col := range at.table.Columns {
		if col.Name == name {
			at.table.Columns = append(at.table.Columns[:i], at.table.Columns[i+1:]...)
			break
		}
	}
	return at
}

func (at *AlterTable) Sql() string {
	qry := bytes.Buffer{}
	qry.WriteString("ALTER TABLE ")
	qry.WriteString(at.table.Name)
	qry.WriteString(" ")

	exprs := 0
	for _, col := range at.adds {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString("ADD COLUMN ")
		col.WriteSql(&qry)
	}

	for _, action := range at.actions {
		if exprs += 1; exprs > 1 {
			qry.WriteString(", ")
		}
		qry.WriteString(action)
	}

	return qry.String()
}

func (at *AlterTable) Args() []interface{} {
	return nil
}
