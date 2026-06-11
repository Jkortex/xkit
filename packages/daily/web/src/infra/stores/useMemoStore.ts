import { defineStore } from 'pinia';
import { ref, shallowRef } from 'vue';
import type { MemoDTO } from '@/application/ports/dto/Memo';
import type { MemoListQuery } from '@/application/ports/MemoListQuery';
import { validateMemoContent } from '@/application/rules/memo-rules';
import { memoGateway } from '@/infra/gateway/HttpMemoGateway';
import type { MemoHistoryDTO } from '@/infra/gateway/HttpMemoGateway';
import { Result, failure, AppError, success } from '@/utils/result';

const serializeQuery = (query: MemoListQuery): string =>
  JSON.stringify({
    search: query.search,
    tag: query.tag,
    from: query.from,
    to: query.to,
    hasResource: query.hasResource,
    tagsAny: query.tagsAny,
    tagsAll: query.tagsAll,
    sort: query.sort,
  });

export const useMemoStore = defineStore('memo', () => {
  const memos = shallowRef<MemoDTO[]>([]);
  const hasMore = ref(true);
  const loading = ref(false);
  const loadingMore = ref(false);
  const activeQuery = ref<MemoListQuery | null>(null);
  const activeQueryKey = ref<string | null>(null);

  const getActiveQuery = () => activeQuery.value;

  const canServe = (query: MemoListQuery) => {
    if (memos.value.length === 0 || !activeQueryKey.value) return false;
    return activeQueryKey.value === serializeQuery(query);
  };

  const setMemos = (query: MemoListQuery, data: MemoDTO[], more: boolean) => {
    activeQuery.value = query;
    activeQueryKey.value = serializeQuery(query);
    memos.value = data;
    hasMore.value = more;
  };

  const appendPage = (data: MemoDTO[], more: boolean) => {
    memos.value = [...memos.value, ...data];
    hasMore.value = more;
  };

  const upsert = (memo: MemoDTO) => {
    const index = memos.value.findIndex((m) => m.uuid === memo.uuid);
    if (index > -1) {
      const newMemos = [...memos.value];
      newMemos[index] = memo;
      memos.value = newMemos;
    } else {
      memos.value = [memo, ...memos.value];
    }
  };

  const remove = (uuid: string) => {
    memos.value = memos.value.filter((m) => m.uuid !== uuid);
  };

  const clear = () => {
    memos.value = [];
    hasMore.value = true;
    activeQuery.value = null;
    activeQueryKey.value = null;
    loading.value = false;
    loadingMore.value = false;
  };

  // --- API actions (formerly MemoService) ---

  const getMemos = async (
    params: MemoListQuery,
  ): Promise<Result<MemoDTO[]>> => {
    const isFirstPage = (params.offset ?? 0) === 0;

    if (isFirstPage && canServe(params)) {
      return success(memos.value);
    }

    if (isFirstPage) loading.value = true;
    else loadingMore.value = true;

    const result = await memoGateway.getMemos(params);
    loading.value = false;
    loadingMore.value = false;

    if (result.kind === 'failure') return result;

    const limit = params.limit ?? result.value.length;
    const more = result.value.length >= limit;

    if (isFirstPage) {
      setMemos(params, result.value, more);
    } else {
      appendPage(result.value, more);
    }

    return result;
  };

  const createMemo = async (
    content: string,
    resourceIds: string[] = [],
    ttl?: string,
  ): Promise<Result<MemoDTO>> => {
    const validationError = validateMemoContent(content);
    if (validationError) {
      return failure(new AppError('VALIDATION_ERROR', validationError));
    }

    const result = await memoGateway.createMemo(content, resourceIds, ttl);
    if (result.kind === 'success') upsert(result.value);
    return result;
  };

  const updateMemo = async (
    uuid: string,
    content: string,
    resourceIds: string[] = [],
    ttl?: string,
  ): Promise<Result<MemoDTO>> => {
    const validationError = validateMemoContent(content);
    if (validationError) {
      return failure(new AppError('VALIDATION_ERROR', validationError));
    }

    const result = await memoGateway.updateMemo(
      uuid,
      content,
      resourceIds,
      ttl,
    );
    if (result.kind === 'success') upsert(result.value);
    return result;
  };

  const deleteMemo = async (uuid: string): Promise<Result<void>> => {
    const result = await memoGateway.deleteMemo(uuid);
    if (result.kind === 'success') remove(uuid);
    return result;
  };

  const getRandomMemo = async (): Promise<Result<MemoDTO>> => {
    return memoGateway.getRandomMemo();
  };

  const listMemoHistory = async (
    uuid: string,
  ): Promise<Result<MemoHistoryDTO[]>> => {
    return memoGateway.listMemoHistory(uuid);
  };

  const rollbackMemo = async (
    uuid: string,
    historyId: string,
  ): Promise<Result<MemoDTO>> => {
    const result = await memoGateway.rollbackMemo(uuid, historyId);
    if (result.kind === 'success') upsert(result.value);
    return result;
  };

  return {
    memos,
    hasMore,
    loading,
    loadingMore,
    getActiveQuery,
    canServe,
    setMemos,
    appendPage,
    upsert,
    remove,
    clear,
    getMemos,
    createMemo,
    updateMemo,
    deleteMemo,
    getRandomMemo,
    listMemoHistory,
    rollbackMemo,
  };
});
