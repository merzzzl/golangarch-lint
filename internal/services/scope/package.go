package scope

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
)

type Service struct{ rules []config.Scope }

func New(rules []config.Scope) *Service { return &Service{rules: rules} }
