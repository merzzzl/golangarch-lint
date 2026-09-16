package controller

import (
	"errors"

	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/services/declarations"
	"github.com/merzzzl/golangarch-lint/internal/services/imports"
	"github.com/merzzzl/golangarch-lint/internal/services/layout"
	"github.com/merzzzl/golangarch-lint/internal/services/scope"
	"github.com/merzzzl/golangarch-lint/internal/services/signatures"
	"github.com/merzzzl/golangarch-lint/internal/services/templater"
)

var ErrNoPackages = errors.New("no packages found")

func New(cfg *config.Config) *Controller {
	analyzeCalls := false

	scopes := make([]config.Scope, 0, len(cfg.Rules))
	layouts := make([]dto.Rule[config.Layout], 0, len(cfg.Rules))
	declarationsRules := make([]dto.Rule[config.Declarations], 0, len(cfg.Rules))
	importsRules := make([]dto.Rule[config.Filter], 0, len(cfg.Rules))

	signaturesRules := make([]dto.Rule[config.Signatures], 0, len(cfg.Rules))
	for i := range cfg.Rules {
		r := &cfg.Rules[i]
		if r.Imports.Allow != nil || len(r.Imports.Deny) > 0 {
			analyzeCalls = true
		}

		scopes = append(scopes, r.Scope)
		layouts = append(layouts, dto.Rule[config.Layout]{Path: r.Scope.Path, Options: r.Layout})
		declarationsRules = append(declarationsRules, dto.Rule[config.Declarations]{Path: r.Scope.Path, Options: r.Declarations})
		importsRules = append(importsRules, dto.Rule[config.Filter]{Path: r.Scope.Path, Options: r.Imports})
		signaturesRules = append(signaturesRules, dto.Rule[config.Signatures]{Path: r.Scope.Path, Options: r.Signatures})
	}

	return &Controller{
		analyzeCalls: analyzeCalls,
		scope:        scope.New(scopes),
		layout:       layout.New(layouts),
		declarations: declarations.New(declarationsRules),
		imports:      imports.New(importsRules),
		signatures:   signatures.New(signaturesRules),
		templater:    templater.New(cfg),
	}
}
