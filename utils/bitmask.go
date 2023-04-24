package utils

// Returns true if `req` is in `opts`
func BitMask[T ~int](opts, req T) bool {
	return req&opts == req
}
