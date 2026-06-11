import { ref } from 'vue';
import { MessagePlugin } from 'tdesign-vue-next';
import { useMemoStore } from '@/infra/stores/useMemoStore';
import { MemoHistoryPresenter } from '@/presentation/presenters/MemoHistoryPresenter';
import type { MemoHistoryVM } from '@/presentation/view-models/MemoHistoryVM';
import { ErrorPresenter } from '@/presentation/presenters/ErrorPresenter';
import type { MemoDTO } from '@/application/ports/dto/Memo';

interface UseMemoHistoryOptions {
  onRollbackSuccess?: (memo: MemoDTO) => void;
}

export function useMemoHistory(options: UseMemoHistoryOptions = {}) {
  const store = useMemoStore();

  const historyList = ref<MemoHistoryVM[]>([]);
  const loading = ref(false);
  const rollingBack = ref(false);
  const error = ref<string | null>(null);

  const fetchHistory = async (uuid: string) => {
    loading.value = true;
    error.value = null;
    const result = await store.listMemoHistory(uuid);
    loading.value = false;

    if (result.kind === 'success') {
      historyList.value = MemoHistoryPresenter.toViewModelList(result.value);
    } else {
      error.value = ErrorPresenter.toMessage(result.error);
    }
  };

  const rollback = async (uuid: string, historyId: string) => {
    rollingBack.value = true;
    const result = await store.rollbackMemo(uuid, historyId);
    rollingBack.value = false;

    if (result.kind === 'success') {
      MessagePlugin.success('已恢复至选定版本');
      options.onRollbackSuccess?.(result.value);
      return true;
    } else {
      MessagePlugin.error(ErrorPresenter.toMessage(result.error));
      return false;
    }
  };

  return {
    historyList,
    loading,
    rollingBack,
    error,
    fetchHistory,
    rollback,
  };
}
