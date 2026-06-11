import type { MemoListQuery } from '@/application/ports/MemoListQuery';
import type { MemoFilterVM } from '@/presentation/view-models/MemoFilterVM';

export const toMemoListQuery = (
  filter: MemoFilterVM,
  page?: { limit?: number; offset?: number },
): MemoListQuery => {
  return {
    search: filter.search,
    tag: filter.tag,
    from: filter.from,
    to: filter.to,
    hasResource: filter.hasResource,
    tagsAny: filter.tagsAny,
    tagsAll: filter.tagsAll,
    sort: filter.sort,
    includeResources: filter.includeResources,
    limit: page?.limit,
    offset: page?.offset,
  };
};
