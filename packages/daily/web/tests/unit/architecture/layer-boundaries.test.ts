import { describe, expect, it } from 'vitest';
import { readdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';

const srcRoot = resolve(process.cwd(), 'src');

const collectFiles = (dir: string): string[] => {
  return readdirSync(dir, { withFileTypes: true }).flatMap((entry) => {
    const path = resolve(dir, entry.name);
    if (entry.isDirectory()) return collectFiles(path);
    if (!path.endsWith('.ts') && !path.endsWith('.vue')) return [];
    return [path];
  });
};

const readFiles = (dir: string): Array<{ path: string; content: string }> => {
  return collectFiles(dir).map((path) => ({
    path,
    content: readFileSync(path, 'utf8'),
  }));
};

interface BoundaryRule {
  name: string;
  scope: string;
  blockedImports: string[];
  allow?: string[];
}

const rules: BoundaryRule[] = [
  {
    name: 'presentation does not import infra HTTP layer directly',
    scope: 'presentation',
    blockedImports: ['@/infra/http'],
  },
  {
    name: 'application does not depend on presentation or infra implementations',
    scope: 'application',
    blockedImports: ['@/presentation/', '@/infra/'],
  },
];

const runRule = (rule: BoundaryRule): void => {
  const files = readFiles(resolve(srcRoot, rule.scope));

  files.forEach(({ path, content }) => {
    rule.blockedImports.forEach((blockedImport) => {
      if (
        rule.allow?.some((allowedImport) => content.includes(allowedImport))
      ) {
        return;
      }
      expect
        .soft(content, `${rule.name}: ${path}`)
        .not.toContain(blockedImport);
    });
  });
};

describe('layer boundaries', () => {
  rules.forEach((rule) => {
    it(rule.name, () => {
      runRule(rule);
    });
  });
});
