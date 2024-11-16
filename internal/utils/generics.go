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

func Contains[T comparable](ss []T, s T) bool {
	return IndexOf(ss, s) > -1
}

func RemoveWithoutOrder[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func MergeMaps[T1 comparable, T2 any](maps ...map[T1]T2) map[T1]T2 {
	result := make(map[T1]T2)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
