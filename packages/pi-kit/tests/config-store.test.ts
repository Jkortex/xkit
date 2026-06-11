import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  mkdtempSync,
  mkdirSync,
  writeFileSync,
  rmSync,
  readFileSync,
  existsSync,
} from 'node:fs';
import { join } from 'node:path';
import { tmpdir, homedir } from 'node:os';
import { ConfigStore } from '../extensions/config-store';

describe('ConfigStore', () => {
  let tmp: string;
  let globalDir: string;
  let cwd: string;
  let store: ConfigStore;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), 'configstore-test-'));
    globalDir = join(tmp, '.pi', 'agent');
    cwd = join(tmp, 'project');
    mkdirSync(globalDir, { recursive: true });
    mkdirSync(join(cwd, '.pi'), { recursive: true });
    vi.stubEnv('HOME', tmp);
    store = new ConfigStore(globalDir, cwd);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    rmSync(tmp, { recursive: true, force: true });
  });

  // ── 作用域自动检测 ──

  describe('defaultScope', () => {
    it('返回 project 当 cwd 下存在 .pi/ 目录', () => {
      expect(store.defaultScope).toBe('project');
    });

    it('返回 global 当 cwd 下不存在 .pi/ 目录', () => {
      const store2 = new ConfigStore(globalDir, join(tmp, 'no-pi-project'));
      expect(store2.defaultScope).toBe('global');
    });
  });

  // ── 默认信任路径 ──

  describe('isPathTrusted', () => {
    it('默认信任 globalDir 所在目录', () => {
      expect(store.isPathTrusted(join(globalDir, 'settings.json'))).toBe(true);
    });

    it('默认信任 cwd 下的 .pi/', () => {
      expect(store.isPathTrusted(join(cwd, '.pi', 'safe-run.json'))).toBe(true);
    });

    it('普通路径默认不信任', () => {
      expect(store.isPathTrusted(join(cwd, 'src', 'index.ts'))).toBe(false);
    });

    it('添加后路径被信任', () => {
      const testPath = join(cwd, 'src');
      store.addTrustedPath(testPath);
      expect(store.isPathTrusted(testPath)).toBe(true);
    });
  });

  // ── 敏感路径 ──

  describe('isPathSensitive', () => {
    it('.env 文件为敏感', () => {
      expect(store.isPathSensitive(join(cwd, '.env'))).toBe(true);
    });

    it('.git/ 下路径为敏感', () => {
      expect(store.isPathSensitive(join(cwd, '.git', 'config'))).toBe(true);
    });

    it('node_modules/ 下路径为敏感', () => {
      expect(
        store.isPathSensitive(join(cwd, 'node_modules', 'lodash', 'index.js')),
      ).toBe(true);
    });

    it('*.pem 文件为敏感', () => {
      expect(store.isPathSensitive(join(cwd, 'keys', 'private.pem'))).toBe(
        true,
      );
    });

    it('settings.json 为敏感', () => {
      expect(store.isPathSensitive(join(cwd, 'settings.json'))).toBe(true);
    });

    it('普通源文件不敏感', () => {
      expect(store.isPathSensitive(join(cwd, 'src', 'index.ts'))).toBe(false);
    });
  });

  // ── 信任路径持久化 ──

  describe('addTrustedPath / removeTrustedPath', () => {
    it('写入文件后重新加载仍可读', () => {
      const testPath = join(cwd, 'src');
      store.addTrustedPath(testPath);

      const store2 = new ConfigStore(globalDir, cwd);
      expect(store2.isPathTrusted(testPath)).toBe(true);
    });

    it('removeTrustedPath 移除信任', () => {
      const testPath = join(cwd, 'src');
      store.addTrustedPath(testPath);
      const result = store.removeTrustedPath(testPath);
      expect(result.found).toBe(true);
      expect(store.isPathTrusted(testPath)).toBe(false);
    });

    it('移除不存在的路径返回 found=false', () => {
      const result = store.removeTrustedPath('/nonexistent');
      expect(result.found).toBe(false);
    });

    it('写入 project 作用域', () => {
      const testPath = join(cwd, 'src');
      const { scope } = store.addTrustedPath(testPath);
      expect(scope).toBe('project');

      const configFile = join(cwd, '.xkit', 'config.json');
      const content = JSON.parse(readFileSync(configFile, 'utf-8'));
      expect(content['safe-run'].trustedPaths).toContain(testPath);
    });

    it('无 .pi/ 时写入 global 作用域', () => {
      const store2 = new ConfigStore(globalDir, join(tmp, 'other'));
      const testPath = join(tmp, 'other', 'src');
      const { scope } = store2.addTrustedPath(testPath);
      expect(scope).toBe('global');
      expect(store2.isPathTrusted(testPath)).toBe(true);
    });
  });

  // ── 危险命令 ──

  describe('isCommandDangerous', () => {
    it('rm -rf 被检测', () => {
      expect(store.isCommandDangerous('rm -rf /tmp/cache')).toBe(true);
    });

    it('sudo 被检测', () => {
      expect(store.isCommandDangerous('sudo apt install')).toBe(true);
    });

    it('chmod 777 被检测', () => {
      expect(store.isCommandDangerous('chmod 777 /etc/passwd')).toBe(true);
    });

    it('base64|bash 被检测', () => {
      expect(
        store.isCommandDangerous('echo "base64string" | base64 -d | bash'),
      ).toBe(true);
    });

    it('git push --force 被检测', () => {
      expect(store.isCommandDangerous('git push --force origin main')).toBe(
        true,
      );
    });

    it('ls 安全', () => {
      expect(store.isCommandDangerous('ls -la')).toBe(false);
    });

    it('cat 安全', () => {
      expect(store.isCommandDangerous('cat file.txt')).toBe(false);
    });

    it('grep 安全', () => {
      expect(store.isCommandDangerous("grep -r 'foo' src/")).toBe(false);
    });

    // ── 组合命令 ──

    it('&& 连接的危险命令被检测', () => {
      expect(store.isCommandDangerous('cd /tmp && rm -rf data')).toBe(true);
    });

    it('|| 连接的危险命令被检测', () => {
      expect(store.isCommandDangerous('cd /tmp || sudo rm -rf data')).toBe(
        true,
      );
    });

    it('; 连接的危险命令被检测', () => {
      expect(store.isCommandDangerous('ls /nonexistent; sudo ls /root')).toBe(
        true,
      );
    });

    it('| 管道前的危险命令被检测', () => {
      expect(store.isCommandDangerous('rm -rf /tmp/cache | echo done')).toBe(
        true,
      );
    });

    it('& 后台命令中的危险命令被检测', () => {
      expect(store.isCommandDangerous('sleep 10 & rm -rf /tmp/cache')).toBe(
        true,
      );
    });

    it('多段组合中仅安全段不触发', () => {
      expect(
        store.isCommandDangerous('ls && cat file.txt && echo hello | grep foo'),
      ).toBe(false);
    });

    it('引号内的危险字符串不误报', () => {
      expect(store.isCommandDangerous('echo "rm -rf is dangerous"')).toBe(
        false,
      );
    });
  });

  // ── 命令白/黑名单 ──

  describe('命令白名单 / 黑名单', () => {
    it('addAllowedCommand 后 isCommandAlwaysAllowed 返回 true', () => {
      const cmd = 'rm -rf /tmp/build';
      store.addAllowedCommand(cmd);
      expect(store.isCommandAlwaysAllowed(cmd)).toBe(true);
    });

    it('未添加的命令不在白名单', () => {
      expect(store.isCommandAlwaysAllowed('sudo something')).toBe(false);
    });

    it('addDeniedCommand 后 isCommandAlwaysDenied 返回 true', () => {
      const cmd = 'sudo rm -rf /';
      store.addDeniedCommand(cmd);
      expect(store.isCommandAlwaysDenied(cmd)).toBe(true);
    });

    it('白名单/黑名单写入文件后持久化', () => {
      store.addAllowedCommand('docker system prune -f');
      store.addDeniedCommand('dd if=/dev/zero of=/dev/sda');

      const store2 = new ConfigStore(globalDir, cwd);
      expect(store2.isCommandAlwaysAllowed('docker system prune -f')).toBe(
        true,
      );
      expect(store2.isCommandAlwaysDenied('dd if=/dev/zero of=/dev/sda')).toBe(
        true,
      );
    });

    // ── 组合命令 ──

    it('白名单命令在 && 组合中仍被识别', () => {
      store.addAllowedCommand('rm -rf /tmp/cache');
      expect(store.isCommandAlwaysAllowed('cd /tmp && rm -rf /tmp/cache')).toBe(
        true,
      );
    });

    it('白名单命令在 ; 组合中仍被识别', () => {
      store.addAllowedCommand('sudo apt update');
      expect(
        store.isCommandAlwaysAllowed('sudo apt update; sudo apt upgrade -y'),
      ).toBe(true);
    });

    it('白名单精确匹配不含类似命令', () => {
      store.addAllowedCommand('rm -rf /tmp/cache-123');
      expect(store.isCommandAlwaysAllowed('rm -rf /tmp/cache-456')).toBe(false);
    });

    it('黑名单命令在 && 组合中仍被识别', () => {
      store.addDeniedCommand('dd if=/dev/zero of=/dev/sda');
      expect(
        store.isCommandAlwaysDenied('echo warn && dd if=/dev/zero of=/dev/sda'),
      ).toBe(true);
    });

    it('黑名单在 || 组合中仍被识别', () => {
      store.addDeniedCommand('rm -rf /important');
      expect(
        store.isCommandAlwaysDenied('cd /important || rm -rf /important'),
      ).toBe(true);
    });
  });

  // ── 自定义模式 ──

  describe('addDangerousPattern / addSensitivePattern', () => {
    it('添加自定义危险模式后生效', () => {
      store.addDangerousPattern('\\bterraform\\s+destroy\\b');
      expect(store.isCommandDangerous('terraform destroy -auto-approve')).toBe(
        true,
      );
    });

    it('添加自定义敏感路径模式后生效', () => {
      store.addSensitivePattern('secrets/');
      expect(store.isPathSensitive(join(cwd, 'secrets', 'key.txt'))).toBe(true);
    });
  });

  // ── 列表查询 ──

  describe('getTrustedPaths', () => {
    it('返回默认 + global + project 三组', () => {
      store.addTrustedPath('/custom/path');
      const result = store.getTrustedPaths();
      // 默认：globalDir 的父目录 + cwd 下 .pi/
      expect(result.default.length).toBeGreaterThan(0);
      // project：刚添加的路径
      expect(result.project.length).toBeGreaterThan(0);
    });
  });

  describe('getDangerousPatterns', () => {
    it('包含默认模式', () => {
      const patterns = store.getDangerousPatterns();
      expect(patterns.length).toBeGreaterThan(10);
      expect(patterns.some((p) => p.includes('rm'))).toBe(true);
    });

    it('包含自定义模式', () => {
      store.addDangerousPattern('custom-pattern');
      const patterns = store.getDangerousPatterns();
      expect(patterns).toContain('custom-pattern');
    });
  });

  describe('getSensitivePatterns', () => {
    it('包含自定义模式', () => {
      store.addSensitivePattern('custom-sensitive');
      const patterns = store.getSensitivePatterns();
      expect(patterns).toContain('custom-sensitive');
    });
  });

  // ── refresh ──

  describe('refresh', () => {
    it('重新加载外部修改', () => {
      const testPath = join(cwd, 'trusted-by-other');
      // 外部直接写入配置
      const configFile = join(cwd, '.pi', 'safe-run.json');
      writeFileSync(
        configFile,
        JSON.stringify({
          trustedPaths: [testPath],
          alwaysAllowCommands: [],
          alwaysDenyCommands: [],
          dangerousPatterns: [],
          sensitivePatterns: [],
        }),
      );
      store.refresh();
      expect(store.isPathTrusted(testPath)).toBe(true);
    });
  });

  // ── 边界情况 ──

  describe('边界情况', () => {
    it('空命令字符串不匹配危险模式', () => {
      expect(store.isCommandDangerous('')).toBe(false);
    });

    it('仅空格的命令不匹配危险模式', () => {
      expect(store.isCommandDangerous('   ')).toBe(false);
    });

    it('不存在的配置文件不会报错', () => {
      // 确保目录存在但无文件
      const cleanDir = join(tmp, 'clean-project');
      mkdirSync(join(cleanDir, '.pi'), { recursive: true });
      const store3 = new ConfigStore(globalDir, cleanDir);
      expect(store3.isPathTrusted('/foo')).toBe(false);
    });
  });
});
