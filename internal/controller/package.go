package controller

import (
	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/services/astcheck"
	"github.com/merzzzl/golangarch-lint/internal/services/fscheck"
	"github.com/merzzzl/golangarch-lint/internal/services/importcheck"
	"github.com/merzzzl/golangarch-lint/internal/services/loader"
	"github.com/merzzzl/golangarch-lint/internal/services/templater"
)

type Controller struct {
	loader      *loader.Service
	fsCheck     *fscheck.Service
	astCheck    *astcheck.Service
	importCheck *importcheck.Service
	templater   *templater.Service
}

func New(cfg *config.Config) *Controller {
	return &Controller{
		loader:      loader.New(),
		fsCheck:     fscheck.New(cfg),
		astCheck:    astcheck.New(cfg),
		importCheck: importcheck.New(cfg),
		templater:   templater.New(cfg),
	}
}
