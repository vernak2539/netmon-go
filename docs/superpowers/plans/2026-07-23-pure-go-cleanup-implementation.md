# Pure Go Repository Cleanup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all legacy Python source files (`*.py`), Python tooling configurations (`pyproject.toml`, `pyrightconfig.json`, `uv.lock`, `.python-version`), and update documentation (`README.md`, `AGENTS.md`, `.gitignore`) so `main-go` is a 100% pure Go repository.

**Architecture:** Pure Go repository structure with zero Python files or dependencies.

**Tech Stack:** Go 1.25, Git.

## Global Constraints

- Remove Python files: `ai.py`, `config.py`, `graphs.py`, `main.py`, `models.py`, `runner.py`, `sqlite.py`, `tg.py`, `pyproject.toml`, `pyrightconfig.json`, `uv.lock`, `.python-version`.
- Update `README.md` to remove `speedtest-cli` external dependency (since `internal/speedtest` uses `speedtest-go`).
- Update `AGENTS.md` to reflect pure Go status of the branch.
- Verify `make build`, `make test`, `make lint` pass cleanly.

---

## Proposed File Structure

- Delete: `ai.py`, `config.py`, `graphs.py`, `main.py`, `models.py`, `runner.py`, `sqlite.py`, `tg.py`
- Delete: `pyproject.toml`, `pyrightconfig.json`, `uv.lock`, `.python-version`
- Modify: `README.md` — remove `speedtest-cli` dependency requirement
- Modify: `AGENTS.md` — update repository overview and branch description
- Modify: `.gitignore` — clean up unnecessary Python entries if any

---

### Task 1: Remove Python Source & Configuration Files

**Files:**
- Delete: `ai.py`, `config.py`, `graphs.py`, `main.py`, `models.py`, `runner.py`, `sqlite.py`, `tg.py`, `pyproject.toml`, `pyrightconfig.json`, `uv.lock`, `.python-version`

- [ ] **Step 1: Remove Python files**

```bash
git rm ai.py config.py graphs.py main.py models.py runner.py sqlite.py tg.py pyproject.toml pyrightconfig.json uv.lock .python-version
```

- [ ] **Step 2: Verify `make test` and `make build` pass**

Run: `make test && make build`
Expected: Passes with zero errors

- [ ] **Step 3: Commit**

```bash
git commit -m "refactor: remove legacy Python source files and package configurations"
```

---

### Task 2: Update `README.md`, `AGENTS.md`, and `.gitignore`

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`
- Modify: `.gitignore`

- [ ] **Step 1: Update `README.md`**

Remove `speedtest-cli` from requirements and tech stack table (replaced by native Go `speedtest-go`).

- [ ] **Step 2: Update `AGENTS.md`**

Update directory tree and description to reflect pure Go codebase.

- [ ] **Step 3: Update `.gitignore`**

Keep clean Go `.gitignore` with `netmon-go` and build artifact rules.

- [ ] **Step 4: Verify `make lint`**

Run: `make lint`
Expected: Clean

- [ ] **Step 5: Commit**

```bash
git add README.md AGENTS.md .gitignore
git commit -m "docs: update README and AGENTS.md for pure Go repository"
```
