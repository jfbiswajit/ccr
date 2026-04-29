# ccr — Claude Code Router

Switch Claude Code between Anthropic and OpenRouter with a single command.

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

**Update:** re-run the same install command to get the latest version.

## Setup (one time)

```sh
ccr init
```

Paste your OpenRouter API key when prompted — get one at [openrouter.ai/keys](https://openrouter.ai/keys).

Then add the printed line to your shell profile:
```sh
source "$HOME/.ccr/env.sh"
```

## Commands

| Command | Description |
|---|---|
| `ccr init` | Set up ccr with your OpenRouter API key |
| `ccr enable` | Switch Claude Code to OpenRouter |
| `ccr disable` | Switch Claude Code back to Anthropic |
| `ccr status` | Show which provider is active |

After `ccr enable` or `ccr disable`, run `source ~/.ccr/env.sh` in your current terminal (or open a new one).

## How it works

- Only the `statusLine` field in `~/.claude/settings.json` is changed — hooks, permissions, and plugins are untouched
- Your original `statusLine` is restored exactly when you run `ccr disable`
- A timestamped backup of `~/.claude/settings.json` is saved to `~/.ccr/backups/` before the first change
- When enabled, Claude Code shows live OpenRouter cost tracking in the statusline

## Uninstall

Run `ccr disable` first, then:

```sh
sudo rm /usr/local/bin/ccr
rm -rf ~/.ccr
```

## Requirements

- [Claude Code](https://claude.ai/code)
- [OpenRouter](https://openrouter.ai) account and API key

## Windows

After `ccr init`, add this to your PowerShell profile instead:
```powershell
. "$HOME\.ccr\env.ps1"
```

After `ccr enable` or `ccr disable`:
```powershell
. $HOME\.ccr\env.ps1
```

If you get a script execution error, run once as admin:
```powershell
Set-ExecutionPolicy RemoteSigned
```
