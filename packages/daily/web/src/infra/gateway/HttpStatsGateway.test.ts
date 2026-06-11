import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AppError, failure, success } from '@/utils/result';

const { requestMock } = vi.hoisted(() => ({
  requestMock: vi.fn(),
}));

vi.mock('../http/FetchClient', () => ({
  httpClient: {
    request: requestMock,
  },
}));

import { HttpStatsGateway } from './HttpStatsGateway';

describe('HttpStatsGateway', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('maps backend payload to frontend dto', async () => {
    requestMock.mockResolvedValue(
      success({
        memos_total: 10,
        tags_total: 5,
        resources_total: 2,
        heatmap: [{ date: '2026-03-06', count: 3 }],
      }),
    );

    const gateway = new HttpStatsGateway();
    const result = await gateway.getStats();

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value).toEqual({
        memosTotal: 10,
        tagsTotal: 5,
        resourcesTotal: 2,
        heatmap: [{ date: '2026-03-06', count: 3 }],
      });
    }
    expect(requestMock).toHaveBeenCalledWith('/stats');
  });

  it('passes through request failures', async () => {
    const err = new AppError('NETWORK_ERROR', 'network');
    requestMock.mockResolvedValue(failure(err));

    const gateway = new HttpStatsGateway();
    const result = await gateway.getStats();

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error).toBe(err);
    }
  });

  it('returns parse error when payload is invalid', async () => {
    requestMock.mockResolvedValue(
      success({
        memos_total: 1,
        tags_total: 1,
        resources_total: 1,
        heatmap: [{ date: 'invalid-date', count: 3 }],
      }),
    );

    const gateway = new HttpStatsGateway();
    const result = await gateway.getStats();

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error).toBeInstanceOf(AppError);
      expect(result.error.code).toBe('SERVER_ERROR');
      expect(result.error.message).toBe('统计数据解析失败，格式不合法');
    }
  });
});
