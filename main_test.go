package main

import (
	"testing"
	"while/eval"
	"while/lexer"
	"while/parser"

	"github.com/gdexlab/go-render/render"
)

func TestMain(t *testing.T) {
	state := lexer.InitState("1 + 4 + 3", "test")
	parser := parser.Parser{State: state, Index: 0}
	expr, err := parser.ParseExpr()
	if err != nil {
		return
	}
	env := eval.Env{Scope: map[string]eval.Val{}, Parent: nil}
	val, err := env.EvalExpr(expr)
	println(render.AsCode(val))
}
