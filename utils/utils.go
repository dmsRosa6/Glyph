package utils

type Layered interface {
	GetLayer() int
}

func InsertSortLayered[T Layered](arr []T, value T) []T {
	idx := 0
	for idx < len(arr) && value.GetLayer() >= arr[idx].GetLayer() {
		idx++
	}

	arr = append(arr, value)
	copy(arr[idx+1:], arr[idx:])
	arr[idx] = value
	return arr
}

func InsertSort[T any](arr []T, value T, less func(a, b T) bool) []T {
	idx := 0
	for idx < len(arr) && !less(value, arr[idx]) {
		idx++
	}

	arr = append(arr, value)
	copy(arr[idx+1:], arr[idx:])
	arr[idx] = value
	return arr
}

