package utils

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

func Find[T any](ss []T, test func(T) bool) (ret *T) {
	for _, s := range ss {
		if test(s) {
			return &s
		}
	}
	return nil
}

func IndexOf[T comparable](ss []T, s T) int {
	for i, v := range ss {
		if v == s {
			return i
		}
	}
	return -1
}
