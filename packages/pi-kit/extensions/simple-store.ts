/**
 * SimpleStore — 极简安全配置存储
 *
 * 设计哲学：信任项目 + git 兜底，只防系统级破坏和写出项目外。
 * - 项目被信任后，内部所有操作放行（包括 rm -rf）
 * - 只拦截：系统破坏命令、写出项目外的文件操作、rm -rf 指向外部路径
 * - 单文件配置，无需双作用域、无需 glob 匹配、无需 shell 解析
 */

import * as fs from 'node:fs';
import * as path from 'node:path';
import * as os from 'node:os';

// ── 系统破坏命令正则 ──
// 这些命令无论在项目内还是项目外都要拦截/确认
const SYSTEM_DANGEROUS_PATTERNS: RegExp[] = [
  /^\s*sudo\b/,
  /^\s*doas\b/,
  /^\s*shutdown\b/,
  /^\s*reboot\b/,
  /^\s*halt\b/,
  /^\s*poweroff\b/,
  /^\s*init\b/,
  /^\s*passwd\b/,
  /^\s*useradd\b/,
  /^\s*userdel\b/,
  /^\s*usermod\b/,
  /^\s*mkfs\b/,
  /^\s*dd\b/,
  /^\s*fdisk\b/,
  /^\s*parted\b/,
  /^\s*mount\b/,
  /^\s*kill\b/,
  /^\s*pkill\b/,
  /^\s*killall\b/,
];

// rm -rf 指向项目外部的检测
const RM_RF_OUTSIDE = /\brm\s+(-rf?|-[a-z]*rf?[a-z]*)\s+(\/|~|\.\.)/;

export interface SimpleConfig {
  trustedProjects: string[];
  alwaysAllowCommands: string[];
  alwaysDenyCommands: string[];
}

export class SimpleStore {
  private configPath: string;

  constructor() {
    const configDir = path.join(os.homedir(), '.pi', 'agent');
    this.configPath = path.join(configDir, 'safe-run.json');
  }

  // ── I/O ──

  private load(): SimpleConfig {
    try {
      const raw = fs.readFileSync(this.configPath, 'utf-8');
      return JSON.parse(raw) as SimpleConfig;
    } catch {
      return { trustedProjects: [], alwaysAllowCommands: [], alwaysDenyCommands: [] };
    }
  }

  private save(config: SimpleConfig): void {
    const dir = path.dirname(this.configPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
    fs.writeFileSync(this.configPath, JSON.stringify(config, null, 2), 'utf-8');
  }

  // ── 项目信任 ──

  isProjectTrusted(projectPath: string): boolean {
    const config = this.load();
    return config.trustedProjects.includes(projectPath);
  }

  trustProject(projectPath: string): void {
    const config = this.load();
    if (!config.trustedProjects.includes(projectPath)) {
      config.trustedProjects.push(projectPath);
      this.save(config);
    }
  }

  untrustProject(projectPath: string): void {
    const config = this.load();
    config.trustedProjects = config.trustedProjects.filter((p) => p !== projectPath);
    this.save(config);
  }

  // ── 路径边界 ──

  static isPathWithinProject(filePath: string, cwd: string): boolean {
    // normalize cwd to handle trailing slashes
    const normalizedCwd = path.resolve(cwd);
    const resolved = path.resolve(normalizedCwd, filePath);
    return resolved === normalizedCwd || resolved.startsWith(normalizedCwd + path.sep);
  }

  // ── 命令危险检测 ──

  static isSystemDangerous(cmd: string): boolean {
    return SYSTEM_DANGEROUS_PATTERNS.some((re) => re.test(cmd));
  }

  static isRmdirOutside(cmd: string): boolean {
    return RM_RF_OUTSIDE.test(cmd);
  }

  // ── 白/黑名单 ──

  isCommandAlwaysAllowed(cmd: string): boolean {
    const config = this.load();
    return config.alwaysAllowCommands.includes(cmd);
  }

  isCommandAlwaysDenied(cmd: string): boolean {
    const config = this.load();
    return config.alwaysDenyCommands.includes(cmd);
  }

  addAllowedCommand(cmd: string): void {
    const config = this.load();
    if (!config.alwaysAllowCommands.includes(cmd)) {
      config.alwaysAllowCommands.push(cmd);
      this.save(config);
    }
  }

  addDeniedCommand(cmd: string): void {
    const config = this.load();
    if (!config.alwaysDenyCommands.includes(cmd)) {
      config.alwaysDenyCommands.push(cmd);
      this.save(config);
    }
  }

  // ── 查询方法（供 extension 命令使用）──

  getTrustedProjects(): string[] {
    return this.load().trustedProjects;
  }

  getAllAllowedCommands(): string[] {
    return this.load().alwaysAllowCommands;
  }

  getAllDeniedCommands(): string[] {
    return this.load().alwaysDenyCommands;
  }
}
