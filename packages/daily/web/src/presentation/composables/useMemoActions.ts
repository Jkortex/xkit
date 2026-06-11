import { useMemoStore } from '@/infra/stores/useMemoStore';
import { useAsyncAction } from '@/presentation/composables/useAsyncAction';

export function useMemoActions() {
  const store = useMemoStore();
  const { loading, error, run } = useAsyncAction();

  const createMemo = async (
    content: string,
    resourceIds: string[] = [],
    ttl?: string,
  ) => {
    return run(() => store.createMemo(content, resourceIds, ttl));
  };

  const updateMemo = async (
    uuid: string,
    content: string,
    resourceIds: string[] = [],
    ttl?: string,
  ) => {
    return run(() => store.updateMemo(uuid, content, resourceIds, ttl));
  };

  const deleteMemo = async (uuid: string) => {
    return run(() => store.deleteMemo(uuid));
  };

  return {
    loading,
    error,
    createMemo,
    updateMemo,
    deleteMemo,
  };
}
