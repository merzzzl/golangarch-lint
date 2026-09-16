package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/controller"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

const (
	mainExitOK         = 0
	mainExitViolations = 1
	mainExitError      = 2
	mainDocsFileMode   = 0o600
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 || (args[0] != "lint" && args[0] != "docs" && args[0] != "migrate") {
		_, _ = fmt.Fprintln(os.Stderr, "usage: golangarch-lint lint|docs|migrate [-config path] [-format text|json] [-output path] [root]")

		os.Exit(mainExitError)
	}

	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configPath := flags.String("config", "", "configuration path")
	format := flags.String("format", "text", "report format: text or json")

	output := flags.String("output", "", "migration destination (default stdout; existing files are never overwritten)")
	if err := flags.Parse(args[1:]); err != nil {
		os.Exit(mainExitError)
	}

	if flags.NArg() > 1 || (*format != "text" && *format != "json") {
		_, _ = fmt.Fprintln(os.Stderr, "invalid arguments")

		os.Exit(mainExitError)
	}

	root := "."
	if flags.NArg() == 1 {
		root = flags.Arg(0)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(mainExitError)
	}

	if args[0] == "migrate" {
		root, source, target := absRoot, *configPath, *output

		if source == "" {
			for _, name := range []string{".golangarch.yml", ".golangarch.yaml"} {
				p := filepath.Join(root, name)
				// #nosec G703 -- Local CLI paths are explicitly selected by the invoking user.
				if _, err := os.Stat(p); err == nil {
					source = p

					break
				}
			}
		}

		// #nosec G703 -- Local CLI paths are explicitly selected by the invoking user.
		data, err := os.ReadFile(source)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			os.Exit(mainExitError)
		}

		migrated, notes, err := (&config.Config{}).Migrate(data)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			os.Exit(mainExitError)
		}

		for _, note := range notes {
			_, _ = fmt.Fprintf(os.Stderr, "migration warning: %s\n", note)
		}

		if target == "" {
			if _, err := os.Stdout.Write(migrated); err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)

				os.Exit(mainExitError)
			}

			os.Exit(mainExitOK)
		}

		// #nosec G703 -- Local CLI paths are explicitly selected by the invoking user.
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mainDocsFileMode)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			os.Exit(mainExitError)
		}

		_, writeErr := file.Write(migrated)
		closeErr := file.Close()

		if writeErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, writeErr)

			os.Exit(mainExitError)
		}

		if closeErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, closeErr)

			os.Exit(mainExitError)
		}

		os.Exit(mainExitOK)
	}

	cfg, err := config.Load(absRoot, *configPath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "config error: %v\n", err)

		os.Exit(mainExitError)
	}

	ctrl := controller.New(cfg)

	if args[0] == "docs" {
		target := filepath.Join(absRoot, "GOLANGARCH.md")
		// #nosec G703 -- Local CLI paths are explicitly selected by the invoking user.
		if err := os.WriteFile(target, []byte(ctrl.Docs()), mainDocsFileMode); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			os.Exit(mainExitError)
		}

		_, _ = fmt.Fprintf(os.Stdout, "GOLANGARCH.md generated at %s\n", target)

		os.Exit(mainExitOK)
	}

	rep := &dto.Report{Violations: []dto.Violation{}}
	if err := ctrl.Lint(absRoot, rep); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)

		os.Exit(mainExitError)
	}

	for i := range rep.Violations {
		if rep.Violations[i].Severity == "warning" {
			rep.WarningCount++
		} else {
			rep.Violations[i].Severity = "error"
			rep.Count++
		}
	}

	sort.Slice(rep.Violations, func(i, j int) bool {
		a, b := rep.Violations[i], rep.Violations[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}

		if a.Pos != b.Pos {
			return a.Pos < b.Pos
		}

		if a.Check != b.Check {
			return a.Check < b.Check
		}

		return a.Message < b.Message
	})

	if *format == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")

		if err := enc.Encode(rep); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)

			os.Exit(mainExitError)
		}
	} else {
		for _, v := range rep.Violations {
			_, _ = fmt.Fprintf(os.Stdout, "%s: %s [%s] %s\n", v.Pos, v.Severity, v.Check, v.Message)
		}

		_, _ = fmt.Fprintf(os.Stdout, "%d error(s), %d warning(s)\n", rep.Count, rep.WarningCount)
	}

	if rep.Count > 0 {
		os.Exit(mainExitViolations)
	}

	os.Exit(mainExitOK)
}
