package declarations

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

type Service struct {
	rules []dto.Rule[config.Declarations]
}

func New(rules []dto.Rule[config.Declarations]) *Service { return &Service{rules: rules} }
