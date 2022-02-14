package lexer

type TokKind int

const (
	TokUnknown TokKind = iota
	TokInt
	TokKw
	TokIdent
	TokLParen
	TokRParen
	TokLBrace
	TokRBrace
	TokEquals
	TokAdd
	TokMul
	TokComma
	TokSemi
	TokEof
)

type Tok struct {
	Kind  TokKind
	Value string
}
