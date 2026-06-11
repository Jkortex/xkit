import { beforeEach, describe, expect, it, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useApiKeyStore } from '@/infra/stores/useApiKeyStore';
import { httpClient } from '@/infra/http/FetchClient';
import { success, failure, AppError } from '@/utils/result';

vi.mock('@/infra/http/FetchClient', () => ({
  httpClient: {
    request: vi.fn(),
  },
}));

describe('ApiKeyStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();
  });

  const mockKey = {
    id: 'key-1',
    label: 'test-key',
    key: 'sk-123',
    expires_at: '2026-12-31T23:59:59Z',
    last_used_at: null,
    created_at: '2026-05-01T12:00:00Z',
  };

  it('fetchKeys should update keys on success', async () => {
    const store = useApiKeyStore();
    vi.mocked(httpClient.request).mockResolvedValue(success([mockKey]));

    await store.fetchKeys();

    expect(store.keys).toHaveLength(1);
    expect(store.keys[0].id).toBe('key-1');
    expect(store.keys[0].label).toBe('test-key');
    expect(store.loading).toBe(false);
  });

  it('createKey should prepend new key to list', async () => {
    const store = useApiKeyStore();
    vi.mocked(httpClient.request).mockResolvedValue(success(mockKey));

    const result = await store.createKey('test-key', 24);

    expect(result.error).toBeNull();
    expect(result.key?.key).toBe('sk-123');
    expect(store.keys).toHaveLength(1);
    expect(store.keys[0].id).toBe('key-1');
  });

  it('createKeyDirect should work and add key to list', async () => {
    const store = useApiKeyStore();
    vi.mocked(httpClient.request).mockResolvedValue(success(mockKey));

    const result = await store.createKeyDirect({
      username: 'user',
      password: 'pwd',
      label: 'direct-key',
    });

    expect(result.error).toBeNull();
    expect(result.key?.id).toBe('key-1');
    expect(store.keys).toHaveLength(1);
    expect(httpClient.request).toHaveBeenCalledWith(
      '/auth/api-keys/direct',
      expect.objectContaining({
        method: 'POST',
      }),
    );
  });

  it('deleteKey should remove key from list', async () => {
    const store = useApiKeyStore();
    store.keys = [
      {
        id: 'key-1',
        label: 'L',
        expiresAt: null,
        lastUsedAt: null,
        createdAt: '',
      },
    ];
    vi.mocked(httpClient.request).mockResolvedValue(success(null));

    const error = await store.deleteKey('key-1');

    expect(error).toBeNull();
    expect(store.keys).toHaveLength(0);
  });

  it('deleteKey should return error message on failure', async () => {
    const store = useApiKeyStore();
    vi.mocked(httpClient.request).mockResolvedValue(
      failure(new AppError('SERVER_ERROR', 'delete failed')),
    );

    const error = await store.deleteKey('key-1');

    expect(error).toBe('delete failed');
  });

  it('revokeCurrent should call revoke endpoint', async () => {
    const store = useApiKeyStore();
    vi.mocked(httpClient.request).mockResolvedValue(success(null));

    const error = await store.revokeCurrent();

    expect(error).toBeNull();
    expect(httpClient.request).toHaveBeenCalledWith(
      '/auth/api-keys/revoke',
      expect.objectContaining({
        method: 'POST',
      }),
    );
  });

  it('clear should reset keys list', async () => {
    const store = useApiKeyStore();
    store.keys = [
      {
        id: 'key-1',
        label: 'L',
        expiresAt: null,
        lastUsedAt: null,
        createdAt: '',
      },
    ];

    store.clear();

    expect(store.keys).toHaveLength(0);
  });
});
