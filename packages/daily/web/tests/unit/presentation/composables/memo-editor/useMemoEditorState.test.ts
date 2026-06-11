// @vitest-environment happy-dom

import { beforeEach, describe, expect, it } from 'vitest';
import { nextTick, ref } from 'vue';
import { useMemoEditorState } from '@/presentation/composables/memo-editor/useMemoEditorState';

describe('useMemoEditorState', () => {
  beforeEach(() => {
    sessionStorage.clear();
  });

  it('loads draft when no edit memo', () => {
    sessionStorage.setItem('daily.create_memo_draft', 'draft content');
    const editMemo = ref(undefined);

    const state = useMemoEditorState({ editMemo });

    expect(state.content.value).toBe('draft content');
    expect(state.attachedResources.value).toEqual([]);
  });

  it('uses edit memo content and resources when editing', async () => {
    const editMemo = ref({
      id: 1,
      content: 'editing',
      tags: [],
      resources: [{ id: 'r1', url: '/1', filename: 'a.png', isImage: true }],
    });

    const state = useMemoEditorState({ editMemo });

    expect(state.content.value).toBe('editing');
    expect(state.attachedResources.value).toHaveLength(1);

    editMemo.value = undefined;
    await nextTick();
    expect(state.attachedResources.value).toEqual([]);
  });

  it('persists and clears draft for create flow', async () => {
    const editMemo = ref(undefined);
    const state = useMemoEditorState({ editMemo });

    state.content.value = 'hello';
    await nextTick();
    expect(sessionStorage.getItem('daily.create_memo_draft')).toBe('hello');

    state.content.value = '   ';
    await nextTick();
    expect(sessionStorage.getItem('daily.create_memo_draft')).toBeNull();

    state.clearAfterSuccess();
    expect(state.content.value).toBe('');
  });
});
