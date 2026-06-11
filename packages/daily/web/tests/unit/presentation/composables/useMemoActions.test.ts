import { beforeEach, describe, expect, it, vi } from 'vitest';
import { failure, success, AppError } from '@/utils/result';
import { setActivePinia, createPinia } from 'pinia';

const { mockMemoGateway } = vi.hoisted(() => ({
  mockMemoGateway: {
    createMemo: vi.fn(),
    updateMemo: vi.fn(),
    deleteMemo: vi.fn(),
    getMemos: vi.fn(),
    getRandomMemo: vi.fn(),
    listMemoHistory: vi.fn(),
    rollbackMemo: vi.fn(),
    getTags: vi.fn(),
    renameTag: vi.fn(),
    mergeTags: vi.fn(),
    upsertTagAlias: vi.fn(),
    listTagAliases: vi.fn(),
    deleteTagAlias: vi.fn(),
    listTagAudits: vi.fn(),
  },
}));

vi.mock('@/infra/gateway/HttpMemoGateway', () => ({
  memoGateway: mockMemoGateway,
}));

vi.mock('@/presentation/presenters/ErrorPresenter', () => ({
  ErrorPresenter: {
    toMessage: vi.fn((err) => `Error: ${err.message}`),
  },
}));

import { useMemoActions } from '@/presentation/composables/useMemoActions';

describe('useMemoActions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    setActivePinia(createPinia());
  });

  it('creates memo successfully', async () => {
    mockMemoGateway.createMemo.mockResolvedValue(success({ id: 1 }));
    const actions = useMemoActions();

    const resultPromise = actions.createMemo('hello', ['r1']);
    expect(actions.loading.value).toBe(true);

    const result = await resultPromise;
    expect(result).toBe(true);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBeNull();
    expect(mockMemoGateway.createMemo).toHaveBeenCalledWith(
      'hello',
      ['r1'],
      undefined,
    );
  });

  it('handles create memo failure', async () => {
    mockMemoGateway.createMemo.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'create failed')),
    );
    const actions = useMemoActions();

    const result = await actions.createMemo('hello');
    expect(result).toBe(false);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBe('Error: create failed');
  });

  it('updates memo successfully', async () => {
    mockMemoGateway.updateMemo.mockResolvedValue(success({ id: 1 }));
    const actions = useMemoActions();

    const resultPromise = actions.updateMemo('1', 'updated', ['r1']);
    expect(actions.loading.value).toBe(true);

    const result = await resultPromise;
    expect(result).toBe(true);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBeNull();
    expect(mockMemoGateway.updateMemo).toHaveBeenCalledWith(
      '1',
      'updated',
      ['r1'],
      undefined,
    );
  });

  it('handles update memo failure', async () => {
    mockMemoGateway.updateMemo.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'update failed')),
    );
    const actions = useMemoActions();

    const result = await actions.updateMemo('1', 'updated', []);
    expect(result).toBe(false);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBe('Error: update failed');
  });

  it('deletes memo successfully', async () => {
    mockMemoGateway.deleteMemo.mockResolvedValue(success(undefined));
    const actions = useMemoActions();

    const resultPromise = actions.deleteMemo('1');
    expect(actions.loading.value).toBe(true);

    const result = await resultPromise;
    expect(result).toBe(true);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBeNull();
    expect(mockMemoGateway.deleteMemo).toHaveBeenCalledWith('1');
  });

  it('handles delete memo failure', async () => {
    mockMemoGateway.deleteMemo.mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'delete failed')),
    );
    const actions = useMemoActions();

    const result = await actions.deleteMemo('1');
    expect(result).toBe(false);
    expect(actions.loading.value).toBe(false);
    expect(actions.error.value).toBe('Error: delete failed');
  });
});
