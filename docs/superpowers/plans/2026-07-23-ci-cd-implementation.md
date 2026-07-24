# GitHub Actions CI/CD Workflows Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `.github/workflows/ci.yml` and `.github/workflows/release.yml` for automated testing, linting, multi-platform cross-compilation, and GitHub Releases.

**Architecture:** CI workflow runs on every push and PR to vet, test (`-race`), and build. Release workflow triggers on `v*` tags, matrix cross-compiling static binaries (`CGO_ENABLED=0`) for `linux/amd64`, `linux/arm64`, `darwin/amd64`, and `darwin/arm64`, uploading artifacts and publishing a GitHub Release with auto-generated release notes via `softprops/action-gh-release@v2`.

**Tech Stack:** GitHub Actions (`actions/checkout@v4`, `actions/setup-go@v5`, `actions/upload-artifact@v4`, `actions/download-artifact@v4`, `softprops/action-gh-release@v2`), Go 1.25.

## Global Constraints

- CI must run `go vet ./...`, `go test -v -race ./...`, and `go build -v ./cmd/netmon`.
- Release matrix targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`.
- Cross-compile with `CGO_ENABLED=0` and `-ldflags="-s -w"`.
- Use `actions/setup-go@v5` with `go-version-file: go.mod`.
- Verify YAML syntax using `actionlint` or python `pyyaml`/`jsonschema` if available.

---

## Proposed File Structure

- Create: `.github/workflows/ci.yml` — Continuous Integration workflow
- Create: `.github/workflows/release.yml` — Tag release workflow

---

### Task 1: Create CI Workflow (`.github/workflows/ci.yml`)

**Files:**
- Create: `.github/workflows/ci.yml`

- [ ] **Step 1: Write `.github/workflows/ci.yml`**

```yaml
name: CI

on:
  push:
    branches: ["*"]
  pull_request:
    branches: ["*"]

jobs:
  test:
    name: Build & Test
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Vet code
        run: go vet ./...

      - name: Run tests with race detector
        run: go test -v -race ./...

      - name: Build binary
        run: go build -v ./cmd/netmon
```

- [ ] **Step 2: Verify YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"`
Expected: No error (returns 0)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add GitHub Actions CI workflow"
```

---

### Task 2: Create Release Workflow (`.github/workflows/release.yml`)

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Write `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - "v*"

jobs:
  build-binaries:
    name: Build Multi-Platform Binary
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - goos: linux
            goarch: amd64
            artifact_name: netmon-linux-amd64
          - goos: linux
            goarch: arm64
            artifact_name: netmon-linux-arm64
          - goos: darwin
            goarch: amd64
            artifact_name: netmon-darwin-amd64
          - goos: darwin
            goarch: arm64
            artifact_name: netmon-darwin-arm64
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Build Binary
        env:
          CGO_ENABLED: 0
          GOOS: ${{ matrix.goos }}
          GOARCH: ${{ matrix.goarch }}
        run: |
          go build -ldflags="-s -w" -o ${{ matrix.artifact_name }} ./cmd/netmon

      - name: Upload Artifact
        uses: actions/upload-artifact@v4
        with:
          name: ${{ matrix.artifact_name }}
          path: ${{ matrix.artifact_name }}

  release:
    name: Create GitHub Release
    needs: build-binaries
    runs-on: ubuntu-latest
    permissions:
      contents: write
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Download all build artifacts
        uses: actions/download-artifact@v4
        with:
          path: release-artifacts
          merge-multiple: true

      - name: Create Release
        uses: softprops/action-gh-release@v2
        with:
          files: release-artifacts/*
          generate_release_notes: true
```

- [ ] **Step 2: Verify YAML syntax & test matrix cross-compilation locally**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/release.yml'))"`
Expected: No error

Test local cross-compilation:
Run: `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /tmp/netmon-linux-amd64 ./cmd/netmon`
Expected: Successfully generates `/tmp/netmon-linux-amd64`

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/release.yml
git commit -m "ci: add multi-platform tag release workflow"
```
