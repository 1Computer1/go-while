package parser

import (
	"errors"
	"fmt"
	"strconv"
	"while/lexer"
	. "while/lexer"
	. "while/parser/ast"
	"while/util"
)

// TODO: Parse ExprFn, check that ExprCall actually works
// TODO: Parse statements and Program

type Parser struct {
	State State
	Index int
}

func (p *Parser) matchToken(kind TokKind) (*L[Tok], error) {
	next, t := lexer.LexToken(p.State)
	if t.Item.Kind == kind {
		p.State = next
		p.Index++
		return &t, nil
	}
	// TODO: Map token kind to readable string
	return nil, errors.New(fmt.Sprintf("Unexpected token at %d:%d", p.State.Pos.Col, p.State.Pos.Line))
}

func (p *Parser) ParseExpr() (*L[Expr], error) {
	return p.parseBinaryAdd()
}

func (p *Parser) parseInt() (*L[Expr], error) {
	t, err := p.matchToken(TokInt)
	if err != nil {
		return nil, err
	}
	ret := ReplaceL(*t, Expr(&ExprInt{
		Value: MapL(*t, func(t Tok) int {
			x, _ := strconv.Atoi(t.Value)
			return x
		}),
	}))
	return &ret, nil
}

func (p *Parser) parseVar() (*L[Expr], error) {
	id, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	ret := ReplaceL(*id, Expr(&ExprVar{
		Name: *id,
	}))
	return &ret, nil
}

func (p *Parser) parseBinaryAdd() (*L[Expr], error) {
	return p.parseBinaryGeneric(map[Op]TokKind{
		OpAdd: TokAdd,
	}, (*Parser).parseBinaryMul)
}

func (p *Parser) parseBinaryMul() (*L[Expr], error) {
	return p.parseBinaryGeneric(map[Op]TokKind{
		OpMul: TokMul,
	}, (*Parser).parseCall)
}

func (p *Parser) parseBinaryGeneric(table map[Op]TokKind, lower func(*Parser) (*L[Expr], error)) (*L[Expr], error) {
	e1, err := lower(p)
	if err != nil {
		return nil, err
	}
	var es []util.Pair[*L[Expr], *L[Op]]
	for {
		var lastErr error
		var lop L[Op]
		for op, kind := range table {
			tok, err := p.matchToken(kind)
			if err != nil {
				lastErr = err
			} else {
				lop = ReplaceL(*tok, op)
				lastErr = nil
				break
			}
		}
		if lastErr != nil {
			break
		}
		e, err := lower(p)
		if err != nil {
			return nil, err
		}
		p := util.Pair[*L[Expr], *L[Op]]{Fst: e, Snd: &lop}
		es = append(es, p)
	}
	b := *e1
	for _, pair := range es {
		b = L[Expr]{
			Item: &ExprBinary{Lhs: b, Op: *pair.Snd, Rhs: *pair.Fst},
			Loc:  Combine(b, pair.Fst),
		}
	}
	return &b, nil
}

func (p *Parser) parseCall() (*L[Expr], error) {
	e, err := p.parseAtom()
	if err == nil {
		return e, nil
	}
	var calls []L[[]L[Expr]]
	for {
		lp, err := p.matchToken(TokLParen)
		if err != nil {
			break
		}
		var args []L[Expr]
		expr1, err := p.ParseExpr()
		if err == nil {
			args = append(args, *expr1)
			for {
				_, err := p.matchToken(TokComma)
				if err != nil {
					break
				}
				expr, err := p.ParseExpr()
				if err != nil {
					return nil, err
				}
				args = append(args, *expr)
			}
			p.matchToken(TokComma)
		}
		rp, err := p.matchToken(TokRParen)
		if err != nil {
			return nil, err
		}
		calls = append(calls, L[[]L[Expr]]{Item: args, Loc: Combine(lp, rp)})
	}
	expr := *e
	for _, call := range calls {
		expr = L[Expr]{
			Item: &ExprCall{Func: expr, Args: call},
			Loc:  Combine(expr, call),
		}
	}
	return &expr, nil
}

func (p *Parser) parseAtom() (*L[Expr], error) {
	e, err := p.parseInt()
	if err == nil {
		return e, nil
	}
	e, err = p.parseVar()
	if err == nil {
		return e, nil
	}
	return nil, err
}

func (p *Parser) parseIdent() (*L[string], error) {
	t, err := p.matchToken(TokIdent)
	if err != nil {
		return nil, err
	}
	ret := MapL(*t, func(t Tok) string {
		return t.Value
	})
	return &ret, nil
}
