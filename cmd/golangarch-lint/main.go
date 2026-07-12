package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/controller"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

type mainArgs struct {
	root   string
	config string
	format string
}

const (
	mainExitOK         = 0
	mainExitViolations = 1
	mainExitError      = 2

	mainDocsFileMode = 0o600
)

func main() {
	if len(os.Args) < 2 || (os.Args[1] != "lint" && os.Args[1] != "docs") {
		_, _ = fmt.Fprintln(os.Stderr, "usage: golangarch-lint lint [-config path] [-format text|json] [root]")
		_, _ = fmt.Fprintln(os.Stderr, "       golangarch-lint docs [-config path] [root]")

		os.Exit(mainExitError)
	}

	cmd := os.Args[1]

	args := mainArgs{
		root:   ".",
		format: "text",
	}

	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-config":
			i++

			if i < len(os.Args) {
				args.config = os.Args[i]
			}
		case "-format":
			i++

			if i < len(os.Args) {
				args.format = os.Args[i]
			}
		default:
			args.root = os.Args[i]
		}
	}

	absRoot, err := filepath.Abs(args.root)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(mainExitError)
	}

	cfg, err := config.Load(absRoot, args.config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config error: %v\n", err)

		os.Exit(mainExitError)
	}

	ctrl := controller.New(cfg)

	if cmd == "docs" {
		target := filepath.Join(absRoot, "GOLANGARCH.md")

		if err := os.WriteFile(target, []byte(ctrl.Docs()), mainDocsFileMode); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error writing GOLANGARCH.md: %v\n", err)

			os.Exit(mainExitError)
		}

		_, _ = fmt.Fprintf(os.Stdout, "GOLANGARCH.md generated at %s\n", target)

		return
	}

	rep := &dto.Report{Violations: []dto.Violation{}}

	if err := ctrl.Lint(absRoot, rep); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(mainExitError)
	}

	sort.Slice(rep.Violations, func(i, j int) bool {
		if rep.Violations[i].Path != rep.Violations[j].Path {
			return rep.Violations[i].Path < rep.Violations[j].Path
		}

		return rep.Violations[i].Pos < rep.Violations[j].Pos
	})

	rep.Count = len(rep.Violations)

	switch args.format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(rep); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error writing json: %v\n", err)

			os.Exit(mainExitError)
		}
	default:
		for _, v := range rep.Violations {
			_, _ = fmt.Fprintf(os.Stdout, "%s: [%s] %s\n", v.Pos, v.Check, v.Message)
		}

		_, _ = fmt.Fprintf(os.Stdout, "%d violation(s)\n", rep.Count)
	}

	if rep.Count > 0 {
		os.Exit(mainExitViolations)
	}
}
