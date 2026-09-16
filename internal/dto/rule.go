package dto

type Rule[T any] struct {
	Path    string
	Options T
}
