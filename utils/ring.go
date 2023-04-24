package utils

// Moves the ring circularly, going from [0, 1, 2, 3] -> [1, 2, 3, 0]
func RingCircleMove[T any, E []T](sl E) E {
	ret := make(E, len(sl))
	tmp := sl[0]
	for i, v := range sl[1:] {
		ret[i] = v
	}
	ret[len(ret)-1] = tmp
	return ret
}
