package sql

import "bytes"
import "strconv"

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
	IdentOpen   rune
	IdentClose  rune
	Placeholder func(n int) string
}

// The SQL dialect defined by ANSI, using the most compatible rules among popular engines where the standard is ambiguous
//
// Other dialects provided for reference:
//
//     var mssql    = sql.Dialect{IdentOpen: '[', IdentClose: ']', Placeholder: sql.PlaceholderQuestion}
//     var mysql    = sql.Dialect{IdentOpen: '`', IdentClose: '`', Placeholder: sql.PlaceholderColon}
//     var oracle   = sql.Dialect{IdentOpen: '"', IdentClose: '"', Placeholder: sql.PlaceholderColon}
//     var postgres = sql.Dialect{IdentOpen: '"', IdentClose: '"', Placeholder: sql.PlaceholderDollar}
//     var sqlite   = sql.Dialect{IdentOpen: '"', IdentClose: '"', Placeholder: sql.PlaceholderQuestion}
//
var Ansi = Dialect{IdentOpen: '"', IdentClose: '"', Placeholder: PlaceholderQuestion}

// PlaceholderColon generates placeholder names in the form :1, :2, :3
func PlaceholderColon(n int) string { return ":" + strconv.Itoa(n) }

// PlaceholderDollar generates placeholders names in the form $1, $2, $3
func PlaceholderDollar(n int) string { return "$" + strconv.Itoa(n) }

// PlaceholderQuestion always returns the question mark "?" as a placeholder
func PlaceholderQuestion(n int) string { return "?" }

func useDialect(dialect *Dialect) *Dialect {
	if dialect == nil {
		return &Ansi
	} else {
		return dialect
	}
}

func (d *Dialect) WriteIdentifier(buf *bytes.Buffer, ident string) {
	buf.WriteRune(d.IdentOpen)
	buf.WriteString(ident)
	buf.WriteRune(d.IdentClose)
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
