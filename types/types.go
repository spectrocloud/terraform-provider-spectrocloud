package types

func Ptr[T any](v T) *T {
	return &v
}

func Val[T any](v *T) T {
	return *v
}
