package fscheck

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Service) Run(root string, rep *dto.Report) error {
	err := filepath.Walk(root, func(p string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)

		if rel == "." {
			return nil
		}

		if s.fsCheckIsHidden(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if s.fsCheckIsIgnored(rel, s.cfg.Ignore) {
			if info.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		if info.IsDir() {
			s.fsCheckDir(rel, rep)

			return nil
		}

		s.fsCheckFile(rel, rep)

		return nil
	})
	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}
