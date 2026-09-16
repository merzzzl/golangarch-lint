package dto

import (
	"go/token"

	"golang.org/x/tools/go/packages"
)

type ASTInput struct {
	LoadedPackages []*packages.Package
	Root           string
	Fset           *token.FileSet
	Packages       []ASTPackage
}
