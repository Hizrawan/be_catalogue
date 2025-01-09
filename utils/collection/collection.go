package collection

func Contains[T comparable](slice []T, value T) bool {
	for _, val := range slice {
		if val == value {
			return true
		}
	}
	return false
}

func ContainsFunc[T any](slice []T, f func(T) bool) bool {
	for _, val := range slice {
		if f(val) {
			return true
		}
	}
	return false
}

func ContainsKey[T comparable, V any](m map[T]V, key T) bool {
	_, ok := m[key]
	return ok
}

func ContainsValue[T comparable, K comparable](m map[K]T, value T) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}
