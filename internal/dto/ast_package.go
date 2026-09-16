package dto

import (
	"go/ast"
	"go/types"
)

type ASTPackage struct {
	Dir       string
	GoFiles   []string
	Files     []*ast.File
	TypesInfo *types.Info
}
