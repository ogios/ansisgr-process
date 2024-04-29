package process

// if endIndex bigger than length, return copy of rest.
// mainly for cut slice easier
func SliceFrom[S ~[]E, E comparable](src S, start, end int) S {
	inputLen := len(src)
	if start < inputLen {
		if end < inputLen {
			s := src[start:end]
			to := make(S, len(s))
			copy(to, s)
			return to
		} else {
			s := src[start:]
			to := make(S, len(s))
			copy(to, s)
			return to
		}
	}
	return make(S, 0)
}
