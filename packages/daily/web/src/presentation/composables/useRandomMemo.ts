import { ref } from 'vue';
import { useMemoStore } from '@/infra/stores/useMemoStore';
import type { MemoVM } from '@/presentation/view-models/MemoVM';
import { MemoPresenter } from '@/presentation/presenters/MemoPresenter';
import { ErrorPresenter } from '@/presentation/presenters/ErrorPresenter';

export function useRandomMemo() {
  const store = useMemoStore();

  const randomMemo = ref<MemoVM | null>(null);
  const loading = ref(false);
  const error = ref<string | null>(null);

  const fetchRandom = async () => {
    loading.value = true;
    error.value = null;
    const result = await store.getRandomMemo();
    loading.value = false;

    if (result.kind === 'success') {
      randomMemo.value = MemoPresenter.toViewModel(result.value);
    } else {
      error.value = ErrorPresenter.toMessage(result.error);
    }
  };

  return {
    randomMemo,
    loading,
    error,
    fetchRandom,
  };
}
