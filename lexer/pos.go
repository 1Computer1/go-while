package lexer

type Pos struct {
	Line  int
	Col   int
	Index int
	File  string
}

func ExtendChar(p Pos, char rune) Pos {
	if char == '\n' {
		return Pos{
			Line:  p.Line + 1,
			Col:   1,
			Index: p.Index,
			File:  p.File,
		}
	} else {
		return Pos{
			Line:  p.Line,
			Col:   p.Col + 1,
			Index: p.Index,
			File:  p.File,
		}
	}
}

func (p *Pos) ExtendCharMut(char rune) {
	p.Index++
	if char == '\n' {
		p.Line++
		p.Col = 1
	} else {
		p.Col++
	}
}

func ExtendString(p Pos, str string) Pos {
	for _, char := range str {
		p.ExtendCharMut(char)
	}
	return p
}

func (p *Pos) ExtendStringMut(str string) {
	for _, char := range str {
		p.ExtendCharMut(char)
	}
}

type Span struct {
	Start Pos
	End   Pos
	File  string
}

type L[T any] struct {
	Item T
	Loc  Span
}

type Spanned interface {
	GetSpan() Span
}

func (s Span) GetSpan() Span {
	return s
}

func (x L[T]) GetSpan() Span {
	return x.Loc
}

func Min(x, y Pos) Pos {
	if x.Index < y.Index {
		return x
	} else {
		return y
	}
}

func Max(x, y Pos) Pos {
	if x.Index > y.Index {
		return x
	} else {
		return y
	}
}

func PosToSpan(x, y Pos) Span {
	return Span{
		Start: x,
		End:   y,
		File:  x.File,
	}
}

func Combine(x, y Spanned) Span {
	return Span{
		Start: Min(x.GetSpan().Start, y.GetSpan().End),
		End:   Max(x.GetSpan().End, y.GetSpan().End),
	}
}

func ReplaceL[T any, U any](l L[T], u U) L[U] {
	return L[U]{
		Item: u,
		Loc:  l.Loc,
	}
}

func MapL[T any, U any](l L[T], f func(T) U) L[U] {
	return L[U]{
		Item: f(l.Item),
		Loc:  l.Loc,
	}
}
