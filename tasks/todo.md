# ccr Task List

## Phase 1: Foundation
- [ ] Task 1:  Project scaffold (Go module, Cobra, Makefile, cross-compile targets)
- [ ] Task 2:  Config module (read/write ~/.ccr/config.json, atomic save, 0600 perms)
- [ ] Task 3:  Settings module (safe read/patch of ~/.claude/settings.json, backup, rollback)

## Checkpoint 1
- [ ] Build succeeds for all platforms, config/settings unit tests pass, no user files touched

## Phase 2: Init Command
- [ ] Task 4:  `ccr init` — prompt + validate API key, save to config
- [ ] Task 5:  `ccr init` — snapshot originalStatusLine, write neutral env.sh, print source instruction
- [ ] Task 6:  `ccr init` — write statusline.ts + statusline.sh to ~/.ccr/

## Checkpoint 2
- [ ] `ccr init` end-to-end: config written, env.sh neutral, ~/.claude/settings.json unchanged

## Phase 3: Enable / Disable / Status
- [ ] Task 7:  `ccr enable` — rewrite env.sh, patch settings.json statusLine, idempotent
- [ ] Task 8:  `ccr disable` — rewrite env.sh neutral, restore original statusLine, idempotent
- [ ] Task 9:  `ccr status` — read-only, print active provider + sync state

## Checkpoint 3
- [ ] enable → disable → enable cycle is clean and idempotent; settings.json restored exactly after disable

## Phase 4: Models Command
- [ ] Task 10: `ccr models` — fetch + display formatted table
- [ ] Task 11: `ccr models [search]` — filter by keyword, highlight match

## Checkpoint 4
- [ ] `ccr models claude` filters correctly; errors handled gracefully

## Phase 5: Polish + Distribution
- [ ] Task 12: Windows support (env.ps1, PowerShell instructions, execution policy)
- [ ] Task 13: GoReleaser + README quickstart

## Final Checkpoint
- [ ] Full flow works on Mac/Linux/Windows
- [ ] Single binary, zero runtime deps
- [ ] settings.json never corrupted; backup exists after first enable
