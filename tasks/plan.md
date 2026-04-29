# Implementation Plan: ccr (Claude Code Router)

## Overview

`ccr` is a cross-platform CLI tool that manages switching Claude Code between Anthropic's native API and OpenRouter. It is distributed as a single compiled Go binary — no runtime, no install friction. Three commands: `init`, `enable`, `disable`, `status`, and `models`.

## Architecture Decisions

- **Runtime:** Go — single compiled binary, zero runtime dependencies, cross-platform (Mac/Linux/Windows) via `GOOS/GOARCH` cross-compilation
- **CLI framework:** Cobra — standard Go CLI library, clean subcommand structure
- **Prompts:** `huh` (Charm) — lightweight, pretty interactive prompts with no external runtime
- **Config:** `~/.ccr/config.json` — stores API key, current state, and snapshot of original Claude settings
- **Env file:** `~/.ccr/env.sh` (Mac/Linux) / `~/.ccr/env.ps1` (Windows) — sourced by shell profile; only `enable`/`disable` rewrites it
- **Statusline:** `~/.ccr/statusline.ts` + `~/.ccr/statusline.sh` — written during `init`, activated only when enabled (requires Node.js/npx, which Claude Code already ships with)
- **State:** `activeProvider: "anthropic" | "openrouter"` in config; enables idempotency

## Non-Destructive Guarantees

These rules are enforced throughout every command:

1. **Never overwrite `~/.claude/settings.json` wholesale.** Always read → patch specific fields → write back. Validate JSON before writing.
2. **Snapshot before first enable.** Before touching `settings.json` for the first time, write a timestamped backup to `~/.ccr/backups/settings.YYYYMMDDHHMMSS.json`.
3. **Store original statusLine.** During `init`, read and save the current `statusLine` value from `settings.json` into config as `originalStatusLine`. `disable` restores it exactly.
4. **Don't touch `ANTHROPIC_API_KEY`.** The disabled env.sh only unsets OpenRouter-specific vars (`ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`). The user's existing `ANTHROPIC_API_KEY` in their shell profile is never touched and naturally re-applies after disable + re-source.
5. **Idempotent commands.** `ccr enable` when already enabled = no-op with a clear message. Same for `disable`.
6. **Never auto-modify shell profiles.** Print the `source` line for the user to add. Don't touch `~/.zshrc` or equivalent.

## Dependency Graph

```
Config module (~/.ccr/config.json)
    │
    ├── init     → writes config, env.sh (neutral), statusline files
    ├── enable   → reads config, writes OpenRouter env.sh, patches settings.json
    ├── disable  → reads config, writes neutral env.sh, restores settings.json
    ├── status   → reads config, prints current state
    └── models   → reads API key from config

Backup module (~/.ccr/backups/)
    │
    └── enable   → snapshots settings.json before first patch

Settings module (~/.claude/settings.json)
    │
    ├── init     → reads originalStatusLine, saves to config
    ├── enable   → patches statusLine.command
    └── disable  → restores statusLine from config

OpenRouter HTTP client
    │
    ├── init     → validates API key
    └── models   → GET /api/v1/models

Shell env file (~/.ccr/env.sh)
    │
    ├── init     → creates neutral file
    ├── enable   → rewrites with OpenRouter vars
    └── disable  → rewrites with unset-only (neutral)
```

---

## Phase 1: Foundation

### Task 1: Project scaffold

**Description:** Initialize the Go module, Cobra CLI entry point, and build/release toolchain. The binary should compile for Mac (arm64/amd64), Linux (amd64), and Windows (amd64).

**Acceptance criteria:**
- [ ] `go build ./...` succeeds with no errors
- [ ] `./ccr --help` prints command list
- [ ] `./ccr --version` prints version
- [ ] `Makefile` targets: `build`, `build-all` (cross-compile), `test`

**Verification:**
- [ ] `make build && ./ccr --help`
- [ ] `make build-all` produces binaries for all three platforms

**Dependencies:** None

**Files likely touched:**
- `main.go`
- `cmd/root.go`
- `Makefile`
- `go.mod`

**Estimated scope:** Small

---

### Task 2: Config module

**Description:** Typed read/write for `~/.ccr/config.json`. Handles first-run defaults, partial updates (merge, not replace), and cross-platform path resolution via `os.UserHomeDir()`.

**Config schema:**
```json
{
  "apiKey": "",
  "activeProvider": "anthropic",
  "originalStatusLine": { "type": "command", "command": "..." }
}
```

**Acceptance criteria:**
- [ ] `LoadConfig()` returns defaults if file doesn't exist
- [ ] `SaveConfig(patch)` merges and persists atomically (write to tmp, rename)
- [ ] All paths use `filepath.Join(home, ".ccr", ...)` — no hardcoded slashes
- [ ] Config directory created with `0700` permissions; config file `0600`

**Verification:**
- [ ] Unit test: load missing file → defaults
- [ ] Unit test: save then load round-trips correctly
- [ ] Unit test: partial save does not wipe unrelated fields

**Dependencies:** Task 1

**Files likely touched:**
- `internal/config/config.go`
- `internal/config/config_test.go`

**Estimated scope:** Small

---

### Task 3: Settings module (safe ~/.claude/settings.json access)

**Description:** Read/write `~/.claude/settings.json` safely. This is the most critical non-destructive piece — must never corrupt or overwrite unrelated fields.

**Acceptance criteria:**
- [ ] `ReadSettings()` returns empty struct if file doesn't exist (not an error)
- [ ] `PatchSettings(patch)` merges only specified fields, leaves all others untouched
- [ ] Validates JSON after write; rolls back to backup if invalid
- [ ] `BackupSettings()` writes a timestamped copy to `~/.ccr/backups/`

**Verification:**
- [ ] Unit test: patch only changes the targeted field
- [ ] Unit test: invalid write triggers rollback
- [ ] Manual: `jq` diff before/after shows only `statusLine` changed

**Dependencies:** Task 2

**Files likely touched:**
- `internal/claudesettings/settings.go`
- `internal/claudesettings/settings_test.go`

**Estimated scope:** Small

---

## Checkpoint: Phase 1

- [ ] `make build` and `make build-all` succeed
- [ ] `./ccr --help` and `./ccr --version` work
- [ ] Config and settings unit tests pass
- [ ] No changes made to any user files yet

---

## Phase 2: Init Command

### Task 4: `ccr init` — prompt and validate API key

**Description:** Ask for the OpenRouter API key (masked input). Validate it by calling `GET /api/v1/models` — if the key works, at least one model comes back. Save to config only on success.

**Acceptance criteria:**
- [ ] Prompts with masked input
- [ ] Validates key before saving
- [ ] Friendly error if key is invalid — does not save
- [ ] Re-running `ccr init` asks for confirmation before overwriting existing config
- [ ] API key saved with `0600` file permissions

**Verification:**
- [ ] Valid key → config written, success message
- [ ] Invalid key → error, no config written
- [ ] Second run → prompts "already initialized, overwrite?"

**Dependencies:** Task 2

**Files likely touched:**
- `cmd/init.go`
- `internal/openrouter/client.go`

**Estimated scope:** Small

---

### Task 5: `ccr init` — snapshot original settings + write env file

**Description:** Read the current `statusLine` from `~/.claude/settings.json` and save it into config as `originalStatusLine` — this is what `disable` will restore. Then write a neutral `~/.ccr/env.sh` (only unsets, no exports) and print the one-time shell source instruction.

**Neutral env.sh (disabled state):**
```sh
# Managed by ccr — do not edit manually
unset ANTHROPIC_BASE_URL
unset ANTHROPIC_AUTH_TOKEN
```

**Acceptance criteria:**
- [ ] `originalStatusLine` captured from settings.json and saved to config
- [ ] If settings.json has no `statusLine`, saves null — disable will leave it absent
- [ ] Neutral `~/.ccr/env.sh` written
- [ ] On Windows, writes `~/.ccr/env.ps1` instead
- [ ] Detects user's shell (via `$SHELL`) and prints correct profile file to add source line to
- [ ] Does not modify the shell profile itself

**Verification:**
- [ ] `source ~/.ccr/env.sh` → `$ANTHROPIC_BASE_URL` is empty
- [ ] Config contains `originalStatusLine` matching what was in settings.json

**Dependencies:** Tasks 3, 4

**Files likely touched:**
- `cmd/init.go`
- `internal/shell/shell.go`

**Estimated scope:** Small

---

### Task 6: `ccr init` — write statusline files

**Description:** Write the OpenRouter cost-tracking statusline to `~/.ccr/statusline.ts` and a thin `~/.ccr/statusline.sh` wrapper. Not activated yet — `enable` does that. Skip if files exist unless `--force` is passed.

**Acceptance criteria:**
- [ ] `~/.ccr/statusline.ts` written with embedded template
- [ ] `~/.ccr/statusline.sh` written and marked executable (`chmod +x`)
- [ ] Second `ccr init` skips rewrite unless `--force` passed
- [ ] Prints note: "Requires Node.js/npx (already included with Claude Code)"

**Verification:**
- [ ] Files exist after init
- [ ] `npx tsx ~/.ccr/statusline.ts` exits without syntax errors (if Node.js present)

**Dependencies:** Task 4

**Files likely touched:**
- `cmd/init.go`
- `internal/statusline/template.go` (embedded file string)

**Estimated scope:** Small

---

## Checkpoint: Phase 2

- [ ] `ccr init` completes end-to-end
- [ ] `~/.ccr/config.json`, `~/.ccr/env.sh`, `~/.ccr/statusline.ts` all written
- [ ] `originalStatusLine` in config matches current `~/.claude/settings.json`
- [ ] `~/.claude/settings.json` is **unchanged** after init

---

## Phase 3: Enable and Disable Commands

### Task 7: `ccr enable`

**Description:** Switch to OpenRouter. Rewrites env.sh with OpenRouter vars, patches `settings.json` to use the OpenRouter statusline, and updates `activeProvider` in config. Idempotent: no-op if already enabled.

**OpenRouter env.sh:**
```sh
# Managed by ccr — do not edit manually
export ANTHROPIC_BASE_URL="https://openrouter.ai/api"
export ANTHROPIC_AUTH_TOKEN="<key from config>"
export ANTHROPIC_API_KEY=""
```

**Acceptance criteria:**
- [ ] No-op (with message) if already enabled
- [ ] Backs up `~/.claude/settings.json` before first patch
- [ ] Rewrites env.sh with OpenRouter vars
- [ ] Patches `settings.json` `statusLine.command` to `sh ~/.ccr/statusline.sh`
- [ ] Updates `activeProvider: "openrouter"` in config
- [ ] Prints: active state + reminder to `source ~/.ccr/env.sh` in current shell

**Verification:**
- [ ] `ccr enable` twice → second run prints "already enabled"
- [ ] `cat ~/.ccr/env.sh` shows OpenRouter exports
- [ ] `jq .statusLine ~/.claude/settings.json` shows ccr statusline path
- [ ] All other fields in settings.json are unchanged

**Dependencies:** Tasks 3, 5

**Files likely touched:**
- `cmd/enable.go`
- `internal/shell/shell.go`
- `internal/claudesettings/settings.go`

**Estimated scope:** Small

---

### Task 8: `ccr disable`

**Description:** Switch back to Anthropic. Rewrites env.sh to neutral (unsets only), restores original `statusLine` in `settings.json`, and updates `activeProvider`. Idempotent.

**Acceptance criteria:**
- [ ] No-op (with message) if already disabled
- [ ] Rewrites env.sh to neutral (no exports, only unsets)
- [ ] Restores `statusLine` in settings.json to `originalStatusLine` from config
- [ ] If `originalStatusLine` was null, removes `statusLine` key from settings.json
- [ ] Updates `activeProvider: "anthropic"` in config
- [ ] Prints: active state + reminder to `source ~/.ccr/env.sh`

**Verification:**
- [ ] `ccr disable` twice → second run prints "already disabled"
- [ ] `source ~/.ccr/env.sh && echo $ANTHROPIC_BASE_URL` → empty
- [ ] `jq .statusLine ~/.claude/settings.json` matches original value from before init

**Dependencies:** Task 7

**Files likely touched:**
- `cmd/disable.go`
- `internal/shell/shell.go`
- `internal/claudesettings/settings.go`

**Estimated scope:** Small

---

### Task 9: `ccr status`

**Description:** Print current state without making any changes. Shows active provider, which env file is in use, and whether settings.json reflects the current state (detects out-of-sync situations).

**Output example:**
```
Provider:  OpenRouter  (active)
Env file:  ~/.ccr/env.sh  [sourced ✓ / not sourced — run: source ~/.ccr/env.sh]
Settings:  ~/.claude/settings.json statusLine → ccr
```

**Acceptance criteria:**
- [ ] Prints active provider
- [ ] Detects if env vars in current shell match config state (sourced or not)
- [ ] Warns if settings.json is out of sync with config state
- [ ] Zero writes — purely read-only

**Verification:**
- [ ] `ccr status` after enable shows OpenRouter
- [ ] `ccr status` after disable shows Anthropic

**Dependencies:** Tasks 2, 3

**Files likely touched:**
- `cmd/status.go`

**Estimated scope:** Small

---

## Checkpoint: Phase 3

- [ ] `ccr enable` → `ccr disable` → `ccr enable` cycle works cleanly
- [ ] All operations are idempotent
- [ ] `~/.claude/settings.json` matches original after disable
- [ ] Backup exists in `~/.ccr/backups/` after first enable

---

## Phase 4: Models Command

### Task 10: `ccr models` — fetch and display

**Description:** Call `GET https://openrouter.ai/api/v1/models` and display a formatted table. Columns: model ID, context length, input cost ($/M tokens), output cost ($/M tokens).

**Acceptance criteria:**
- [ ] Uses API key from config
- [ ] Renders a clean table (not raw JSON)
- [ ] Handles missing API key, network failure, and HTTP errors gracefully
- [ ] Exits with a clear error if `ccr init` hasn't been run

**Verification:**
- [ ] `ccr models` prints at least 10 rows with a valid API key
- [ ] `ccr models` with no config → "run ccr init first"

**Dependencies:** Task 2

**Files likely touched:**
- `cmd/models.go`
- `internal/openrouter/client.go`

**Estimated scope:** Small

---

### Task 11: `ccr models [search]` — filter

**Description:** Optional positional argument filters results by case-insensitive substring match on model ID or name.

**Acceptance criteria:**
- [ ] `ccr models claude` returns only Claude models
- [ ] `ccr models gemini` returns only Gemini models
- [ ] No matches → "no models matched 'X'"
- [ ] Matching portion highlighted in output

**Verification:**
- [ ] `ccr models claude` is a strict subset of `ccr models`

**Dependencies:** Task 10

**Files likely touched:**
- `cmd/models.go`

**Estimated scope:** Small

---

## Checkpoint: Phase 4

- [ ] `ccr models` and `ccr models [search]` work with valid API key
- [ ] All error states handled with friendly messages

---

## Phase 5: Polish + Distribution

### Task 12: Windows support

**Description:** On Windows, write `~/.ccr/env.ps1` instead of `env.sh`. Print PowerShell profile instructions. Detect and warn if execution policy would block the script.

**Acceptance criteria:**
- [ ] `runtime.GOOS == "windows"` path writes `.ps1`
- [ ] Enabled `env.ps1` uses `$env:` syntax
- [ ] Disabled `env.ps1` uses `Remove-Item Env:` syntax
- [ ] Prints `Set-ExecutionPolicy` fix if needed

**Verification:**
- [ ] Manual test on Windows — `ccr init` and `ccr enable` complete without errors

**Dependencies:** Tasks 5, 7, 8

**Files likely touched:**
- `internal/shell/shell.go`

**Estimated scope:** Small

---

### Task 13: Distribution + README

**Description:** GitHub Releases with pre-built binaries via GoReleaser. README covers: install options, `ccr init`, `ccr enable`/`disable`, `ccr status`, `ccr models`, and the one-time shell source step.

**Acceptance criteria:**
- [ ] `goreleaser release --snapshot` produces binaries for all platforms
- [ ] `go install` works from the module path
- [ ] README has quickstart in under 10 lines

**Verification:**
- [ ] Download release binary, run `./ccr --help` — no install needed

**Dependencies:** All previous tasks

**Files likely touched:**
- `.goreleaser.yaml`
- `README.md`

**Estimated scope:** Small

---

## Final Checkpoint

- [ ] Full flow: `ccr init` → `ccr enable` → Claude Code uses OpenRouter + new statusline
- [ ] `ccr disable` → Claude Code settings.json exactly matches original
- [ ] `ccr enable`/`disable` are idempotent
- [ ] `ccr models claude` filters correctly
- [ ] Zero runtime dependencies — single binary runs on Mac/Linux/Windows
- [ ] `~/.claude/settings.json` is never corrupted; backup exists

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Env vars don't propagate to parent shell | High | Remind on every enable/disable; never auto-source |
| `~/.claude/settings.json` format changes | Medium | Merge-only patch; backup before first write; validate after write |
| OpenRouter API key stored in plaintext | Medium | `0600` file permissions on config; note in docs |
| `npx tsx` unavailable for statusline | Low | Print warning during init; Node.js ships with Claude Code |
| Windows PowerShell execution policy | Medium | Detect and print fix instruction during init |
| User runs `ccr disable` after manually editing settings.json | Low | Status command warns if out-of-sync; backup always available |

## Open Questions

- Should `ccr init` store the user's Anthropic API key so `disable` can restore it to env.sh? (Currently assumes the user's profile handles it)
- Homebrew tap for Mac install — worth adding in v1?
