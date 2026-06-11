/**
 * safe-run.ts 测试（简化版）
 *
 * 测试基于项目信任的简单安全模型。
 * 注意：safe-run 不再注册任何 slash 命令，只通过事件自动拦截。
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { mkdtempSync, mkdirSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { SimpleStore } from '../extensions/simple-store.js';

// ── 最小 mock ──

type ToolCallHandler = (event: any, ctx: any) => Promise<any>;

interface MockExtensionAPI {
  on: ReturnType<typeof vi.fn>;
  registerCommand: ReturnType<typeof vi.fn>;
  registerTool: ReturnType<typeof vi.fn>;
  toolHandlers: Map<string, ToolCallHandler[]>;
}

function createMockAPI(): MockExtensionAPI {
  const handlers = new Map<string, ToolCallHandler[]>();

  return {
    on: vi.fn((event: string, handler: ToolCallHandler) => {
      let list = handlers.get(event);
      if (!list) {
        list = [];
        handlers.set(event, list);
      }
      list.push(handler);
    }),
    registerCommand: vi.fn(),
    registerTool: vi.fn(),
    toolHandlers: handlers,
  };
}

function createMockCtx(
  overrides?: Partial<{
    hasUI: boolean;
    ui_select_result: string | null;
    cwd: string;
  }>,
) {
  const cwd = overrides?.cwd ?? process.cwd();
  return {
    cwd,
    hasUI: overrides?.hasUI ?? true,
    ui: {
      select: vi.fn(async (_title: string, _options: string[]) => {
        return overrides?.ui_select_result ?? null;
      }),
      notify: vi.fn((_msg: string, _level: string) => {}),
    },
  };
}

// ── ──

describe('safe-run 集成层（简化版）', () => {
  let tmp: string;
  let cwd: string;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), 'saferun-simple-test-'));
    cwd = join(tmp, 'project');
    mkdirSync(cwd, { recursive: true });

    vi.stubEnv('HOME', tmp);
    vi.spyOn(process, 'cwd').mockReturnValue(cwd as any);
    vi.resetModules();
  });

  afterEach(() => {
    rmSync(tmp, { recursive: true, force: true });
    vi.unstubAllEnvs();
    vi.restoreAllMocks();
  });

  describe('注册', () => {
    it('注册了 tool_call 事件处理器', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const toolCalls = api.on.mock.calls.filter(
        ([e]: [string]) => e === 'tool_call',
      );
      expect(toolCalls.length).toBeGreaterThanOrEqual(1);
    });

    it('不再注册任何 slash 命令', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      expect(api.registerCommand).not.toHaveBeenCalled();
    });
  });

  describe('session_start 项目信任', () => {
    it('启动时未信任项目显示询问', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('session_start')![0];
      const event = { reason: 'startup' };
      const selectMock = vi.fn().mockResolvedValue('🚫 暂不信任，每次写入询问');
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      await handler(event, ctx);
      expect(selectMock).toHaveBeenCalled();
    });

    it('已信任项目不显示询问', async () => {
      // 先通过 SimpleStore 信任项目
      const store = new SimpleStore();
      store.trustProject(cwd);

      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('session_start')![0];
      const event = { reason: 'startup' };
      const selectMock = vi.fn();
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      await handler(event, ctx);
      expect(selectMock).not.toHaveBeenCalled();
    });

    it('非 startup 的 session_start 跳过', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('session_start')![0];
      const selectMock = vi.fn();
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      for (const reason of ['new', 'resume', 'fork', 'reload']) {
        await handler({ reason }, ctx);
      }
      expect(selectMock).not.toHaveBeenCalled();
    });

    it('无 UI 时跳过', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('session_start')![0];
      const event = { reason: 'startup' };
      const selectMock = vi.fn();
      const ctx = createMockCtx({ cwd, hasUI: false });
      ctx.ui.select = selectMock;

      await handler(event, ctx);
      expect(selectMock).not.toHaveBeenCalled();
    });

    it('选择信任后持久化', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('session_start')![0];
      const event = { reason: 'startup' };
      const selectMock = vi.fn().mockResolvedValue('🔓 信任此项目');
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      await handler(event, ctx);

      // 再次调用不应再询问
      const selectMock2 = vi.fn();
      const ctx2 = createMockCtx({ cwd, hasUI: true });
      ctx2.ui.select = selectMock2;
      await handler(event, ctx2);
      expect(selectMock2).not.toHaveBeenCalled();
    });
  });

  describe('bash 工具拦截', () => {
    it('安全命令直接放行', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'ls -la' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ cwd });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
    });

    it('npm install 直接放行', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'npm install' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ cwd });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
    });

    it('rm -rf node_modules 放行（项目内）', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'rm -rf node_modules' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ cwd });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
    });

    it('系统破坏命令无 UI 时拦截', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'sudo rm -rf /' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ hasUI: false, cwd });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拦截'),
      });
    });

    it('系统破坏命令用户拒绝时拦截', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'sudo rm -rf /' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({
        hasUI: true,
        ui_select_result: '🚫 拒绝',
        cwd,
      });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拒绝'),
      });
    });

    it('系统破坏命令用户始终允许后放行', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'shutdown -h now' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({
        hasUI: true,
        ui_select_result: '🔓 始终允许此命令',
        cwd,
      });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();

      // 第二次应直接放行
      const result2 = await handler(event, ctx);
      expect(result2).toBeUndefined();
    });

    it('系统破坏命令用户始终拒绝后拦截', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'dd if=/dev/zero of=/dev/sda' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({
        hasUI: true,
        ui_select_result: '🔒 始终拒绝此命令',
        cwd,
      });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拒绝'),
      });

      // 第二次应直接拦截
      const result2 = await handler(event, ctx);
      expect(result2).toEqual({
        block: true,
        reason: expect.stringContaining('禁止'),
      });
    });

    it('rm -rf /tmp 拦截（外部路径）', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'rm -rf /tmp/cache' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ hasUI: false, cwd });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拦截'),
      });
    });

    it('rm -rf ../other 拦截（父目录）', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'bash',
        input: { command: 'rm -rf ../other-project/dist' },
        toolCallId: '1',
      };
      const ctx = createMockCtx({ hasUI: false, cwd });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拦截'),
      });
    });
  });

  describe('write/edit 工具拦截', () => {
    it('信任项目内路径直接放行', async () => {
      // 先通过 SimpleStore 信任项目
      const store = new SimpleStore();
      store.trustProject(cwd);

      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'write',
        input: { path: join(cwd, 'src/index.ts'), content: 'test' },
        toolCallId: '2',
      };
      const ctx = createMockCtx({ cwd });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
    });

    it('信任项目内 .env 也直接放行（不再特殊处理敏感文件）', async () => {
      const store = new SimpleStore();
      store.trustProject(cwd);

      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'write',
        input: { path: join(cwd, '.env'), content: 'KEY=val' },
        toolCallId: '2',
      };
      const ctx = createMockCtx({ cwd });

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
    });

    it('信任项目但写到外部时询问', async () => {
      const store = new SimpleStore();
      store.trustProject(cwd);

      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'write',
        input: { path: '/tmp/outside.txt', content: 'test' },
        toolCallId: '2',
      };
      const selectMock = vi.fn().mockResolvedValue('🚫 拒绝');
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拒绝'),
      });
      expect(selectMock).toHaveBeenCalled();
    });

    it('未信任项目 + 无 UI 时拦截', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'write',
        input: { path: join(cwd, 'untracked.txt'), content: 'test' },
        toolCallId: '2',
      };
      const ctx = createMockCtx({ hasUI: false, cwd });

      const result = await handler(event, ctx);
      expect(result).toEqual({
        block: true,
        reason: expect.stringContaining('拦截'),
      });
    });

    it('未信任项目 + 有 UI 时询问', async () => {
      const api = createMockAPI();
      const mod = await import('../extensions/safe-run.js');
      mod.default(api as any);

      const handler = api.toolHandlers.get('tool_call')![0];
      const event = {
        toolName: 'edit',
        input: { path: join(cwd, 'src/file.ts'), content: 'edit' },
        toolCallId: '2',
      };
      const selectMock = vi.fn().mockResolvedValue('✅ 允许一次');
      const ctx = createMockCtx({ cwd, hasUI: true });
      ctx.ui.select = selectMock;

      const result = await handler(event, ctx);
      expect(result).toBeUndefined();
      expect(selectMock).toHaveBeenCalled();
    });
  });
});
