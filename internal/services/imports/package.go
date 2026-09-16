package imports

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

type Service struct{ rules []dto.Rule[config.Filter] }

func New(rules []dto.Rule[config.Filter]) *Service { return &Service{rules: rules} }
