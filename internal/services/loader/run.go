package loader

import (
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

// Run loads Go packages under root and converts them into check inputs.
func (s *Service) Run(root string) (*dto.ASTInput, map[string][]dto.ImportEntry, error) {
	pkgs, err := s.loadPackages(root)
	if err != nil {
		return nil, nil, err
	}

	if len(pkgs) == 0 {
		return nil, nil, ErrNoPackages
	}

	return s.buildASTInput(root, pkgs), s.collectImports(root, pkgs), nil
}
