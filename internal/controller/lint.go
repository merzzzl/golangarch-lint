package controller

import (
	"fmt"

	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Controller) Lint(root string, rep *dto.Report) error {
	if err := s.fsCheck.Run(root, rep); err != nil {
		return fmt.Errorf("fs check: %w", err)
	}

	input, imports, err := s.loader.Run(root)
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}

	s.astCheck.Run(input, rep)
	s.importCheck.Run(imports, rep)

	return nil
}
