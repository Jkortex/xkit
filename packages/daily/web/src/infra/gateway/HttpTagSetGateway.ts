import type { TagSetGroupDTO, TagSetDTO } from '@/application/ports/dto/TagSet';
import { httpClient } from '../http/FetchClient';
import { Result, success } from '@/utils/result';

interface BackendTagSetGroupResponse {
  id: string;
  name: string;
  weight: number;
  created_at: string;
  updated_at: string;
}

interface BackendTagSetResponse {
  id: string;
  name: string;
  group_id: string | null;
  tags_any: string[];
  tags_all: string[];
  tags_exclude: string[];
  weight: number;
  last_used_at: string | null;
  created_at: string;
  updated_at: string;
}

function toGroupDTO(raw: BackendTagSetGroupResponse): TagSetGroupDTO {
  return {
    id: raw.id,
    name: raw.name,
    weight: raw.weight,
    created_at: raw.created_at,
    updated_at: raw.updated_at,
  };
}

function toTagSetDTO(raw: BackendTagSetResponse): TagSetDTO {
  return {
    id: raw.id,
    name: raw.name,
    group_id: raw.group_id,
    tags_any: raw.tags_any || [],
    tags_all: raw.tags_all || [],
    tags_exclude: raw.tags_exclude || [],
    weight: raw.weight,
    last_used_at: raw.last_used_at,
    created_at: raw.created_at,
    updated_at: raw.updated_at,
  };
}

export class HttpTagSetGateway {
  async listGroups(): Promise<Result<TagSetGroupDTO[]>> {
    const result =
      await httpClient.request<BackendTagSetGroupResponse[]>('/tag-set-groups');
    if (result.kind === 'failure') return result;
    return success(result.value.map(toGroupDTO));
  }

  async createGroup(name: string, weight = 0): Promise<Result<TagSetGroupDTO>> {
    const result = await httpClient.request<BackendTagSetGroupResponse>(
      '/tag-set-groups',
      {
        method: 'POST',
        body: JSON.stringify({ name, weight }),
      },
    );
    if (result.kind === 'failure') return result;
    return success(toGroupDTO(result.value));
  }

  async updateGroup(
    id: string,
    name?: string,
    weight?: number,
  ): Promise<Result<TagSetGroupDTO>> {
    const body: Record<string, unknown> = {};
    if (name !== undefined) body.name = name;
    if (weight !== undefined) body.weight = weight;
    const result = await httpClient.request<BackendTagSetGroupResponse>(
      `/tag-set-groups/${id}`,
      {
        method: 'PATCH',
        body: JSON.stringify(body),
      },
    );
    if (result.kind === 'failure') return result;
    return success(toGroupDTO(result.value));
  }

  async deleteGroup(id: string): Promise<Result<void>> {
    const result = await httpClient.request<void>(`/tag-set-groups/${id}`, {
      method: 'DELETE',
    });
    return result;
  }

  async listTagSets(groupID?: string): Promise<Result<TagSetDTO[]>> {
    const path = groupID
      ? `/tag-sets?group_id=${encodeURIComponent(groupID)}`
      : '/tag-sets';
    const result = await httpClient.request<BackendTagSetResponse[]>(path);
    if (result.kind === 'failure') return result;
    return success(result.value.map(toTagSetDTO));
  }

  async getTagSet(id: string): Promise<Result<TagSetDTO>> {
    const result = await httpClient.request<BackendTagSetResponse>(
      `/tag-sets/${id}`,
    );
    if (result.kind === 'failure') return result;
    return success(toTagSetDTO(result.value));
  }

  async createTagSet(params: {
    name: string;
    group_id?: string;
    tags_any?: string[];
    tags_all?: string[];
    tags_exclude?: string[];
    weight?: number;
  }): Promise<Result<TagSetDTO>> {
    const body: Record<string, unknown> = { name: params.name };
    if (params.group_id !== undefined) body.group_id = params.group_id;
    if (params.tags_any !== undefined) body.tags_any = params.tags_any;
    if (params.tags_all !== undefined) body.tags_all = params.tags_all;
    if (params.tags_exclude !== undefined)
      body.tags_exclude = params.tags_exclude;
    if (params.weight !== undefined) body.weight = params.weight;
    const result = await httpClient.request<BackendTagSetResponse>(
      '/tag-sets',
      {
        method: 'POST',
        body: JSON.stringify(body),
      },
    );
    if (result.kind === 'failure') return result;
    return success(toTagSetDTO(result.value));
  }

  async updateTagSet(
    id: string,
    params: {
      name?: string;
      group_id?: string | null;
      tags_any?: string[];
      tags_all?: string[];
      tags_exclude?: string[];
      weight?: number;
    },
  ): Promise<Result<TagSetDTO>> {
    const body: Record<string, unknown> = {};
    if (params.name !== undefined) body.name = params.name;
    if (params.group_id !== undefined) body.group_id = params.group_id;
    if (params.tags_any !== undefined) body.tags_any = params.tags_any;
    if (params.tags_all !== undefined) body.tags_all = params.tags_all;
    if (params.tags_exclude !== undefined)
      body.tags_exclude = params.tags_exclude;
    if (params.weight !== undefined) body.weight = params.weight;
    const result = await httpClient.request<BackendTagSetResponse>(
      `/tag-sets/${id}`,
      {
        method: 'PATCH',
        body: JSON.stringify(body),
      },
    );
    if (result.kind === 'failure') return result;
    return success(toTagSetDTO(result.value));
  }

  async deleteTagSet(id: string): Promise<Result<void>> {
    const result = await httpClient.request<void>(`/tag-sets/${id}`, {
      method: 'DELETE',
    });
    return result;
  }

  async touchTagSet(id: string): Promise<Result<void>> {
    const result = await httpClient.request<void>(`/tag-sets/${id}/touch`, {
      method: 'POST',
    });
    return result;
  }
}

export const tagSetGateway = new HttpTagSetGateway();
