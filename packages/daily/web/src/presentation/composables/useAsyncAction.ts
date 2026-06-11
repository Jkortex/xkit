import { ref } from 'vue';
import type { Result } from '@/utils/result';
import { ErrorPresenter } from '@/presentation/presenters/ErrorPresenter';

interface UseAsyncActionResult {
  loading: ReturnType<typeof ref<boolean>>;
  error: ReturnType<typeof ref<string | null>>;
  run: <T>(action: () => Promise<Result<T>>) => Promise<boolean>;
}

export function useAsyncAction(): UseAsyncActionResult {
  const loading = ref(false);
  const error = ref<string | null>(null);

  const run = async <T>(action: () => Promise<Result<T>>): Promise<boolean> => {
    loading.value = true;
    error.value = null;
    try {
      const result = await action();
      if (result.kind === 'success') return true;
      error.value = ErrorPresenter.toMessage(result.error);
      return false;
    } finally {
      loading.value = false;
    }
  };

  return { loading, error, run };
}
