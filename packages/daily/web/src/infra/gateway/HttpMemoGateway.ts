export interface TagStatDTO {
  readonly name: string;
  readonly count: number;
}

export interface RenameTagDTO {
  readonly from: string;
  readonly to: string;
  readonly affected_memos: number;
  readonly merged: boolean;
}

export interface MergeTagsDTO {
  readonly sources: string[];
  readonly target: string;
  readonly affected_memos: number;
  readonly merged_sources: number;
  readonly skipped_sources: string[];
}

export interface TagAliasDTO {
  readonly alias: string;
  readonly canonical: string;
}

export interface TagAuditDTO {
  readonly action: string;
  readonly summary: string;
  readonly affected_memos: number;
  readonly created_at: string;
}

export interface MemoHistoryDTO {
  readonly id: string;
  readonly memo_uuid: string;
  readonly content: string;
  readonly tags: string[];
  readonly resource_ids: string[];
  readonly created_at: string;
}
import type { MemoListQuery } from '@/application/ports/MemoListQuery';
import type { MemoDTO } from '@/application/ports/dto/Memo';
import type { BackendMemoDTO } from './dto/BackendMemoDTO';
import { httpClient } from '../http/FetchClient';
import { transformMemo } from './transform/memo-transform';
import { Result, success, failure, AppError } from '@/utils/result';

export class HttpMemoGateway {
  async getMemos(params: MemoListQuery): Promise<Result<MemoDTO[]>> {
    const query = new URLSearchParams();
    if (params.search) query.append('search', params.search);
    if (params.tag) query.append('tag', params.tag);
    if (params.from) query.append('from', params.from);
    if (params.to) query.append('to', params.to);
    if (params.hasResource !== undefined) {
      query.append('has_resource', String(params.hasResource));
    }
    if (params.tagsAny && params.tagsAny.length > 0) {
      query.append('tags_any', params.tagsAny.join(','));
    }
    if (params.tagsAll && params.tagsAll.length > 0) {
      query.append('tags_all', params.tagsAll.join(','));
    }
    if (params.tagsExclude && params.tagsExclude.length > 0) {
      query.append('tags_exclude', params.tagsExclude.join(','));
    }
    if (params.sort) query.append('sort', params.sort);
    if (params.includeResources !== undefined) {
      query.append('include_resources', String(params.includeResources));
    }
    if (params.limit) query.append('limit', params.limit.toString());
    if (params.offset) query.append('offset', params.offset.toString());

    const result = await httpClient.request<BackendMemoDTO[]>(
      `/memos?${query.toString()}`,
    );
    if (result.kind === 'failure') return result;

    try {
      const dtos = result.value.map(transformMemo);
      return success(dtos);
    } catch (error) {
      return failure(
        new AppError('SERVER_ERROR', '解析后端数据时发生错误', error),
      );
    }
  }

  async createMemo(
    content: string,
    resourceIds: string[],
    ttl?: string,
  ): Promise<Result<MemoDTO>> {
    const result = await httpClient.request<BackendMemoDTO>('/memos', {
      method: 'POST',
      body: JSON.stringify({
        content,
        resource_ids: resourceIds,
        ttl,
      }),
    });

    if (result.kind === 'failure') return result;

    try {
      return success(transformMemo(result.value));
    } catch (error) {
      return failure(
        new AppError('SERVER_ERROR', '创建笔记成功，但解析响应失败', error),
      );
    }
  }

  async getRandomMemo(): Promise<Result<MemoDTO>> {
    const result = await httpClient.request<BackendMemoDTO>('/memos/random');
    if (result.kind === 'failure') return result;

    try {
      return success(transformMemo(result.value));
    } catch (error) {
      return failure(new AppError('SERVER_ERROR', '获取随机笔记失败', error));
    }
  }

  async updateMemo(
    uuid: string,
    content: string,
    resourceIds: string[],
    ttl?: string,
  ): Promise<Result<MemoDTO>> {
    const result = await httpClient.request<BackendMemoDTO>(`/memos/${uuid}`, {
      method: 'PATCH',
      body: JSON.stringify({
        content,
        resource_ids: resourceIds,
        ttl,
      }),
    });
    if (result.kind === 'failure') return result;
    try {
      return success(transformMemo(result.value));
    } catch (error) {
      return failure(new AppError('SERVER_ERROR', '更新成功但解析失败', error));
    }
  }

  async deleteMemo(uuid: string): Promise<Result<void>> {
    return await httpClient.request<void>(`/memos/${uuid}`, {
      method: 'DELETE',
    });
  }

  async getTags(): Promise<Result<TagStatDTO[]>> {
    return await httpClient.request<TagStatDTO[]>('/tags');
  }

  async renameTag(from: string, to: string): Promise<Result<RenameTagDTO>> {
    return await httpClient.request<RenameTagDTO>('/tags/rename', {
      method: 'POST',
      body: JSON.stringify({ from, to }),
    });
  }

  async mergeTags(
    sources: string[],
    target: string,
  ): Promise<Result<MergeTagsDTO>> {
    return await httpClient.request<MergeTagsDTO>('/tags/merge', {
      method: 'POST',
      body: JSON.stringify({ sources, target }),
    });
  }

  async upsertTagAlias(
    alias: string,
    canonical: string,
  ): Promise<Result<TagAliasDTO>> {
    return await httpClient.request<TagAliasDTO>('/tags/aliases', {
      method: 'POST',
      body: JSON.stringify({ alias, canonical }),
    });
  }

  async listTagAliases(): Promise<Result<TagAliasDTO[]>> {
    return await httpClient.request<TagAliasDTO[]>('/tags/aliases');
  }

  async deleteTagAlias(alias: string): Promise<Result<void>> {
    return await httpClient.request<void>(
      `/tags/aliases/${encodeURIComponent(alias)}`,
      {
        method: 'DELETE',
      },
    );
  }

  async listTagAudits(limit = 20, action = ''): Promise<Result<TagAuditDTO[]>> {
    const query = new URLSearchParams({ limit: String(limit) });
    if (action) query.append('action', action);
    return await httpClient.request<TagAuditDTO[]>(
      `/tags/audits?${query.toString()}`,
    );
  }

  async listMemoHistory(uuid: string): Promise<Result<MemoHistoryDTO[]>> {
    return await httpClient.request<MemoHistoryDTO[]>(`/memos/${uuid}/history`);
  }

  async rollbackMemo(
    uuid: string,
    historyId: string,
  ): Promise<Result<MemoDTO>> {
    const result = await httpClient.request<BackendMemoDTO>(
      `/memos/${uuid}/rollback/${historyId}`,
      { method: 'POST' },
    );
    if (result.kind === 'failure') return result;
    try {
      return success(transformMemo(result.value));
    } catch (error) {
      return failure(new AppError('SERVER_ERROR', '回滚成功但解析失败', error));
    }
  }
}

export const memoGateway = new HttpMemoGateway();
