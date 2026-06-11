/**
 * Daily Sidecar Extension — 静默同步对话到 daily memo
 *
 * 模式：Sidecar
 * - 不注册任何 tool / command
 * - 不拦截、不修改任何事件
 * - 监听到用户输入后，异步翻译并写入 daily
 * - 完全不影响 pi 的行为和上下文
 *
 * 翻译：OpenAI 兼容 API（配置见 ~/.daily/config.json pulse 段）
 *   openai.base_url:  默认 https://token.sensenova.cn/v1
 *   openai.model:     默认 sensenova-6.7-flash-lite
 *   openai.api_key:   配置或 EN_OPENAI_API_KEY
 */

import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { exec as execCb } from 'node:child_process';
import { promisify } from 'node:util';
import { writeFile, mkdtemp, rm, readFile, appendFile, readdir, mkdir } from 'node:fs/promises';
import { existsSync, mkdirSync, readFileSync, appendFileSync, readdirSync, rmSync } from 'node:fs';
import { join, basename, isAbsolute } from 'node:path';
import { tmpdir, homedir } from 'node:os';

const exec = promisify(execCb);

// ── 类型 ──

interface OpenAIConfig {
  base_url?: string;
  api_key?: string;
  model?: string;
}

interface PulseConfig {
  openai?: OpenAIConfig;
}

interface TranslateResult {
  translation: string;
}

// ── 默认值 ──

const DEFAULT_OPENAI_BASE_URL = 'https://token.sensenova.cn/v1';
const DEFAULT_OPENAI_MODEL = 'sensenova-6.7-flash-lite';

// ── 日志（独立文件，按 session 轮转）──

const LOG_DIR = join(homedir(), '.pi', 'agent');

class DailyLogger {
  private logPath: string;

  constructor() {
    if (!existsSync(LOG_DIR)) {
      mkdirSync(LOG_DIR, { recursive: true });
    }

    const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
    this.logPath = join(LOG_DIR, `daily-${timestamp}.log`);

    // 写入 session 起始标记
    appendFileSync(this.logPath, JSON.stringify({
      timestamp: new Date().toISOString(),
      level: 'info',
      logger: 'daily',
      event: 'session_start',
      details: { logFile: this.logPath },
    }) + '\n', 'utf-8');

    // 轮转：保留最近 3 个日志文件
    rotateLogs();
  }

  log(level: 'info' | 'warn' | 'error', event: string, details: Record<string, unknown>): void {
    const entry = JSON.stringify({
      timestamp: new Date().toISOString(),
      level,
      logger: 'daily',
      event,
      details,
    }) + '\n';
    try {
      appendFileSync(this.logPath, entry, 'utf-8');
    } catch (err) {
      console.error('[daily] log write failed:', err);
    }
  }
}

function rotateLogs(): void {
  try {
    const files = readdirSync(LOG_DIR)
      .filter(f => f.startsWith('daily-') && f.endsWith('.log'))
      .sort()
      .reverse(); // 最新在前

    if (files.length > 3) {
      for (const old of files.slice(3)) {
        rmSync(join(LOG_DIR, old), { force: true });
      }
    }
  } catch {
    // 轮转失败不影响主流程
  }
}

const logger = new DailyLogger();



// ── 配置加载（与 pulse 共用 strata 格式）──

function loadPulseConfig(): PulseConfig {
  try {
    const configPath = join(homedir(), '.daily', 'config.json');
    if (!existsSync(configPath)) return {};
    const raw = JSON.parse(readFileSync(configPath, 'utf-8'));
    return (raw as any).pulse ?? {};
  } catch {
    return {};
  }
}

// ── 翻译 ──

async function translateOpenAI(text: string): Promise<TranslateResult> {
  const cfg = loadPulseConfig();
  const baseURL = (cfg.openai?.base_url || DEFAULT_OPENAI_BASE_URL).replace(/\/+$/, '');
  const model = cfg.openai?.model || DEFAULT_OPENAI_MODEL;
  const apiKey = cfg.openai?.api_key || process.env.EN_OPENAI_API_KEY || '';

  if (!apiKey) {
    throw new Error('OpenAI API key not configured (EN_OPENAI_API_KEY or ~/.daily/config.json pulse.openai.api_key)');
  }

  const systemPrompt = 'You are a professional translator. Translate the user\'s Chinese text into natural, idiomatic English. Your response MUST be the translation only, with no explanations, no markdown, no extra text.';

  const payload = {
    model,
    messages: [
      { role: 'system', content: systemPrompt },
      { role: 'user', content: text },
    ],
    temperature: 0.3,
    max_tokens: 2000,
  };

  const res = await fetch(`${baseURL}/chat/completions`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${apiKey}`,
    },
    body: JSON.stringify(payload),
    signal: AbortSignal.timeout(30_000),
  });

  if (!res.ok) {
    const body = await res.text().catch(() => '');
    throw new Error(`OpenAI API ${res.status}: ${body.slice(0, 200)}`);
  }

  const data = (await res.json()) as {
    choices?: Array<{ message?: { content?: string } }>;
  };

  const content = data.choices?.[0]?.message?.content;
  if (!content) {
    throw new Error('OpenAI: no content in response');
  }

  return { translation: content.trim() };
}

// ── 文件同步模式 ──

const FILE_SYNC_PATTERNS = [
  /plan/i,
  /spec/i,
  /task/i,
  /walkthrough/i,
  /handoff/i,
  /design/i,
  /req(?:uirement)?/i,
  /readme/i,
  /todo/i,
  /changelog/i,
  /roadmap/i,
];

interface GitInfo {
  repo: string;
  branch: string;
}

async function getGitInfo(cwd: string): Promise<GitInfo> {
  try {
    const { stdout: remoteUrl } = await exec('git remote get-url origin', {
      cwd,
      encoding: 'utf-8',
      timeout: 3000,
      maxBuffer: 1024,
    });

    const match = remoteUrl.trim().match(/[\/:]([^\/]+\/[^\/]+?)(?:\.git)?$/);
    const repo = match ? match[1] : '';

    let branch = '';
    try {
      const { stdout: b } = await exec('git rev-parse --abbrev-ref HEAD', {
        cwd,
        encoding: 'utf-8',
        timeout: 3000,
        maxBuffer: 1024,
      });
      branch = b.trim();
    } catch { /* no branch info */ }

    if (repo) return { repo, branch };
    return { repo: basename(cwd), branch };
  } catch {
    return { repo: basename(cwd), branch: '' };
  }
}

async function getFileSyncTags(cwd: string, fileTag: string): Promise<string[]> {
  const tags = ['file-sync', fileTag];
  const { repo, branch } = await getGitInfo(cwd);
  if (repo) tags.push(repo);
  if (branch) tags.push(branch);
  return tags;
}

function formatFileMemo(filePath: string, content: string): string {
  const now = new Date();
  const date = now.toISOString().slice(0, 10);
  const time = now.toTimeString().slice(0, 8);
  return [
    `# ${basename(filePath)} — ${date} ${time}`,
    '',
    `路径: \`${filePath}\``,
    '',
    '---',
    '',
    content,
    '',
  ].join('\n');
}

// ── 判断 ──

function containsChinese(text: string): boolean {
  return /[\u4e00-\u9fff]/.test(text);
}

function shouldTranslate(text: string): boolean {
  const trimmed = text.trim();
  if (trimmed.length < 3) return false;
  if (!containsChinese(trimmed)) return false;

  const skipWords = ['好的', '谢谢', '你好', '嗯', '对', '是', 'ok', '好', '是的', '没错', '可以', '行'];
  for (const w of skipWords) {
    if (trimmed === w) return false;
  }
  return true;
}

// ── 写入 daily memo ──

async function syncToDaily(content: string, tags: string[] = ['pi']): Promise<void> {
  const tmpDir = await mkdtemp(join(tmpdir(), 'daily-'));
  const tmpFile = join(tmpDir, 'memo.md');
  await writeFile(tmpFile, content, 'utf-8');

  const args = ['daily', 'memo', 'create', '--file', tmpFile];
  for (const tag of tags) {
    args.push('--tag', tag);
  }

  logger.log('info', 'daily_sync_start', { contentLength: content.length, tags });

  try {
    const { stdout } = await exec(args.join(' '), {
      encoding: 'utf-8',
      timeout: 10_000,
      maxBuffer: 1024 * 1024,
    });
    logger.log('info', 'daily_sync_ok', { stdout: stdout.trim() });
  } catch (err: any) {
    const stderr = err.stderr?.trim() || err.message || String(err);
    logger.log('error', 'daily_sync_failed', { error: stderr });
  } finally {
    try { await rm(tmpDir, { recursive: true, force: true }); } catch { /* ignore */ }
  }
}

// ── 格式化 markdown ──

function formatMemo(original: string, translated: string): string {
  const now = new Date();
  const date = now.toISOString().slice(0, 10);
  const time = now.toTimeString().slice(0, 8);
  return [
    `# Pi Prompt — ${date} ${time}`,
    '',
    '## 原始输入',
    '',
    original,
    '',
    '## English Translation',
    '',
    translated,
    '',
  ].join('\n');
}

// ── Sidecar 入口 ──

export default function (pi: ExtensionAPI) {
  // ── 1. 翻译输入 → daily ──
  pi.on('input', (event, ctx) => {
    // 只处理交互式用户输入
    if (event.source !== 'interactive') return;

    const text = event.text?.trim();
    if (!text) return;
    if (text.startsWith('/')) return;
    if (!shouldTranslate(text)) return;

    logger.log('info', 'daily_input_captured', { text: text.slice(0, 80) });

    // 异步处理，不阻塞 pi 事件循环
    (async () => {
      const startTime = Date.now();
      const cfg = loadPulseConfig();
      const model = cfg.openai?.model || DEFAULT_OPENAI_MODEL;

      try {
        const result = await translateOpenAI(text);
        const elapsed = Date.now() - startTime;
        const memo = formatMemo(text, result.translation);
        const tags = await getFileSyncTags(ctx.cwd, 'prompt');
        await syncToDaily(memo, tags);
        logger.log('info', 'daily_flow_complete', {
          original: text.slice(0, 40),
          translated: result.translation.slice(0, 40),
          elapsedMs: elapsed,
        });
      } catch (err: any) {
        logger.log('error', 'daily_flow_failed', { error: err.message, text: text.slice(0, 40) });
        // 保存原文
        const memo = formatMemo(text, `*（翻译失败: ${err.message}）*`);
        const tags = await getFileSyncTags(ctx.cwd, 'prompt').catch(() => ['pi', 'translate-failed']);
        await syncToDaily(memo, tags).catch(() => {});
      }
    })();
  });

  // ── 2. 文件写入 → daily ──
  pi.on('tool_result', (event, ctx) => {
    if (event.isError) return;
    if (event.toolName !== 'write' && event.toolName !== 'edit') return;

    const input = event.input as Record<string, unknown> | undefined;
    const toolPath = typeof input?.path === 'string' ? input.path : undefined;
    if (!toolPath) return;

    const fileName = basename(toolPath);
    const match = FILE_SYNC_PATTERNS.find(p => p.test(fileName));
    if (!match) return;

    const fileTag = match.source.replace(/\^|\$|\\?/g, '').toLowerCase();

    logger.log('info', 'daily_file_match', { fileName, filePath: toolPath, tag: fileTag });

    (async () => {
      try {
        const fullPath = isAbsolute(toolPath) ? toolPath : join(ctx.cwd, toolPath);
        const content = await readFile(fullPath, 'utf-8');
        const memo = formatFileMemo(toolPath, content);
        const tags = await getFileSyncTags(ctx.cwd, fileTag);
        await syncToDaily(memo, tags);
        logger.log('info', 'daily_file_synced', { fileName, tags });
      } catch (err: any) {
        logger.log('error', 'daily_file_sync_failed', { fileName, error: err.message });
      }
    })();
  });
}
