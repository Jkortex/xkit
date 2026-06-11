import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { TagStatDTO } from '@/infra/gateway/HttpMemoGateway';
import { memoGateway } from '@/infra/gateway/HttpMemoGateway';

export const useTagStore = defineStore('tag', () => {
  const tags = ref<TagStatDTO[]>([]);
  const loading = ref(false);

  const setTags = (data: TagStatDTO[]) => {
    tags.value = data;
  };

  const fetchTags = async () => {
    loading.value = true;
    const result = await memoGateway.getTags();
    loading.value = false;
    if (result.kind === 'success') {
      tags.value = result.value;
    }
  };

  return {
    tags,
    loading,
    setTags,
    fetchTags,
  };
});
