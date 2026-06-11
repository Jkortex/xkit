import { httpClient } from '@/infra/http/FetchClient';
import { success, type Result } from '@/utils/result';
import type {
  AuthUserDTO,
  InviteDTO,
  InviteSummaryDTO,
  VerifyInviteDTO,
} from '@/application/ports/dto/Auth';
export interface LoginPayload {
  username: string;
  password: string;
}

export interface RegisterByInvitePayload extends LoginPayload {
  code: string;
}

interface BackendInviteResponse {
  id: string;
  code?: string;
  role: 'admin' | 'member';
  status: 'active' | 'used' | 'revoked' | 'expired';
  expires_at: string;
  created_at?: string;
  used_at?: string;
  revoked_at?: string;
}

interface BackendInviteSummaryResponse {
  active_member_count: number;
  active_admin_count: number;
  member_limit: number;
  admin_limit: number;
}

interface BackendVerifyInviteResponse {
  valid: boolean;
  role: string;
  expires_at: string;
}

function transformInvite(raw: BackendInviteResponse): InviteDTO {
  return {
    id: raw.id,
    code: raw.code,
    role: raw.role,
    status: raw.status,
    expiresAt: raw.expires_at,
    createdAt: raw.created_at,
    usedAt: raw.used_at,
    revokedAt: raw.revoked_at,
  };
}

export class HttpAuthGateway {
  async me(): Promise<Result<AuthUserDTO>> {
    return httpClient.request<AuthUserDTO>('/auth/me');
  }

  async login(payload: LoginPayload): Promise<Result<AuthUserDTO>> {
    return httpClient.request<AuthUserDTO>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  }

  async logout(): Promise<Result<void>> {
    return httpClient.request<void>('/auth/logout', {
      method: 'POST',
    });
  }

  async createInvite(
    role: 'admin' | 'member',
    ttlHours: number,
  ): Promise<Result<InviteDTO>> {
    const res = await httpClient.request<BackendInviteResponse>(
      '/auth/invites',
      {
        method: 'POST',
        body: JSON.stringify({ role, ttl_hours: ttlHours }),
      },
    );
    if (res.kind === 'failure') return res;
    return success(transformInvite(res.value));
  }

  async listInvites(
    status?: string,
    limit = 100,
  ): Promise<Result<InviteDTO[]>> {
    const params = new URLSearchParams({ limit: String(limit) });
    if (status) params.set('status', status);
    const res = await httpClient.request<BackendInviteResponse[]>(
      `/auth/invites?${params}`,
    );
    if (res.kind === 'failure') return res;
    return success(res.value.map(transformInvite));
  }

  async revokeInvite(id: string): Promise<Result<void>> {
    return httpClient.request<void>(`/auth/invites/${id}/revoke`, {
      method: 'POST',
    });
  }

  async getInviteSummary(): Promise<Result<InviteSummaryDTO>> {
    const res = await httpClient.request<BackendInviteSummaryResponse>(
      '/auth/invites/summary',
    );
    if (res.kind === 'failure') return res;
    return success({
      activeMemberCount: res.value.active_member_count,
      activeAdminCount: res.value.active_admin_count,
      memberLimit: res.value.member_limit,
      adminLimit: res.value.admin_limit,
    });
  }

  async verifyInvite(code: string): Promise<Result<VerifyInviteDTO>> {
    const res = await httpClient.request<BackendVerifyInviteResponse>(
      `/auth/invites/${encodeURIComponent(code)}/verify`,
    );
    if (res.kind === 'failure') return res;
    return success({
      valid: res.value.valid,
      role: res.value.role,
      expiresAt: res.value.expires_at,
    });
  }

  async registerByInvite(
    payload: RegisterByInvitePayload,
  ): Promise<Result<AuthUserDTO>> {
    return httpClient.request<AuthUserDTO>('/auth/register-by-invite', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  }
}

export const authGateway = new HttpAuthGateway();
