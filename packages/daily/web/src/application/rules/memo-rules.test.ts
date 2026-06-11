import { describe, it, expect } from 'vitest';
import { validateMemoContent, validateMemoDates } from './memo-rules';

describe('Memo Rules', () => {
  it('should invalidate empty content', () => {
    expect(validateMemoContent('')).toBe('笔记内容不能为空');
    expect(validateMemoContent('   ')).toBe('笔记内容不能为空');
  });

  it('should validate non-empty content', () => {
    expect(validateMemoContent('hello')).toBeNull();
  });

  it('should invalidate update time before create time', () => {
    const created = new Date('2026-01-02');
    const updated = new Date('2026-01-01');
    expect(validateMemoDates(created, updated)).toBe(
      '更新时间不能早于创建时间',
    );
  });

  it('should validate correct date sequence', () => {
    const created = new Date('2026-01-01');
    const updated = new Date('2026-01-02');
    expect(validateMemoDates(created, updated)).toBeNull();
  });
});
