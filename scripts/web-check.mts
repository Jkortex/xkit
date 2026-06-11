#!/usr/bin/env tsx
/**
 * daily/web 文档端点检查脚本
 *
 * 检查 web CAPABILITIES.md 中记录的后端 API 端点是否全部被 server/api/ 文档覆盖。
 * 用法: pnpm web-check
 */

import { execSync } from 'child_process';
import { existsSync } from 'fs';
import { dirname, join, resolve } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const WEB_ROOT = resolve(__dirname, '../packages/daily/web');
const WEB_DOC = join(WEB_ROOT, 'docs/business/CAPABILITIES.md');
const SERVER_API_DIR = join(WEB_ROOT, '../server/docs/api');

function extractWebEndpoints(): string[] {
  if (!existsSync(WEB_DOC)) {
    console.error(`Error: ${WEB_DOC} not found.`);
    process.exit(1);
  }
  const content = execSync(`cat "${WEB_DOC}"`, { encoding: 'utf-8' });
  const endpoints = new Set<string>();
  // 匹配 `METHOD /path` 格式
  const re = /`(GET|POST|PATCH|DELETE) ([^`]+)`/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(content))) {
    const ep = `${m[1]} ${m[2]}`
      .replace(/\/api\/v1/, '')
      .replace(/\/auth\//, '/')
      .replace(/\?.*$/, '');
    endpoints.add(ep);
  }
  return [...endpoints].sort();
}

function extractServerEndpoints(): string[] {
  if (!existsSync(SERVER_API_DIR)) {
    console.warn(
      `Warning: ${SERVER_API_DIR} not found. Skipping server comparison.`,
    );
    return [];
  }
  const output = execSync(
    `grep -roh '\\(GET\\|POST\\|PATCH\\|DELETE\\) /[^ ]*' "${SERVER_API_DIR}" 2>/dev/null || true`,
    { encoding: 'utf-8', shell: 'bash' },
  ).trim();
  if (!output) return [];
  const endpoints = new Set<string>();
  for (const line of output.split('\n')) {
    const cleaned = line
      .replace(/^###\s+/, '')
      .replace(/`/g, '')
      .replace(/\?.*$/, '')
      .replace(/\/api\/v1/, '')
      .replace(/\/auth\//, '/');
    endpoints.add(cleaned);
  }
  return [...endpoints].sort();
}

// ── 主流程 ──────────────────────────────────────────────

const webEndpoints = extractWebEndpoints();
console.log('=== Endpoints in web CAPABILITIES.md ===');
console.log(webEndpoints.join('\n'));
console.log();

const serverEndpoints = extractServerEndpoints();
if (serverEndpoints.length === 0) {
  console.log('Warning: No server endpoints found. Skipping comparison.');
  console.log('Extra in web (not checked): all endpoints listed above.');
  process.exit(0);
}

console.log('=== Endpoints in server/api/ ===');
console.log(serverEndpoints.join('\n'));
console.log();

const serverSet = new Set(serverEndpoints);

const missing = webEndpoints.filter((ep) => !serverSet.has(ep));

console.log('=== Missing from server docs (in web but not in server/api/) ===');
if (missing.length > 0) {
  console.log(missing.join('\n'));
  console.log();
  console.error(
    'ERROR: web CAPABILITIES.md references endpoints not found in server/api/ docs.',
  );
  console.error('This may mean:');
  console.error('  1. Server endpoint removed but web doc not updated');
  console.error('  2. Route path changed but web doc not synchronized');
  console.error(
    '  3. Server doc missing the endpoint (add it to server/docs/api/)',
  );
  process.exit(1);
} else {
  console.log('(none - all web endpoints covered by server docs)');
}
