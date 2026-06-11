import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AppError, failure, success } from '@/utils/result';

const { requestMock, requestBlobMock } = vi.hoisted(() => ({
  requestMock: vi.fn(),
  requestBlobMock: vi.fn(),
}));

vi.mock('../http/FetchClient', () => ({
  httpClient: {
    request: requestMock,
    requestBlob: requestBlobMock,
  },
}));

import { HttpResourceGateway } from './HttpResourceGateway';

describe('HttpResourceGateway', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('maps upload response', async () => {
    requestMock.mockResolvedValue(
      success({
        id: 'res-1',
        filename: 'a.png',
        size: 10,
        mime_type: 'image/png',
        created_at: '2026-03-06T00:00:00Z',
      }),
    );
    const file = new File(['x'], 'a.png', { type: 'image/png' });
    const gateway = new HttpResourceGateway();

    const result = await gateway.upload(file);

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value.id).toBe('res-1');
      expect(result.value.filename).toBe('a.png');
      expect(result.value.mimeType).toBe('image/png');
    }
    expect(requestMock).toHaveBeenCalledWith('/resources', {
      method: 'POST',
      body: expect.any(FormData),
    });
  });

  it('passes through import failures', async () => {
    const err = new AppError('SERVER_ERROR', 'boom');
    requestMock.mockResolvedValue(failure(err));
    const gateway = new HttpResourceGateway();
    const file = new File(['x'], 'backup.zip', { type: 'application/zip' });

    const result = await gateway.importData(file);

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error).toBe(err);
    }
  });

  it('maps import report fields', async () => {
    requestMock.mockResolvedValue(
      success({
        message: 'ok',
        memos_imported: 2,
        resources_imported: 3,
        memos_skipped: 1,
        resources_skipped: 0,
        report: {
          memos: { imported: 2, skipped: 1, details: [] },
          resources: { imported: 3, skipped: 0, details: [] },
        },
      }),
    );
    const gateway = new HttpResourceGateway();
    const file = new File(['x'], 'backup.zip', { type: 'application/zip' });

    const result = await gateway.importData(file);

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value).toMatchObject({
        memosImported: 2,
        resourcesImported: 3,
        memosSkipped: 1,
        resourcesSkipped: 0,
      });
    }
  });

  it('delegates export to requestBlob', async () => {
    const blob = new Blob(['zip']);
    requestBlobMock.mockResolvedValue(success(blob));
    const gateway = new HttpResourceGateway();

    const result = await gateway.exportData();

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value).toBe(blob);
    }
    expect(requestBlobMock).toHaveBeenCalledWith('/system/export');
  });
});
