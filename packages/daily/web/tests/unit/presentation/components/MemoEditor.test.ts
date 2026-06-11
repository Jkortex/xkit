// @vitest-environment happy-dom

import { beforeEach, describe, expect, it, vi } from 'vitest';
import { defineComponent, h, ref } from 'vue';
import { mount } from '@vue/test-utils';

const { createMemoMock, updateMemoMock } = vi.hoisted(() => ({
  createMemoMock: vi.fn(),
  updateMemoMock: vi.fn(),
}));

vi.mock('@/presentation/composables/useMemoActions', () => ({
  useMemoActions: () => ({
    createMemo: createMemoMock,
    updateMemo: updateMemoMock,
    loading: ref(false),
  }),
}));

const { uploadFileMock } = vi.hoisted(() => ({
  uploadFileMock: vi.fn(),
}));

vi.mock('@/presentation/composables/useResource', () => ({
  useResource: () => ({
    uploadFile: uploadFileMock,
    uploading: ref(false),
    error: ref(null),
  }),
}));

vi.mock('tdesign-vue-next', () => {
  const Button = defineComponent({
    name: 'TButton',
    emits: ['click'],
    setup(_, { slots, emit }) {
      return () =>
        h('button', { onClick: () => emit('click') }, slots.default?.());
    },
  });
  const Textarea = defineComponent({
    name: 'TTextarea',
    props: ['modelValue'],
    emits: ['update:modelValue', 'keydown'],
    setup(props, { emit }) {
      return () =>
        h('textarea', {
          value: props.modelValue,
          onInput: (e: any) => emit('update:modelValue', e.target.value),
          onKeydown: (e: KeyboardEvent) =>
            emit('keydown', props.modelValue, { e }),
        });
    },
  });
  return {
    Button,
    Textarea,
    MessagePlugin: {
      success: vi.fn(),
      error: vi.fn(),
    },
  };
});

vi.mock('lucide-vue-next', () => {
  const Icon = defineComponent({
    name: 'IconStub',
    setup: () => () => h('span'),
  });
  return { Paperclip: Icon, Send: Icon, X: Icon };
});

import MemoEditor from '@/presentation/components/MemoEditor.vue';

describe('MemoEditor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    sessionStorage.clear();
  });

  it('submits create memo and clears editor on success', async () => {
    createMemoMock.mockResolvedValue(true);
    const wrapper = mount(MemoEditor);

    const textarea = wrapper.find('textarea');
    await textarea.setValue('hello world');

    const saveBtn = wrapper
      .findAll('button')
      .find((b) => b.text().includes('保存'));
    await saveBtn?.trigger('click');

    expect(createMemoMock).toHaveBeenCalledWith('hello world', []);
    expect(wrapper.emitted('success')).toBeTruthy();
    expect(textarea.element.value).toBe('');
  });

  it('submits update memo and emits cancel-edit on success', async () => {
    updateMemoMock.mockResolvedValue(true);
    const wrapper = mount(MemoEditor, {
      props: {
        editMemo: { uuid: '1', id: 1, content: 'existing', resources: [] },
      },
    });

    const textarea = wrapper.find('textarea');
    await textarea.setValue('updated content');

    const saveBtn = wrapper
      .findAll('button')
      .find((b) => b.text().includes('更新'));
    await saveBtn?.trigger('click');

    expect(updateMemoMock).toHaveBeenCalledWith('1', 'updated content', []);
    expect(wrapper.emitted('success')).toBeTruthy();
    expect(wrapper.emitted('cancel-edit')).toBeTruthy();
  });

  it('does not clear on failed submit', async () => {
    createMemoMock.mockResolvedValue(false);
    const wrapper = mount(MemoEditor);

    const textarea = wrapper.find('textarea');
    await textarea.setValue('hello world');

    const saveBtn = wrapper
      .findAll('button')
      .find((b) => b.text().includes('保存'));
    await saveBtn?.trigger('click');

    expect(createMemoMock).toHaveBeenCalledWith('hello world', []);
    expect(wrapper.emitted('success')).toBeFalsy();
    expect(textarea.element.value).toBe('hello world');
  });

  it('saves via Ctrl+Enter shortcut', async () => {
    createMemoMock.mockResolvedValue(true);
    const wrapper = mount(MemoEditor);

    const textarea = wrapper.find('textarea');
    await textarea.setValue('shortcut save');

    await textarea.trigger('keydown', { key: 'Enter', ctrlKey: true });

    expect(createMemoMock).toHaveBeenCalledWith('shortcut save', []);
    expect(wrapper.emitted('success')).toBeTruthy();
  });
});
