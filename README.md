# ccr — Claude Code Router

Switch Claude Code between Anthropic and OpenRouter with a single command. When you hit Anthropic's 5-hour limit, `ccr enable` routes through OpenRouter in seconds. `ccr disable` puts everything back exactly as it was.

## Install

**Mac (Apple Silicon):**
```sh
curl -L https://github.com/jfbiswajit/ccr/releases/latest/download/ccr-mac-arm64 -o ccr && chmod +x ccr && sudo mv ccr /usr/local/bin/
```

**Mac (Intel):**
```sh
curl -L https://github.com/jfbiswajit/ccr/releases/latest/download/ccr-mac-amd64 -o ccr && chmod +x ccr && sudo mv ccr /usr/local/bin/
```

**Linux:**
```sh
curl -L https://github.com/jfbiswajit/ccr/releases/latest/download/ccr-linux-amd64 -o ccr && chmod +x ccr && sudo mv ccr /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
Invoke-WebRequest -Uri https://github.com/jfbiswajit/ccr/releases/latest/download/ccr-windows-amd64.exe -OutFile ccr.exe
```

## Setup (one time)

```sh
ccr init
```

Paste your OpenRouter API key when prompted — get one at [openrouter.ai/keys](https://openrouter.ai/keys). It validates the key, sets everything up, and prints one line to add to your shell profile:

```sh
# Add to ~/.zshrc or ~/.bashrc (printed by ccr init)
source "$HOME/.ccr/env.sh"
```

Then reload your shell or open a new terminal.

## Usage

**When you hit the Anthropic 5-hour limit:**
```sh
ccr enable
source ~/.ccr/env.sh   # or open a new terminal
```

**When the limit resets:**
```sh
ccr disable
source ~/.ccr/env.sh   # or open a new terminal
```

**Check which provider is active:**
```sh
ccr status
```

**Browse available OpenRouter models:**
```sh
ccr models
ccr models claude    # filter by keyword
```

## How it works

- **Non-destructive** — only the `statusLine` field in `~/.claude/settings.json` is ever changed. Your hooks, permissions, plugins, and all other settings are untouched.
- **Fully reversible** — your original `statusLine` is snapshotted during `init` and restored exactly when you run `ccr disable`.
- **Safe** — a timestamped backup of `~/.claude/settings.json` is written to `~/.ccr/backups/` before the first change.
- **Statusline** — when enabled, Claude Code shows live OpenRouter cost tracking at the bottom of the screen.

## What gets installed

```
~/.ccr/
├── config.json          # API key and current state
├── env.sh               # env vars managed by ccr (rewritten on enable/disable)
├── statusline.ts        # OpenRouter cost tracking statusline
├── statusline.sh        # wrapper to run statusline.ts
└── backups/             # timestamped backups of ~/.claude/settings.json
```

## Requirements

- [Claude Code](https://claude.ai/code)
- [OpenRouter](https://openrouter.ai) account and API key

## Windows

After `ccr init`, add this to your PowerShell profile instead:

```powershell
. "$HOME\.ccr\env.ps1"
```

After `ccr enable` or `ccr disable`, run:
```powershell
. $HOME\.ccr\env.ps1
```

If you get a script execution error, run this once in PowerShell as admin:
```powershell
Set-ExecutionPolicy RemoteSigned
```
