package main

import (
	"testing"
	"while/eval"
	"while/lexer"
	"while/parser"
)

const source = `
let a = 2 + 1 * 2;
print(a);
a = a + 2;
print(a);
let f = fn(x) {
	return x + 2;
};
print(f(a));
`

func TestMain(t *testing.T) {
	state := lexer.InitState(source, "test")
	parser := parser.New(state)
	prog, err := parser.ParseProgram()
	if err != nil {
		t.Error(err)
		return
	}
	env := eval.TopLevel()
	err = env.EvalProgram(prog)
	if err != nil {
		t.Error(err)
		return
	}
}
