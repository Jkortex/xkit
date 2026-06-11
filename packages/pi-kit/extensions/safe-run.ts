/**
 * Safe Run Extension (简化版)
 *
 * 设计：信任项目 + git 兜底
 * - 信任的项目内：读写全放行（包括 rm -rf node_modules 等）
 * - 项目外写入：询问确认
 * - 系统破坏命令（sudo, shutdown 等）：询问确认
 * - rm -rf 指向外部路径（/、~、..）：询问确认
 * - 其余全部放行
 */

import type { ExtensionAPI } from '@earendil-works/pi-coding-agent';
import { isToolCallEventType } from '@earendil-works/pi-coding-agent';
import * as path from 'node:path';
import { SimpleStore } from './simple-store.js';
import { FileLogger } from './file-logger.js';

const logger = new FileLogger();

export default function (pi: ExtensionAPI) {
  const store = new SimpleStore();

  // ── Session start: 询问是否信任当前工作目录 ──
  pi.on('session_start', async (event, ctx) => {
    if (event.reason !== 'startup') return;

    const cwd = ctx.cwd;
    const alreadyTrusted = store.isProjectTrusted(cwd);
    logger.log('info', 'session_start', { cwd, alreadyTrusted, reason: event.reason });
    if (alreadyTrusted) return;
    if (!ctx.hasUI) {
      logger.log('info', 'session_start_no_ui', { cwd, action: 'skip_trust_prompt' });
      return;
    }

    const choice = await ctx.ui.select(
      `🛡️ 信任此项目？\n\n  ${cwd}\n\n信任后 agent 可自由读写项目内文件。项目内的 rm -rf 等操作不受限制（git 可恢复），但系统破坏命令和写出项目外的操作仍会提醒。`,
      ['🔓 信任此项目', '🚫 暂不信任，每次写入询问'],
    );

    if (choice?.startsWith('🔓')) {
      store.trustProject(cwd);
      logger.log('info', 'project_trusted', { cwd, source: 'session_start_prompt' });
      ctx.ui.notify(`已信任项目: ${cwd}`, 'info');
    } else {
      logger.log('info', 'project_not_trusted', { cwd, action: 'per_write_prompt' });
    }
  });

  // ── Tool call: 统一入口，按工具名分派 ──
  pi.on('tool_call', async (event, ctx) => {
    if (isToolCallEventType('bash', event)) {
      return handleBash(event, ctx, store);
    }
    if (
      isToolCallEventType('write', event) ||
      isToolCallEventType('edit', event)
    ) {
      return handleFile(event, ctx, store);
    }
  });

  /** 处理文件写入/编辑的路径保护 */
  async function handleFile(
    event: any,
    ctx: any,
    store: SimpleStore,
  ): Promise<{ block: true; reason: string } | undefined> {
    const cwd = ctx.cwd;
    const toolPath = event.input.path;
    const filePath = path.resolve(cwd, toolPath);
    const toolName = event.toolName as string;
    const label = toolName === 'write' ? '写入' : '编辑';

    const projectTrusted = store.isProjectTrusted(cwd);
    const insideProject = SimpleStore.isPathWithinProject(filePath, cwd);

    logger.log('info', 'file_operation', {
      toolName,
      filePath,
      cwd,
      projectTrusted,
      insideProject,
    });

    // 项目已信任 + 在项目内 → 直接放行
    if (projectTrusted && insideProject) {
      logger.log('info', 'file_allowed', { filePath, reason: 'trusted_project_inside' });
      return;
    }

    // 项目已信任但写到项目外 → 询问
    // 项目未信任 → 询问
    if (!ctx.hasUI) {
      logger.log('warn', 'file_blocked', {
        filePath,
        reason: 'no_ui_non_interactive',
        projectTrusted,
      });
      return {
        block: true,
        reason: `未信任路径的 ${label} 操作已在非交互模式拦截: ${filePath}`,
      };
    }

    const outsideHint = projectTrusted ? '（项目外路径）' : '（项目未信任）';
    const choice = await ctx.ui.select(
      `🛡️ ${label}确认 ${outsideHint}\n\n  ${filePath}\n\n请选择操作:`,
      ['✅ 允许一次', '🚫 拒绝', '🔓 信任此路径'],
    );

    if (!choice || choice.startsWith('🚫')) {
      logger.log('info', 'file_blocked', { filePath, reason: 'user_rejected', choice });
      return {
        block: true,
        reason: `用户拒绝 ${label} 路径: ${filePath}`,
      };
    }

    logger.log('info', 'file_allowed', { filePath, reason: 'user_allowed', choice });
    // "信任此路径" — 但我们现在不维护路径白名单了，
    // 所以只是放行本次。用户可以改为信任项目本身。
    // 放行
    return;
  }

  /** 处理危险命令确认 */
  async function handleBash(
    event: { input: { command: string } },
    ctx: any,
    store: SimpleStore,
  ): Promise<{ block: true; reason: string } | undefined> {
    const cmd = event.input.command.trim();
    if (!cmd) {
      logger.log('warn', 'bash_empty_command', {});
      return;
    }

    logger.log('info', 'bash_operation', { command: cmd });

    // 白/黑名单（精确匹配整条命令）
    if (store.isCommandAlwaysAllowed(cmd)) {
      logger.log('info', 'bash_allowed', { command: cmd, reason: 'whitelisted' });
      return;
    }
    if (store.isCommandAlwaysDenied(cmd)) {
      logger.log('info', 'bash_blocked', { command: cmd, reason: 'blacklisted' });
      return { block: true, reason: `命令已被管理员禁止执行: ${cmd}` };
    }

    // 判断是否需要拦截
    let dangerType: string | null = null;
    if (SimpleStore.isSystemDangerous(cmd)) {
      dangerType = '系统破坏命令';
    } else if (SimpleStore.isRmdirOutside(cmd)) {
      dangerType = '项目外路径删除';
    }

    if (!dangerType) {
      logger.log('info', 'bash_allowed', { command: cmd, reason: 'safe_command' });
      return; // 安全，放行
    }

    logger.log('info', 'bash_danger_detected', { command: cmd, dangerType });

    if (!ctx.hasUI) {
      logger.log('warn', 'bash_blocked', {
        command: cmd,
        dangerType,
        reason: 'no_ui_non_interactive',
      });
      return {
        block: true,
        reason: `${dangerType}已在非交互模式拦截: ${cmd}`,
      };
    }

    const choice = await ctx.ui.select(
      `⚠️ ${dangerType}\n\n  ${cmd}\n\n请选择操作:`,
      ['✅ 允许一次', '🚫 拒绝', '🔓 始终允许此命令', '🔒 始终拒绝此命令'],
    );

    if (!choice || choice.startsWith('🚫')) {
      logger.log('info', 'bash_blocked', { command: cmd, dangerType, reason: 'user_rejected' });
      return { block: true, reason: '用户拒绝执行' };
    }
    if (choice.startsWith('🔓')) {
      store.addAllowedCommand(cmd);
      logger.log('info', 'bash_allowlisted', { command: cmd, dangerType });
      ctx.ui.notify(`已添加至允许白名单`, 'info');
      return;
    }
    if (choice.startsWith('🔒')) {
      store.addDeniedCommand(cmd);
      logger.log('info', 'bash_denylisted', { command: cmd, dangerType });
      ctx.ui.notify(`已添加至拒绝黑名单`, 'info');
      return { block: true, reason: '用户选择始终拒绝此命令' };
    }

    // fallback（理论上不会走到这里）
    logger.log('warn', 'bash_unexpected_choice', { command: cmd, dangerType, choice });
    return;
  }

}
