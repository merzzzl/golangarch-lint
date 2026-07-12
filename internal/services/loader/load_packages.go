package loader

import (
	"fmt"

	"golang.org/x/tools/go/packages"
)

// LoadPackages loads and type-checks all Go packages under root.
func (*Service) loadPackages(root string) ([]*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedSyntax |
			packages.NeedTypes |
			packages.NeedTypesInfo |
			packages.NeedImports,
		Dir:   root,
		Tests: false,
	}, "./...")
	if err != nil {
		return nil, fmt.Errorf("loading packages: %w", err)
	}

	for _, pkg := range pkgs {
		for _, e := range pkg.Errors {
			return nil, fmt.Errorf("package error: %w", e)
		}
	}

	return pkgs, nil
}
