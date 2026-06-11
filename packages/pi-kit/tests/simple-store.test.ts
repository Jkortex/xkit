import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { mkdtempSync, rmSync } from 'node:fs';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { SimpleStore } from '../extensions/simple-store';

describe('SimpleStore', () => {
  let tmp: string;

  beforeEach(() => {
    tmp = mkdtempSync(join(tmpdir(), 'simple-store-test-'));
    vi.stubEnv('HOME', tmp);
  });

  afterEach(() => {
    vi.unstubAllEnvs();
    rmSync(tmp, { recursive: true, force: true });
  });

  describe('project trust', () => {
    it('new project is not trusted by default', () => {
      const store = new SimpleStore();
      expect(store.isProjectTrusted('/some/path')).toBe(false);
    });

    it('trusted project is recognized', () => {
      const store = new SimpleStore();
      store.trustProject('/my/project');
      expect(store.isProjectTrusted('/my/project')).toBe(true);
    });

    it('persists across instances', () => {
      const store = new SimpleStore();
      store.trustProject('/my/project');

      const store2 = new SimpleStore();
      expect(store2.isProjectTrusted('/my/project')).toBe(true);
    });

    it('untrust removes trust', () => {
      const store = new SimpleStore();
      store.trustProject('/my/project');
      store.untrustProject('/my/project');
      expect(store.isProjectTrusted('/my/project')).toBe(false);
    });
  });

  describe('path boundary', () => {
    it('path inside project is within project', () => {
      expect(SimpleStore.isPathWithinProject('/project/src/index.ts', '/project')).toBe(true);
    });

    it('path outside project is not within project', () => {
      expect(SimpleStore.isPathWithinProject('/other/file.txt', '/project')).toBe(false);
    });

    it('path equal to cwd is within project', () => {
      expect(SimpleStore.isPathWithinProject('/project', '/project')).toBe(true);
    });

    it('path escaping project root via ../ is detected', () => {
      const { resolve } = require('node:path');
      const resolved = resolve('/project/src', '../../etc');
      expect(SimpleStore.isPathWithinProject(resolved, '/project')).toBe(false);
    });

    it('path within project via ../ is still inside', () => {
      const { resolve } = require('node:path');
      const resolved = resolve('/project/src', '../outside');
      expect(SimpleStore.isPathWithinProject(resolved, '/project')).toBe(true);
    });

    it('cwd with trailing slash still works', () => {
      expect(SimpleStore.isPathWithinProject('/project/file.ts', '/project/')).toBe(true);
    });
  });

  describe('system dangerous commands', () => {
    it('sudo is dangerous', () => {
      expect(SimpleStore.isSystemDangerous('sudo apt install')).toBe(true);
    });

    it('shutdown is dangerous', () => {
      expect(SimpleStore.isSystemDangerous('shutdown -h now')).toBe(true);
    });

    it('ls is not dangerous', () => {
      expect(SimpleStore.isSystemDangerous('ls -la')).toBe(false);
    });

    it('npm install is not dangerous', () => {
      expect(SimpleStore.isSystemDangerous('npm install')).toBe(false);
    });

    it('rm -rf is not system dangerous', () => {
      expect(SimpleStore.isSystemDangerous('rm -rf node_modules')).toBe(false);
    });
  });

  describe('rm -rf outside project detection', () => {
    it('rm -rf with absolute path is outside', () => {
      expect(SimpleStore.isRmdirOutside('rm -rf /tmp/cache')).toBe(true);
    });

    it('rm -rf ~ is outside', () => {
      expect(SimpleStore.isRmdirOutside('rm -rf ~/.config')).toBe(true);
    });

    it('rm -rf .. is outside', () => {
      expect(SimpleStore.isRmdirOutside('rm -rf ../other-project')).toBe(true);
    });

    it('rm -rf with relative path is allowed', () => {
      expect(SimpleStore.isRmdirOutside('rm -rf node_modules')).toBe(false);
    });

    it('rm without -rf is not checked', () => {
      expect(SimpleStore.isRmdirOutside('rm file.txt')).toBe(false);
    });

    it('rm -rf . is allowed', () => {
      expect(SimpleStore.isRmdirOutside('rm -rf .')).toBe(false);
    });
  });

  describe('allow/deny lists', () => {
    it('new command is not in allow list', () => {
      const store = new SimpleStore();
      expect(store.isCommandAlwaysAllowed('some-cmd')).toBe(false);
    });

    it('added allowed command is recognized', () => {
      const store = new SimpleStore();
      store.addAllowedCommand('my-cmd');
      expect(store.isCommandAlwaysAllowed('my-cmd')).toBe(true);
    });

    it('added denied command is recognized', () => {
      const store = new SimpleStore();
      store.addDeniedCommand('bad-cmd');
      expect(store.isCommandAlwaysDenied('bad-cmd')).toBe(true);
    });

    it('allow list persists across instances', () => {
      const store = new SimpleStore();
      store.addAllowedCommand('persistent-cmd');

      const store2 = new SimpleStore();
      expect(store2.isCommandAlwaysAllowed('persistent-cmd')).toBe(true);
    });
  });
});
