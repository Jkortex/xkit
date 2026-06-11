import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

export interface MemoVM {
  readonly uuid: string;
  readonly content: string;
  readonly tags: string[];
  readonly resources: ResourceVM[];
  readonly expiresAt?: string;
  readonly headline?: string;
  readonly relativeTime: string; // 比如 "3分钟前"
  readonly displayDate: string; // 比如 "2026-03-01"
}
