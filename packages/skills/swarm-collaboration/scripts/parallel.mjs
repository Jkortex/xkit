#!/usr/bin/env node
/**
 * parallel.mjs — 并行派发子任务给 pi agent
 *
 * 用法:
 *   node parallel.mjs <tasks.json>             人类可读输出
 *   node parallel.mjs <tasks.json> --json        JSON 输出
 *   cat tasks.json | node parallel.mjs           从 stdin 读
 *
 * tasks.json 格式:
 *   { "concurrency": 2, "tasks": [{ "id": "...", "instruction": "...", ... }] }
 *
 * 行为:
 *   - 每个任务 spawn 独立的 `pi -p "..."` 进程（无 shell）
 *   - 指令自动追加 handoff 要求，子 agent 写入 .agents/handoff-{id}.md
 *   - 并发由信号量控制
 *   - 全部成功退出 0，有失败退出 1
 */

import { readFileSync } from 'node:fs';
import { spawn } from 'node:child_process';
import { access, readFile, writeFile, mkdir } from 'node:fs/promises';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));

// ── 类型 ────────────────────────────────────────────

/** @typedef {{ id: string, instruction: string, model?: string, thinking?: string, files?: string[] }} TaskSpec */
/** @typedef {{ concurrency: number, tasks: TaskSpec[] }} TasksConfig */
/** @typedef {{ task_id: string, status: 'success'|'failed', output: string, handoff: string, error?: string, duration_ms: number }} TaskResult */

// ── 主入口 ──────────────────────────────────────────

async function main() {
  const args = process.argv.slice(2);
  const jsonMode = args.includes('--json');
  const filePath = args.find(a => !a.startsWith('-'));

  // 读取 tasks.json
  let raw;
  if (filePath) {
    raw = readFileSync(filePath, 'utf-8');
  } else if (!process.stdin.isTTY) {
    // 从 stdin 读（管道）
    raw = await readStdin();
  } else {
    console.error('Usage: node parallel.mjs <tasks.json> [--json]');
    process.exit(1);
  }

  /** @type {TasksConfig} */
  let config;
  try {
    config = JSON.parse(raw);
  } catch (e) {
    console.error(`Invalid tasks.json: ${e.message}`);
    process.exit(1);
  }

  const { tasks, concurrency = 2 } = config;
  if (!tasks?.length) {
    console.error('tasks.json: empty tasks array');
    process.exit(1);
  }

  if (!jsonMode) {
    console.log(`Tasks: ${tasks.length}, concurrency: ${concurrency}\n`);
  }

  // 确保 .agents 目录存在
  await mkdir('.agents', { recursive: true });

  // 信号量 + 执行
  const sem = new Semaphore(concurrency);
  const results = /** @type {TaskResult[]} */ ([]);
  let aborted = false;

  const running = tasks.map(async (task) => {
    await sem.acquire();
    if (aborted) {
      results.push({ task_id: task.id, status: 'failed', output: '', handoff: '', error: 'cancelled', duration_ms: 0 });
      sem.release();
      return;
    }

    const start = Date.now();
    try {
      const result = await runTask(task);
      result.duration_ms = Date.now() - start;
      results.push(result);
    } catch (err) {
      results.push({ task_id: task.id, status: 'failed', output: '', handoff: '', error: err.message, duration_ms: Date.now() - start });
    } finally {
      sem.release();
    }
  });

  await Promise.all(running);

  // 输出结果
  const successCount = results.filter(r => r.status === 'success').length;

  if (jsonMode) {
    process.stdout.write(JSON.stringify(results, null, 2) + '\n');
  } else {
    for (const r of results) {
      const icon = r.status === 'success' ? '✓' : '✗';
      console.log(`${icon} [${r.task_id}] ${r.status} (${r.duration_ms}ms)`);
      if (r.handoff) {
        const lines = r.handoff.split('\n').slice(0, 8).join('\n');
        console.log(`   ${lines.replace(/\n/g, '\n   ')}`);
      }
      if (r.error) console.log(`   error: ${r.error}`);
      console.log('');
    }
    console.log(`---\n${successCount}/${tasks.length} succeeded`);
  }

  process.exit(successCount === tasks.length ? 0 : 1);
}

// ── 执行单个任务 ────────────────────────────────────

/**
 * @param {TaskSpec} task
 * @returns {Promise<TaskResult>}
 */
async function runTask(task) {
  const handoffFile = `.agents/handoff-${task.id}.md`;

  // 构建指令：追加 handoff 要求
  const handoffPrompt = `
完成后，将完成总结写入 ${handoffFile} 文件，格式如下：

\`\`\`
# Handoff: ${task.id}

## 完成内容
- ...

## 修改的文件
- ...

## 遗留问题
- ...
\`\`\`

任务内容:
${task.instruction}`.trim();

  // 构建 pi 参数
  const piArgs = ['-p', handoffPrompt, '--no-session'];
  if (task.model) piArgs.push('--model', task.model);
  if (task.thinking) piArgs.push('--thinking', task.thinking);

  // spawn（无 shell，防注入）
  const child = spawn('pi', piArgs, {
    stdio: ['ignore', 'pipe', 'pipe'],
    env: { ...process.env },
    signal: AbortSignal.timeout(300_000), // 5min 超时
  });

  const stdoutChunks = [];
  const stderrChunks = [];

  child.stdout.on('data', (chunk) => stdoutChunks.push(chunk));
  child.stderr.on('data', (chunk) => stderrChunks.push(chunk));

  const exitCode = await new Promise((resolve) => {
    child.on('close', resolve);
    child.on('error', (err) => {
      // If spawn itself fails, reject will be caught by the promise wrapper below
      stderrChunks.push(Buffer.from(err.message));
      resolve(1);
    });
  });

  const output = Buffer.concat(stdoutChunks).toString('utf-8').trim();
  const stderr = Buffer.concat(stderrChunks).toString('utf-8').trim();
  const combined = [output, stderr].filter(Boolean).join('\n').slice(0, 2000);

  // 读 handoff 文件（不管成功失败都尝试）
  let handoff = '';
  try {
    await access(handoffFile);
    handoff = (await readFile(handoffFile, 'utf-8')).trim();
  } catch {
    // handoff 文件不存在
  }

  // 判定结果
  if (exitCode === 0) {
    return { task_id: task.id, status: 'success', output, handoff, duration_ms: 0 };
  }

  // 进程失败但手写了 handoff → 按成功算（子 agent 完成了工作但 pi 退出码异常）
  if (handoff) {
    return { task_id: task.id, status: 'success', output, handoff, duration_ms: 0 };
  }

  return { task_id: task.id, status: 'failed', output, handoff, error: `exit code ${exitCode}`, duration_ms: 0 };
}

// ── 工具 ────────────────────────────────────────────

class Semaphore {
  /** @param {number} max */
  constructor(max) {
    this._max = max;
    this._count = max;
    /** @type {Array<() => void>} */
    this._queue = [];
  }

  acquire() {
    if (this._count > 0) { this._count--; return Promise.resolve(); }
    return new Promise(resolve => this._queue.push(resolve));
  }

  release() {
    const next = this._queue.shift();
    if (next) setTimeout(next, 0);
    else this._count++;
  }
}

/** @returns {Promise<string>} */
function readStdin() {
  return new Promise((resolve, reject) => {
    const chunks = [];
    process.stdin.setEncoding('utf-8');
    process.stdin.on('data', c => chunks.push(c));
    process.stdin.on('end', () => resolve(chunks.join('')));
    process.stdin.on('error', reject);
  });
}

main().catch(err => {
  console.error(`Fatal: ${err.message}`);
  process.exit(1);
});
