package sqltest

import (
	"database/sql/driver"
	"errors"
	"io"

	"github.com/reflexionhealth/vanilla/sql/sqltest/ast"
	"github.com/reflexionhealth/vanilla/sql/sqltest/parser"
)

type Driver struct{}

func (d *Driver) Open(name string) (driver.Conn, error) {
	return new(Conn), nil
}

type Conn struct {
	Closed bool
}

func (c *Conn) Prepare(query string) (driver.Stmt, error) {
	prep := parser.Make([]byte(query), parser.Ruleset{})
	stmt, err := prep.ParseStatement()
	return &Stmt{Ast: stmt}, err
}

func (c *Conn) Close() error {
	// TODO: Return an error if not all Rows created by the connection have been closed
	c.Closed = true
	return nil
}

func (c *Conn) Begin() (driver.Tx, error) {
	return nil, errors.New("TODO: Implement Conn.Begin() for testing of transactions")
}

type Stmt struct {
	Ast    ast.Stmt
	Closed bool
}

func (s *Stmt) Close() error {
	s.Closed = true
	return nil
}

func (s *Stmt) NumInput() int {
	return -1
}

func (s *Stmt) Exec(args []driver.Value) (driver.Result, error) {
	return nil, errors.New("TODO: Implement Stmt.Exec() for testing of INSERTs, UPDATEs")
}

func (s *Stmt) Query(args []driver.Value) (driver.Rows, error) {
	slct, ok := s.Ast.(*ast.SelectStmt)
	if !ok {
		return nil, errors.New("called Query() but statement is not a SELECT")
	}

	var columns []string
	for _, expr := range slct.Selection {
		if ident, ok := expr.(*ast.Identifier); ok {
			columns = append(columns, ident.Name)
		} else {
			columns = append(columns, "")
		}
	}

	return &Rows{columns: columns}, nil
}

type Rows struct {
	Closed  bool
	Scanned int // count of scanned rows

	columns []string
	rows    [][]driver.Value
}

func (r *Rows) Columns() []string {
	return r.columns
}

func (r *Rows) Close() error {
	r.Closed = true
	return nil
}

func (r *Rows) Next(dest []driver.Value) error {
	if r.Scanned < len(r.rows) {
		src := r.rows[r.Scanned]
		for i := range src {
			dest[i] = src[i]
		}
		r.Scanned += 1
		return nil
	} else {
		return io.EOF
	}
}
