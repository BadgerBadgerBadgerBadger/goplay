package util

// ApplyOnType takes the address of type `t` and an array of functions
// that do something with this type, and applies the functions to the type.
func ApplyOnType[T any](t *T, tFuncs []func(*T)) {
	for _, optFunc := range tFuncs {
		optFunc(t)
	}
}
