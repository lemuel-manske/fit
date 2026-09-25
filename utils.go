package fit

func Map[T, U any](data []T, f func(T) U) []U {
	res := make([]U, 0, len(data))

	for _, e := range data {
		res = append(res, f(e))
	}

	return res
}

func MapValues[K comparable, V, U any](data map[K]V, f func(V) U) map[K]U {
	res := make(map[K]U, len(data))

	for k, v := range data {
		res[k] = f(v)
	}

	return res
}
