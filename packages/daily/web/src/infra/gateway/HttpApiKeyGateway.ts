import { httpClient } from '@/infra/http/FetchClient';
import { success, type Result } from '@/utils/result';
import type { ApiKeyDTO } from '@/application/ports/dto/ApiKey';
export interface CreateApiKeyPayload {
  label: string;
  ttlHours?: number;
}

export interface CreateApiKeyDirectPayload {
  username: string;
  password: string;
  label: string;
  ttlHours?: number;
}

interface BackendApiKeyResponse {
  id: string;
  label: string;
  key?: string;
  expires_at: string | null;
  last_used_at: string | null;
  created_at: string;
}

function transform(raw: BackendApiKeyResponse): ApiKeyDTO {
  return {
    id: raw.id,
    label: raw.label,
    key: raw.key,
    expiresAt: raw.expires_at,
    lastUsedAt: raw.last_used_at,
    createdAt: raw.created_at,
  };
}

export class HttpApiKeyGateway {
  async list(): Promise<Result<ApiKeyDTO[]>> {
    const res =
      await httpClient.request<BackendApiKeyResponse[]>('/auth/api-keys');
    if (res.kind === 'success') {
      return success(res.value.map(transform));
    }
    return res;
  }

  async create(payload: CreateApiKeyPayload): Promise<Result<ApiKeyDTO>> {
    const res = await httpClient.request<BackendApiKeyResponse>(
      '/auth/api-keys',
      {
        method: 'POST',
        body: JSON.stringify({
          label: payload.label,
          ttl_hours: payload.ttlHours,
        }),
      },
    );
    if (res.kind === 'success') {
      return success(transform(res.value));
    }
    return res;
  }

  async createDirect(
    payload: CreateApiKeyDirectPayload,
  ): Promise<Result<ApiKeyDTO>> {
    const res = await httpClient.request<BackendApiKeyResponse>(
      '/auth/api-keys/direct',
      {
        method: 'POST',
        body: JSON.stringify({
          username: payload.username,
          password: payload.password,
          label: payload.label,
          ttl_hours: payload.ttlHours,
        }),
      },
    );
    if (res.kind === 'success') {
      return success(transform(res.value));
    }
    return res;
  }

  async delete(id: string): Promise<Result<void>> {
    return httpClient.request<void>(`/auth/api-keys/${id}`, {
      method: 'DELETE',
    });
  }

  async revokeCurrent(): Promise<Result<void>> {
    return httpClient.request<void>('/auth/api-keys/revoke', {
      method: 'POST',
    });
  }
}

export const apiKeyGateway = new HttpApiKeyGateway();
