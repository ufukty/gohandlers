package produce

func ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

// use with symbol tables
func declare(variable *bool) {
	*variable = true
}
