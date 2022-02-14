package lexer

import (
	"unicode"
)

type State struct {
	Source []rune
	Pos    Pos
}

func InitState(source string, file string) State {
	return State{
		Source: []rune(source),
		Pos: Pos{
			Line: 1,
			Col:  1,
			File: file,
		},
	}
}

func MakeToken(kind TokKind, start Pos, state State) L[Tok] {
	return L[Tok]{
		Item: Tok{
			Kind:  kind,
			Value: string(state.Source[start.Index:state.Pos.Index]),
		},
		Loc: Span{
			Start: start,
			End:   state.Pos,
			File:  state.Pos.File,
		},
	}
}

func LexToken(st State) (State, L[Tok]) {
	state := &st

	if state.Pos.Index >= len(state.Source) {
		return *state, MakeToken(TokEof, state.Pos, *state)
	}

	start := state.Pos
	char := func() rune { return state.Source[state.Pos.Index] }

	simpleTokens := map[rune]TokKind{
		'(': TokLParen,
		')': TokRParen,
		'{': TokLBrace,
		'}': TokRBrace,
		'+': TokAdd,
		'*': TokMul,
		'=': TokEquals,
		',': TokComma,
		';': TokSemi,
	}

	simpleKind, isSimple := simpleTokens[char()]
	switch {
	case isSimple:
		state.Pos.ExtendCharMut(char())
		return *state, MakeToken(simpleKind, start, *state)
	case unicode.IsNumber(char()):
		for {
			if state.Pos.Index >= len(state.Source) {
				break
			}
			c := char()
			if unicode.IsNumber(c) {
				state.Pos.ExtendCharMut(c)
			} else {
				break
			}
		}
		return *state, MakeToken(TokInt, start, *state)
	case unicode.IsLetter(char()):
		for {
			if state.Pos.Index >= len(state.Source) {
				break
			}
			c := char()
			if unicode.IsLetter(c) {
				state.Pos.ExtendCharMut(c)
			} else {
				break
			}
		}
		var kind TokKind
		switch string(state.Source[start.Index:state.Pos.Index]) {
		case "fn", "return":
			kind = TokKw
			break
		default:
			kind = TokIdent
		}
		return *state, MakeToken(kind, start, *state)
	case unicode.IsSpace(char()):
		state.Pos.ExtendCharMut(char())
		return LexToken(*state)
	default:
		state.Pos.ExtendCharMut(char())
		return *state, MakeToken(TokUnknown, start, *state)
	}
}

func LexAll(state State) []L[Tok] {
	var tokens []L[Tok]
	for {
		next, tok := LexToken(state)
		if tok.Item.Kind == TokEof {
			tokens = append(tokens, tok)
			return tokens
		}
		tokens = append(tokens, tok)
		state = next
	}
}
