package utils

// Returns the greatest common divisor between a & b.
func GreatestCommonDivisor(a, b int) int {
	// condition: b <= a

	if b > a {
		a, b = b, a
	}

	for b != 0 {
		a, b = b, a%b
	}

	if a == 0 {
		return 1
	}

	return a
}
