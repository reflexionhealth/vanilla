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

func (e *BinaryExpr) ImplementsExpr() {}
func (e *UnaryExpr) ImplementsExpr()  {}
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
	DISTINCT
	DISTINCT_ROW
)

type SelectStmt struct {
	Type     SelectType
	Select   []Expr
	Star     bool
	From     *Identifier
	Where    Expr
	Having   Expr
	GroupBy  string
	Grouping Direction
	OrderBy  string
	Ordering Direction
	Limit    int
	Offset   int
}

type InsertStmt struct{}
type UpdateStmt struct{}

type Identifier struct {
	Name   string
	Quoted bool
}

func Name(name string) *Identifier   { return &Identifier{name, false} }
func Quoted(name string) *Identifier { return &Identifier{name, true} }

type Literal struct {
	Raw string
}

func Lit(raw string) *Literal { return &Literal{raw} }

type BinaryExpr struct {
	Left     Expr
	Operator OpType
	Right    Expr
}

func Binary(left Expr, op OpType, right Expr) *BinaryExpr {
	return &BinaryExpr{left, op, right}
}

type UnaryOperator int

const ()

type UnaryExpr struct {
	Operator OpType
	Subexpr  Expr
}

func Unary(op OpType, subexpr Expr) *UnaryExpr {
	return &UnaryExpr{op, subexpr}
}
