import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { TagSetGroupDTO, TagSetDTO } from '@/application/ports/dto/TagSet';
import { tagSetGateway } from '@/infra/gateway/HttpTagSetGateway';
import type { Result } from '@/utils/result';

export const useTagSetStore = defineStore('tagSet', () => {
  const groups = ref<TagSetGroupDTO[]>([]);
  const tagSets = ref<TagSetDTO[]>([]);
  const loading = ref(false);
  const groupLoading = ref(false);

  // --- Groups ---

  async function fetchGroups(): Promise<Result<TagSetGroupDTO[]>> {
    groupLoading.value = true;
    const result = await tagSetGateway.listGroups();
    groupLoading.value = false;
    if (result.kind === 'success') groups.value = result.value;
    return result;
  }

  async function createGroup(
    name: string,
    weight = 0,
  ): Promise<Result<TagSetGroupDTO>> {
    const result = await tagSetGateway.createGroup(name, weight);
    if (result.kind === 'success') groups.value.push(result.value);
    return result;
  }

  async function updateGroup(
    id: string,
    name?: string,
    weight?: number,
  ): Promise<Result<TagSetGroupDTO>> {
    const result = await tagSetGateway.updateGroup(id, name, weight);
    if (result.kind === 'success') {
      const idx = groups.value.findIndex((g) => g.id === id);
      if (idx !== -1) groups.value[idx] = result.value;
    }
    return result;
  }

  async function deleteGroup(id: string): Promise<Result<void>> {
    const result = await tagSetGateway.deleteGroup(id);
    if (result.kind === 'success') {
      groups.value = groups.value.filter((g) => g.id !== id);
      tagSets.value = tagSets.value.filter((s) => s.group_id !== id);
    }
    return result;
  }

  // --- TagSets ---

  async function fetchTagSets(groupID?: string): Promise<Result<TagSetDTO[]>> {
    loading.value = true;
    const result = await tagSetGateway.listTagSets(groupID);
    loading.value = false;
    if (result.kind === 'success') {
      if (!groupID) tagSets.value = result.value;
    }
    return result;
  }

  async function getTagSet(id: string): Promise<Result<TagSetDTO>> {
    return tagSetGateway.getTagSet(id);
  }

  async function createTagSet(params: {
    name: string;
    group_id?: string;
    tags_any?: string[];
    tags_all?: string[];
    tags_exclude?: string[];
    weight?: number;
  }): Promise<Result<TagSetDTO>> {
    const result = await tagSetGateway.createTagSet(params);
    if (result.kind === 'success') tagSets.value.push(result.value);
    return result;
  }

  async function updateTagSet(
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
    const result = await tagSetGateway.updateTagSet(id, params);
    if (result.kind === 'success') {
      const idx = tagSets.value.findIndex((s) => s.id === id);
      if (idx !== -1) tagSets.value[idx] = result.value;
    }
    return result;
  }

  async function deleteTagSet(id: string): Promise<Result<void>> {
    const result = await tagSetGateway.deleteTagSet(id);
    if (result.kind === 'success') {
      tagSets.value = tagSets.value.filter((s) => s.id !== id);
    }
    return result;
  }

  async function touchTagSet(id: string): Promise<Result<void>> {
    return tagSetGateway.touchTagSet(id);
  }

  return {
    groups,
    tagSets,
    loading,
    groupLoading,
    fetchGroups,
    createGroup,
    updateGroup,
    deleteGroup,
    fetchTagSets,
    getTagSet,
    createTagSet,
    updateTagSet,
    deleteTagSet,
    touchTagSet,
  };
});
