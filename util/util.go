package util

type Pair[A any, B any] struct {
	Fst A
	Snd B
}

func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	m2 := map[K]V{}
	for k, v := range m {
		m2[k] = v
	}
	return m2
}
