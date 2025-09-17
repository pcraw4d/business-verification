package classification

// absFloat64 returns the absolute value of a float64
func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// maxFloat64 returns the maximum of two float64 values
func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
