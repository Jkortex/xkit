import { ref } from 'vue';
import { defineStore } from 'pinia';
import { apiKeyGateway } from '@/infra/gateway/HttpApiKeyGateway';
import type { ApiKeyDTO } from '@/application/ports/dto/ApiKey';

export type ApiKeyVM = ApiKeyDTO;

export const useApiKeyStore = defineStore('apiKey', () => {
  const keys = ref<ApiKeyVM[]>([]);
  const loading = ref(false);

  const fetchKeys = async () => {
    loading.value = true;
    const res = await apiKeyGateway.list();
    loading.value = false;
    if (res.kind === 'success') {
      keys.value = res.value;
    }
  };

  const createKey = async (label: string, ttlHours?: number) => {
    const res = await apiKeyGateway.create({ label, ttlHours });
    if (res.kind === 'success') {
      keys.value = [res.value, ...keys.value];
      return { key: res.value, error: null };
    }
    return { key: null, error: res.error.message };
  };

  const createKeyDirect = async (payload: {
    username: string;
    password: string;
    label: string;
    ttlHours?: number;
  }) => {
    const res = await apiKeyGateway.createDirect(payload);
    if (res.kind === 'success') {
      keys.value = [res.value, ...keys.value];
      return { key: res.value, error: null };
    }
    return { key: null, error: res.error.message };
  };

  const deleteKey = async (id: string) => {
    const res = await apiKeyGateway.delete(id);
    if (res.kind === 'success') {
      keys.value = keys.value.filter((k) => k.id !== id);
      return null;
    }
    return res.error.message;
  };

  const revokeCurrent = async () => {
    const res = await apiKeyGateway.revokeCurrent();
    return res.kind === 'success' ? null : res.error.message;
  };

  const clear = () => {
    keys.value = [];
  };

  return {
    keys,
    loading,
    fetchKeys,
    createKey,
    createKeyDirect,
    deleteKey,
    revokeCurrent,
    clear,
  };
});
