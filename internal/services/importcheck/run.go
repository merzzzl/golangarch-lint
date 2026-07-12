package importcheck

import (
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Service) Run(imports map[string][]dto.ImportEntry, rep *dto.Report) {
	for pkgRel, entries := range imports {
		for _, imp := range entries {
			s.importCheckSingle(pkgRel, imp, rep)
		}
	}
}
