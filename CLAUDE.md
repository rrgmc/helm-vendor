# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`helm-vendor` is a Go CLI (Go 1.25, module `github.com/rrgmc/helm-vendor`) for vendoring Helm charts into a repo, driven by a `helm-vendor.yaml` config. It uses the Helm v3 Go SDK (`helm.sh/helm/v3`) directly, not the `helm` binary, and `urfave/cli/v3` for the command line.

## Commands

```shell
go build ./...                 # build
go vet ./...                   # static checks (no separate linter is configured)
go test ./...                  # tests (also `task test`); there are no _test.go files yet
go test ./internal/file -run TestName   # run a single test
go install github.com/rrgmc/helm-vendor # or `task install`
```

To try it by hand, run from `example/`: `go run .. -c helm-vendor.yaml info`. This needs network access to the chart repositories. `example/data` (the `outputPath` there) is gitignored.

Releases: `task release-version VERSION=vX.Y.Z` tags and pushes. The tag push triggers `.github/workflows/release.yml`, which runs GoReleaser (`.goreleaser.yaml`, CGO disabled).

## Architecture

- `main.go` defines every CLI subcommand and its flags inline and parses the arguments. There are two kinds of command:
  - **Config-based** (`info`, `fetch`, `upgrade`): these go through `newCmd` → `cmd.NewFromFile`. The output root is the directory of the config file joined with `outputPath` from the config. Each chart's `path` is relative to that root. These commands are methods on `*cmd.Cmd`.
  - **Standalone** (`download`, `dependency`, `values-diff`, `values-render`): plain functions in `internal/cmd` that don't read the config. `dependency` and the `values-*` commands work on the chart in the current working directory.
- `internal/cmd`: one file per command. `upgrade.go` holds the main algorithm, and the README's "Upgrade process" section describes it. In order, it:
  1. downloads the chart for the local `Chart.yaml` version;
  2. builds a unified diff (`internal/diff.Builder`, using go-udiff) of the pristine files against the local ones, and also covers new local files in directories the chart contains;
  3. writes the diff as `helm-vendor-<path>-<version>.diff`;
  4. deletes only the files the old chart contained, so custom files survive;
  5. copies in the new chart;
  6. with `--apply-patch`, applies the diff using go-gitdiff (`internal/diff.Patcher`). Conflicts are written to `*_conflict.diff` files instead of failing.

  `--ignore-current` skips steps 1–4.
- `internal/helm`: a wrapper over the Helm SDK.
  - `Repository` either loads an HTTP repo's `index.yaml` or, for `oci://` URLs, uses a registry client with `index == nil`. OCI version listing goes through registry tags and is best effort. `GetChart` falls back to building the chart reference by hand.
  - `Chart.Download` pulls the chart and expands it into a temp dir, or into a given path. It returns `ChartFiles`, which is rooted at `<dir>/<chartName>`. `Close()` removes the temp dir.
- `internal/file`: filesystem helpers.
  - `Iter` is an `iter.Seq2[Info, error]` over a directory.
  - `IterFilter` applies the config's `files.ignore` doublestar globs. The globs match chart-relative paths, and they are applied to both the old and the new chart file sets.
  - There are also helpers for copying files and generating unique filenames.
- `internal/config` holds the YAML config structs, and `internal/yaml` is a small decode helper.

## Conventions

- File access is scoped with `*os.Root` (`os.OpenRoot` / `root.OpenRoot`), for the output root, each chart root and the downloaded chart. Keep new file operations going through the roots rather than raw `os` path calls.
- Iteration uses Go 1.23+ range-over-func iterators (`iter.Seq`/`iter.Seq2`).
- User-facing progress is printed directly with `fmt.Printf`. Errors are wrapped with `fmt.Errorf("...: %w", err)`.
