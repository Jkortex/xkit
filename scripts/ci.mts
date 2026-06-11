#!/usr/bin/env tsx
/**
 * xkit 本地 CI 检查脚本
 *
 * 用法:
 *   pnpm run ci           全量检查
 *   pnpm run ci --fix     自动修复格式问题
 */

import { execSync } from 'child_process';
import { dirname } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = dirname(__dirname);

const args = process.argv.slice(2);
const fix = args.includes('--fix');

type CheckResult = { name: string; pass: boolean };

const results: CheckResult[] = [];

function run(name: string, cmd: string) {
  console.log(`\n=== ${name} ===`);
  try {
    execSync(cmd, { stdio: 'inherit', cwd: ROOT });
    console.log(`  PASS`);
    results.push({ name, pass: true });
  } catch {
    console.log(`  FAIL`);
    results.push({ name, pass: false });
  }
}

// ── 检查步骤 ──────────────────────────────────────────

// 1. TypeScript 类型检查
run('TypeScript (daily/web)', 'pnpm typecheck');

// 2. Vitest (hotkeys)
run('Vitest (hotkeys)', 'pnpm test:hotkeys');

// 3. Go vet
run('Go vet', 'pnpm run xkit vet');

// 4. Go 测试 (SQLite)
run('Go test (daily/server)', 'pnpm run xkit test');

// 5. Go 测试 (daily-cli)
run('Go test (daily-cli)', 'pnpm test:cli');

// 6. Lint
run('Lint (oxlint)', 'pnpm lint');

// 7. Format
if (fix) {
  run('Format (oxfmt fix)', 'pnpm fmt:fix');
} else {
  run('Format (oxfmt check)', 'pnpm fmt');
}

// ── 汇总 ──────────────────────────────────────────────

console.log(`\n${'='.repeat(40)}`);

const failed = results.filter((r) => !r.pass);
if (failed.length === 0) {
  console.log('All checks passed.');
  process.exit(0);
} else {
  console.log(`Failed checks (${failed.length}):`);
  for (const f of failed) {
    console.log(`  - ${f.name}`);
  }
  process.exit(1);
}
