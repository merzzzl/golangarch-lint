package config

import "errors"

var (
	ErrConfigNotFound   = errors.New("config not found")
	ErrPathRequired     = errors.New("path is required")
	ErrOverlappingPaths = errors.New("overlapping paths")
	ErrInvalidMode      = errors.New("invalid mode")
	ErrInvalidAllowMode = errors.New("invalid allow mode")
	ErrModuleNotFound   = errors.New("module path not found in go.mod")
)
