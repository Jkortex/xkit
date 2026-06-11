import type { MemoHistoryDTO } from '@/infra/gateway/HttpMemoGateway';
import type { MemoHistoryVM } from '@/presentation/view-models/MemoHistoryVM';

export const MemoHistoryPresenter = {
  toViewModel(dto: MemoHistoryDTO): MemoHistoryVM {
    const date = new Date(dto.created_at);
    return {
      id: dto.id,
      content: dto.content,
      tags: dto.tags,
      resourceIds: dto.resource_ids,
      relativeTime: this.formatRelativeTime(date),
      absoluteTime: date.toLocaleString('zh-CN', { hour12: false }),
    };
  },

  toViewModelList(dtos: MemoHistoryDTO[]): MemoHistoryVM[] {
    return dtos.map((dto) => this.toViewModel(dto));
  },

  formatRelativeTime(date: Date): string {
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const minute = 60 * 1000;
    const hour = 60 * minute;
    const day = 24 * hour;

    if (diffMs < minute) return '刚刚';
    if (diffMs < hour) return `${Math.floor(diffMs / minute)}分钟前`;
    if (diffMs < day) return `${Math.floor(diffMs / hour)}小时前`;
    if (diffMs < 7 * day) return `${Math.floor(diffMs / day)}天前`;
    return date.toLocaleDateString('zh-CN');
  },
};
