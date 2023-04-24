package xp

const (
	A = 5
	B = 50
	C = 100
)

func LevelUpRequirement(lvl int) int {
	return A*(lvl*lvl) + (B * lvl) + C
}

// The total amount of XP that is required for lvl
func LvlXP(lvl int) int {
	// req = a*n^2 + b*n + c, where n = lvl
	// Sn1 = a*(n(n+1)(2n+1)/6), where n = lvl - 1, since minimum value of lvl = 0
	// Sn2 = n(U1+Un)/2, where n = lvl, because fuck you <3
	// Sn2 = n(c+(c+b*(n-1)))/2
	// Sn = ((c+(c+b*(n-1))) * n/2) + (a*(n-1)*(n)*(2*(n-1)+1)/6), where n = lvl
	// Sn = \frac{n}{2}(2c+b(n-1))+\frac{a}{6}n(n-1)(2n-1)
	// Sn = nc + \frac{b(n^2-n)}{2}+\frac{a}{6}n(n-1)(2n-1)
	// Sn = nc + (n^2-n)(\frac{b}{2}+\frac{a(2n-1)}{6})
	// Sn = nc + (n^2-n)(\frac{3b+a(2n-1)}{6})
	// Sn = n(c+\frac{(n-1)(3b+a(2n-1))}{6})

	n := lvl

	return int(float64(n) * (float64(A) + float64((n-1)*(3*B+C*(2*n-1)))/6))
}
