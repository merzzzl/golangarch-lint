package importcheck

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
)

type Service struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Service {
	return &Service{cfg: cfg}
}
