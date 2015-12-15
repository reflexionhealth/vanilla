package sql

import "bytes"

// Dialect contains the rules necessary to generate SQL for a specific database engine.
// Specifying a Dialect is optional, the ANSI dialect is used by default.
//
// A Dialect can be specified for a statement like:
//     sql.Select("*").From("example").Dialect(&dialect)
//
// However it may be more convenient to use the following the pattern instead:
//
//     var mysql Dialect
//     mysql.Select("*").From("example")
//
// TODO: Tests for Dialect et al.
type Dialect struct {
	IdentifierOpen  rune
	IdentifierClose rune
}

var Ansi = Dialect{IdentifierOpen: '"', IdentifierClose: '"'}

func useDialect(dialect *Dialect) *Dialect {
	if dialect == nil {
		return &Ansi
	} else {
		return dialect
	}
}

// Not sure whether I really want to define these here
//var MsSql = Dialect{IdentifierOpen: '[', IdentifierClose: ']'}
//var Mysql = Dialect{IdentifierOpen: '`', IdentifierClose: '`'}
//var Postgres = Dialect{IdentifierOpen: '"', IdentifierClose: '"'}

func (d *Dialect) WriteIdentifier(buf *bytes.Buffer, ident string) {
	buf.WriteRune(d.IdentifierOpen)
	buf.WriteString(ident)
	buf.WriteRune(d.IdentifierClose)
}

func (d *Dialect) CreateTable(name string) *CreateTableStmt {
	return CreateTable(name).Dialect(d)
}

func (d *Dialect) AlterTable(name string) *AlterTableStmt {
	return AlterTable(name).Dialect(d)
}

func (d *Dialect) Select(selection string) *SelectStmt {
	return Select(selection).Dialect(d)
}

func (d *Dialect) SelectColumns(columns []Column) *SelectStmt {
	return SelectColumns(columns).Dialect(d)
}
