import { beforeEach, describe, expect, it, vi } from 'vitest';
import { AppError, failure, success } from '@/utils/result';

const { mockGateway } = vi.hoisted(() => ({
  mockGateway: {
    upload: vi.fn(),
    importData: vi.fn(),
    exportData: vi.fn(),
  },
}));

vi.mock('@/infra/gateway/HttpResourceGateway', () => ({
  resourceGateway: mockGateway,
}));

vi.mock('@/presentation/presenters/ErrorPresenter', () => ({
  ErrorPresenter: {
    toMessage: vi.fn((err) => `Error: ${err.message}`),
  },
}));

import { useResource } from '@/presentation/composables/useResource';

describe('useResource', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uploads a file successfully', async () => {
    const dto = {
      id: '1',
      filename: 'test.png',
      mimeType: 'image/png',
      size: 100,
      createdAt: new Date(),
    };
    mockGateway.upload.mockResolvedValue(success(dto));
    const resource = useResource();

    const file = new File([''], 'test.png');
    const result = await resource.uploadFile(file);

    expect(result).toEqual(dto);
    expect(resource.error.value).toBeNull();
  });

  it('handles upload failure', async () => {
    mockGateway.upload.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'upload failed')),
    );
    const resource = useResource();

    const file = new File([''], 'test.png');
    const result = await resource.uploadFile(file);

    expect(result).toBeNull();
    expect(resource.error.value).toBe('Error: upload failed');
  });

  it('imports backup successfully', async () => {
    const report = {
      message: 'OK',
      memosImported: 5,
      resourcesImported: 3,
      memosSkipped: 0,
      resourcesSkipped: 0,
      report: {
        memos: { imported: 5, skipped: 0, details: [] },
        resources: { imported: 3, skipped: 0, details: [] },
      },
    };
    mockGateway.importData.mockResolvedValue(success(report));
    const resource = useResource();

    const file = new File([''], 'backup.zip');
    const result = await resource.importBackup(file);

    expect(result).toEqual({
      message: 'OK',
      memosImported: 5,
      resourcesImported: 3,
      memosSkipped: 0,
      resourcesSkipped: 0,
      report: {
        memos: { imported: 5, skipped: 0, details: [] },
        resources: { imported: 3, skipped: 0, details: [] },
      },
    });
  });

  it('handles import failure', async () => {
    mockGateway.importData.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'import failed')),
    );
    const resource = useResource();

    const file = new File([''], 'backup.zip');
    const result = await resource.importBackup(file);

    expect(result).toBeNull();
    expect(resource.error.value).toBe('Error: import failed');
  });

  it('exports backup successfully', async () => {
    const blob = new Blob(['data']);
    mockGateway.exportData.mockResolvedValue(success(blob));
    const resource = useResource();

    const result = await resource.exportBackup();

    expect(result).toBe(blob);
  });

  it('handles export failure', async () => {
    mockGateway.exportData.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'export failed')),
    );
    const resource = useResource();

    const result = await resource.exportBackup();

    expect(result).toBeNull();
    expect(resource.error.value).toBe('Error: export failed');
  });
});
