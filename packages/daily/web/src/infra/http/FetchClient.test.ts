import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { FetchClient } from './FetchClient';

describe('FetchClient', () => {
  const originalFetch = globalThis.fetch;
  const fetchMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    globalThis.fetch = fetchMock as typeof fetch;
    Object.defineProperty(globalThis, 'localStorage', {
      value: {
        getItem: vi.fn().mockReturnValue(null),
      },
      configurable: true,
    });
  });

  afterEach(() => {
    globalThis.fetch = originalFetch;
  });

  it('returns success result on json payload', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ id: 1 }), {
        status: 200,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    const client = new FetchClient('/api');

    const result = await client.request<{ id: number }>('/memos');

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value).toEqual({ id: 1 });
    }
  });

  it('returns null for 204 response', async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 204 }));
    const client = new FetchClient('/api');

    const result = await client.request<null>('/memos');

    expect(result.kind).toBe('success');
    if (result.kind === 'success') {
      expect(result.value).toBeNull();
    }
  });

  it('maps 401 to auth error', async () => {
    fetchMock.mockResolvedValue(new Response(null, { status: 401 }));
    const client = new FetchClient('/api');

    const result = await client.request('/memos');

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error.code).toBe('AUTH_ERROR');
    }
  });

  it('extracts backend error from json body', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ error: 'tag not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      }),
    );
    const client = new FetchClient('/api');

    const result = await client.request('/tags');

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error.code).toBe('SERVER_ERROR');
      expect(result.error.message).toBe('tag not found');
    }
  });

  it('maps backend INVALID_INPUT code to VALIDATION_ERROR', async () => {
    fetchMock.mockResolvedValue(
      new Response(
        JSON.stringify({ error: 'invalid date format', code: 'INVALID_INPUT' }),
        {
          status: 400,
          headers: { 'Content-Type': 'application/json' },
        },
      ),
    );
    const client = new FetchClient('/api');

    const result = await client.request('/tags');

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error.code).toBe('VALIDATION_ERROR');
      expect(result.error.serverCode).toBe('INVALID_INPUT');
      expect(result.error.message).toBe('invalid date format');
    }
  });

  it('falls back to status message for non-json error body', async () => {
    fetchMock.mockResolvedValue(
      new Response('bad gateway', {
        status: 502,
        headers: { 'Content-Type': 'text/plain' },
      }),
    );
    const client = new FetchClient('/api');

    const result = await client.request('/tags');

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error.message).toBe('服务器返回错误: 502');
    }
  });

  it('maps network failure to network error', async () => {
    fetchMock.mockRejectedValue(new Error('offline'));
    const client = new FetchClient('/api');

    const result = await client.request('/tags');

    expect(result.kind).toBe('failure');
    if (result.kind === 'failure') {
      expect(result.error.code).toBe('NETWORK_ERROR');
    }
  });

  it('does not send json content-type for FormData body', async () => {
    fetchMock.mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 }),
    );
    const client = new FetchClient('/api');
    const formData = new FormData();
    formData.append('file', new Blob(['x']), 'a.txt');

    await client.request('/upload', { method: 'POST', body: formData });

    const [, requestInit] = fetchMock.mock.calls[0] as [string, RequestInit];
    const headers = requestInit.headers as Record<string, string>;
    expect(headers['Content-Type']).toBeUndefined();
  });

  it('supports blob response and unwraps server error', async () => {
    fetchMock
      .mockResolvedValueOnce(new Response(new Blob(['zip']), { status: 200 }))
      .mockResolvedValueOnce(
        new Response(JSON.stringify({ error: 'export failed' }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' },
        }),
      );
    const client = new FetchClient('/api');

    const successResult = await client.requestBlob('/system/export');
    const failureResult = await client.requestBlob('/system/export');

    expect(successResult.kind).toBe('success');
    expect(failureResult.kind).toBe('failure');
    if (failureResult.kind === 'failure') {
      expect(failureResult.error.message).toBe('export failed');
    }
  });
});
