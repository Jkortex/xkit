/**
 * ConfigStore — 安全规则配置存储与查询
 *
 * 零 pi 依赖，可独立测试与复用。
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

// ============================================================
// Types
// ============================================================

export interface SafeRunConfig {
  trustedPaths: string[];
  alwaysAllowCommands: string[];
  alwaysDenyCommands: string[];
  dangerousPatterns: string[];
  sensitivePatterns: string[];
}

export type Scope = 'global' | 'project';

// ============================================================
// Defaults
// ============================================================

const DEFAULT_DANGEROUS_PATTERNS: string[] = [
  '\\brm\\s+-rf?\\b',
  '\\bsudo\\b',
  '\\bchmod\\b.*777',
  '\\bchown\\b',
  '\\bdd\\b',
  '\\bmkfs\\b',
  '\\bcurl\\b.*\\|',
  '\\bwget\\b.*-O\\s*-\\s*\\|',
  '\\bbase64\\b.*\\|\\s*bash\\b',
  '\\beval\\b',
  '\\bexec\\b',
  '\\bpasswd\\b',
  '\\bkill\\b',
  '\\bpkill\\b',
  '\\bshutdown\\b',
  '\\breboot\\b',
  '\\binit\\s+0\\b',
  '\\b>\\s+/dev/',
  '\\bmv\\s+.*\\s+/dev/null',
  '\\b:\\(\\)\\s*\\{',
  '\\bsystemctl\\b',
  '\\bjournalctl\\b.*--vacuum',
  '\\bdpkg\\b',
  '\\bapt\\s+(remove|purge|autoremove)',
  '\\bpip\\s+uninstall',
  '\\bnpm\\s+(uninstall|remove)',
  '\\byarn\\s+remove',
  '\\bpnpm\\s+(remove|uninstall)',
  '\\bdocker\\s+(rm|rmi|system\\s+prune)',
  '\\bgit\\s+push\\s+--force',
  '\\bgit\\s+reset\\s+--hard',
];

const DEFAULT_SENSITIVE_PATTERNS: string[] = [
  '.env',
  '.env.*',
  '.git/',
  'node_modules/',
  '*.pem',
  '*.key',
  'config.*.json',
  'settings.json',
  'safe-run.json',
  'dist/',
  'build/',
  '.next/',
  'package-lock.json',
  'pnpm-lock.yaml',
  'yarn.lock',
];

// ============================================================
// ConfigStore
// ============================================================

export class ConfigStore {
  private readonly globalDir: string;
  private readonly cwd: string;
  private readonly defaultTrustedPaths: string[];

  private mergedGlobal!: SafeRunConfig;
  private mergedProject!: SafeRunConfig;
  private cachedGlobalMtime = -1;
  private cachedProjectMtime = -1;

  /**
   * @param globalDir  全局配置目录，通常是 `~/.pi/agent`
   * @param cwd        项目工作目录
   */
  constructor(globalDir: string, cwd: string) {
    this.globalDir = globalDir;
    this.cwd = cwd;

    const globalPiDir = path.dirname(globalDir); // ~/.pi
    this.defaultTrustedPaths = [
      globalPiDir.endsWith(path.sep) ? globalPiDir : globalPiDir + path.sep,
      path.join(cwd, '.pi') + path.sep,
    ];

    this.ensureFresh();
  }

  // ── 作用域 ──

  get defaultScope(): Scope {
    return fs.existsSync(path.join(this.cwd, '.pi')) ? 'project' : 'global';
  }

  // ── 懒加载：仅在文件变更时重新读取 ──

  private getMtime(filePath: string): number {
    try {
      return fs.statSync(filePath).mtimeMs;
    } catch {
      return 0;
    }
  }

  private ensureFresh(): void {
    const globalPath = this.getConfigPath('global');
    const projectPath = this.getConfigPath('project');

    const globalMtime = this.getMtime(globalPath);
    const projectMtime = this.getMtime(projectPath);

    if (globalMtime !== this.cachedGlobalMtime) {
      this.mergedGlobal = this.loadScope('global');
      this.cachedGlobalMtime = globalMtime;
    }
    if (projectMtime !== this.cachedProjectMtime) {
      this.mergedProject = this.loadScope('project');
      this.cachedProjectMtime = projectMtime;
    }
  }

  /** 标记缓存过期，下次查询时自动重载 */
  private markDirty(): void {
    this.cachedGlobalMtime = -1;
    this.cachedProjectMtime = -1;
  }

  /** 立即重新加载（对测试/调试友好） */
  refresh(): void {
    this.ensureFresh();
  }

  // ── 路径匹配 ──

  private expandPath(p: string): string {
    if (p.startsWith('~')) p = path.join(os.homedir(), p.slice(1));
    return path.resolve(this.cwd, p);
  }

  private pathMatches(targetPath: string, pattern: string): boolean {
    const target = this.expandPath(targetPath);

    if (pattern.startsWith('*.')) return target.endsWith(pattern.slice(1));

    const normPattern = this.expandPath(pattern);
    if (normPattern.endsWith('/') || normPattern.endsWith(path.sep)) {
      const prefix = normPattern.replace(/[/\\]$/, '');
      return target.startsWith(prefix + path.sep) || target === prefix;
    }
    return target === normPattern || target.startsWith(normPattern + path.sep);
  }

  // ── 文件 I/O ──

  private xkitConfigPath(scope: Scope): string {
    const dir =
      scope === 'global'
        ? path.join(os.homedir(), '.xkit')
        : path.join(this.cwd, '.xkit');
    return path.join(dir, 'config.json');
  }

  private legacyConfigPath(scope: Scope): string {
    const dir =
      scope === 'global' ? this.globalDir : path.join(this.cwd, '.pi');
    return path.join(dir, 'safe-run.json');
  }

  /** Read config: try ~/.xkit/config.json → safe-run namespace, fall back to legacy safe-run.json. */
  private readXkitNamespace(scope: Scope): SafeRunConfig | null {
    try {
      const raw = JSON.parse(
        fs.readFileSync(this.xkitConfigPath(scope), 'utf-8'),
      );
      return raw['safe-run'] ?? null;
    } catch {
      return null;
    }
  }

  /** Write config to ~/.xkit/config.json (unified). Merge into existing file, keep other namespaces intact. */
  private writeXkitNamespace(scope: Scope, safeRunConfig: SafeRunConfig): void {
    const configPath = this.xkitConfigPath(scope);
    let merged: Record<string, unknown> = {};
    try {
      merged = JSON.parse(fs.readFileSync(configPath, 'utf-8'));
    } catch {
      /* new file */
    }
    merged['safe-run'] = safeRunConfig;
    const dir = path.dirname(configPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(configPath, JSON.stringify(merged, null, 2), 'utf-8');
  }

  private getConfigPath(scope: Scope): string {
    const xkitPath = this.xkitConfigPath(scope);
    if (fs.existsSync(xkitPath)) return xkitPath;
    return this.legacyConfigPath(scope);
  }

  private loadScope(scope: Scope): SafeRunConfig {
    // Try unified config first, then legacy
    const fromXkit = this.readXkitNamespace(scope);
    if (fromXkit) return fromXkit;

    try {
      return JSON.parse(fs.readFileSync(this.legacyConfigPath(scope), 'utf-8'));
    } catch {
      return {
        trustedPaths: [],
        alwaysAllowCommands: [],
        alwaysDenyCommands: [],
        dangerousPatterns: [],
        sensitivePatterns: [],
      };
    }
  }

  private saveScope(scope: Scope, config: SafeRunConfig): void {
    this.writeXkitNamespace(scope, config);
  }

  // ── 命令解析 ──

  /** 按 &&, ||, ;, |, & 分割命令，保留引号内完整性 */
  private splitCommand(cmd: string): string[] {
    const segments: string[] = [];
    let cur = '';
    let inS = false,
      inD = false,
      inB = false;

    for (let i = 0; i < cmd.length; i++) {
      const ch = cmd[i];
      if (ch === "'" && !inD && !inB) {
        inS = !inS;
        cur += ch;
        continue;
      }
      if (ch === '"' && !inS && !inB) {
        inD = !inD;
        cur += ch;
        continue;
      }
      if (ch === '`' && !inS && !inD) {
        inB = !inB;
        cur += ch;
        continue;
      }

      if (!inS && !inD && !inB) {
        const next = cmd[i + 1];
        const isSep =
          (ch === '&' && next === '&') ||
          (ch === '|' && next === '|') ||
          ch === ';' ||
          (ch === '|' && next !== '|') ||
          (ch === '&' && next !== '&' && next !== undefined);
        if (isSep) {
          const t = cur.trim();
          if (t) segments.push(t);
          cur = '';
          if ((ch === '&' && next === '&') || (ch === '|' && next === '|')) i++;
          continue;
        }
      }
      cur += ch;
    }
    const t = cur.trim();
    if (t) segments.push(t);
    return segments;
  }

  /** 移除引号内内容，防止引号内字符串触发模式匹配 */
  private stripQuoted(s: string): string {
    return s
      .replace(/'[^']*'/g, '')
      .replace(/"[^"]*"/g, '')
      .replace(/`[^`]*`/g, '');
  }

  /** 提取命令中所有被引号包围的字符串段 */
  private extractQuotedSegments(cmd: string): string[] {
    const segments: string[] = [];
    const regex = /(["'])(.*?)\1/g;
    let match: RegExpExecArray | null;
    while ((match = regex.exec(cmd)) !== null) {
      segments.push(match[2]);
    }
    return segments;
  }

  // ── 查询 ──

  isPathTrusted(filePath: string): boolean {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.trustedPaths,
      ...this.mergedProject.trustedPaths,
      ...this.defaultTrustedPaths,
    ].some((p) => this.pathMatches(filePath, p));
  }

  isPathSensitive(filePath: string): boolean {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.sensitivePatterns,
      ...this.mergedProject.sensitivePatterns,
      ...DEFAULT_SENSITIVE_PATTERNS,
    ].some((p) => this.pathMatches(filePath, p));
  }

  /**
   * 检查命令是否在允许列表中。
   * 匹配规则：任一命令段以允许列表中的条目开头（后跟空格或完全一致），
   * 避免子串误匹配（如 "rm" 不匹配 "remove"）。
   */
  isCommandAlwaysAllowed(cmd: string): boolean {
    this.ensureFresh();
    const segments = this.splitCommand(cmd);
    const allowed = [
      ...this.mergedGlobal.alwaysAllowCommands,
      ...this.mergedProject.alwaysAllowCommands,
    ];
    return segments.some((seg) =>
      allowed.some((c) => seg === c || seg.startsWith(c + ' ')),
    );
  }

  /**
   * 检查命令是否在拒绝列表中。
   * 匹配规则同 isCommandAlwaysAllowed。
   */
  isCommandAlwaysDenied(cmd: string): boolean {
    this.ensureFresh();
    const segments = this.splitCommand(cmd);
    const denied = [
      ...this.mergedGlobal.alwaysDenyCommands,
      ...this.mergedProject.alwaysDenyCommands,
    ];
    return segments.some((seg) =>
      denied.some((c) => seg === c || seg.startsWith(c + ' ')),
    );
  }

  isCommandDangerous(cmd: string): boolean {
    this.ensureFresh();
    // 危险模式需要跨管道检测（如 curl ... | bash），所以用完整命令但去掉引号内容
    const stripped = this.stripQuoted(cmd);
    const patterns = [
      ...this.mergedGlobal.dangerousPatterns,
      ...this.mergedProject.dangerousPatterns,
      ...DEFAULT_DANGEROUS_PATTERNS,
    ];
    const dangerous = patterns.some((p) => {
      try {
        return new RegExp(p, 'i').test(stripped);
      } catch {
        return false;
      }
    });
    if (dangerous) return true;

    // 二次检测：对 meta-commands（bash -c, sh -c 等），检查引号内是否藏有危险操作
    if (/^\s*(?:ba)?sh\s+-c\s*$/i.test(stripped)) {
      const quoted = this.extractQuotedSegments(cmd);
      for (const seg of quoted) {
        if (
          patterns.some((p) => {
            try {
              return new RegExp(p, 'i').test(seg);
            } catch {
              return false;
            }
          })
        ) {
          return true;
        }
      }
    }

    return false;
  }

  // ── 可变操作（自动检测作用域）──

  addTrustedPath(filePath: string): { scope: Scope } {
    const scope = this.defaultScope;
    const config = this.loadScope(scope);
    if (!config.trustedPaths.includes(filePath)) {
      config.trustedPaths.push(filePath);
      this.saveScope(scope, config);
      this.markDirty();
    }
    return { scope };
  }

  removeTrustedPath(filePath: string): { found: boolean; scope?: Scope } {
    for (const scope of ['project', 'global'] as Scope[]) {
      const config = this.loadScope(scope);
      const idx = config.trustedPaths.indexOf(filePath);
      if (idx !== -1) {
        config.trustedPaths.splice(idx, 1);
        this.saveScope(scope, config);
        this.markDirty();
        return { found: true, scope };
      }
    }
    return { found: false };
  }

  addAllowedCommand(cmd: string): { scope: Scope } {
    const scope = this.defaultScope;
    const config = this.loadScope(scope);
    if (config.alwaysAllowCommands.includes(cmd)) return { scope };
    config.alwaysAllowCommands.push(cmd);
    this.saveScope(scope, config);
    this.markDirty();
    return { scope };
  }

  addDeniedCommand(cmd: string): { scope: Scope } {
    const scope = this.defaultScope;
    const config = this.loadScope(scope);
    if (config.alwaysDenyCommands.includes(cmd)) return { scope };
    config.alwaysDenyCommands.push(cmd);
    this.saveScope(scope, config);
    this.markDirty();
    return { scope };
  }

  addDangerousPattern(pattern: string): { scope: Scope } {
    const scope = this.defaultScope;
    const config = this.loadScope(scope);
    config.dangerousPatterns.push(pattern);
    this.saveScope(scope, config);
    this.markDirty();
    return { scope };
  }

  addSensitivePattern(pattern: string): { scope: Scope } {
    const scope = this.defaultScope;
    const config = this.loadScope(scope);
    config.sensitivePatterns.push(pattern);
    this.saveScope(scope, config);
    this.markDirty();
    return { scope };
  }

  // ── 列表查询 ──

  getTrustedPaths(): {
    global: string[];
    project: string[];
    default: string[];
  } {
    this.ensureFresh();
    return {
      global: this.mergedGlobal.trustedPaths,
      project: this.mergedProject.trustedPaths,
      default: [...this.defaultTrustedPaths],
    };
  }

  getAlwaysAllowCommands(): string[] {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.alwaysAllowCommands,
      ...this.mergedProject.alwaysAllowCommands,
    ];
  }

  getAlwaysDenyCommands(): string[] {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.alwaysDenyCommands,
      ...this.mergedProject.alwaysDenyCommands,
    ];
  }

  getDangerousPatterns(): string[] {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.dangerousPatterns,
      ...this.mergedProject.dangerousPatterns,
      ...DEFAULT_DANGEROUS_PATTERNS,
    ];
  }

  getSensitivePatterns(): string[] {
    this.ensureFresh();
    return [
      ...this.mergedGlobal.sensitivePatterns,
      ...this.mergedProject.sensitivePatterns,
      ...DEFAULT_SENSITIVE_PATTERNS,
    ];
  }
}
