import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  copyAuditSummary,
  formatAbsoluteAuditTime,
  formatRelativeAuditTime,
} from '@/presentation/composables/tag-governance/tagGovernanceAuditUtils';
import { MessagePlugin } from 'tdesign-vue-next';

vi.mock('tdesign-vue-next', () => ({
  MessagePlugin: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

describe('tagGovernanceAuditUtils', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-08T08:00:00.000Z'));
  });

  it('formats relative time for recent minutes', () => {
    const value = formatRelativeAuditTime('2026-03-08T07:55:00.000Z');

    expect(value).toBe('5分钟前');
  });

  it('returns raw text when absolute time is invalid', () => {
    const value = formatAbsoluteAuditTime('invalid-date');

    expect(value).toBe('invalid-date');
  });

  it('copies summary and shows success message', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined);
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      configurable: true,
    });

    await copyAuditSummary('hello');

    expect(writeText).toHaveBeenCalledWith('hello');
    expect(MessagePlugin.success).toHaveBeenCalledWith('已复制');
  });

  it('shows error message when copy fails', async () => {
    const writeText = vi.fn().mockRejectedValue(new Error('copy failed'));
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText },
      configurable: true,
    });

    await copyAuditSummary('hello');

    expect(MessagePlugin.error).toHaveBeenCalledWith('复制失败');
  });
});
