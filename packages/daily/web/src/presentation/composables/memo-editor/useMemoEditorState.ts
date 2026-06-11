import { ref, watch, type Ref } from 'vue';
import type { ResourceVM } from '@/presentation/view-models/ResourceVM';

const CREATE_DRAFT_KEY = 'daily.create_memo_draft';

export interface EditableMemoInput {
  uuid: string;
  content: string;
  tags: string[];
  resources: ResourceVM[];
}

interface UseMemoEditorStateOptions {
  editMemo: Ref<EditableMemoInput | undefined>;
}

interface UseMemoEditorStateResult {
  content: Ref<string>;
  attachedResources: Ref<ResourceVM[]>;
  clearAfterSuccess: () => void;
}

export function useMemoEditorState(
  options: UseMemoEditorStateOptions,
): UseMemoEditorStateResult {
  const content = ref('');
  const attachedResources = ref<ResourceVM[]>([]);

  watch(
    () => options.editMemo.value,
    (memo) => {
      if (memo) {
        const tagsLine = (memo.tags || []).map((t) => `#${t}`).join(' ');
        content.value = tagsLine
          ? `${memo.content}\n${tagsLine}`
          : memo.content;
        attachedResources.value = [...memo.resources];
        return;
      }
      const draft = sessionStorage.getItem(CREATE_DRAFT_KEY);
      content.value = draft ?? '';
      attachedResources.value = [];
    },
    { immediate: true },
  );

  watch(content, (value) => {
    if (options.editMemo.value) return;
    const trimmed = value.trim();
    if (!trimmed) {
      sessionStorage.removeItem(CREATE_DRAFT_KEY);
      return;
    }
    sessionStorage.setItem(CREATE_DRAFT_KEY, value);
  });

  const clearAfterSuccess = (): void => {
    content.value = '';
    attachedResources.value = [];
    if (!options.editMemo.value) sessionStorage.removeItem(CREATE_DRAFT_KEY);
  };

  return {
    content,
    attachedResources,
    clearAfterSuccess,
  };
}
