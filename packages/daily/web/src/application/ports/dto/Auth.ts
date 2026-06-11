export interface AuthUserDTO {
  id: number;
  username: string;
  role: 'admin' | 'member';
  status: 'active' | 'disabled';
}

export interface InviteDTO {
  id: string;
  code?: string;
  role: 'admin' | 'member';
  status: 'active' | 'used' | 'revoked' | 'expired';
  expiresAt: string;
  createdAt?: string;
  usedAt?: string;
  revokedAt?: string;
}

export interface InviteSummaryDTO {
  activeMemberCount: number;
  activeAdminCount: number;
  memberLimit: number;
  adminLimit: number;
}

export interface VerifyInviteDTO {
  valid: boolean;
  role: string;
  expiresAt: string;
}
