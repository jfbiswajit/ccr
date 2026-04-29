package statusline

import (
	"os"
	"path/filepath"
)

const tsTemplate = `#!/usr/bin/env npx tsx

/**
 * OpenRouter cost tracking statusline for Claude Code
 * Displays: Provider: model - $cost - cache discount: $saved
 * Requires: ANTHROPIC_AUTH_TOKEN set to your OpenRouter API key
 */

import { existsSync, readFileSync, writeFileSync } from 'node:fs';

interface GenerationData {
  total_cost: number;
  cache_discount: number | null;
  provider_name: string;
  model: string;
}

interface State {
  seen_ids: string[];
  total_cost: number;
  total_cache_discount: number;
  last_provider: string;
  last_model: string;
}

async function fetchGeneration(id: string, apiKey: string): Promise<GenerationData | null> {
  try {
    const res = await fetch(` + "`" + `https://openrouter.ai/api/v1/generation?id=${id}` + "`" + `, {
      headers: { Authorization: ` + "`" + `Bearer ${apiKey}` + "`" + ` },
    });
    if (!res.ok) return null;
    const json = await res.json();
    const data = json?.data;
    if (!data || typeof data.total_cost !== 'number') return null;
    return data;
  } catch {
    return null;
  }
}

function extractGenerationIds(transcriptPath: string): string[] {
  try {
    const content = readFileSync(transcriptPath, 'utf-8');
    const ids: string[] = [];
    for (const line of content.split('\n')) {
      if (!line.trim()) continue;
      try {
        const entry = JSON.parse(line);
        const messageId = entry?.message?.id;
        if (typeof messageId === 'string' && messageId.startsWith('gen-')) {
          ids.push(messageId);
        }
      } catch {
        // skip malformed lines
      }
    }
    return [...new Set(ids)];
  } catch {
    return [];
  }
}

function loadState(statePath: string): State {
  const def: State = { seen_ids: [], total_cost: 0, total_cache_discount: 0, last_provider: '', last_model: '' };
  if (!existsSync(statePath)) return def;
  try {
    const parsed = JSON.parse(readFileSync(statePath, 'utf-8'));
    if (!Array.isArray(parsed.seen_ids)) return def;
    return {
      seen_ids: parsed.seen_ids,
      total_cost: typeof parsed.total_cost === 'number' ? parsed.total_cost : 0,
      total_cache_discount: typeof parsed.total_cache_discount === 'number' ? parsed.total_cache_discount : 0,
      last_provider: typeof parsed.last_provider === 'string' ? parsed.last_provider : '',
      last_model: typeof parsed.last_model === 'string' ? parsed.last_model : '',
    };
  } catch {
    return def;
  }
}

function shortModelName(model: string): string {
  return model.replace(/^[^/]+\//, '').replace(/-\d{8}$/, '');
}

async function main(): Promise<void> {
  const apiKey = process.env.ANTHROPIC_AUTH_TOKEN ?? process.env.ANTHROPIC_API_KEY ?? '';
  if (!apiKey) {
    process.stdout.write('Set ANTHROPIC_AUTH_TOKEN to your OpenRouter key');
    return;
  }

  let inputData = '';
  for await (const chunk of process.stdin) inputData += chunk;

  const input = JSON.parse(inputData);
  const { session_id, transcript_path } = input ?? {};
  if (typeof session_id !== 'string' || typeof transcript_path !== 'string') {
    process.stdout.write('Invalid statusline input');
    return;
  }

  const statePath = ` + "`" + `/tmp/claude-openrouter-cost-${session_id}.json` + "`" + `;
  const state = loadState(statePath);
  const allIds = extractGenerationIds(transcript_path);
  const seenSet = new Set(state.seen_ids);
  const newIds = allIds.filter((id) => !seenSet.has(id));

  let fetchFailed = 0;
  for (const id of newIds) {
    const gen = await fetchGeneration(id, apiKey);
    if (!gen) { fetchFailed++; continue; }
    state.total_cost += gen.total_cost ?? 0;
    state.total_cache_discount += gen.cache_discount ?? 0;
    if (gen.provider_name) state.last_provider = gen.provider_name;
    if (gen.model) state.last_model = gen.model;
    state.seen_ids.push(id);
  }

  writeFileSync(statePath, JSON.stringify(state, null, 2));

  const green = '\x1b[32m', red = '\x1b[31m', reset = '\x1b[0m';
  const shortModel = shortModelName(state.last_model);
  let indicator = '';
  if (newIds.length > 0) {
    indicator = fetchFailed === 0
      ? ` + "`" + `\nusage tracking: ${green}up-to-date${reset}` + "`" + `
      : ` + "`" + `\nusage tracking: ${red}behind${reset}` + "`" + `;
  }

  const label = state.last_provider ? ` + "`" + `${state.last_provider}: ${shortModel}` + "`" + ` : 'OpenRouter';
  process.stdout.write(
    ` + "`" + `${label} - $${state.total_cost.toFixed(4)} - cache discount: $${state.total_cache_discount.toFixed(2)}${indicator}` + "`" + `
  );
}

main().catch((err) => process.stdout.write(` + "`" + `error: ${err.message}` + "`" + `));
`

const shTemplate = `#!/bin/sh
# Managed by ccr
exec npx tsx "$(dirname "$0")/statusline.ts"
`

// WriteFiles writes statusline.ts and statusline.sh to dir, skipping if they already exist.
func WriteFiles(dir string) error {
	return writeFiles(dir, false)
}

// WriteFilesForce writes statusline.ts and statusline.sh to dir, overwriting if they exist.
func WriteFilesForce(dir string) error {
	return writeFiles(dir, true)
}

func writeFiles(dir string, force bool) error {
	tsPath := filepath.Join(dir, "statusline.ts")
	shPath := filepath.Join(dir, "statusline.sh")

	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if force || !fileExists(tsPath) {
		if err := os.WriteFile(tsPath, []byte(tsTemplate), 0644); err != nil {
			return err
		}
	}

	if force || !fileExists(shPath) {
		if err := os.WriteFile(shPath, []byte(shTemplate), 0755); err != nil {
			return err
		}
	}

	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
