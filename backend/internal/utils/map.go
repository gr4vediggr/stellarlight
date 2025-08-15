package utils

func MapToSlice[T any, M comparable](m map[M]T) []T {
	slice := make([]T, 0, len(m))
	for _, v := range m {
		slice = append(slice, v)
	}
	return slice
}
