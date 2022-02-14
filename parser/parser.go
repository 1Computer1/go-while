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

type Parser struct {
	State State
	Index int
}

func New(state lexer.State) Parser {
	return Parser{
		State: state,
		Index: 0,
	}
}

func (p *Parser) matchToken(kind TokKind) (*L[Tok], error) {
	next, t := lexer.LexToken(p.State)
	if t.Item.Kind == kind {
		p.State = next
		p.Index++
		return &t, nil
	}
	// TODO: Map token kind to readable string
	return nil, errors.New(fmt.Sprintf("Unexpected token at %d:%d", p.State.Pos.Line, p.State.Pos.Col))
}

func (p *Parser) matchKw(kw string) (*L[Tok], error) {
	next, t := lexer.LexToken(p.State)
	if t.Item.Kind == TokKw && t.Item.Value == kw {
		p.State = next
		p.Index++
		return &t, nil
	}
	// TODO: Map token kind to readable string
	return nil, errors.New(fmt.Sprintf("Unexpected token at %d:%d", p.State.Pos.Line, p.State.Pos.Col))
}

func (p *Parser) ParseProgram() (*Program, error) {
	ss, err := p.parseStmts()
	if err != nil {
		return nil, err
	}
	_, err = p.matchToken(TokEof)
	if err != nil {
		return nil, err
	}
	return &Program{Body: ss}, nil
}

func (p *Parser) parseStmts() ([]L[Stmt], error) {
	var ss []L[Stmt]
	for {
		s, err := p.parseStmt()
		if err != nil {
			break
		}
		ss = append(ss, *s)
	}
	return ss, nil
}

func (p *Parser) parseStmt() (*L[Stmt], error) {
	s, err := p.parseStmtLet()
	if err == nil {
		return s, nil
	}
	s, err = p.parseStmtReturn()
	if err == nil {
		return s, nil
	}
	s, err = p.parseStmtExpr()
	if err == nil {
		return s, nil
	}
	return nil, err
}

func (p *Parser) parseStmtExpr() (*L[Stmt], error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	semi, err := p.matchToken(TokSemi)
	if err != nil {
		return nil, err
	}
	ret := L[Stmt]{
		Item: &StmtExpr{Body: *expr},
		Loc:  lexer.Combine(expr, semi),
	}
	return &ret, nil
}

func (p *Parser) parseStmtLet() (*L[Stmt], error) {
	lettok, err := p.matchKw("let")
	if err != nil {
		return nil, err
	}
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	_, err = p.matchToken(TokEquals)
	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	semi, err := p.matchToken(TokSemi)
	if err != nil {
		return nil, err
	}
	ret := L[Stmt]{
		Item: &StmtLet{Name: *name, Body: *expr},
		Loc:  lexer.Combine(lettok, semi),
	}
	return &ret, nil
}

func (p *Parser) parseStmtReturn() (*L[Stmt], error) {
	lettok, err := p.matchKw("return")
	if err != nil {
		return nil, err
	}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	semi, err := p.matchToken(TokSemi)
	if err != nil {
		return nil, err
	}
	ret := L[Stmt]{
		Item: &StmtReturn{Body: *expr},
		Loc:  lexer.Combine(lettok, semi),
	}
	return &ret, nil
}

func (p *Parser) ParseFullExpr() (*L[Expr], error) {
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	_, err = p.matchToken(TokEof)
	if err != nil {
		return nil, err
	}
	return expr, nil
}

func (p *Parser) parseExpr() (*L[Expr], error) {
	return p.parseBinaryAssign()
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

func (p *Parser) parseBinaryAssign() (*L[Expr], error) {
	return p.parseBinaryGeneric(map[Op]TokKind{
		OpAssign: TokEquals,
	}, (*Parser).parseBinaryAdd)
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
	if err != nil {
		return nil, err
	}
	var calls []L[[]L[Expr]]
	for {
		lp, err := p.matchToken(TokLParen)
		if err != nil {
			break
		}
		var args []L[Expr]
		expr1, err := p.parseExpr()
		if err == nil {
			args = append(args, *expr1)
			for {
				_, err := p.matchToken(TokComma)
				if err != nil {
					break
				}
				expr, err := p.parseExpr()
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
	e, err := p.parseFn()
	if err == nil {
		return e, nil
	}
	e, err = p.parseInt()
	if err == nil {
		return e, nil
	}
	e, err = p.parseVar()
	if err == nil {
		return e, nil
	}
	return nil, err
}

func (p *Parser) parseFn() (*L[Expr], error) {
	fntok, err := p.matchKw("fn")
	if err != nil {
		return nil, err
	}
	lp, err := p.matchToken(TokLParen)
	if err != nil {
		return nil, err
	}
	var args []L[string]
	arg1, err := p.parseIdent()
	if err == nil {
		args = append(args, *arg1)
		for {
			_, err := p.matchToken(TokComma)
			if err != nil {
				break
			}
			arg, err := p.parseIdent()
			if err != nil {
				return nil, err
			}
			args = append(args, *arg)
		}
		p.matchToken(TokComma)
	}
	rp, err := p.matchToken(TokRParen)
	if err != nil {
		return nil, err
	}
	lb, err := p.matchToken(TokLBrace)
	if err != nil {
		return nil, err
	}
	ss, err := p.parseStmts()
	if err != nil {
		return nil, err
	}
	rb, err := p.matchToken(TokRBrace)
	if err != nil {
		return nil, err
	}
	ret := L[Expr]{
		Item: &ExprFn{
			Args: L[[]L[string]]{
				Item: args,
				Loc:  lexer.Combine(lp, rp),
			},
			Body: L[[]L[Stmt]]{
				Item: ss,
				Loc:  lexer.Combine(lb, rb),
			},
		},
		Loc: lexer.Combine(fntok, rp),
	}
	return &ret, nil
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
