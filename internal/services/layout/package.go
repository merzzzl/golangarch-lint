package layout

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

type Service struct{ rules []dto.Rule[config.Layout] }

func New(rules []dto.Rule[config.Layout]) *Service { return &Service{rules: rules} }
