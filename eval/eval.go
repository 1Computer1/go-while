package eval

import (
	"errors"
	"fmt"
	"while/lexer"
	. "while/parser/ast"
)

type (
	Val interface {
		value()
	}

	ValInt int

	ValFn struct {
		Closure Scope
		Body    ExprFn
	}
)

func (*ValInt) value() {}
func (*ValFn) value()  {}

type Scope = map[string]Val

type Env struct {
	Scope  Scope
	Parent *Env
}

// TODO: Evaluate ExprCall, ExprFn
// TODO: Evaluate statements and Program

func (env *Env) EvalExpr(expr *lexer.L[Expr]) (Val, error) {
	switch e := expr.Item.(type) {
	case *ExprInt:
		ret := ValInt(e.Value.Item)
		return &ret, nil
	case *ExprVar:
		ret, ok := env.Scope[e.Name.Item]
		if !ok {
			return nil, errors.New(fmt.Sprintf(
				"Variable %s is undefined at %d:%d-%d:%d",
				e.Name.Item,
				e.Name.Loc.Start.Line,
				e.Name.Loc.Start.Col,
				e.Name.Loc.End.Line,
				e.Name.Loc.End.Col,
			))
		}
		return ret, nil
	case *ExprBinary:
		if e.Op.Item == OpAssign {
			switch lhsn := e.Lhs.Item.(type) {
			case *ExprVar:
				rhs, err := env.EvalExpr(&e.Rhs)
				if err != nil {
					return nil, err
				}
				env.Scope[lhsn.Name.Item] = rhs
				return rhs, nil
			}
			return nil, errors.New(fmt.Sprintf(
				"Cannot assign to non-variable at %d:%d-%d:%d",
				e.Lhs.Loc.Start.Line,
				e.Lhs.Loc.Start.Col,
				e.Lhs.Loc.End.Line,
				e.Lhs.Loc.End.Col,
			))
		}
		lhs, err := env.EvalExpr(&e.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := env.EvalExpr(&e.Rhs)
		if err != nil {
			return nil, err
		}
		switch lhsv := lhs.(type) {
		case *ValInt:
			switch rhsv := rhs.(type) {
			case *ValInt:
				var ret ValInt
				switch e.Op.Item {
				case OpAdd:
					ret = ValInt(*lhsv + *rhsv)
					break
				case OpMul:
					ret = ValInt(*lhsv * *rhsv)
					break
				}
				return &ret, nil
			}
		}
		return nil, errors.New(fmt.Sprintf(
			"Type mismatch at %d:%d-%d:%d",
			e.Lhs.Loc.Start.Line,
			e.Lhs.Loc.Start.Col,
			e.Lhs.Loc.End.Line,
			e.Lhs.Loc.End.Col,
		))
	}
	return nil, errors.New("Unsupported expression node")
}
