package signatures

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

type Service struct{ rules []dto.Rule[config.Signatures] }

func New(rules []dto.Rule[config.Signatures]) *Service { return &Service{rules: rules} }
