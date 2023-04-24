package utils

// Returns the greatest common divisor between a & b.
func GreatestCommonDivisor(a, b int) int {
	if b > a {
		a, b = b, a
	}

	r := a % b
	if r == 0 {
		return b
	}

	return GreatestCommonDivisor(b, r)
}
