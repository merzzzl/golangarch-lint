package scope

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Service) Run(root string, rep *dto.Report) (*dto.Scope, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving root: %w", err)
	}

	result := &dto.Scope{Root: root, Files: map[string][]int{}, FileDirectories: map[string][]int{}, Subdirs: map[string][]int{}, Directories: map[string][]int{}}

	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return fmt.Errorf("relative path: %w", err)
		}

		rel = filepath.ToSlash(rel)
		if rel == "." {
			result.Directories[rel], _ = s.directoryRules(rel, rel)

			return nil
		}

		if strings.HasPrefix(entry.Name(), ".") {
			if entry.IsDir() {
				return filepath.SkipDir
			}

			return nil
		}

		parent := filepath.ToSlash(filepath.Dir(rel))

		directoryRules, covered := s.directoryRules(parent, rel)
		if entry.IsDir() {
			result.Subdirs[rel] = directoryRules
			result.Directories[rel], _ = s.directoryRules(rel, rel)

			return nil
		}

		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
			return nil
		}

		if !covered {
			rep.Violations = append(rep.Violations, dto.Violation{Severity: "error", Check: "fs-coverage", Path: rel, Pos: rel, Message: fmt.Sprintf("directory %q is not covered by any scope", parent)})
		}

		result.Files[rel] = s.selectRules(rel, parent)
		result.FileDirectories[rel] = directoryRules

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scoping project: %w", err)
	}

	return result, nil
}

func (s *Service) directoryRules(dir, path string) ([]int, bool) {
	var selected []int

	covered := false

	for i, r := range s.rules {
		if r.Matches(dir) {
			covered = true

			if !r.Ignored(path) {
				selected = append(selected, i)
			}
		}
	}

	return selected, covered
}

func (s *Service) selectRules(file, dir string) []int {
	var files, dirs []int

	for i, r := range s.rules {
		if r.Matches(file) {
			files = append(files, i)
		} else if r.Matches(dir) {
			dirs = append(dirs, i)
		}
	}

	selected := dirs
	if len(files) > 0 {
		selected = files
	}

	var active []int

	for _, i := range selected {
		if !s.rules[i].Ignored(file) {
			active = append(active, i)
		}
	}

	return active
}
