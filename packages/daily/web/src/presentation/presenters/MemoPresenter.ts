import type { MemoDTO } from '@/application/ports/dto/Memo';
import type { MemoVM } from '../view-models/MemoVM';
import { ResourcePresenter } from '@/presentation/presenters/ResourcePresenter';

const formatRelativeTime = (date: Date): string => {
  const now = new Date();
  const diffInSeconds = Math.floor((now.getTime() - date.getTime()) / 1000);

  if (diffInSeconds < 60) return '刚刚';
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}分钟前`;
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}小时前`;
  return date.toLocaleDateString();
};

/**
 * 笔记表现适配器
 * 职责：将稳定 DTO 映射为 UI 专用的 ViewModel
 */
export const MemoPresenter = {
  toViewModel(dto: MemoDTO): MemoVM {
    return {
      uuid: dto.uuid,
      content: dto.content,
      tags: dto.tags,
      resources: dto.resources.map(ResourcePresenter.toViewModel),
      expiresAt: dto.expiresAt?.toLocaleString(),
      headline: dto.headline,
      relativeTime: formatRelativeTime(dto.createdAt),
      displayDate: dto.createdAt.toLocaleDateString(),
    };
  },
};
