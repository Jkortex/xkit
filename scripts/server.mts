#!/usr/bin/env tsx
/**
 * daily/server 管理脚本
 *
 * 用法:
 *   pnpm server <task>
 *
 * 可用 task:
 *   dev          SQLite 本地开发（构建前端 + 启动）
 *   test         Go 测试 (SQLite)
 *   build        编译 Go 二进制
 *   build-web    构建前端 + 拷贝到 embed 目录
 *   vet          go vet
 *   sqlc-gen     sqlc 代码生成
 *   doc-check    路由文档一致性检查
 *   clean        清理构建产物
 *   help         显示帮助
 */

import { execSync } from 'child_process';
import { existsSync, mkdirSync, writeFileSync, rmSync, cpSync } from 'fs';
import { dirname, join, resolve } from 'path';
import { fileURLToPath } from 'url';

// ── 路径常量 ──────────────────────────────────────────────

const __dirname = dirname(fileURLToPath(import.meta.url));
const SERVER_ROOT = resolve(__dirname, '../packages/daily/server');
const WEB_ROOT = resolve(__dirname, '../packages/daily/web');
const ROUTER_FILE = join(SERVER_ROOT, 'internal/infrastructure/api/router.go');
const DOC_DIR = join(SERVER_ROOT, 'docs/api');

// ── 工具函数 ──────────────────────────────────────────────

function run(
  cmd: string,
  opts?: { cwd?: string; env?: Record<string, string>; allowFail?: boolean },
) {
  const cwd = opts?.cwd ?? SERVER_ROOT;
  const env = { ...process.env, ...opts?.env };
  try {
    execSync(cmd, { stdio: 'inherit', cwd, env });
  } catch (e) {
    if (opts?.allowFail) return;
    throw e;
  }
}

// ── Task 函数 ──────────────────────────────────────────────

function ensureDist() {
  const distDir = join(SERVER_ROOT, 'internal/infrastructure/api/static/dist');
  const indexFile = join(distDir, 'index.html');
  if (!existsSync(distDir)) mkdirSync(distDir, { recursive: true });
  if (!existsSync(indexFile)) {
    writeFileSync(
      indexFile,
      '<!DOCTYPE html><html><head><meta charset="utf-8"><title>Daily</title></head><body>Frontend not built. Run: pnpm build:web</body></html>',
    );
  }
}

// buildWeb 构建前端并拷贝到 go:embed 目录
function buildWeb() {
  ensureDist();
  run('pnpm exec vite build', { cwd: WEB_ROOT });
  const srcDist = join(WEB_ROOT, 'dist');
  const destDist = join(SERVER_ROOT, 'internal/infrastructure/api/static/dist');
  cpSync(srcDist, destDist, { recursive: true });
}

function taskVet() {
  run('go vet ./cmd/daily/...');
}

function taskTest() {
  run('go test -v ./internal/... ./tests/... ./cmd/...', {
    allowFail: true,
  });
}

function taskBuild() {
  run('go build -o daily ./cmd/daily/');
}

function taskSqlcGen() {
  run('sqlc generate');
}

function taskDev() {
  buildWeb();
  run('go run ./cmd/daily/ web');
}

function taskClean() {
  const dataDir = join(SERVER_ROOT, 'data');
  if (existsSync(dataDir)) rmSync(dataDir, { recursive: true, force: true });
}

// ── 路由文档检查（从 docroute-check.sh 移植）──────────────

function extractCodeRoutes(): string[] {
  if (!existsSync(ROUTER_FILE)) {
    console.error(`Error: ${ROUTER_FILE} not found.`);
    process.exit(1);
  }
  const content = execSync(`cat "${ROUTER_FILE}"`, { encoding: 'utf-8' });
  const routes = new Set<string>();
  const re = /\b(GET|POST|PATCH|DELETE)\("\/([^"]+)/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(content))) {
    const method = m[1];
    const path = m[2].replace(/:[^/]+/g, ':param'); // 统一 :param 占位
    routes.add(`${method} /${path}`);
  }
  return [...routes].sort();
}

function extractDocRoutes(): string[] {
  if (!existsSync(DOC_DIR)) return [];
  const output = execSync(
    `grep -roh '\\(GET\\|POST\\|PATCH\\|DELETE\\) /[^ ]*' "${DOC_DIR}" 2>/dev/null || true`,
    { encoding: 'utf-8', shell: 'bash' },
  ).trim();
  if (!output) return [];
  const routes = new Set<string>();
  for (const line of output.split('\n')) {
    const cleaned = line
      .replace(/^###\s+/, '')
      .replace(/`/g, '')
      .replace(/\?.*$/, '')
      .replace(/\/api\/v1/, '')
      .replace(/\/auth\//, '/');
    routes.add(cleaned);
  }
  return [...routes].sort();
}

function taskDocCheck() {
  const codeRoutes = extractCodeRoutes();
  const docRoutes = extractDocRoutes();

  console.log('=== Routes in router.go (code) ===');
  console.log(codeRoutes.join('\n'));
  console.log();

  console.log('=== Routes in api/ (docs) ===');
  console.log(docRoutes.join('\n'));
  console.log();

  const codeSet = new Set(codeRoutes);
  const docSet = new Set(docRoutes);

  const missing = codeRoutes.filter((r) => !docSet.has(r));
  const extra = docRoutes.filter((r) => !codeSet.has(r));

  console.log('=== Missing from docs (in code but not documented) ===');
  if (missing.length > 0) {
    console.log(missing.join('\n'));
    console.log();
    console.error(
      `ERROR: router.go has ${missing.length} route(s) not covered by docs/api/.`,
    );
    process.exit(1);
  } else {
    console.log('(none - all routes documented)');
  }

  console.log();
  console.log('=== Extra in docs (documented but not in code) ===');
  if (extra.length > 0) {
    console.log(
      `${extra.length} route(s) in docs but not in code (may be intentional).`,
    );
  } else {
    console.log('(none - docs in sync)');
  }
}

// ── 帮助 ──────────────────────────────────────────────────

function showHelp() {
  console.log(`daily/server 管理脚本

用法: pnpm server <task>

可用 task:
  dev          SQLite 本地开发（构建前端 + 启动）
  test         Go 测试 (SQLite)
  test-cli     Go 测试 (daily-cli)
  build        编译 Go 二进制
  build-web    构建前端 + 拷贝到 embed 目录
  vet          go vet
  sqlc-gen     sqlc 代码生成
  doc-check    路由文档一致性检查
  clean        清理构建产物
  help         显示此帮助`);
}

// ── 入口 ──────────────────────────────────────────────────

const task = process.argv[2] ?? 'help';

const tasks: Record<string, () => void> = {
  dev: taskDev,
  test: taskTest,
  'test-cli': () => run('go test -v ./cmd/daily/...'),
  build: taskBuild,
  'build-web': buildWeb,
  vet: taskVet,
  'sqlc-gen': taskSqlcGen,
  'doc-check': taskDocCheck,
  clean: taskClean,
  help: showHelp,
};

const fn = tasks[task];
if (!fn) {
  console.error(`Unknown task: ${task}`);
  console.error(`Run "pnpm server help" to see available tasks.`);
  process.exit(1);
}
fn();
