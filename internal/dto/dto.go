package dto

import (
	"go/ast"
	"go/token"
	"go/types"
)

// --- Violation ---

type Violation struct {
	Check   string `json:"check"`
	Rule    string `json:"rule"`
	Path    string `json:"path"`
	Pos     string `json:"pos"`
	Message string `json:"message"`
}

// --- Report ---

type Report struct {
	Violations []Violation `json:"violations"`
	Count      int         `json:"count"`
}

// --- Import entry ---

type ImportEntry struct {
	Path string
	Pos  string
	File string
}

// --- AST input ---

type ASTInput struct {
	Root     string
	Fset     *token.FileSet
	Packages []ASTPackage
}

type ASTPackage struct {
	Dir       string
	GoFiles   []string
	Files     []*ast.File
	TypesInfo *types.Info
}
