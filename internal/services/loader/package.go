package loader

import "errors"

type Service struct{}

var ErrNoPackages = errors.New("no packages found")

func New() *Service {
	return &Service{}
}
