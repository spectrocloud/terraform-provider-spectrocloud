package ptr

func To[T any](v T) *T {
	return &v
}

func DeRef[T any](v *T) T {
	return *v
}
