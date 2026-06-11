export interface ApiKeyDTO {
  id: string;
  label: string;
  key?: string;
  expiresAt: string | null;
  lastUsedAt: string | null;
  createdAt: string;
}
