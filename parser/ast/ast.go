package parser

import (
	. "while/lexer"
)

type Op int

const (
	OpAdd Op = iota
	OpMul
	OpAssign
)

type (
	Program struct {
		Body L[[]L[Stmt]]
	}

	Stmt interface {
		stmt()
	}

	// StmtExpr <body>;
	StmtExpr struct {
		Body L[Expr]
	}

	// StmtLet let <name> = <body>;
	StmtLet struct {
		Name L[string]
		Body L[Expr]
	}

	Expr interface {
		expr()
	}

	// ExprInt <value>
	ExprInt struct {
		Value L[int]
	}

	// ExprVar <name>
	ExprVar struct {
		Name L[string]
	}

	// ExprCall <func>(<args>, ...)
	ExprCall struct {
		Func L[Expr]
		Args L[[]L[Expr]]
	}

	// ExprBinary <lhs> <op> <rhs>
	ExprBinary struct {
		Lhs L[Expr]
		Op  L[Op]
		Rhs L[Expr]
	}

	// ExprFn fn(<args>, ...) { <body> }
	ExprFn struct {
		Args L[[]L[string]]
		Body L[[]L[Stmt]]
	}
)

func (*StmtExpr) stmt() {}
func (*StmtLet) stmt()  {}

func (*ExprInt) expr()    {}
func (*ExprVar) expr()    {}
func (*ExprCall) expr()   {}
func (*ExprBinary) expr() {}
func (*ExprFn) expr()     {}
