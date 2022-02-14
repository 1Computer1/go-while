package eval

import (
	"errors"
	"fmt"
	"while/lexer"
	. "while/parser/ast"
	"while/util"
)

type (
	Val interface {
		value()
	}

	ValNil struct{}

	ValInt struct {
		Value int
	}

	ValFn struct {
		Closure *Scope
		Body    *ExprFn
	}

	ValPrimFn struct {
		Len  int
		Body func([]Val) (Val, error)
	}
)

func (*ValNil) value()    {}
func (*ValInt) value()    {}
func (*ValFn) value()     {}
func (*ValPrimFn) value() {}

type VarMap = map[string]Val

type Scope struct {
	Vars   VarMap
	Parent *Scope
}

type ControlKind int

const (
	ControlNone ControlKind = iota
	ControlReturn
	ControlErr
)

func TopLevel() Scope {
	return Scope{
		Vars: map[string]Val{
			"print": &ValPrimFn{
				Len: 1,
				Body: func(args []Val) (Val, error) {
					x := args[0]
					switch v := x.(type) {
					case *ValInt:
						fmt.Printf("%d\n", v.Value)
					case *ValFn:
						fmt.Printf("[function]\n")
					case *ValPrimFn:
						fmt.Printf("[primitive function]\n")
					}
					return &ValNil{}, nil
				},
			},
		},
		Parent: nil,
	}
}

func (scope *Scope) getVar(name string) (Val, bool) {
	val, ok := scope.Vars[name]
	if !ok {
		if scope.Parent != nil {
			return scope.Parent.getVar(name)
		}
		return nil, false
	}
	return val, true
}

func (scope *Scope) setVar(name string, val Val) bool {
	_, ok := scope.Vars[name]
	if !ok {
		if scope.Parent != nil {
			return scope.Parent.setVar(name, val)
		}
		return false
	}
	scope.Vars[name] = val
	return true
}

func (scope *Scope) EvalProgram(prog *Program) error {
	_, _, err := scope.evalStmts(prog.Body)
	return err
}

func (scope *Scope) evalStmts(stmts []lexer.L[Stmt]) (ControlKind, Val, error) {
	for _, stmt := range stmts {
		ctrl, val, err := scope.evalStmt(&stmt)
		if err != nil {
			return ControlErr, nil, err
		}
		if ctrl == ControlReturn {
			return ControlReturn, val, nil
		}
	}
	return ControlNone, &ValNil{}, nil
}

func (scope *Scope) evalStmt(stmt *lexer.L[Stmt]) (ControlKind, Val, error) {
	switch s := stmt.Item.(type) {
	case *StmtExpr:
		_, err := scope.EvalExpr(&s.Body)
		if err != nil {
			return ControlErr, nil, err
		}
		return ControlNone, &ValNil{}, nil
	case *StmtLet:
		_, ok := scope.Vars[s.Name.Item]
		if ok {
			return ControlNone, nil, errors.New(fmt.Sprintf(
				"Variable %s is redefined at %d:%d-%d:%d",
				s.Name.Item,
				s.Name.Loc.Start.Line,
				s.Name.Loc.Start.Col,
				s.Name.Loc.End.Line,
				s.Name.Loc.End.Col,
			))
		}
		val, err := scope.EvalExpr(&s.Body)
		if err != nil {
			return ControlErr, nil, err
		}
		scope.Vars[s.Name.Item] = val
		return ControlNone, &ValNil{}, nil
	case *StmtReturn:
		val, err := scope.EvalExpr(&s.Body)
		if err != nil {
			return ControlErr, nil, err
		}
		return ControlReturn, val, nil
	}
	return ControlErr, nil, errors.New("Unsupported statement node")
}

func (scope *Scope) EvalExpr(expr *lexer.L[Expr]) (Val, error) {
	switch e := expr.Item.(type) {
	case *ExprInt:
		ret := ValInt{Value: e.Value.Item}
		return &ret, nil
	case *ExprVar:
		ret, ok := scope.getVar(e.Name.Item)
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
				rhs, err := scope.EvalExpr(&e.Rhs)
				if err != nil {
					return nil, err
				}
				ok := scope.setVar(lhsn.Name.Item, rhs)
				if !ok {
					return nil, errors.New(fmt.Sprintf(
						"Variable %s is undefined at %d:%d-%d:%d",
						lhsn.Name.Item,
						lhsn.Name.Loc.Start.Line,
						lhsn.Name.Loc.Start.Col,
						lhsn.Name.Loc.End.Line,
						lhsn.Name.Loc.End.Col,
					))
				}
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
		lhs, err := scope.EvalExpr(&e.Lhs)
		if err != nil {
			return nil, err
		}
		rhs, err := scope.EvalExpr(&e.Rhs)
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
					ret = ValInt{Value: lhsv.Value + rhsv.Value}
					break
				case OpMul:
					ret = ValInt{Value: lhsv.Value * rhsv.Value}
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
	case *ExprCall:
		f, err := scope.EvalExpr(&e.Func)
		if err != nil {
			return nil, err
		}
		var vs []Val
		for _, arg := range e.Args.Item {
			v, err := scope.EvalExpr(&arg)
			if err != nil {
				return nil, err
			}
			vs = append(vs, v)
		}
		switch fv := f.(type) {
		case *ValFn:
			if len(vs) != len(fv.Body.Args.Item) {
				return nil, errors.New(fmt.Sprintf(
					"Invalid function call at %d:%d-%d:%d",
					e.Args.Loc.Start.Line,
					e.Args.Loc.Start.Col,
					e.Args.Loc.End.Line,
					e.Args.Loc.End.Col,
				))
			}
			vars := util.CopyMap(fv.Closure.Vars)
			for i, v := range vs {
				vars[fv.Body.Args.Item[i].Item] = v
			}
			fscope := Scope{Vars: vars, Parent: fv.Closure.Parent}
			ctrl, val, err := fscope.evalStmts(fv.Body.Body.Item)
			if err != nil {
				return nil, err
			}
			if ctrl == ControlReturn {
				return val, nil
			}
			return &ValNil{}, nil
		case *ValPrimFn:
			if len(vs) != fv.Len {
				return nil, errors.New(fmt.Sprintf(
					"Invalid function call at %d:%d-%d:%d",
					e.Args.Loc.Start.Line,
					e.Args.Loc.Start.Col,
					e.Args.Loc.End.Line,
					e.Args.Loc.End.Col,
				))
			}
			return fv.Body(vs)
		}
		return nil, errors.New(fmt.Sprintf(
			"Not a function at %d:%d-%d:%d",
			e.Func.Loc.Start.Line,
			e.Func.Loc.Start.Col,
			e.Func.Loc.End.Line,
			e.Func.Loc.End.Col,
		))
	case *ExprFn:
		return &ValFn{Closure: scope, Body: e}, nil
	}
	return nil, errors.New("Unsupported expression node")
}
