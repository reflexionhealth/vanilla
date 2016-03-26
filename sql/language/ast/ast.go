package ast

type Stmt interface {
	ImplementsStmt()
}

func (s *SelectStmt) ImplementsStmt() {}
func (s *InsertStmt) ImplementsStmt() {}
func (s *UpdateStmt) ImplementsStmt() {}

type Expr interface {
	ImplementsExpr()
}

func (i *Identifier) ImplementsExpr() {}
func (l *Literal) ImplementsExpr()    {}

type Direction int

const (
	ASC Direction = iota
	DESC
)

type SelectType int

const (
	SELECT_ALL SelectType = iota
	SELECT_DISTINCT
	SELECT_DISTINCTROW
)

type SelectStmt struct {
	Type      SelectType
	Selection []Expr
	Star      bool
	From      Identifier
	Where     Expr
	Having    Expr
	GroupBy   string
	Grouping  Direction
	OrderBy   string
	Ordering  Direction
	Limit     int
	Offset    int
}

type InsertStmt struct{}
type UpdateStmt struct{}

type Identifier struct {
	Name   string
	Quoted bool
}

type Literal struct {
	Raw string
}

type BinaryOperator int

const (
	AND BinaryOperator = iota
	OR
	XOR
	IN
	IS
	LESS
	LESSEQ
	GRTR
	GRTREQ
	EQUAL
	ADD
	SUBTRACT
	MULTIPLY
	DIVIDE
	MODULO
)

type BinaryExpr struct {
	Left  Expr
	Oper  BinaryOperator
	Right Expr
}

type UnaryOperator int

const (
	NOT UnaryOperator = iota
	ISNULL
	NOTNULL
	NEGATIVE
)

type UnaryExpr struct {
	Expr Expr
	Oper UnaryOperator
}
